package catalog

import (
	"encoding/json"
	"fintech/internal/common"
	"fintech/internal/providers/bond"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

var GET_BOND_PRODUCTS = "vas/products/code"
var GET_BOND_PRODUCT_ITEMS = "vas/product_items"

func (s *Service) HandleGetBondProducts(c *fiber.Ctx) error {

	serviceCode := c.Params("serviceCode")
	if serviceCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"BILL_MISSING_SERVICE_CODE",
			"serviceCode is required",
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

	url := os.Getenv("BOND_SANDBOX_API") + GET_BOND_PRODUCTS + "/" + serviceCode

	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_PROVIDER_REQUEST_CREATE_FAILED",
			"Internal Server Error",
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
			"Failed to call products provider",
			common.ErrorDetail{Details: err.Error()},
		))
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			"BILL_PROVIDER_RESPONSE_READ_FAILED",
			"Failed to call provider response",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	// fmt.Printf("product items provider response status=%d body=%s\n", resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			"PRODUCT_ITEMS_FETCH_FAILED",
			"Failed to fetch product",
			struct {
				ProviderStatus int    `json:"providerStatus"`
				ProviderBody   string `json:"providerBody"`
			}{
				ProviderStatus: resp.StatusCode,
				ProviderBody:   string(respBody),
			},
		))
	}

	var providerResponse bond.Response[[]bond.ProductsResponse]

	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &providerResponse); err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
				fiber.StatusBadGateway,
				"BILL_PROVIDER_RESPONSE_INVALID",
				"Invalid  provider response",
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
		"BILL_SUCCESS",
		"Product fetched successfully",
		struct {
			ServiceCode string                  `json:"serviceCode"`
			Products    []bond.ProductsResponse `json:"products"`
		}{
			ServiceCode: serviceCode,
			Products:    providerResponse.Data,
		},
	))

}

func (s *Service) HandleGetBondProductItems(c *fiber.Ctx) error {

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"BILL_MISSING_SERVICE_CODE",
			"id is required",
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

	url := os.Getenv("BOND_SANDBOX_API") + GET_BOND_PRODUCT_ITEMS + "/" + id

	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"BILL_PROVIDER_REQUEST_CREATE_FAILED",
			"Failed to call provider",
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
			"Failed to call provider",
			common.ErrorDetail{Details: err.Error()},
		))
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			"BILL_PROVIDER_RESPONSE_READ_FAILED",
			"Failed to read response",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	// fmt.Printf("product items provider response status=%d body=%s\n", resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.Status(fiber.StatusBadGateway).JSON(common.Failure(
			fiber.StatusBadGateway,
			"PRODUCT_ITEMS_FETCH_FAILED",
			"Failed to fetch product items",
			struct {
				ProviderStatus int    `json:"providerStatus"`
				ProviderBody   string `json:"providerBody"`
			}{
				ProviderStatus: resp.StatusCode,
				ProviderBody:   string(respBody),
			},
		))
	}

	var providerResponse bond.Response[[]bond.ProductItemsResponse]

	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &providerResponse); err != nil {
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

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"BILL_SUCCESS",
		"Product items fetched successfully",
		struct {
			ProductID    string                      `json:"productID"`
			ProductItems []bond.ProductItemsResponse `json:"productItems"`
		}{
			ProductID:    id,
			ProductItems: providerResponse.Data,
		},
	))

}
