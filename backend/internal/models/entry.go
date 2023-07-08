package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const EntryLockExpirationDelay = 10 * time.Minute

type Entry struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();->"`
	BandID   uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_entry_band_id_member_id,where:deleted_at IS NULL"`
	MemberID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_entry_band_id_member_id,where:deleted_at IS NULL"`

	CreatedAt time.Time      `gorm:"<-:create;not null"`
	ExpiresAt time.Time      `gorm:"index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Confirmed bool `gorm:"not null;default:false"`
}
