package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID      uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email   string    `gorm:"not null"`
	IsAdmin bool      `gorm:"not null"`

	CreatedAt time.Time      `gorm:"<-:create;not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Members []Member
}
