package public

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/SuperPingPong/tournoi/internal/auth"
	"github.com/SuperPingPong/tournoi/internal/models"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SetMemberBandsInput struct {
	BandIDs []uuid.UUID `binding:"required"`
}

func (api *API) SetMemberBands(ctx *gin.Context) {
	claims := jwt.ExtractClaims(ctx)
	userID := uuid.MustParse(claims[auth.IdentityKey].(string))

	var user models.User
	if err := api.db.Preload("Members").First(&user, userID).Error; err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get user %s: %w", userID, err))
		return
	}

	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid member id: %s", ctx.Param("id")))
		return
	}

	var input SetMemberBandsInput
	err = ctx.ShouldBindJSON(&input)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
		return
	}

	if len(input.BandIDs) > 3 {

	}

	var bands []models.Band
	for _, bandID := range input.BandIDs {
		var band models.Band
		if api.db.First(&band, bandID).Error != nil {
			ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("band %s not found", bandID))
			return
		}
		bands = append(bands, band)
	}

	var member models.Member
	if !userHasMember(user, memberID) || api.db.Preload("Bands").First(&member, memberID).Error != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("member %s not found", memberID))
		return
	}

	err = api.db.Transaction(func(tx *gorm.DB) error {
		for _, band := range member.Bands {
			if !lo.Contains(input.BandIDs, band.ID) {
				if err = tx.Model(&member).Association("Bands").Delete(band); err != nil {
					return fmt.Errorf("failed to delete member band relationship: %w", err)
				}
			}
		}

		for _, band := range bands {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.BandMember{
				BandID:   band.ID,
				MemberID: member.ID,
			}).Error; err != nil {
				return fmt.Errorf("failed to create member band relationship (%s, %s): %w", member.ID, band.ID, err)
			}
		}

		return nil
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err = api.db.Preload("Bands").First(&member, memberID).Error; err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to reload member: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, &member)
}

func userHasMember(user models.User, memberID uuid.UUID) bool {
	var hasMember bool
	for _, member := range user.Members {
		if member.ID == memberID {
			hasMember = true
		}
	}
	return hasMember
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
