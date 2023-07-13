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

	"github.com/getsentry/sentry-go"
)

type ListMembersEntry struct {
	BandID    uuid.UUID
	BandName  string
	BandPrice int
	CreatedAt time.Time
}

type ListMembersUser struct {
	UserID    uuid.UUID
	UserEmail string
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
	User       ListMembersUser
}

type ListMembersMembers struct {
	Members []ListMembersMember
	IsAdmin bool
	Total   int
}

func (api *API) ListMembers(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		sentry.CaptureException(fmt.Errorf("invalid page: %s", ctx.Query("page")))
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid page: %s", ctx.Query("page")))
		return
	}
	pageSize, err := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	if err != nil {
		sentry.CaptureException(fmt.Errorf("invalid page size: %s", ctx.Query("page_size")))
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid page size: %s", ctx.Query("page_size")))
		return
	}
	validOrderBy := map[string]string{
		"created_at_asc":  "members.created_at ASC",
		"created_at_desc": "members.created_at DESC",
		"last_name_asc":   "members.last_name ASC",
		"last_name_desc":  "members.last_name DESC",
	}
	orderBy, valid := validOrderBy[ctx.DefaultQuery("order_by", "created_at_desc")]
	if !valid {
		sentry.CaptureException(fmt.Errorf("invalid page size: %s", ctx.Query("page_size")))
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid page size: %s", ctx.Query("page_size")))
		return
	}

	user, err := ExtractUserFromContext(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var totalCount int64
	if err := api.db.
		Model(&models.Member{}).
		Scopes(FilterByUserID(user)).
		Scopes(searchMembersScope(ctx.Query("search"), *user)).
		Scopes(filterByPermitID(ctx.Query("permit_id"))).
		Joins("JOIN users ON users.id = members.user_id").
		Select("COUNT(*) AS total_count").
		Count(&totalCount).Error; err != nil {
		sentry.CaptureException(fmt.Errorf("failed to count members: %w", err))
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to count members: %w", err))
		return
	}

	var members []models.Member
	if err := api.db.
		Model(&models.Member{}).
		Scopes(FilterByUserID(user)).
		Scopes(searchMembersScope(ctx.Query("search"), *user)).
		Scopes(filterByPermitID(ctx.Query("permit_id"))).
		Scopes(Paginate(page, pageSize)).
		Joins("JOIN users ON users.id = members.user_id").
		Select("members.*").
		Order(orderBy).
		Find(&members).
		Error; err != nil {
		sentry.CaptureException(fmt.Errorf("failed to list members: %w", err))
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list members: %w", err))
		return
	}

	result := ListMembersMembers{
		Members: []ListMembersMember{},
		IsAdmin: user.IsAdmin,
		Total:   int(totalCount),
	}

	for _, member := range members {
		var memberEntries []ListMembersEntry
		if err := api.db.Model(&models.Entry{}).
			Select("entries.band_id, bands.name AS band_name, bands.price AS band_price, entries.created_at").
			Joins("JOIN bands ON bands.id = entries.band_id").
			Where("entries.member_id = ? AND entries.confirmed IS TRUE", member.ID.String()).
			Order("bands.created_at ASC").
			Scan(&memberEntries).Error; err != nil {
			sentry.CaptureException(fmt.Errorf("failed to list members: %w", err))
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list members: %w", err))
			return
		}
		var memberUser ListMembersUser
		if user.IsAdmin {
			if err := api.db.Model(&models.User{}).
				Select("users.id AS user_id, users.email AS user_email").
				Joins("JOIN members ON members.user_id = users.id").
				Where("members.id = ?", member.ID.String()).
				Scan(&memberUser).Error; err != nil {
				sentry.CaptureException(fmt.Errorf("failed to list members: %w", err))
				ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list members: %w", err))
				return
			}
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
			User:       memberUser,
		})
	}

	ctx.JSON(http.StatusOK, &result)
}

