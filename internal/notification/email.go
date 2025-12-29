package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/sendgrid"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// EmailService handles email sending
type EmailService struct {
	sendgridClient *sendgrid.Client
	fromEmail      string
	fromName       string
}

// NewEmailService creates a new email service
func NewEmailService(sendgridClient *sendgrid.Client, fromEmail, fromName string) *EmailService {
	return &EmailService{
		sendgridClient: sendgridClient,
		fromEmail:      fromEmail,
		fromName:       fromName,
	}
}

// SendEmail sends an email using SendGrid
func (e *EmailService) SendEmail(ctx context.Context, to, subject string, data map[string]interface{}) error {
	if e.sendgridClient == nil {
		logger.Warnf("SendGrid client not configured, skipping email to %s", to)
		return nil // Don't fail if email not configured
	}

	// Build HTML body from data
	htmlBody := e.buildHTMLBody(subject, data)

	// Send email via SendGrid
	email := &sendgrid.Email{
		To:       []string{to},
		Subject:  subject,
		HTMLBody: htmlBody,
		Data:     data,
	}

	return e.sendgridClient.SendEmail(email)
}

// SendBookingConfirmationEmail sends booking confirmation email
func (e *EmailService) SendBookingConfirmationEmail(ctx context.Context, to, userName string, bookingData map[string]interface{}) error {
	subject := "Booking Confirmation - " + bookingData["booking_reference"].(string)
	data := map[string]interface{}{
		"user_name":         userName,
		"booking_reference": bookingData["booking_reference"],
		"hotel_name":        bookingData["hotel_name"],
		"check_in":          bookingData["check_in"],
		"check_out":         bookingData["check_out"],
		"guests":            bookingData["guests"],
		"total_amount":      fmt.Sprintf("%.2f", bookingData["total_amount"].(float64)),
		"currency":          bookingData["currency"],
		"payment_type":      bookingData["payment_type"],
	}

	return e.SendEmail(ctx, to, subject, data)
}

// SendPaymentConfirmationEmail sends payment success email
func (e *EmailService) SendPaymentConfirmationEmail(ctx context.Context, to, userName string, paymentData map[string]interface{}) error {
	subject := "Payment Successful - " + paymentData["booking_reference"].(string)
	data := map[string]interface{}{
		"user_name":         userName,
		"booking_reference": paymentData["booking_reference"],
		"payment_method":    paymentData["payment_method"],
		"amount":            fmt.Sprintf("%.2f", paymentData["amount"].(float64)),
		"currency":          paymentData["currency"],
		"payment_date":      time.Now().Format("2006-01-02 15:04:05"),
	}

	return e.SendEmail(ctx, to, subject, data)
}

// SendBookingConfirmedEmail sends final confirmation with voucher
func (e *EmailService) SendBookingConfirmedEmail(ctx context.Context, to, userName string, bookingData map[string]interface{}) error {
	subject := "Booking Confirmed - Voucher Attached - " + bookingData["booking_reference"].(string)
	data := map[string]interface{}{
		"user_name":          userName,
		"booking_reference":  bookingData["booking_reference"],
		"supplier_reference": bookingData["supplier_reference"],
		"hotel_name":         bookingData["hotel_name"],
		"check_in":           bookingData["check_in"],
		"check_out":          bookingData["check_out"],
		"voucher_code":       bookingData["voucher_code"],
		"special_requests":   bookingData["special_requests"],
	}

	return e.SendEmail(ctx, to, subject, data)
}

// SendCancellationEmail sends booking cancellation email
func (e *EmailService) SendCancellationEmail(ctx context.Context, to, userName string, cancellationData map[string]interface{}) error {
	subject := "Booking Cancelled - " + cancellationData["booking_reference"].(string)
	data := map[string]interface{}{
		"user_name":         userName,
		"booking_reference": cancellationData["booking_reference"],
		"cancellation_date":  time.Now().Format("2006-01-02 15:04:05"),
		"refund_amount":     fmt.Sprintf("%.2f", cancellationData["refund_amount"].(float64)),
		"currency":          cancellationData["currency"],
		"refund_status":     cancellationData["refund_status"],
	}

	return e.SendEmail(ctx, to, subject, data)
}

// buildHTMLBody builds HTML email body from data
func (e *EmailService) buildHTMLBody(subject string, data map[string]interface{}) string {
	// Professional HTML email template
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + subject + `</title>
</head>
<body style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f6f6f6;">
    <div style="background-color: #ffffff; padding: 30px; border-radius: 10px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <!-- Header -->
        <div style="text-align: center; margin-bottom: 30px; padding-bottom: 20px; border-bottom: 2px solid #4CAF50;">
            <h1 style="color: #4CAF50; margin: 0; font-size: 28px;">Bookingkuy</h1>
        </div>

        <!-- Title -->
        <h2 style="color: #333; margin-top: 0;">` + subject + `</h2>
        <p style="color: #666; font-size: 16px;">Thank you for choosing Bookingkuy!</p>

        <!-- Content -->
        <div style="background-color: #f9f9f9; padding: 20px; border-radius: 5px; margin: 20px 0;">
`

	// Add data as formatted key-value pairs
	for key, value := range data {
		formattedKey := formatKey(key)
		formattedValue := formatValue(value)
		html += `            <div style="margin-bottom: 15px;">
                <div style="color: #888; font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px;">` + formattedKey + `</div>
                <div style="color: #333; font-size: 16px; font-weight: 500;">` + formattedValue + `</div>
            </div>
`
	}

	html += `        </div>

        <!-- Footer -->
        <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #e0e0e0; text-align: center; color: #999; font-size: 12px;">
            <p style="margin: 5px 0;">&copy; 2025 Bookingkuy. All rights reserved.</p>
            <p style="margin: 5px 0;">Need help? Contact us at support@bookingkuy.com</p>
        </div>
    </div>
</body>
</html>`

	return html
}

// formatKey converts database key to readable format
func formatKey(key string) string {
	// Convert snake_case to Title Case
	result := ""
	for i, ch := range key {
		if ch == '_' {
			continue
		}
		if i == 0 || (i > 0 && key[i-1] == '_') {
			result += string(toUpper(ch))
		} else {
			result += string(ch)
		}
	}
	return result
}

// toUpper converts a rune to uppercase if it's lowercase
func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
}

// formatValue formats value for display
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case float64:
		return fmt.Sprintf("%.2f", v)
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprintf("%v", v)
	}
}
