package airtime

import (
	"errors"
	"fintech/internal/bills"
	"fintech/internal/models"

	"gorm.io/gorm"
)

type Service struct {
	Deps bills.Dependencies
}

var AIRTIME_ENDPOINT = "vas/pay/airtime"

func NewService(deps bills.Dependencies) *Service {
	return &Service{Deps: deps}
}

func (s *Service) createPendingDebit(userID string, amount int64, reference, category, description string) error {
	return s.Deps.DB.Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet

		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&wallet).Error; err != nil {
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

func (s *Service) markSuccess(reference, outboundProviderRef, providerRef string) error {
	return s.Deps.DB.Model(&models.Transaction{}).
		Where("reference = ?", reference).
		Updates(map[string]any{
			"status":                models.Success,
			"outbound_provider_ref": outboundProviderRef,
			"provider_ref":          providerRef,
		}).Error
}

func (s *Service) markFailedAndRefund(userID string, amount int64, reference, _ string, outboundProviderRef string) error {
	return s.Deps.DB.Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			return err
		}

		wallet.Balance += amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		description := "Airtime purchase failed"
		return tx.Model(&models.Transaction{}).
			Where("reference = ?", reference).
			Updates(map[string]any{
				"status":                models.Failed,
				"outbound_provider_ref": outboundProviderRef,
				"description":           description,
			}).Error
	})
}
