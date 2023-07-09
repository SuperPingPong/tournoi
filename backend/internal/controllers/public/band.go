package public

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/SuperPingPong/tournoi/internal/auth"
	"github.com/SuperPingPong/tournoi/internal/models"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type SetMemberEntries struct {
	BandIDs   []uuid.UUID `binding:"required"`
	SessionID uuid.UUID   `binding:"required"`
}

var sessionExpiredError = errors.New("missing lock for entry")

func (api *API) SetMemberEntries(ctx *gin.Context) {
	claims := jwt.ExtractClaims(ctx)
	userID := uuid.MustParse(claims[auth.IdentityKey].(string))

	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid member id: %s", ctx.Param("id")))
		return
	}

	var input SetMemberEntries
	err = ctx.ShouldBindJSON(&input)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
		return
	}

	// Get the current member
	var member models.Member
	err = api.db.Where(&models.Member{ID: member.ID, UserID: userID}).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("member %s not found", memberID))
			return
		}

		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get member: %w", err))
		return
	}

	// List possible bands for the current member
	var bands []models.Band
	if api.db.Scopes(possibleBandsScope(member)).Where("id IN ?", input.BandIDs).Find(&bands).Error != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to find bands %v", bands))
		return
	}
	if len(bands) != len(input.BandIDs) {
		missingBands := lo.Filter(input.BandIDs, func(bandID uuid.UUID, _ int) bool {
			return !lo.Contains(mapBandIDs(bands), bandID)
		})
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("bands %v not found", missingBands))
		return
	}

	err = api.db.Transaction(func(tx *gorm.DB) error {
		// Delete the unwanted entries.
		if err = api.db.Where("member_id = ? AND band_id NOT IN ?", member.ID, input.BandIDs).Delete(&models.Entry{}).Error; err != nil {
			return fmt.Errorf("failed to delete entry: %w", err)
		}

		// Find existing entries for the member
		var existingEntries []models.Entry
		if err = tx.Where("member_id = ?", member.ID).Find(&existingEntries).Error; err != nil {
			return fmt.Errorf("failed to list member entries: %w", err)
		}

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
		if len(entriesToConfirm) != len(input.BandIDs)-confirmedEntriesCount {
			return sessionExpiredError
		}

		if err = tx.Model(models.Entry{}).
			Where("id IN ?", entriesToConfirm).
			Updates(models.Entry{Confirmed: true}).Error; err != nil {
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

	ctx.Status(http.StatusOK)
}

func (api *API) ListBands(ctx *gin.Context) {
	var bands []models.Band
	var filteredBands *gorm.DB

	day, err := strconv.Atoi(ctx.Query("day"))
	if err != nil {
		filteredBands = api.db
	} else {
		filteredBands = api.db.Where("day = ?", day)
	}

	err = filteredBands.Find(&bands).Error
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list bands: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"bands": bands})
}
