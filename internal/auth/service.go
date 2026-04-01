package auth

import (
	"fintech/internal/models"

	"gorm.io/gorm"
)

type AuthService struct {
	DB *gorm.DB
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
