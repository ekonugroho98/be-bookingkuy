package midtrans

import (
	"fmt"
)

// Mapper handles conversion between Midtrans and payment service models
type Mapper struct{}

// NewMapper creates a new mapper
func NewMapper() *Mapper {
	return &Mapper{}
}

// PaymentInput represents minimal payment data needed for Midtrans mapping
type PaymentInput struct {
	OrderID   string
	BookingID string
	Amount    int
}

// ToChargeRequest converts payment request to Midtrans charge request
func (m *Mapper) ToChargeRequest(payment *PaymentInput, customerDetails *CustomerDetails) *ChargeRequest {
	return &ChargeRequest{
		PaymentType: PaymentTypeGopay, // Default payment type
		TransactionDetails: TransactionDetails{
			OrderID:     payment.OrderID,
			GrossAmount: int64(payment.Amount),
		},
		CustomerDetails: *customerDetails,
		ItemDetails:     m.buildItemDetails(payment),
		CustomExpiry: &CustomExpiry{
			ExpiryDuration: 60, // 60 minutes
			ExpiryUnit:     "minute",
		},
	}
}

// ToChargeRequestWithPaymentType converts payment request with specific payment type
func (m *Mapper) ToChargeRequestWithPaymentType(
	payment *PaymentInput,
	customerDetails *CustomerDetails,
	paymentType PaymentType,
) *ChargeRequest {
	req := m.ToChargeRequest(payment, customerDetails)
	req.PaymentType = paymentType

	// Add payment type specific details
	switch paymentType {
	case PaymentTypeBankTransfer:
		req.BankTransfer = &BankTransferDetails{
			Bank: "bca", // Default to BCA
		}
	case PaymentTypeQRIS:
		req.QRIS = &QRISDetails{
			Acquirer: "gopay",
		}
	case PaymentTypeGopay:
		req.Gopay = &GopayDetails{
			EnableCallback: true,
		}
	}

	return req
}

// buildItemDetails builds item details from payment metadata
func (m *Mapper) buildItemDetails(payment *PaymentInput) []ItemDetails {
	// If metadata contains booking details, use them
	// For now, create a single item
	price := int64(payment.Amount)

	return []ItemDetails{
		{
			ID:       "BOOKING-" + payment.BookingID,
			Price:    price,
			Quantity: 1,
			Name:     "Hotel Booking",
			Category: "Travel",
		},
	}
}

// ToCustomerDetails converts user data to Midtrans customer details
func (m *Mapper) ToCustomerDetails(
	firstName, lastName, email, phone string,
	billingAddress, shippingAddress *Address,
) *CustomerDetails {
	return &CustomerDetails{
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		Phone:           phone,
		BillingAddress:  billingAddress,
		ShippingAddress: shippingAddress,
	}
}

// GetPaymentAmount extracts amount from webhook gross amount string
func (m *Mapper) GetPaymentAmount(grossAmount string) int {
	// Midtrans returns gross amount as string like "100000.00"
	// Parse it to integer (in IDR, no decimal)
	var amount float64
	if _, err := fmt.Sscanf(grossAmount, "%f", &amount); err == nil {
		return int(amount)
	}
	return 0
}
