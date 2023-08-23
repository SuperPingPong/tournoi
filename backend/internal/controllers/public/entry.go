package public

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/SuperPingPong/tournoi/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ListBandAvailabilitiesInput struct {
	MemberID uuid.UUID `binding:"required"`
}

type BandAvailability struct {
	models.Band
	Available int
	Waiting   int
}

func (api *API) ListBandAvailabilities(ctx *gin.Context) {
	user, err := ExtractUserFromContext(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid member ID: %s", ctx.Param("id")))
		return
	}

	// Get the current member
	var member models.Member
	err = api.db.
		Scopes(FilterByUserID(user)).
		Where("id = ?", memberID).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("member %s not found", memberID))
			return
		}

		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get member: %w", err))
		return
	}

	// List possible bands for the current member
	var possibleBands []models.Band
	if err = api.db.Scopes(possibleBandsScope(member)).Find(&possibleBands).Error; err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list available bands: %w", err))
		return
	}
	possibleBandIDs := lo.Map(possibleBands, func(b models.Band, index int) uuid.UUID {
		return b.ID
	})

	// Delete existing locks for the current member
	if err = api.db.Where("member_id = ? AND band_id IN ? AND confirmed IS FALSE", member.ID, possibleBandIDs).Delete(&models.Entry{}).Error; err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to delete locked entries for available bands: %w", err))
		return
	}

	sessionID := uuid.New()
	var entries []models.Entry
	var bandAvailabilities []BandAvailability
	err = api.db.Transaction(func(tx *gorm.DB) error {
		// List existing confirmed entries and locks
		if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("(confirmed IS TRUE OR (expires_at > ? AND confirmed IS FALSE)) AND band_id IN ? ", time.Now(), possibleBandIDs).Find(&entries).Error; err != nil {
			return fmt.Errorf("failed to list entries for available bands: %w", err)
		}

		// Count each bands' existing entries
		bandCounts := lo.SliceToMap(possibleBands, func(b models.Band) (uuid.UUID, int) {
			return b.ID, 0
		})
		for _, entry := range entries {
			bandCounts[entry.BandID] += 1
		}

		for _, band := range possibleBands {
			// Compute each bands' available spots and number of people in the waiting list
			bandAvailabilities = append(bandAvailabilities, BandAvailability{
				Band:      band,
				Available: int(math.Max(float64(band.MaxEntries-bandCounts[band.ID]), 0)),
				Waiting:   int(math.Max(float64(bandCounts[band.ID]-band.MaxEntries), 0)),
			})

			// Lock a position
			if err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.Entry{
				BandID:    band.ID,
				MemberID:  member.ID,
				ExpiresAt: time.Now().Add(models.EntryLockExpirationDelay),
				Confirmed: false,
				SessionID: sessionID,
				CreatedBy: uuid.NullUUID{UUID: user.ID, Valid: true},
			}).Error; err != nil {
				return fmt.Errorf("failed to lock entries: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}

	ctx.JSON(http.StatusOK, gin.H{"bands": bandAvailabilities, "session_id": sessionID})
}

type EntriesHistory struct {
	ID             uuid.UUID
	BandId         string
	BandName       string
	EventTime      time.Time
	EventType      string
	EventBy        string
	EventByIsAdmin bool `gorm:"column:event_by_is_admin"`
}

func (api *API) GetMemberEntriesHistory(ctx *gin.Context) {
	user, err := ExtractUserFromContext(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if !user.IsAdmin {
		ctx.AbortWithError(http.StatusForbidden, fmt.Errorf("user %s should be admin", user.ID.String()))
		return
	}

	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid member ID: %s", ctx.Param("id")))
		return
	}

	var history []EntriesHistory
	query := `
        SELECT
            entries.id,
            entries.band_id,
            entries.created_at AS event_time,
            'created' AS event_type,
            users.email AS event_by,
            bands.name AS band_name,
            users.is_admin AS event_by_is_admin
        FROM
            entries
        JOIN
            users ON entries.created_by = users.id
        JOIN
            bands ON entries.band_id = bands.id
        WHERE
            entries.confirmed = ?
            AND entries.member_id = ?
        UNION ALL
        SELECT
            entries.id,
            entries.band_id,
            entries.deleted_at AS event_time,
            'deleted' AS event_type,
            users.email AS event_by,
            bands.name AS band_name,
            users.is_admin AS event_by_is_admin
        FROM
            entries
        JOIN
            users ON entries.deleted_by = users.id
        JOIN
            bands ON entries.band_id = bands.id
        WHERE
            entries.confirmed = ?
            AND entries.deleted_at IS NOT NULL
            AND entries.member_id = ?
        ORDER BY
            event_time DESC
    `
	if err := api.db.Raw(query, true, memberID, true, memberID).Scan(&history).Error; err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get entries: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"history": history})

}

func possibleBandsScope(member models.Member) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("(sex = ? OR sex = 'ALL') AND max_points >= ?", member.Sex, member.Points)
	}
}

func mapBandIDs(bands []models.Band) []uuid.UUID {
	return lo.Map(bands, func(band models.Band, _ int) uuid.UUID {
		return band.ID
	})
}

type SetMemberEntriesInput struct {
	BandIDs   []uuid.UUID `binding:"required"`
	SessionID uuid.UUID   `binding:"required"`
}

var sessionExpiredError = errors.New("missing lock for entry")

