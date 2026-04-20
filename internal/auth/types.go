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
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	Password    string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
	Reference string `json:"reference"`
	OTP       string `json:"otp"`
}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

type ResendVerificationEmailRequest struct {
	Email string `json:"email"`
}

type ResendLoginOTPRequest struct {
	Reference string `json:"reference"`
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
