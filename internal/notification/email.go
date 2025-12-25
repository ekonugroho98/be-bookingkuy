package notification

import (
	"context"
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/sendgrid"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// EmailService handles email sending
type EmailService struct {
	sendgridClient *sendgrid.Client
}

// NewEmailService creates a new email service
func NewEmailService(sendgridClient *sendgrid.Client) *EmailService {
	return &EmailService{
		sendgridClient: sendgridClient,
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

// buildHTMLBody builds HTML email body from data
func (e *EmailService) buildHTMLBody(subject string, data map[string]interface{}) string {
	// Simple HTML template
	// In production, use proper template engine
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>` + subject + `</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f4f4f4; padding: 20px; border-radius: 5px;">
        <h2 style="color: #333;">` + subject + `</h2>
`

	// Add data as key-value pairs
	for key, value := range data {
		html += `        <p style="color: #666;"><strong>` + key + `:</strong> ` + fmt.Sprintf("%v", value) + `</p>
`
	}

	html += `
    </div>
    <footer style="margin-top: 20px; text-align: center; color: #999; font-size: 12px;">
        <p>&copy; 2025 Bookingkuy. All rights reserved.</p>
    </footer>
</body>
</html>`

	return html
}
