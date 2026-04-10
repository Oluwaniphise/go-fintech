package bills

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fintech/internal/auth"
	"fintech/internal/common"
	"fintech/internal/models"
	"fintech/internal/providers/bond"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var AIRTIME_ENDPOINT = "vas/pay/airtime"

func (s *BillService) HandleAirtimePurchase(c *fiber.Ctx) error {
	req := new(AirtimeRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"BILL_INVALID_REQUEST",
			"Invalid request",
			common.ErrorDetail{Details: "request body could not be parsed"},
		))
	}

	if req.ProductCode == "" || req.ProductItemCode == "" || req.CustomerVendID == "" || req.CustomerEmail == "" || req.CustomerPhone == "" || req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"BILL_MISSING_FIELDS",
			"productCode, productItemCode, customerVendId, customerEmail, customerPhoneNumber and amount are required",
			nil,
		))
	}

	appID := c.Get("app-id")
	appSecret := c.Get("app-secret")

	if appID == "" || appSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"BILL_MISSING_HEADERS",
			"app-id and app-secret headers are required",
			nil,
		))
	}

	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.Failure(
			fiber.StatusUnauthorized,
			"AUTH_UNAUTHORIZED",
			err.Error(),
			nil,
		))
	}

	internalRef, err := generateSecureReference("AIR")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_REFERENCE_GENERATION_FAILED",
			"Failed to generate transaction reference",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	if err := s.createPendingDebit(userID, req.Amount, internalRef, "AIRTIME", "Airtime purchase"); err != nil {
		if err.Error() == "insufficient funds" {
			return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
				fiber.StatusBadRequest,
				"BILL_INSUFFICIENT_FUNDS",
				"insufficient funds",
				nil,
			))
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.Failure(
				fiber.StatusNotFound,
				"WALLET_NOT_FOUND",
				"Wallet not found",
				nil,
			))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_WALLET_DEBIT_FAILED",
			"Failed to debit wallet",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	outboundProviderRef, err := generateSecureReference("BOND")
	if err != nil {
		_ = s.markFailedAndRefund(userID, req.Amount, internalRef, "failed to generate provider reference", "")
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_PROVIDER_REFERENCE_FAILED",
			"Failed to generate provider reference",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	payload := bondAirtimePayload{
		Reference:       outboundProviderRef,
		ProductCode:     req.ProductCode,
		ProductItemCode: req.ProductItemCode,
		CustomerVendID:  req.CustomerVendID,
		CustomerEmail:   req.CustomerEmail,
		CustomerPhone:   req.CustomerPhone,
		Amount:          req.Amount,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_PROVIDER_REQUEST_PREP_FAILED",
			"Failed to prepare airtime request",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	httpReq, err := http.NewRequest(http.MethodPost, os.Getenv("BOND_SANDBOX_API")+AIRTIME_ENDPOINT, bytes.NewBuffer(body))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_PROVIDER_REQUEST_CREATE_FAILED",
			"Failed to create airtime request",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("app-id", appID)
	httpReq.Header.Set("app-secret", appSecret)

	resp, err := client.Do(httpReq)
	if err != nil {
		_ = s.markFailedAndRefund(userID, req.Amount, internalRef, "failed to call airtime provider", outboundProviderRef)
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			"BILL_PROVIDER_CALL_FAILED",
			"Failed to call airtime provider",
			common.ErrorDetail{Details: err.Error()},
		))
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = s.markFailedAndRefund(userID, req.Amount, internalRef, "failed to read airtime provider response", outboundProviderRef)
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			"BILL_PROVIDER_RESPONSE_READ_FAILED",
			"Failed to read airtime provider response",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	// fmt.Printf("airtime provider response status=%d body=%s\n", resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = s.markFailedAndRefund(userID, req.Amount, internalRef, string(respBody), outboundProviderRef)
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			"BILL_AIRTIME_PURCHASE_FAILED",
			"Airtime purchase failed",
			struct {
				ProviderStatus int    `json:"providerStatus"`
				ProviderBody   string `json:"providerBody"`
			}{
				ProviderStatus: resp.StatusCode,
				ProviderBody:   string(respBody),
			},
		))
	}

	var providerResponse bond.Response[bond.AirtimeData]
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &providerResponse); err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
				fiber.StatusBadGateway,
				"BILL_PROVIDER_RESPONSE_INVALID",
				"Invalid airtime provider response",
				struct {
					ProviderBody string `json:"providerBody"`
				}{
					ProviderBody: string(respBody),
				},
			))
		}
	}

	if err := s.markSuccess(internalRef, outboundProviderRef, providerResponse.Data.PaymentReference); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_FINALIZATION_FAILED",
			"Airtime was purchased but failed to finalize transaction",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"BILL_AIRTIME_PURCHASE_SUCCESS",
		"Airtime purchase successful",
		struct {
			Reference         string                          `json:"reference"`
			ProviderReference string                          `json:"providerReference"`
			ProviderResponse  bond.Response[bond.AirtimeData] `json:"providerResponse"`
		}{
			Reference:         internalRef,
			ProviderReference: outboundProviderRef,
			ProviderResponse:  providerResponse,
		},
	))
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

func (s *BillService) markSuccess(reference, outboundProviderRef, providerRef string) error {
	return s.DB.Model(&models.Transaction{}).
		Where("reference = ?", reference).
		Updates(map[string]any{
			"status":                models.Success,
			"outbound_provider_ref": outboundProviderRef,
			"provider_ref":          providerRef,
		}).Error
}

func (s *BillService) markFailedAndRefund(userID string, amount int64, reference, failureReason, outboundProviderRef string) error {
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
				"status":                models.Failed,
				"outbound_provider_ref": outboundProviderRef,
				"description":           description,
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
