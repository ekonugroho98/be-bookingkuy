package midtrans

// Midtrans API types and structures

// PaymentType represents Midtrans payment method types
type PaymentType string

const (
	PaymentTypeCreditCard      PaymentType = "credit_card"
	PaymentTypeBankTransfer    PaymentType = "bank_transfer"
	PaymentTypeGopay           PaymentType = "gopay"
	PaymentTypeQRIS            PaymentType = "qris"
	PaymentTypeShopeePay       PaymentType = "shopeepay"
)

// TransactionStatus represents Midtrans transaction status
type TransactionStatus string

const (
	StatusPending       TransactionStatus = "pending"
	StatusAuthorize     TransactionStatus = "authorize"
	StatusCapture       TransactionStatus = "capture"
	StatusSettlement    TransactionStatus = "settlement"
	StatusDeny          TransactionStatus = "deny"
	StatusPendingCancel TransactionStatus = "pending_cancel"
	StatusCancel        TransactionStatus = "cancel"
	StatusExpire        TransactionStatus = "expire"
	StatusFailure       TransactionStatus = "failure"
	StatusRefund        TransactionStatus = "refund"
	StatusPartialRefund TransactionStatus = "partial_refund"
)

// TransactionDetails represents transaction details
type TransactionDetails struct {
	OrderID    string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

// CustomerDetails represents customer information
type CustomerDetails struct {
	FirstName string        `json:"first_name"`
	LastName  string        `json:"last_name"`
	Email     string        `json:"email"`
	Phone     string        `json:"phone"`
	BillingAddress *Address `json:"billing_address,omitempty"`
	ShippingAddress *Address `json:"shipping_address,omitempty"`
}

// Address represents address details
type Address struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Address    string `json:"address"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Phone      string `json:"phone"`
	CountryCode string `json:"country_code"`
}

// ItemDetails represents item/product details
type ItemDetails struct {
	ID       string  `json:"id"`
	Price    int64   `json:"price"`
	Quantity int     `json:"quantity"`
	Name     string  `json:"name"`
	Category string  `json:"category,omitempty"`
}

// CreditCardDetails represents credit card specific options
type CreditCardDetails struct {
	Secure     bool   `json:"secure,omitempty"`
	Bank       string `json:"bank,omitempty"`
	Installment *Installment `json:"installment,omitempty"`
	TokenID    string `json:"token_id,omitempty"`
	SavedTokenID string `json:"saved_token_id,omitempty"`
}

// Installment represents installment options
type Installment struct {
	Required bool     `json:"required"`
	Terms    map[string][]int `json:"terms"` // e.g., {"bni": [3, 6, 12], "mandiri": [6, 12]}
}

// BankTransferDetails represents bank transfer specific options
type BankTransferDetails struct {
	Bank string `json:"bank"` // bca, bni, permata, bri
}

// GopayDetails represents Gopay specific options
type GopayDetails struct {
	EnableCallback bool   `json:"enable_callback"`
	CallbackURL    string `json:"callback_url,omitempty"`
}

// QRISDetails represents QRIS specific options
type QRISDetails struct {
	Acquirer string `json:"acquirer,omitempty"` // gopay, shopee
}

// ChargeRequest represents request to charge/create transaction
type ChargeRequest struct {
	PaymentType       PaymentType           `json:"payment_type"`
	TransactionDetails TransactionDetails   `json:"transaction_details"`
	CustomerDetails   CustomerDetails       `json:"customer_details,omitempty"`
	ItemDetails       []ItemDetails         `json:"item_details,omitempty"`
	CreditCard        *CreditCardDetails    `json:"credit_card,omitempty"`
	BankTransfer      *BankTransferDetails  `json:"bank_transfer,omitempty"`
	Gopay             *GopayDetails         `json:"gopay,omitempty"`
	QRIS              *QRISDetails          `json:"qris,omitempty"`
	CustomExpiry      *CustomExpiry         `json:"custom_expiry,omitempty"`
}

// CustomExpiry represents custom expiry time
type CustomExpiry struct {
	OrderTime       string `json:"order_time,omitempty"`       // ISO 8601
	ExpiryDuration  int    `json:"expiry_duration,omitempty"`  // in minutes
	ExpiryUnit      string `json:"expiry_unit,omitempty"`      // "minute", "hour", "day"
}

// ChargeResponse represents response from charge API
type ChargeResponse struct {
	Status        string           `json:"status"`
	Code          string           `json:"code"`
	Message       string           `json:"message"`
	TransactionID string           `json:"transaction_id,omitempty"`
	OrderID       string           `json:"order_id,omitempty"`
	GrossAmount   string           `json:"gross_amount,omitempty"`
	Currency      string           `json:"currency,omitempty"`
	PaymentType   string           `json:"payment_type,omitempty"`
	RedirectURL   string           `json:"redirect_url,omitempty"`
	TokenID       string           `json:"token_id,omitempty"`
	PaymentURL    string           `json:"payment_url,omitempty"`
	VANumbers     []VANumber       `json:"va_numbers,omitempty"`
	BillKey       string           `json:"bill_key,omitempty"`
	BillerCode    string           `json:"biller_code,omitempty"`
	Actions       []Action         `json:"actions,omitempty"`
}

// VANumber represents virtual account number
type VANumber struct {
	Bank     string `json:"bank"`
	VANumber string `json:"va_number"`
}

// Action represents payment action
type Action struct {
	Name       string `json:"name"`
	Method     string `json:"method"`
	URL        string `json:"url"`
}

// GetTransactionStatusResponse represents response from status API
type GetTransactionStatusResponse struct {
	Status        string           `json:"status"`
	Code          string           `json:"code"`
	Message       string           `json:"message"`
	TransactionID string           `json:"transaction_id,omitempty"`
	OrderID       string           `json:"order_id,omitempty"`
	GrossAmount   string           `json:"gross_amount,omitempty"`
	Currency      string           `json:"currency,omitempty"`
	PaymentType   string           `json:"payment_type,omitempty"`
	TransactionStatus string       `json:"transaction_status,omitempty"`
	FraudStatus   string           `json:"fraud_status,omitempty"`
	TransactionTime string         `json:"transaction_time,omitempty"`
	SettlementTime  string         `json:"settlement_time,omitempty"`
	ExpiryTime      string         `json:"expiry_time,omitempty"`
}

// WebhookPayload represents Midtrans HTTP notification
type WebhookPayload struct {
	TransactionID        string            `json:"transaction_id,omitempty"`
	Status               string           `json:"status"`
	StatusCode           string           `json:"status_code,omitempty"`
	SignatureKey         string           `json:"signature_key,omitempty"`
	PaymentType          string           `json:"payment_type,omitempty"`
	TransactionStatus    TransactionStatus `json:"transaction_status,omitempty"`
	FraudStatus          string           `json:"fraud_status,omitempty"`
	OrderID              string           `json:"order_id"`
	GrossAmount          string           `json:"gross_amount,omitempty"`
	Currency             string           `json:"currency,omitempty"`
	TransactionTime      string           `json:"transaction_time,omitempty"`
	TransactionChannel   string           `json:"transaction_channel,omitempty"`
	PaymentAmounts       []PaymentAmount  `json:"payment_amounts,omitempty"`
	CustomFields         map[string]interface{} `json:"custom_fields,omitempty"`
}

// PaymentAmount represents payment amount breakdown
type PaymentAmount struct {
	Bank        string `json:"bank,omitempty"`
	Amount      string `json:"amount,omitempty"`
}

// CancelResponse represents response from cancel API
type CancelResponse struct {
	Status        string `json:"status"`
	Code          string `json:"code"`
	Message       string `json:"message"`
	TransactionID string `json:"transaction_id,omitempty"`
}
