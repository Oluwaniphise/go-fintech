package bills

type BillPurchaseConfig struct {
	Endpoint                   string
	Category                   string
	Description                string
	RequireValidationReference bool
	FailureCode                string
	FailureMessage             string
	SuccessCode                string
	SuccessMessage             string
}

type BillPurchaseRequest struct {
	ValidationReference string `json:"validationReference" validate:"omitempty,uuid4"`
	ProductCode         string `json:"productCode" validate:"required,max=100"`
	ProductItemCode     string `json:"productItemCode" validate:"required,max=100"`
	CustomerVendID      string `json:"customerVendId" validate:"required,max=100"`
	CustomerEmail       string `json:"customerEmail" validate:"required,email,max=255"`
	CustomerPhoneNumber string `json:"customerPhoneNumber" validate:"required,e164,max=20"`
	Amount              int64  `json:"amount" validate:"required,gt=0"`
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
