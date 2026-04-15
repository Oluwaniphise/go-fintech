package transactions

import (
	"fintech/internal/common"

	"github.com/gofiber/fiber/v2"
)

func (s *TransactionService) HandleGetUserTransactions(c *fiber.Ctx) error {

	result, err := s.GetUserTransactions(c)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not fetch transactions"})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch transaction stats"})
	}

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"TRANSACTION_STATS_SUCCESS",
		"Transaction stats fetched successfully.",
		stats,
	))
}
