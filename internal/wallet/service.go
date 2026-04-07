package wallet

import (
	"fintech/internal/auth"
	"fintech/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type WalletService struct {
	DB *gorm.DB
}

func (s *WalletService) GetBalance(c *fiber.Ctx) error {
	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var wallet models.Wallet
	if err := s.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Wallet not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch wallet balance"})
	}

	return c.JSON(fiber.Map{
		"user_id":  wallet.UserId,
		"balance":  wallet.Balance,
		"currency": wallet.Currency,
	})
}
func (s *WalletService) CreditWallet(c *fiber.Ctx, amount int64, reference string, desc string) error {
	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
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
