package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshSession struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	TokenHash  string    `gorm:"not null;uniqueIndex"`
	ExpiresAt  time.Time `gorm:"not null"`
	RevokedAt  *time.Time
	LastUsedAt *time.Time
	UserAgent  string
	IPAddress  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
