package public

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/SuperPingPong/tournoi/internal/auth"
	"github.com/SuperPingPong/tournoi/internal/models"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (api *API) ListMembers(ctx *gin.Context) {
	claims := jwt.ExtractClaims(ctx)
	userID := uuid.MustParse(claims[auth.IdentityKey].(string))

	var members []models.Member
	err := api.db.Where("user_id = ?", userID).Find(&members).Error
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get member: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"members": members})
}

type CreateMemberInput struct {
	PermitID string `binding:"required,min=2"`
}

func (api *API) CreateMember(ctx *gin.Context) {
	claims := jwt.ExtractClaims(ctx)
	userID := uuid.MustParse(claims[auth.IdentityKey].(string))

	var input CreateMemberInput
	err := ctx.ShouldBindJSON(&input)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
		return
	}

	data, err := api.GetFFTTPlayerData(input.PermitID)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get member data from FFTT: %w", err))
		return
	}

	member := models.Member{
		UserID:     userID,
		PermitID:   data.PermitID,
		FirstName:  data.FirstName,
		LastName:   data.LastName,
		Sex:        data.Sex,
		Points:     data.Points,
		Category:   data.Category,
		ClubName:   data.ClubName,
		PermitType: data.PermitType,
	}
	err = api.db.Create(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			ctx.AbortWithError(http.StatusConflict, fmt.Errorf("member with permit %s already exists", input.PermitID))
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to create member: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, &member)
}

type UpdateMemberInput struct {
	PermitID string `binding:"required,min=2"`
}

func (api *API) UpdateMember(ctx *gin.Context) {
	claims := jwt.ExtractClaims(ctx)
	userID := uuid.MustParse(claims[auth.IdentityKey].(string))

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid member id: %s", ctx.Param("id")))
		return
	}

	var input UpdateMemberInput
	err = ctx.ShouldBindJSON(&input)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
		return
	}

	member := models.Member{}
	err = api.db.First(&member, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("member %s not found", id))
			return
		}

		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get member: %w", err))
		return
	}

	if member.UserID != userID {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("member %s not found", id))
		return
	}

	if input.PermitID != "" {
		data, err := api.GetFFTTPlayerData(input.PermitID)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get member data from FFTT: %w", err))
			return
		}

		member.PermitID = data.PermitID
		member.FirstName = data.FirstName
		member.LastName = data.LastName
		member.Sex = data.Sex
		member.Points = data.Points
		member.Category = data.Category
		member.ClubName = data.ClubName
		member.PermitType = data.PermitType
	}

	err = api.db.Save(&member).Error
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to update member: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, &member)
}
