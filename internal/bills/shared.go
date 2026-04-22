package bills

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"fintech/internal/auth"
	"fintech/internal/common"
	"fintech/internal/providers/bond"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type BillPurchaseConfig struct {
	Endpoint                   string
	Category                   string
	Description                string
	RequireValidationReference bool
	FailureCode                string
	FailureMessage             string
	SuccessCode                string
	SuccessMessage             string
	DataKey                    string
}

type BillPurchaseRequest struct {
	ValidationReference string `json:"validationReference"`
	ProductCode         string `json:"productCode"`
	ProductItemCode     string `json:"productItemCode"`
	CustomerVendID      string `json:"customerVendId"`
	CustomerEmail       string `json:"customerEmail"`
	CustomerPhoneNumber string `json:"customerPhoneNumber"`
	Amount              int64  `json:"amount"`
}

type billPurchasePayload struct {
	ValidationReference string `json:"validationReference,omitempty"`
	Reference           string `json:"reference"`
	ProductCode         string `json:"productCode"`
	ProductItemCode     string `json:"productItemCode"`
	CustomerVendID      string `json:"customerVendId"`
	CustomerEmail       string `json:"customerEmail"`
	CustomerPhoneNumber string `json:"customerPhoneNumber"`
	Amount              int64  `json:"amount"`
}

type billPurchaseData struct {
	PaymentReference string `json:"paymentReference"`
	VendStatus       string `json:"vendStatus"`
	CustomerVendID   string `json:"customerVendId"`
	TransactionData  any    `json:"transactionData"`
}

type ValidateBillResponse struct {
	Reference     string  `json:"reference"`
	CustomerName  string  `json:"customerName"`
	MinimumAmount float64 `json:"minimumAmount"`
	MaximumAmount float64 `json:"maximumAmount"`

	CustomerData struct {
		CustomerName   string  `json:"customerName"`
		Address        *string `json:"address"`
		CanVend        bool    `json:"canVend"`
		ArrearsBalance float64 `json:"arrearsBalance"`
	}
}

type ValidateBillRequest struct {
	ProductCode     string `json:"productCode"`
	ProductItemCode string `json:"productItemCode"`
	CustomerVendID  string `json:"customerVendId"`
}

var ErrInsufficientFunds = errors.New("Insufficient funds")

func HandleBillPurchase(
	c *fiber.Ctx,
	deps Dependencies,
	helpers *Helpers,
	cfg BillPurchaseConfig,
) error {
	req := new(BillPurchaseRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"BILL_INVALID_REQUEST",
			"Invalid request",
			common.ErrorDetail{Details: "request body could not be parsed"},
		))
	}

	if req.ProductCode == "" || req.ProductItemCode == "" || req.CustomerVendID == "" || req.CustomerEmail == "" || req.CustomerPhoneNumber == "" || req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"BILL_MISSING_FIELDS",
			"productCode, productItemCode, customerVendId, customerEmail, customerPhoneNumber and amount are required",
			nil,
		))
	}

	if cfg.RequireValidationReference && req.ValidationReference == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"BILL_MISSING_FIELDS",
			"validationReference is required",
			nil,
		))
	}

	appID := os.Getenv("BOND_APP_ID")
	appSecret := os.Getenv("BOND_APP_SECRET")
	if appID == "" || appSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"BILL_MISSING_HEADERS",
			"app-id and app-secret headers are required",
			nil,
		))
	}

	client := deps.HTTPClient
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

	internalRef, err := GenerateSecureReference()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_REFERENCE_GENERATION_FAILED",
			"Failed to generate reference",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	if err := helpers.CreatePendingDebit(userID, req.Amount, internalRef, cfg.Category, cfg.Description); err != nil {
		if errors.Is(err, ErrInsufficientFunds) {
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

	outboundProviderRef, err := GenerateSecureReference()
	if err != nil {
		_ = helpers.MarkFailedAndRefund(userID, req.Amount, internalRef, "", "failed to generate provider reference")
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_PROVIDER_REFERENCE_FAILED",
			"Failed to generate provider reference",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	payload := billPurchasePayload{
		ValidationReference: req.ValidationReference,
		Reference:           outboundProviderRef,
		ProductCode:         req.ProductCode,
		ProductItemCode:     req.ProductItemCode,
		CustomerVendID:      req.CustomerVendID,
		CustomerEmail:       req.CustomerEmail,
		CustomerPhoneNumber: req.CustomerPhoneNumber,
		Amount:              req.Amount,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		_ = helpers.MarkFailedAndRefund(userID, req.Amount, internalRef, outboundProviderRef, "failed to prepare provider request")
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_PROVIDER_REQUEST_PREP_FAILED",
			"Failed to prepare request",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	httpReq, err := http.NewRequest(http.MethodPost, os.Getenv("BOND_SANDBOX_API")+cfg.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		_ = helpers.MarkFailedAndRefund(userID, req.Amount, internalRef, outboundProviderRef, "failed to create provider request")
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_PROVIDER_REQUEST_CREATE_FAILED",
			"Failed to create request",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("app-id", appID)
	httpReq.Header.Set("app-secret", appSecret)

	resp, err := client.Do(httpReq)
	if err != nil {
		_ = helpers.MarkFailedAndRefund(userID, req.Amount, internalRef, outboundProviderRef, "failed to call provider")
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			"BILL_PROVIDER_CALL_FAILED",
			"Failed to call provider",
			common.ErrorDetail{Details: err.Error()},
		))
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = helpers.MarkFailedAndRefund(userID, req.Amount, internalRef, outboundProviderRef, "failed to read provider response")
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			"BILL_PROVIDER_RESPONSE_READ_FAILED",
			"Failed to read provider response",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = helpers.MarkFailedAndRefund(userID, req.Amount, internalRef, outboundProviderRef, string(respBody))
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			cfg.FailureCode,
			cfg.FailureMessage,
			struct {
				ProviderStatus int    `json:"providerStatus"`
				ProviderBody   string `json:"providerBody"`
			}{
				ProviderStatus: resp.StatusCode,
				ProviderBody:   string(respBody),
			},
		))
	}

	var providerResponse bond.Response[billPurchaseData]
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &providerResponse); err != nil {
			_ = helpers.MarkFailedAndRefund(userID, req.Amount, internalRef, outboundProviderRef, "invalid provider response")
			return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
				fiber.StatusBadGateway,
				"BILL_PROVIDER_RESPONSE_INVALID",
				"Invalid provider response",
				struct {
					ProviderBody string `json:"providerBody"`
					Error        string `json:"error"`
				}{
					ProviderBody: string(respBody),
					Error:        err.Error(),
				},
			))
		}
	}

	if err := helpers.MarkSuccess(internalRef, outboundProviderRef, providerResponse.Data.PaymentReference); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_FINALIZATION_FAILED",
			cfg.Description+" was successful but failed to finalize transaction",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	// dataKey := cfg.DataKey
	// if dataKey == "" {
	// 	dataKey = "data"
	// }

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		cfg.SuccessCode,
		cfg.SuccessMessage,
		map[string]any{
			"reference":         internalRef,
			"providerReference": outboundProviderRef,
			"data":              providerResponse.Data,
		},
	))
}

