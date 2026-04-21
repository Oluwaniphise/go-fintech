package electricity

import (
	"bytes"
	"encoding/json"
	"fintech/internal/auth"
	"fintech/internal/bills"
	"fintech/internal/common"
	"fintech/internal/providers/bond"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

var VALIDATE_ELECTRICITY_ENDPOINT = "vas/validate/electricity"

func (s *Service) HandleValidateElectricityPurchase(c *fiber.Ctx) error {
	req := new(ElectricityRequestValidate)

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

	client := s.Deps.HTTPClient
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

	internalRef, err := bills.GenerateSecureReference("ELEC")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_REFERENCE_GENERATION_FAILED",
			"Failed to generate transaction reference",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	outboundProviderRef, err := bills.GenerateSecureReference("BOND")
	if err != nil {
		// _ = s.Helpers.MarkFailedAndRefund(userID, req.Amount, internalRef, "failed to generate provider reference", "")
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_PROVIDER_REFERENCE_FAILED",
			"Failed to generate provider reference",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	payload := ElectricityRequestValidate{
		ProductCode:     req.ProductCode,
		ProductItemCode: req.ProductItemCode,
		CustomerVendID:  req.CustomerVendID,
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

	httpReq, err := http.NewRequest(http.MethodPost, os.Getenv("BOND_SANDBOX_API")+VALIDATE_ELECTRICITY_ENDPOINT, bytes.NewBuffer(body))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_PROVIDER_REQUEST_CREATE_FAILED",
			"Failed to valid meter token",
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
			"Failed to call electricity provider",
			common.ErrorDetail{Details: err.Error()},
		))
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			"BILL_PROVIDER_RESPONSE_READ_FAILED",
			"Failed to validate electricity provider response",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	// fmt.Printf("electricity validate provider response status=%d body=%s\n", resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			"BILL_ELECTRICITY_VERIFICATION_FAILED",
			"Electricity verification failed",
			struct {
				ProviderStatus int    `json:"providerStatus"`
				ProviderBody   string `json:"providerBody"`
			}{
				ProviderStatus: resp.StatusCode,
				ProviderBody:   string(respBody),
			},
		))
	}

	var providerResponse bond.Response[bond.ValidateElectricityResponse]
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &providerResponse); err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
				fiber.StatusBadGateway,
				"BILL_PROVIDER_RESPONSE_INVALID",
				"Invalid electricity provider response",
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
		"BILL_VALIDATE_ELECTRICITY_SUCCESS",
		"Meter No validated successfully",
		struct {
			Reference         string                           `json:"reference"`
			ProviderReference string                           `json:"providerReference"`
			ElectricityData   bond.ValidateElectricityResponse `json:"electricityData"`
		}{
			Reference:         internalRef,
			ProviderReference: outboundProviderRef,
			ElectricityData:   providerResponse.Data,
		},
	))

}
