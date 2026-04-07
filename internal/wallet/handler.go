package wallet

import (
	"fintech/internal/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type WalletService struct {
	DB *gorm.DB
}

func (s *WalletService) GetBalance(c *fiber.Ctx) error {
	// Fiber JWT middleware stores the user token in c.Locals("user").
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
	}

	userID, ok := claims["user_id"]
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user_id missing in token"})
	}

	var wallet models.Wallet
	if err := s.DB.Where("user_id = ?", fmt.Sprint(userID)).First(&wallet).Error; err != nil {
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
