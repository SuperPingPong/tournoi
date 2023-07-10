package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Member struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();->"`
	PermitID   string    `gorm:"not null;unique;index;where:deleted_at IS NULL"`
	FirstName  string    `gorm:"not null"`
	LastName   string    `gorm:"not null"`
	Sex        string    `gorm:"not null"`
	Points     float64   `gorm:"not null"`
	Category   string
	ClubName   string
	PermitType string

	CreatedAt time.Time      `gorm:"<-:create;not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserID uuid.UUID
}
