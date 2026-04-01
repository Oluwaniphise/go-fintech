package models

import "github.com/google/uuid"

type TransactionType string

const (
	Credit TransactionType = "CREDIT"
	Debit  TransactionType = "DEBIT"
)

type TransactionStatus string

const (
	Pending TransactionStatus = "PENDING"
	Success TransactionStatus = "SUCCESS"
	Failed  TransactionStatus = "FAILED"
)

type Transaction struct {
	Base
	WalletID    uuid.UUID         `gorm:"type:uuid;not null" json:"wallet_id"`
	Amount      int64             `gorm:"not null" json:"amount"`
	Type        TransactionType   `gorm:"type:varchar(10)" json:"type"`
	Status      TransactionStatus `gorm:"type:varchar(10);default:'PENDING'" json:"status"`
	Reference   string            `gorm:"uniqueIndex;not null" json:"reference"` //  Internal ID
	ProviderRef string            `json:"provider_reference"`                    // The 3rd party ID
	Category    string            `json:"category"`                              // e.g., "AIRTIME", "DATA", "TRANSFER"
	Description string            `json:"description"`
}
