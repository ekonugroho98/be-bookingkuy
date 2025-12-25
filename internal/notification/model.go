package notification

// Notification represents a notification
type Notification struct {
	ID       string `json:"id"`
	Type     string `json:"type"`     // email, sms, push
	To       string `json:"to"`
	Subject  string `json:"subject,omitempty"`
	Message  string `json:"message"`
	Status   string `json:"status"`   // pending, sent, failed
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// SMSTemplate represents an SMS template
type SMSTemplate struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
}
