package auth

import (
	"errors"
	"fintech/internal/email"
	"fintech/internal/models"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	DB *gorm.DB
}

type LoginResult struct {
	Token string
	User  models.User
}

func (s *AuthService) RegisterUser(firstName, lastName, email, phone, password string) (*models.User, error) {

	hashedPassword, _ := HashPassword(password)

	user := &models.User{
		FirstName:   firstName,
		LastName:    lastName,
		Email:       email,
		PhoneNumber: phone,
		Password:    hashedPassword,
		IsVerified:  false,
	}

	// use a DB transaction: both must succeed or both fail

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// create user

		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// create the associated Wallet for this user
		wallet := &models.Wallet{
			UserId:  user.ID,
			Balance: 0, // new users start at 0
		}

		if err := tx.Create(wallet).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	_ = s.SendVerificationEmail(*user)

	return user, err
}

func (s *AuthService) Login(email, password string) (*LoginResult, error) {
	var user models.User

	// 1. find user

	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err //user not found
	}

	// 2. check password

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return nil, err // wrong password
	}

	// 3. create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), //expires in 3 days
	})

	// 4. sign token with a secret key from .env
	tokenString, err := token.SignedString([]byte(os.Getenv(("JWT_SECRET"))))

	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token: tokenString,
		User:  user,
	}, nil
}

func (s *AuthService) SendVerificationEmail(user models.User) error {
	rawToken, err := generateEmailVerificationToken()
	if err != nil {
		return err
	}

	tokenHash := hashEmailVerificationToken(rawToken)

	verificationToken := models.EmailVerificationToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	if err := s.DB.Create(&verificationToken).Error; err != nil {
		return err
	}

	baseURL := os.Getenv("EMAIL_VERIFICATION_URL")

	verificationURL, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	query := verificationURL.Query()
	query.Set("token", rawToken)
	verificationURL.RawQuery = query.Encode()

	emailService := email.NewEmailService()

	return emailService.SendVerificationEmail(user.Email, verificationURL.String())
}

func (s *AuthService) VerifyEmail(rawToken string) error {
	if rawToken == "" {
		return errors.New("verification token is required")
	}

	tokenHash := hashEmailVerificationToken(rawToken)

	var verificationToken models.EmailVerificationToken

	err := s.DB.Where("token_hash = ?", tokenHash).First(&verificationToken).Error
	if err != nil {
		return err
	}

	if verificationToken.UsedAt != nil {
		return errors.New("verification token has already been used")
	}

	if time.Now().After(verificationToken.ExpiresAt) {
		return errors.New("verification token has expired")
	}

	now := time.Now()

	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).
			Where("id = ?", verificationToken.UserID).
			Update("is_verified", true).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.EmailVerificationToken{}).
			Where("id = ?", verificationToken.ID).
			Update("used_at", &now).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *AuthService) ResendVerificationEmail(userEmail string) error {
	var user models.User

	if err := s.DB.Where("email = ?", userEmail).First(&user).Error; err != nil {
		return err
	}

	if user.IsVerified {
		return errors.New("email is already verified")
	}

	// Optional: invalidate old unused tokens for this user.
	if err := s.DB.Model(&models.EmailVerificationToken{}).
		Where("user_id = ? AND used_at IS NULL", user.ID).
		Update("used_at", time.Now()).Error; err != nil {
		return err
	}

	return s.SendVerificationEmail(user)
}
