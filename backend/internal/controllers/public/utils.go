package public

import (
	"fmt"

	"github.com/SuperPingPong/tournoi/internal/auth"
	"github.com/SuperPingPong/tournoi/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Paginate(page int, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func ExtractUserFromContext(ctx *gin.Context) (*models.User, error) {
	userValue, ok := ctx.Get(auth.IdentityKey)
	if !ok {
		return nil, fmt.Errorf("failed to get current user")
	}

	user, ok := userValue.(*models.User)
	if !ok {
		return nil, fmt.Errorf("failed to extract current user from context")
	}

	return user, nil
}

func FilterByUserID(user *models.User, columnName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if user.IsAdmin {
			return db
		}
		return db.Where(fmt.Sprintf("%s = ?", columnName), user.ID)
	}
}
