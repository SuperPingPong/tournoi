package models

import (
	"github.com/google/uuid"
)

type BandStatus string

type Band struct {
	ID   uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();->"`
	Name string    `gorm:"not null"`

	Members []*Member `gorm:"many2many:bands__members"`
}
