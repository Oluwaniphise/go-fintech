package wallet

import (
	"errors"
	"fintech/internal/common"

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
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"AUTH_INVALID_REQUEST",
			"Invalid request",
			common.ErrorDetail{Details: "request body could not be parsed"},
		))
	}

	if err := s.CreditWallet(c, req.Amount, req.Description); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.Failure(
				fiber.StatusNotFound,
				"WALLET_NOT_FOUND",
				"Wallet not found",
				common.ErrorDetail{Details: err.Error()},
			))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"WALLET_CREDIT_FAILED",
			"Failed to credit wallet",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusInternalServerError).JSON(common.Success(
		fiber.StatusOK,
		"WALLET_CREDIT_SUCCESS",
		"Wallet credited successfully",
		struct{}{},
	))
}