func (api *API) SetMemberEntries(ctx *gin.Context) {
	user, err := ExtractUserFromContext(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid member id: %s", ctx.Param("id")))
		return
	}

	var input SetMemberEntriesInput
	err = ctx.ShouldBindBodyWith(&input, binding.JSON)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
		return
	}

	// Get the requested member
	var member models.Member
	err = api.db.
		Scopes(FilterByUserID(user)).
		Where("id = ?", memberID).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("member %s not found", memberID))
			return
		}

		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get member: %w", err))
		return
	}
	fmt.Printf("member: %v\n", member)

	// List possible bands for the current member
	var bands []models.Band
	if api.db.Scopes(possibleBandsScope(member)).Where("id IN ?", input.BandIDs).Find(&bands).Error != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to find bands %v", bands))
		return
	}

	fmt.Printf("bands: %v\n", bands)
	if len(bands) != len(input.BandIDs) {
		missingBands := lo.Filter(input.BandIDs, func(bandID uuid.UUID, _ int) bool {
			return !lo.Contains(mapBandIDs(bands), bandID)
		})
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("bands %v not found", missingBands))
		return
	}

	// Enforce bands limit
	if err = enforceBandsLimit(bands); err != nil {
		ctx.AbortWithError(http.StatusConflict, err)
		return
	}

	err = api.db.Transaction(func(tx *gorm.DB) error {
		// Delete the unwanted entries.
		inputBandIDs := input.BandIDs
		// We add uuid.Nil to input.BandIDs when it is empty since "band_id NOT IN (NULL)" doesn't match any entry
		if len(input.BandIDs) == 0 {
		    inputBandIDs = []uuid.UUID{uuid.Nil}
		}
		if err = tx.
			Where("member_id = ? AND band_id NOT IN ?", member.ID, inputBandIDs).
			Updates(&models.Entry{DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true}, DeletedBy: uuid.NullUUID{UUID: user.ID, Valid: true}}).Error; err != nil {
			return fmt.Errorf("failed to delete entry: %w", err)
		}

		// Find existing entries for the member
		var existingEntries []models.Entry
		if err = tx.Where("member_id = ?", member.ID).Find(&existingEntries).Error; err != nil {
			return fmt.Errorf("failed to list member entries: %w", err)
		}
		fmt.Printf("existingEntries: %v\n", existingEntries)

		// We need to make sure that every band ID from the input is either confirmed or has a lock in the current session
		var confirmedEntriesCount int
		requestedEntries := map[uuid.UUID]models.Entry{}
		for _, entry := range existingEntries {
			if entry.Confirmed {
				confirmedEntriesCount += 1
				requestedEntries[entry.BandID] = entry
			} else if entry.SessionID == input.SessionID {
				requestedEntries[entry.BandID] = entry
			}
		}
		fmt.Printf("confirmedEntries: %v\n", existingEntries)

		var entriesToConfirm []uuid.UUID
		for _, bandID := range input.BandIDs {
			requestedEntry, ok := requestedEntries[bandID]
			// Reject request if no entry (confirmed or locked) exists for the band ID
			if !ok {
				return sessionExpiredError
			}
			// Confirm the entry if not already confirmed and not expired
			if !requestedEntry.Confirmed && requestedEntry.ExpiresAt.After(time.Now()) {
				entriesToConfirm = append(entriesToConfirm, requestedEntry.ID)
			}
		}

		// The number of entries to confirm should be the difference between the number of expected bands
		// and the number of already confirmed entries. Otherwise, it means that some entries were expired.
		fmt.Printf("entriesToConfirm: %v", entriesToConfirm)
		if len(entriesToConfirm) != len(input.BandIDs)-confirmedEntriesCount {
			return sessionExpiredError
		}

		if err = tx.Model(models.Entry{}).
			Where("id IN ?", entriesToConfirm).
			Updates(models.Entry{Confirmed: true, ConfirmedBy: uuid.NullUUID{UUID: user.ID, Valid: true}}).Error; err != nil {
			return fmt.Errorf("failed to confirm entry: %w", err)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, sessionExpiredError) {
			ctx.AbortWithError(http.StatusConflict, err)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Only send email if it's the first registration
	if !member.HasBeenNotified {
		err = sendEmailHTML(user.Email, member.LastName, member.FirstName)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to send email: %w", err))
			return
		}
		if err = api.db.Model(models.Member{}).
			Where("id", memberID).
			Updates(models.Member{HasBeenNotified: true}).Error; err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to update member: %w", err))
		}

	}

	ctx.Status(http.StatusOK)
}

var (
	limitThreeBandsPerDayReachedError = errors.New("can't have more than three bands per day")
	limitSameColorPerDayReachedError  = errors.New("can't have more than two bands of the same color the same day")
)

// Enforce a limit of max 3 bands per day and
// ensures that no two bands of the same color are scheduled on the same day
func enforceBandsLimit(bands []models.Band) error {
	countPerDay := make(map[int]int)
	colorCountPerDay := make(map[int]map[string]int)
	for _, band := range bands {
		countPerDay[band.Day] += 1

		colorCount, ok := colorCountPerDay[band.Day]
		if !ok {
			colorCount = make(map[string]int)
			colorCountPerDay[band.Day] = colorCount
		}
		colorCount[band.Color] += 1

		if countPerDay[band.Day] > 3 {
			return limitThreeBandsPerDayReachedError
		}

		if colorCountPerDay[band.Day][band.Color] > 1 {
			return limitSameColorPerDayReachedError
		}
	}
	return nil
}
