package public

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/SuperPingPong/tournoi/internal/auth"
	"github.com/SuperPingPong/tournoi/internal/models"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
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
	claims := jwt.ExtractClaims(ctx)
	userID := uuid.MustParse(claims[auth.IdentityKey].(string))

	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid member ID: %s", ctx.Param("id")))
		return
	}

	// Get the current member
	member := models.Member{}
	err = api.db.
		Joins("LEFT JOIN entries ON entries.member_id = members.id").
		Where("members.id = ? AND members.user_id = ?", memberID, userID).
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
	possibleBandsFilters := api.db.Where("(sex = ? OR sex = 'ALL') AND max_points > ?", member.Sex, member.Points)
	if err = possibleBandsFilters.Find(&possibleBands).Error; err != nil {
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

	var entries []models.Entry
	var bandAvailabilities []BandAvailability
	err = api.db.Transaction(func(tx *gorm.DB) error {
		// List existing confirmed entries and locks
		if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("expires_at > ? AND band_id IN ? ", time.Now(), possibleBandIDs).Find(&entries).Error; err != nil {
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
			if err = tx.Create(&models.Entry{
				BandID:    band.ID,
				MemberID:  member.ID,
				ExpiresAt: time.Now().Add(models.EntryLockExpirationDelay),
				Confirmed: false,
			}).Error; err != nil {
				return fmt.Errorf("failed to lock entries: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}

	ctx.JSON(http.StatusOK, gin.H{"bands": bandAvailabilities})
}
