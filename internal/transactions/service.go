package transactions

import (
	"fintech/internal/auth"
	"fintech/internal/common"
	"fintech/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionService struct {
	DB *gorm.DB
}

type TransactionStats struct {
	TotalTransactions      int64 `json:"totalTransactions"`
	SuccessfulTransactions int64 `json:"successfulTransactions"`
	FailedTransactions     int64 `json:"failedTransactions"`
	PendingTransactions    int64 `json:"pendingTransactions"`
	TotalCredits           int64 `json:"totalCredits"`
	TotalDebits            int64 `json:"totalDebits"`
	SuccessfulCreditAmount int64 `json:"successfulCreditAmount"`
	SuccessfulDebitAmount  int64 `json:"successfulDebitAmount"`
}

type PaginatedTransactionsResponse struct {
	Transactions []models.Transaction  `json:"transactions"`
	Meta         common.PaginationMeta `json:"meta"`
}

func (s *TransactionService) GetUserTransactions(c *fiber.Ctx) (*PaginatedTransactionsResponse, error) {
	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	parsedUserID, err := uuid.Parse(userID)

	if err != nil {
		return nil, err
	}

	pagination := common.ParsePagination(c)
	var transactions []models.Transaction
	var total int64

	err = s.DB.Model(&models.Transaction{}).
		Joins("JOIN wallets ON wallets.id = transactions.wallet_id").
		Where("wallets.user_id = ?", parsedUserID).
		Count(&total).Error

	if err != nil {
		return nil, err
	}

	err = s.DB.Model(&models.Transaction{}).
		Joins("JOIN wallets ON wallets.id = transactions.wallet_id").
		Where("wallets.user_id = ?", parsedUserID).
		Order("transactions.created_at DESC").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&transactions).Error

	if err != nil {
		return nil, err
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	return &PaginatedTransactionsResponse{
		Transactions: transactions,
		Meta:         common.BuildPaginationMeta(pagination.Page, pagination.Limit, total),
	}, nil
}

func (s *TransactionService) GetUserTransactionStats(c *fiber.Ctx) (TransactionStats, error) {
	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		return TransactionStats{}, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return TransactionStats{}, err
	}

	var stats TransactionStats

	err = s.DB.Model(&models.Transaction{}).
		Select(`
			COUNT(transactions.id) AS total_transactions,
			SUM(CASE WHEN transactions.status = ? THEN 1 ELSE 0 END) AS successful_transactions,
			SUM(CASE WHEN transactions.status = ? THEN 1 ELSE 0 END) AS failed_transactions,
			SUM(CASE WHEN transactions.status = ? THEN 1 ELSE 0 END) AS pending_transactions,
			COALESCE(SUM(CASE WHEN transactions.type = ? THEN transactions.amount ELSE 0 END), 0) AS total_credits,
			COALESCE(SUM(CASE WHEN transactions.type = ? THEN transactions.amount ELSE 0 END), 0) AS total_debits,
			COALESCE(SUM(CASE WHEN transactions.type = ? AND transactions.status = ? THEN transactions.amount ELSE 0 END), 0) AS successful_credit_amount,
			COALESCE(SUM(CASE WHEN transactions.type = ? AND transactions.status = ? THEN transactions.amount ELSE 0 END), 0) AS successful_debit_amount
		`,
			models.Success,
			models.Failed,
			models.Pending,
			models.Credit,
			models.Debit,
			models.Credit, models.Success,
			models.Debit, models.Success,
		).
		Joins("JOIN wallets ON wallets.id = transactions.wallet_id").
		Where("wallets.user_id = ?", parsedUserID).
		Scan(&stats).Error

	if err != nil {
		return TransactionStats{}, err
	}

	return stats, nil
}
