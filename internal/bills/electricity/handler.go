package electricity

import (
	"bytes"
	"encoding/json"
	"fintech/internal/common"
	"fmt"
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

	// userID, err := auth.GetUserIDFromContext(c)
	// if err != nil {
	// 	return c.Status(fiber.StatusUnauthorized).JSON(common.Failure(
	// 		fiber.StatusUnauthorized,
	// 		"AUTH_UNAUTHORIZED",
	// 		err.Error(),
	// 		nil,
	// 	))
	// }

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
			"Failed to create airtime request",
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

	fmt.Printf("electricity validate provider response status=%d body=%s\n", resp.StatusCode, string(respBody))

	return c.JSON(string(respBody))

}