func HandleValidateBill(c *fiber.Ctx, deps Dependencies, cfg BillPurchaseConfig) error {
	req := new(ValidateBillRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"BILL_INVALID_REQUEST",
			"Invalid request",
			common.ErrorDetail{Details: "request body could not be parsed"},
		))
	}

	if req.ProductCode == "" || req.ProductItemCode == "" || req.CustomerVendID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"BILL_MISSING_FIELDS",
			"productCode, productItemCode, customerVendId are required",
			nil,
		))
	}

	appID := os.Getenv("BOND_APP_ID")
	appSecret := os.Getenv("BOND_APP_SECRET")

	if appID == "" || appSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"BILL_MISSING_HEADERS",
			"app-id and app-secret headers are required",
			nil,
		))
	}

	client := deps.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	_, err := auth.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.Failure(
			fiber.StatusUnauthorized,
			"AUTH_UNAUTHORIZED",
			err.Error(),
			nil,
		))
	}

	internalRef, err := GenerateSecureReference()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_REFERENCE_GENERATION_FAILED",
			"Failed to generate reference",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	outboundProviderRef, err := GenerateSecureReference()
	if err != nil {
		// _ = s.Helpers.MarkFailedAndRefund(userID, req.Amount, internalRef, "failed to generate provider reference", "")
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_PROVIDER_REFERENCE_FAILED",
			"Failed to generate provider reference",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	payload := ValidateBillRequest{
		ProductCode:     req.ProductCode,
		ProductItemCode: req.ProductItemCode,
		CustomerVendID:  req.CustomerVendID,
	}

	body, err := json.Marshal(payload)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_PROVIDER_REQUEST_PREP_FAILED",
			"Failed to prepare bill request",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	httpReq, err := http.NewRequest(http.MethodPost, os.Getenv("BOND_SANDBOX_API")+cfg.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_PROVIDER_REQUEST_CREATE_FAILED",
			"Failed to validate vendID",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("app-id", appID)
	httpReq.Header.Set("app-secret", appSecret)

	resp, err := client.Do(httpReq)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			"BILL_PROVIDER_CALL_FAILED",
			"Failed to call bill provider",
			common.ErrorDetail{Details: err.Error()},
		))
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			"BILL_PROVIDER_RESPONSE_READ_FAILED",
			"Failed to validate bill provider response",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	// fmt.Printf("electricity validate provider response status=%d body=%s\n", resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			cfg.FailureCode,
			cfg.FailureMessage,
			struct {
				ProviderStatus int    `json:"providerStatus"`
				ProviderBody   string `json:"providerBody"`
			}{
				ProviderStatus: resp.StatusCode,
				ProviderBody:   string(respBody),
			},
		))
	}

	var providerResponse bond.Response[ValidateBillResponse]
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &providerResponse); err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
				fiber.StatusBadGateway,
				"BILL_PROVIDER_RESPONSE_INVALID",
				"Invalid bill provider response",
				struct {
					ProviderBody string `json:"providerBody"`
					Error        string `json:"error"`
				}{
					ProviderBody: string(respBody),
					Error:        err.Error(),
				},
			))
		}
	}

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		cfg.SuccessCode,
		cfg.SuccessMessage,
		struct {
			Reference         string               `json:"reference"`
			ProviderReference string               `json:"providerReference"`
			Data              ValidateBillResponse `json:"data"`
		}{
			Reference:         internalRef,
			ProviderReference: outboundProviderRef,
			Data:              providerResponse.Data,
		},
	))

}
