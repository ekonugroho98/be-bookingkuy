package sendgrid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Client represents SendGrid API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	fromEmail  string
	fromName   string
}

// Config represents SendGrid configuration
type Config struct {
	APIKey    string
	BaseURL   string
	FromEmail string
	FromName  string
	Timeout   time.Duration
}

// NewClient creates a new SendGrid client
func NewClient(config Config) *Client {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.sendgrid.com/v3"
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		apiKey:     config.APIKey,
		baseURL:    baseURL,
		fromEmail:  config.FromEmail,
		fromName:   config.FromName,
		httpClient: &http.Client{Timeout: config.Timeout},
	}
}

// Email represents an email
type Email struct {
	To       []string
	Subject  string
	HTMLBody string
	TextBody string
	Data     map[string]interface{}
}

// SendEmail sends an email using SendGrid API
func (c *Client) SendEmail(email *Email) error {
	// Build SendGrid request
	payload := c.buildSendGridRequest(email)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/mail/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	logger.Debugf("Sending email via SendGrid: %s", email.Subject)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		respBody, _ := io.ReadAll(resp.Body)
		logger.Errorf("SendGrid error: %s", string(respBody))
		return fmt.Errorf("SendGrid returned status %d: %s", resp.StatusCode, string(respBody))
	}

	logger.Infof("Email sent successfully to %v", email.To)
	return nil
}

// SendGridPersonalization represents SendGrid personalization
type SendGridPersonalization struct {
	To       []SendGridEmail `json:"to"`
	Subject  string          `json:"subject"`
	Cc       []SendGridEmail `json:"cc,omitempty"`
	Bcc      []SendGridEmail `json:"bcc,omitempty"`
}

// SendGridEmail represents SendGrid email address
type SendGridEmail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// SendGridContent represents SendGrid content
type SendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// SendGridRequest represents SendGrid API request
type SendGridRequest struct {
	Personalizations []SendGridPersonalization `json:"personalizations"`
	From             SendGridEmail             `json:"from"`
	ReplyTo          *SendGridEmail            `json:"reply_to,omitempty"`
	Content          []SendGridContent         `json:"content"`
}

// buildSendGridRequest builds SendGrid API request from Email
func (c *Client) buildSendGridRequest(email *Email) *SendGridRequest {
	// Build personalizations
	personalizations := make([]SendGridPersonalization, 0, len(email.To))
	for _, to := range email.To {
		personalizations = append(personalizations, SendGridPersonalization{
			To: []SendGridEmail{
				{Email: to},
			},
			Subject: email.Subject,
		})
	}

	// Build content
	content := []SendGridContent{}

	// Prefer HTML over text
	if email.HTMLBody != "" {
		content = append(content, SendGridContent{
			Type:  "text/html",
			Value: email.HTMLBody,
		})
	} else if email.TextBody != "" {
		content = append(content, SendGridContent{
			Type:  "text/plain",
			Value: email.TextBody,
		})
	}

	return &SendGridRequest{
		Personalizations: personalizations,
		From: SendGridEmail{
			Email: c.fromEmail,
			Name:  c.fromName,
		},
		Content: content,
	}
}
