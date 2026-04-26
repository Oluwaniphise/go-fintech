package auth

import (
	"fintech/internal/models"

	"gorm.io/gorm"
)

type AuthService struct {
	DB *gorm.DB
}

type LoginResult struct {
	// Token string
	User models.User
}

type LoginOTPStartResult struct {
	Reference string
}

type RegisterRequest struct {
	FirstName   string `json:"firstName" validate:"required,min=2,max=100"`
	LastName    string `json:"lastName"  validate:"required,min=2,max=100"`
	Email       string `json:"email" validate:"required,email,max=255"`
	PhoneNumber string `json:"phoneNumber" validate:"required,e164,max=20"`
	Password    string `json:"password" validate:"required,password,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	User        struct {
		ID          string `json:"id"`
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
	} `json:"user"`
}

type LoginStartResponse struct {
	Reference string `json:"reference"`
}

type VerifyLoginOTPRequest struct {
	Reference string `json:"reference" validate:"required,uuid4"`
	OTP       string `json:"otp" validate:"required,len=6,numeric"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type ResendVerificationEmailRequest struct {
	Email string `json:"email"  validate:"required,email,max=255"`
}

type ResendLoginOTPRequest struct {
	Reference string `json:"reference"  validate:"required,uuid4"`
}

type ResendLoginOTPResponse struct {
	Reference string `json:"reference"`
}

type ResendLoginOTPResult struct {
	Reference string
}

type MeResponse struct {
	ID          string `json:"id"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	IsVerified  bool   `json:"isVerified"`
}
