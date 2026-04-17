package bills

import (
	"errors"
	"fintech/internal/models"
	"net/http"

	"gorm.io/gorm"
)

type Dependencies struct {
	DB         *gorm.DB
	HTTPClient *http.Client
}

type Helpers struct {
	Deps Dependencies
}

func NewHelpers(deps Dependencies) *Helpers {
	return &Helpers{Deps: deps}
}

func (h *Helpers) CreatePendingDebit(userID string, amount int64, reference, category, description string) error {
	return h.Deps.DB.Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet

		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", userID).
			First(&wallet).Error; err != nil {
			return err
		}

		if wallet.Balance < amount {
			return errors.New("insufficient funds")
		}

		wallet.Balance -= amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        models.Debit,
			Status:      models.Pending,
			Reference:   reference,
			Description: description,
			Category:    category,
		}

		return tx.Create(&transaction).Error
	})
}

func (h *Helpers) MarkSuccess(reference, outboundProviderRef, providerRef string) error {
	return h.Deps.DB.Model(&models.Transaction{}).
		Where("reference = ?", reference).
		Updates(map[string]any{
			"status":                models.Success,
			"outbound_provider_ref": outboundProviderRef,
			"provider_ref":          providerRef,
		}).Error
}

func (h *Helpers) MarkFailedAndRefund(userID string, amount int64, reference, outboundProviderRef, description string) error {
	return h.Deps.DB.Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet

		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", userID).
			First(&wallet).Error; err != nil {
			return err
		}

		wallet.Balance += amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		return tx.Model(&models.Transaction{}).
			Where("reference = ?", reference).
			Updates(map[string]any{
				"status":                models.Failed,
				"outbound_provider_ref": outboundProviderRef,
				"description":           description,
			}).Error
	})
}
