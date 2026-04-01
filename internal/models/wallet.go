package models

import "github.com/google/uuid"

type Wallet struct {
	Base
	UserId uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`

	// balance in kobo
	Balance  int64  `gorm:"default:0" json:"balance"`
	Currency string `gorm:"size:3;default:'NGN'" json:"currency"`
}
