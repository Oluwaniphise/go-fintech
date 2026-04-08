package bills

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fintech/internal/auth"
	"fintech/internal/models"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type BillService struct {
	DB         *gorm.DB
	HTTPClient *http.Client
}

type AirtimeRequest struct {
	ProductCode     string `json:"productCode"`
	ProductItemCode string `json:"productItemCode"`
	CustomerVendID  string `json:"customerVendId"`
	CustomerEmail   string `json:"customerEmail"`
	CustomerPhone   string `json:"customerPhoneNumber"`
	Amount          int64  `json:"amount"`
}

type bondAirtimePayload struct {
	Reference       string `json:"reference"`
	ProductCode     string `json:"productCode"`
	ProductItemCode string `json:"productItemCode"`
	CustomerVendID  string `json:"customerVendId"`
	CustomerEmail   string `json:"customerEmail"`
	CustomerPhone   string `json:"customerPhoneNumber"`
	Amount          int64  `json:"amount"`
}

func (s *BillService) HandleAirtimePurchase(c *fiber.Ctx) error {
	req := new(AirtimeRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.ProductCode == "" || req.ProductItemCode == "" || req.CustomerVendID == "" || req.CustomerEmail == "" || req.CustomerPhone == "" || req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "productCode, productItemCode, customerVendId, customerEmail, customerPhoneNumber and amount are required"})
	}

	appID := c.Get("app-id")
	appSecret := c.Get("app-secret")

	if appID == "" || appSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "app-id and app-secret headers are required"})
	}

	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	internalRef, err := generateSecureReference("AIR")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate transaction reference"})
	}

	if err := s.createPendingDebit(userID, req.Amount, internalRef, "AIRTIME", "Airtime purchase"); err != nil {
		if err.Error() == "insufficient funds" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "insufficient funds"})
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Wallet not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to debit wallet"})
	}

	providerRef, err := generateSecureReference("BOND")
	if err != nil {
		_ = s.markFailedAndRefund(userID, req.Amount, internalRef, "failed to generate provider reference", "")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate provider reference"})
	}

	payload := bondAirtimePayload{
		Reference:       providerRef,
		ProductCode:     req.ProductCode,
		ProductItemCode: req.ProductItemCode,
		CustomerVendID:  req.CustomerVendID,
		CustomerEmail:   req.CustomerEmail,
		CustomerPhone:   req.CustomerPhone,
		Amount:          req.Amount,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to prepare airtime request"})
	}

	httpReq, err := http.NewRequest(http.MethodPost, "https://sandboxapi.iusebond.com/api/v1/vas/pay/airtime", bytes.NewBuffer(body))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create airtime request"})
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("app-id", appID)
	httpReq.Header.Set("app-secret", appSecret)

	resp, err := client.Do(httpReq)
	if err != nil {
		_ = s.markFailedAndRefund(userID, req.Amount, internalRef, "failed to call airtime provider", providerRef)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Failed to call airtime provider"})
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = s.markFailedAndRefund(userID, req.Amount, internalRef, "failed to read airtime provider response", providerRef)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Failed to read airtime provider response"})
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = s.markFailedAndRefund(userID, req.Amount, internalRef, string(respBody), providerRef)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":          "Airtime purchase failed",
			"providerStatus": resp.StatusCode,
			"providerBody":   string(respBody),
		})
	}

	if err := s.markSuccess(internalRef, providerRef); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Airtime was purchased but failed to finalize transaction"})
	}

	var providerResponse map[string]any
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &providerResponse); err != nil {
			providerResponse = map[string]any{
				"raw": string(respBody),
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":           "Airtime purchase successful",
		"reference":         internalRef,
		"providerReference": providerRef,
		"providerResponse":  providerResponse,
	})
}

func (s *BillService) createPendingDebit(userID string, amount int64, reference, category, description string) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
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

func (s *BillService) markSuccess(reference, providerRef string) error {
	return s.DB.Model(&models.Transaction{}).
		Where("reference = ?", reference).
		Updates(map[string]any{
			"status":       models.Success,
			"provider_ref": providerRef,
		}).Error
}

func (s *BillService) markFailedAndRefund(userID string, amount int64, reference, failureReason, providerRef string) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			return err
		}

		wallet.Balance += amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		description := "Airtime purchase failed: " + failureReason
		return tx.Model(&models.Transaction{}).
			Where("reference = ?", reference).
			Updates(map[string]any{
				"status":       models.Failed,
				"provider_ref": providerRef,
				"description":  description,
			}).Error
	})
}

func generateSecureReference(prefix string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", errors.New("unable to generate secure random reference")
	}

	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(randomBytes)), nil
}
