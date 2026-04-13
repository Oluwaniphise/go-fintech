package auth

import (
	"fintech/internal/models"
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
