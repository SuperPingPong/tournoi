package auth

import (
	"time"

	"github.com/SuperPingPong/tournoi/internal/models"

	"gorm.io/gorm"
)

type AuthBusiness struct {
	db *gorm.DB
}

func NewAuthBusiness(db *gorm.DB) *AuthBusiness {
	c := &AuthBusiness{
		db: db,
	}
	return c
}

type LoginRequest struct {
	Email  string `json:"email" binding:"required,email"`
	Secret string `json:"secret" binding:"required,len=6,numeric"`
}

func (a *AuthBusiness) Login(email string, secret string) (*models.User, error) {
	var otp models.OTP
	var user models.User

	err := a.db.Where("email = ? AND secret = ? AND expires_at > ? AND deleted_at IS NULL", email, secret, time.Now()).First(&otp).Error
	if err != nil {
		return nil, err
	}

	err = a.db.FirstOrCreate(&user, models.User{
		Email: email,
	}).Error
	if err != nil {
		return nil, err
	}

	a.db.Delete(&otp)

	return &user, nil
}
