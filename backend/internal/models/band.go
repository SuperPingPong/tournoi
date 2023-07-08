package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Band struct {
	ID   uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();->"`
	Name string    `gorm:"not null;unique"`
	Day  int       `gorm:"not null"`
}

type Entry struct {
	BandID    uuid.UUID `gorm:"primaryKey"`
	MemberID  uuid.UUID `gorm:"primaryKey"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type EntryLock struct {
	BandID    uuid.UUID `gorm:"primaryKey"`
	MemberID  uuid.UUID `gorm:"primaryKey"`
	CreatedAt time.Time
	ExpiresAt time.Time `gorm:"index"`
}
