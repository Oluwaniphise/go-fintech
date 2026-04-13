package transactions

import (
	"fintech/internal/auth"
	"fintech/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionService struct {
	DB *gorm.DB
}

func (s *TransactionService) GetUserTransactions(c *fiber.Ctx) ([]models.Transaction, error) {
	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	uuid, err := uuid.Parse(userID)

	if err != nil {
		return nil, err
	}

	var transactions []models.Transaction

	err = s.DB.Model(&models.Transaction{}).
		Joins("JOIN wallets ON wallets.id = transactions.wallet_id").
		Where("wallets.user_id = ?", uuid).
		Order("transactions.created_at DESC").
		Find(&transactions).Error

	if err != nil {
		return nil, err
	}

	return transactions, nil
}
