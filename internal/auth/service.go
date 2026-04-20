package auth

import (
	"errors"
	emailService "fintech/internal/email"
	"fintech/internal/models"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailNotVerified  = errors.New("email not verified")
	ErrInvalidOTP        = errors.New("invalid otp")
	ErrOTPExpired        = errors.New("otp expired")
	ErrOTPUsed           = errors.New("otp already used")
	ErrOTPWaitFor1Minute = errors.New("please wait before requesting another OTP")
)

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

func (s *AuthService) StartLogin(email, password string) (*LoginOTPStartResult, error) {
	var user models.User

	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, err
	}

	if !user.IsVerified {
		return nil, ErrEmailNotVerified
	}

	otp, err := generateLoginOTP()
	if err != nil {
		return nil, err
	}

	loginOTP := models.LoginOTPToken{
		UserID:    user.ID,
		CodeHash:  hashLoginOTP(otp),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := s.DB.Create(&loginOTP).Error; err != nil {
		return nil, err
	}

	emailService := emailService.NewEmailService()
	if err := emailService.SendLoginOTPEmail(user.Email, otp); err != nil {
		return nil, err
	}

	return &LoginOTPStartResult{
		Reference: loginOTP.ID.String(),
	}, nil
}

func (s *AuthService) VerifyLoginOTP(Reference, otp string) (*LoginResult, error) {
	referenceUUID, err := uuid.Parse(Reference)
	if err != nil {
		return nil, ErrInvalidOTP
	}

	var otpRecord models.LoginOTPToken
	if err := s.DB.Where("id = ?", referenceUUID).First(&otpRecord).Error; err != nil {
		return nil, ErrInvalidOTP
	}

	if otpRecord.UsedAt != nil {
		return nil, ErrOTPUsed
	}

	if time.Now().After(otpRecord.ExpiresAt) {
		return nil, ErrOTPExpired
	}

	if otpRecord.AttemptCount >= 5 {
		return nil, ErrInvalidOTP
	}

	if otpRecord.CodeHash != hashLoginOTP(otp) {
		_ = s.DB.Model(&models.LoginOTPToken{}).
			Where("id = ?", otpRecord.ID).
			Update("attempt_count", gorm.Expr("attempt_count + 1")).Error
		return nil, ErrInvalidOTP
	}

	var user models.User
	if err := s.DB.Where("id = ?", otpRecord.UserID).First(&user).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	if err := s.DB.Model(&models.LoginOTPToken{}).
		Where("id = ?", otpRecord.ID).
		Update("used_at", &now).Error; err != nil {
		return nil, err
	}

	// tokenString, err := s.SignJWT(user)
	// if err != nil {
	// 	return nil, err
	// }

	return &LoginResult{
		// Token: tokenString,
		User: user,
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

	emailService := emailService.NewEmailService()

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

func (s *AuthService) SignJWT(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func (s *AuthService) ResendLoginOTP(reference string) (*ResendLoginOTPResult, error) {
	referenceUUID, err := uuid.Parse(reference)
	if err != nil {
		return nil, errors.New("invalid reference id")
	}

	var existing models.LoginOTPToken
	if err := s.DB.Where("id = ?", referenceUUID).First(&existing).Error; err != nil {
		return nil, errors.New("login reference not found")
	}

	var user models.User
	if err := s.DB.Where("id = ?", existing.UserID).First(&user).Error; err != nil {
		return nil, err
	}

	if time.Since(existing.CreatedAt) < 60*time.Second {
		return nil, ErrOTPWaitFor1Minute
	}

	otp, err := generateLoginOTP()
	if err != nil {
		return nil, err
	}

	newToken := models.LoginOTPToken{
		UserID:    user.ID,
		CodeHash:  hashLoginOTP(otp),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		if err := tx.Model(&models.LoginOTPToken{}).
			Where("user_id = ? AND used_at IS NULL", user.ID).
			Update("used_at", &now).Error; err != nil {
			return err
		}

		if err := tx.Create(&newToken).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	emailService := emailService.NewEmailService()
	if err := emailService.SendLoginOTPEmail(user.Email, otp); err != nil {
		return nil, err
	}

	return &ResendLoginOTPResult{
		Reference: newToken.ID.String(),
	}, nil
}