func searchMembersScope(search string, user models.User) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if search == "" {
			return db
		}
		if user.IsAdmin {
			return db.Where(
				"members.last_name ILIKE ? OR members.first_name ILIKE ? OR members.club_name ILIKE ? OR users.email ILIKE ?",
				"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%",
			)
		}
		return db.Where(
			"members.last_name ILIKE ? OR members.first_name ILIKE ? OR members.club_name ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%",
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
		sentry.CaptureException(fmt.Errorf("invalid member id: %s", ctx.Param("id")))
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid member id: %s", ctx.Param("id")))
		return
	}

	user, err := ExtractUserFromContext(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	member := models.Member{}
	err = api.db.First(&member, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sentry.CaptureException(fmt.Errorf("member %s not found", id))
			ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("member %s not found", id))
			return
		}

		sentry.CaptureException(fmt.Errorf("failed to get member: %w", err))
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get member: %w", err))
		return
	}

	if member.UserID != userID && user.IsAdmin == false {
		sentry.CaptureException(fmt.Errorf("member %s not found", id))
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
		sentry.CaptureException(fmt.Errorf("invalid input: %w", err))
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
		return
	}

	data, err := api.GetFFTTPlayerData(input.PermitID)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("failed to get member data from FFTT: %w", err))
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
			sentry.CaptureException(fmt.Errorf("member with permit %s already exists", input.PermitID))
			ctx.AbortWithError(http.StatusConflict, fmt.Errorf("member with permit %s already exists", input.PermitID))
			return
		}
		sentry.CaptureException(fmt.Errorf("failed to create member: %w", err))
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
		sentry.CaptureException(fmt.Errorf("invalid member id: %s", ctx.Param("id")))
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid member id: %s", ctx.Param("id")))
		return
	}

	var input UpdateMemberInput
	err = ctx.ShouldBindJSON(&input)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("invalid input: %w", err))
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid input: %w", err))
		return
	}

	user, err := ExtractUserFromContext(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	member := models.Member{}
	err = api.db.First(&member, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sentry.CaptureException(fmt.Errorf("member %s not found", id))
			ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("member %s not found", id))
			return
		}

		sentry.CaptureException(fmt.Errorf("failed to get member: %w", err))
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get member: %w", err))
		return
	}

	if member.UserID != userID || user.IsAdmin == false {
		sentry.CaptureException(fmt.Errorf("member %s not found", id))
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("member %s not found", id))
		return
	}

	if input.PermitID != "" {
		data, err := api.GetFFTTPlayerData(input.PermitID)
		if err != nil {
			sentry.CaptureException(fmt.Errorf("failed to get member data from FFTT: %w", err))
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
		sentry.CaptureException(fmt.Errorf("failed to update member: %w", err))
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to update member: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, &member)
}

func (api *API) DeleteMember(ctx *gin.Context) {
	claims := jwt.ExtractClaims(ctx)
	userID := uuid.MustParse(claims[auth.IdentityKey].(string))

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		sentry.CaptureException(fmt.Errorf("invalid member id: %s", ctx.Param("id")))
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid member id: %s", ctx.Param("id")))
		return
	}

	user, err := ExtractUserFromContext(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	member := models.Member{}
	err = api.db.First(&member, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sentry.CaptureException(fmt.Errorf("member %s not found", id))
			ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("member %s not found", id))
			return
		}

		sentry.CaptureException(fmt.Errorf("failed to get member: %w", err))
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get member: %w", err))
		return
	}

	if member.UserID != userID && user.IsAdmin == false {
		sentry.CaptureException(fmt.Errorf("member %s not found", id))
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("member %s not found", id))
		return
	}

	// Delete member and related entries
	err = api.db.Transaction(func(tx *gorm.DB) error {
		// Delete entries
		if err := tx.
			Where("member_id = ?", id).
			Delete(&models.Entry{}).
			Error; err != nil {
			return fmt.Errorf("failed to delete entries: %w", err)
		}

		// Delete member
		if err := tx.
			Where("id = ?", id).
			Delete(&models.Member{}, id).
			Error; err != nil {
			return fmt.Errorf("failed to delete member: %w", err)
		}

		return nil
	})

	if err != nil {
		sentry.CaptureException(fmt.Errorf("failed to make delete transaction: %w", err))
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to make delete transaction: %w", err))
		return
	}

	ctx.Status(http.StatusNoContent)
}
