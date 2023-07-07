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

	Members []*Member `gorm:"many2many:bands__members;joinForeignKey:BandID"`
}

type BandMember struct {
	BandID    uuid.UUID `gorm:"primaryKey"`
	MemberID  uuid.UUID `gorm:"primaryKey"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (BandMember) TableName() string {
	return "bands__members"
}
