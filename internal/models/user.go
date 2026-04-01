package models

type User struct {
	Base
	FirstName   string `gorm:"size:100;not null" json:"firstName"`
	LastName    string `gorm:"size:100;not null" json:"lastName"`
	Email       string `gorm:"size:255; uniqueIndex;not null" json:"email"`
	Password    string `gorm:"not null" json:"-"` //hidden from JSON
	PhoneNumber string `gorm:"size:20;uniqueIndex;not null" json:"phoneNumber"`
	IsVerified  bool   `gorm:"default:false" json:"is_verified"`

	// Relationships

	Wallet Wallet `json:"wallet"`
}
