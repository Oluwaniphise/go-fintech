package models

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerificationToken struct {
	Base
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	TokenHash string    `gorm:"not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
}
