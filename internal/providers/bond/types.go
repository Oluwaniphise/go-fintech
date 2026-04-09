package bond

type Response[T any] struct {
	Success            bool     `json:"success"`
	StatusCode         int      `json:"statusCode"`
	ResponseCode       string   `json:"responseCode"`
	Message            string   `json:"message"`
	UserMessage        string   `json:"userMessage"`
	ValidationMessages []string `json:"validationMessages"`
	Data               T        `json:"data"`
}

type AirtimeData struct {
	PaymentReference string `json:"paymentReference"`
	VendStatus       string `json:"vendStatus"`
	CustomerVendID   string `json:"customerVendId"`
	TransactionData  any    `json:"transactionData"`
}
