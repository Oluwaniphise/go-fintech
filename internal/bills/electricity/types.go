package electricity

type ElectricityRequestValidate struct {
	ProductCode     string `json:"productCode"`
	ProductItemCode string `json:"productItemCode"`
	CustomerVendID  string `json:"customerVendId"`
}
