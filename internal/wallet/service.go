package wallet

import (
	"errors"
	"fintech/internal/auth"
	"fintech/internal/models"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrBalanceRecordNotFound = errors.New("Balance record not found")
	ErrInsufficientFunds     = errors.New("Insufficient funds")
)

type WalletService struct {
	DB *gorm.DB
}

func (s *WalletService) GetBalance(userID string) (*models.Wallet, error) {

	var wallet models.Wallet
	if err := s.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBalanceRecordNotFound
		}
		return nil, err
	}

	return &wallet, nil

}

func generateTransactionReference() string {
	return "TXN_" + strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", ""))
}

func (s *WalletService) CreditWallet(c *fiber.Ctx, amount int64, desc string) error {
	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet

		if err := tx.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			return err
		}

		// 1. update the wallet balance

		wallet.Balance += amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		// 2. create transaction record (Type)
		reference := generateTransactionReference()

		transaction := models.Transaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        models.Credit,
			Status:      models.Success,
			Reference:   reference,
			Description: desc,
			Category:    "FUNDING",
		}

		return tx.Create(&transaction).Error

	})
}

func (s *WalletService) DebitWallet(c *fiber.Ctx, amount int64, desc string) error {

	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet

		// 1. Get wallet and lock the row so no other process can change it
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			return err
		}

		// 1. check balance

		if wallet.Balance < amount {
			return ErrInsufficientFunds
		}

		// 3. deduct balance

		wallet.Balance -= amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		reference := generateTransactionReference()
		// 4. create transaction record

		transaction := models.Transaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        models.Debit,
			Status:      models.Pending, // we start as pending until API confirms
			Reference:   reference,
			Description: desc,
			Category:    "DEBIT",
		}

		return tx.Create(&transaction).Error
	})
}
