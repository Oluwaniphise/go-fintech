package wallet

import (
	"errors"
	"fintech/internal/auth"
	"fintech/internal/common"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CreditWalletRequest struct {
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Description string `json:"desc" validate:"omitempty,max=255"`
}
type DebitWalletRequest struct {
	Amount      int64  `json:"amount"`
	Description string `json:"desc"`
}

func (s *WalletService) HandleDebitWallet(c *fiber.Ctx) error {
	req := new(DebitWalletRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"AUTH_INVALID_REQUEST",
			"Invalid request",
			common.ErrorDetail{Details: "request body could not be parsed"},
		))
	}

	if err := s.DebitWallet(c, req.Amount, req.Description); err != nil {
		if errors.Is(err, ErrInsufficientFunds) {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(common.Failure(
				fiber.StatusUnprocessableEntity,
				"INSUFFICIENT FUNDS",
				"Insufficient funds",
				common.ErrorDetail{Details: err.Error()},
			))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"WALLET_DEBIT_FAILED",
			"Failed to debit wallet",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"WALLET_DEBIT_SUCCESS",
		"Wallet debited successfully",
		struct{}{},
	))
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

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"WALLET_CREDIT_SUCCESS",
		"Wallet credited successfully",
		struct{}{},
	))
}

func (s *WalletService) HandleGetBalance(c *fiber.Ctx) error {
	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.Failure(
			fiber.StatusUnauthorized,
			"AUTH_UNAUTHORIZED",
			"Unauthorized: Please login to continue",
			common.ErrorDetail{Details: err.Error()},
		))

	}

	wallet, err := s.GetBalance(userID)
	if err != nil {
		if errors.Is(err, ErrBalanceRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.Failure(
				fiber.StatusNotFound,
				"RECORD_NOT_FOUND",
				"Wallet not found",
				common.ErrorDetail{Details: err.Error()},
			))
		}

		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"WALLET_BALANCE_ERROR",
			"Failed to fetch wallet balance",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"WALLET_BALANCE_SUCCESS",
		"Wallet balance fetched successfully",
		struct {
			UserID   string `json:"userId"`
			Balance  int64  `json:"balance"`
			Currency string `json:"currency"`
		}{
			UserID:   wallet.UserId.String(),
			Balance:  wallet.Balance,
			Currency: wallet.Currency,
		},
	))
}
