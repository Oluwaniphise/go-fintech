package electricity

type ElectricityRequestValidate struct {
	ProductCode     string `json:"productCode"`
	ProductItemCode string `json:"productItemCode"`
	CustomerVendID  string `json:"customerVendId"`
}

type PurchaseElectricityRequest struct {
	ValidationReference string `json:"validationReference"`
	ProductCode         string `json:"productCode"`
	ProductItemCode     string `json:"productItemCode"`
	CustomerVendID      string `json:"customerVendId"`
	CustomerEmail       string `json:"customerEmail"`
	CustomerPhoneNumber string `json:"customerPhoneNumber"`
	Amount              int64  `json:"amount"`
}

type PurchaseElectricityResponse struct {
	Reference           string `json:"reference"`
	ProductCode         string `json:"productCode"`
	ProductItemCode     string `json:"productItemCode"`
	CustomerVendID      string `json:"customerVendId"`
	CustomerEmail       string `json:"customerEmail"`
	CustomerPhoneNumber string `json:"customerPhoneNumber"`
	Amount              int64  `json:"amount"`
}

type PurchaseElectricityPayload struct {
	ValidationReference string `json:"validationReference"`
	Reference           string `json:"reference"`
	ProductCode         string `json:"productCode"`
	ProductItemCode     string `json:"productItemCode"`
	CustomerVendID      string `json:"customerVendId"`
	CustomerEmail       string `json:"customerEmail"`
	CustomerPhoneNumber string `json:"customerPhoneNumber"`
	Amount              int64  `json:"amount"`
}
