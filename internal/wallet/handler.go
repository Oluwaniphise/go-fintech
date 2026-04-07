package wallet

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CreditWalletRequest struct {
	Amount      int64  `json:"amount"`
	Description string `json:"desc"`
}

func (s *WalletService) HandleCreditWallet(c *fiber.Ctx) error {
	req := new(CreditWalletRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := s.CreditWallet(c, req.Amount, req.Description); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Wallet not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to credit wallet"})
	}

	return c.JSON(fiber.Map{"message": "Wallet credited successfully"})
}
