package transactions

import (
	"fintech/internal/common"
	"fintech/internal/models"

	"github.com/gofiber/fiber/v2"
)

func (s *TransactionService) HandleGetUserTransactions(c *fiber.Ctx) error {

	transactions, err := s.GetUserTransactions(c)

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not fetch transactions"})
	}

	return c.Status(fiber.StatusOK).JSON((common.Success(
		fiber.StatusOK,
		"TRANSACTION_SUCCESS",
		"Transactions fetched successfully.",
		transactions,
	)))
}
