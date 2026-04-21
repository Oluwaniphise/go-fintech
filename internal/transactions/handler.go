package transactions

import (
	"fintech/internal/auth"
	"fintech/internal/common"

	"github.com/gofiber/fiber/v2"
)

func (s *TransactionService) HandleGetUserTransactions(c *fiber.Ctx) error {

	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.Failure(
			fiber.StatusUnauthorized,
			"AUTH_UNAUTHORIZED",
			"Unauthorized: Please login to continue",
			common.ErrorDetail{Details: err.Error()},
		))

	}

	result, err := s.GetUserTransactions(c, userID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"TRANSACTIONS_FAILED",
			"Could not fetch transactions",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusOK).JSON((common.Success(
		fiber.StatusOK,
		"TRANSACTION_SUCCESS",
		"Transactions fetched successfully.",
		result,
	)))
}

func (s *TransactionService) HandleGetUserTransactionStats(c *fiber.Ctx) error {
	stats, err := s.GetUserTransactionStats(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"TRANSACTION_STATS_ERROR",
			"Failed to fetch transaction stats",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"TRANSACTION_STATS_SUCCESS",
		"Transaction stats fetched successfully.",
		stats,
	))
}
