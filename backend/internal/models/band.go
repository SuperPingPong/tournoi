package models

import (
	"github.com/google/uuid"
)

const (
	BandSex_M   string = "M"
	BandSex_F          = "F"
	BandSex_ALL        = "ALL"
)

type Band struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();->"`
	Name       string    `gorm:"not null;unique"`
	Day        int       `gorm:"not null"`
	Sex        string    `gorm:"not null"`
	MaxPoints  float64   `gorm:"not null"`
	MaxEntries int       `gorm:"not null"`
}
