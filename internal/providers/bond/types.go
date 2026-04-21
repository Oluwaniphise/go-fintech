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

type ValidateElectricityResponse struct {
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

type ProductsResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	DisplayImage string `json:"displayImage"`
	New          bool   `json:"new"`
	SkipBoolean  bool   `json:"skipValidation"`
}

type ProductItemsResponse struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Code          string  `json:"code"`
	Amount        float64 `json:"amount"`
	MinAmount     float64 `json:"minAmount"`
	MaxAmount     float64 `json:"maxAmount"`
	IsFixedAmount bool    `json:"isFixedAmount"`
	New           bool    `json:"new"`
	Duration      int64   `json:"duration"`
}
