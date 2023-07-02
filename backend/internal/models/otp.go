package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OTP struct {
	ID     uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email  string    `gorm:"not null"`
	Secret string    `gorm:"not null"`

	CreatedAt time.Time      `gorm:"<-:create;not null"`
	ExpiresAt time.Time      `gorm:"index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
