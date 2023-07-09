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
	"gorm.io/gorm"
)

type ListMembersEntry struct {
	BandID    uuid.UUID
	BandName  string
	CreatedAt time.Time
}

type ListMembersMember struct {
	ID         uuid.UUID
	PermitID   string
	FirstName  string
	LastName   string
	Sex        string
	Points     float64
	Category   string
	ClubName   string
	PermitType string
	Entries    []ListMembersEntry
}

type ListMembersMembers struct {
	Members []ListMembersMember
	Total   int
}

func (api *API) ListMembers(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid page: %s", ctx.Query("page")))
		return
	}
	pageSize, err := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid page size: %s", ctx.Query("page_size")))
		return
	}
	validOrderBy := map[string]string{
		"created_at_asc":  "created_at ASC",
		"created_at_desc": "created_at ASC",
		"last_name_desc":  "last_name DESC",
		"last_name_asc":   "last_name ASC",
	}
	orderBy, valid := validOrderBy[ctx.DefaultQuery("order_by", "created_at_desc")]
	if !valid {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid page size: %s", ctx.Query("page_size")))
		return
	}

	user, err := ExtractUserFromContext(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var members []models.Member
	var totalCount int64
	if err := api.db.
		Scopes(FilterByUserID(user)).
		Scopes(searchMembersScope(ctx.Query("search"))).
		Scopes(filterByPermitID(ctx.Query("permit_id"))).
		Scopes(Paginate(page, pageSize)).
		Joins("JOIN users ON users.id = members.user_id").
		Where(&models.Member{UserID: user.ID}).
		Select("members.*, COUNT(*) OVER () AS total_count").
		Find(&members).
		Count(&totalCount).
		Order(orderBy).Error; err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list members: %w", err))
		return
	}

	result := ListMembersMembers{
		Members: []ListMembersMember{},
		Total:   int(totalCount),
	}
	for _, member := range members {
		var memberEntries []ListMembersEntry
		if err := api.db.Model(&models.Entry{}).
			Select("entries.band_id, bands.name AS band_name, entries.created_at").
			Joins("JOIN bands ON bands.id = entries.band_id").
			Where("entries.member_id = ? AND entries.confirmed IS TRUE", member.ID.String()).
			Scan(&memberEntries).Error; err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list members: %w", err))
			return
		}
		result.Members = append(result.Members, ListMembersMember{
			ID:         member.ID,
			PermitID:   member.PermitID,
			FirstName:  member.FirstName,
			LastName:   member.LastName,
			Sex:        member.Sex,
			Points:     member.Points,
			Category:   member.Category,
			ClubName:   member.ClubName,
			PermitType: member.PermitType,
			Entries:    memberEntries,
		})
	}

	ctx.JSON(http.StatusOK, &result)
}

func searchMembersScope(search string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if search == "" {
			return db
		}
		return db.Where(
			"members.last_name ILIKE ? OR members.first_name ILIKE ? OR members.club_name ILIKE ? OR users.email ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%",
		)
	}
}

func filterByPermitID(permitID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if permitID == "" {
			return db
		}
		return db.Where("members.permit_id = ?", permitID)
	}
}

func (api *API) GetMember(ctx *gin.Context) {
	claims := jwt.ExtractClaims(ctx)
	userID := uuid.MustParse(claims[auth.IdentityKey].(string))

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid member id: %s", ctx.Param("id")))
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

	ctx.JSON(http.StatusOK, &member)
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
