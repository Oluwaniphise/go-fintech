package models

import (
	"time"

	"github.com/google/uuid"
)

type LoginOTPToken struct {
	Base
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`
	CodeHash     string    `gorm:"not null"`
	ExpiresAt    time.Time `gorm:"not null"`
	UsedAt       *time.Time
	AttemptCount int `gorm:"default:0"`
}
