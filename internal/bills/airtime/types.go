package airtime

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
