# Notifications & Webhooks Configuration Guide

## Overview

Bookingkuy backend supports comprehensive notification system including:
- **Email Notifications** via SendGrid
- **SMS Notifications** (Twilio/Nexmo - framework ready)
- **Webhook Delivery** to external systems
- **Queue-based Async Processing** via RabbitMQ

---

## Email Notifications

### Setup SendGrid

1. **Get SendGrid API Key**
   - Sign up at https://sendgrid.com/
   - Create API key with "Mail Send" permissions
   - Copy your API key

2. **Configure Environment Variables**

   Add to your `.env` file:
   ```bash
   SENDGRID_API_KEY=SG.your-sendgrid-api-key-here
   SENDGRID_FROM_EMAIL=noreply@bookingkuy.com
   SENDGRID_FROM_NAME=Bookingkuy
   ```

3. **SendGrid Client Initialization** (in `cmd/api/main.go`)

   ```go
   import (
       "github.com/ekonugroho98/be-bookingkuy/internal/sendgrid"
       "github.com/ekonugroho98/be-bookingkuy/internal/notification"
   )

   // Initialize SendGrid client
   sendgridClient := sendgrid.NewClient(
       config.GetString("sendgrid.api_key"),
       config.GetBool("sendgrid.is_production"), // false for sandbox mode
   )

   // Initialize Email Service
   emailService := notification.NewEmailService(
       sendgridClient,
       config.GetString("sendgrid.from_email"),
       config.GetString("sendgrid.from_name"),
   )
   ```

### Available Email Templates

The system includes professional HTML email templates for:

1. **Booking Confirmation** (`SendBookingConfirmationEmail`)
   - Sent when user creates a booking
   - Includes: booking reference, hotel details, dates, total amount

2. **Payment Confirmation** (`SendPaymentConfirmationEmail`)
   - Sent when payment is successful
   - Includes: payment method, amount, payment date

3. **Booking Confirmed** (`SendBookingConfirmedEmail`)
   - Sent when supplier confirms booking
   - Includes: supplier reference, voucher code

4. **Cancellation Notice** (`SendCancellationEmail`)
   - Sent when booking is cancelled
   - Includes: refund amount, refund status

### Customizing Email Templates

Email templates are defined in `internal/notification/email.go`:

```go
func (e *EmailService) buildHTMLBody(subject string, data map[string]interface{}) string {
    // Customize HTML template here
    // Uses inline CSS for email client compatibility
}
```

**Template Features:**
- Responsive design (600px max width)
- Professional Bookingkuy branding
- Inline CSS for maximum email client compatibility
- Dynamic data insertion

### Sending Emails Programmatically

```go
import "github.com/ekonugroho98/be-bookingkuy/internal/notification"

// Send booking confirmation
bookingData := map[string]interface{}{
    "booking_reference": "BKG-ABC123",
    "hotel_name":        "Grand Hotel Bali",
    "check_in":          "2025-01-15",
    "check_out":         "2025-01-17",
    "guests":            2,
    "total_amount":      3000000,
    "currency":          "IDR",
    "payment_type":      "PAY_NOW",
}

err := emailService.SendBookingConfirmationEmail(
    ctx,
    "user@example.com",
    "John Doe",
    bookingData,
)
```

---

## SMS Notifications

### Current Status

SMS service framework is implemented but actual provider integration requires configuration.

### Setup Twilio (Optional)

1. **Get Twilio Credentials**
   - Sign up at https://www.twilio.com/
   - Get Account SID and Auth Token
   - Purchase a phone number

2. **Configure Environment Variables**

   ```bash
   TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   TWILIO_AUTH_TOKEN=your-auth-token
   TWILIO_PHONE_NUMBER=+1234567890
   ```

3. **Implement SMS Service** (in `internal/notification/sms.go`)

   Current implementation logs only. To integrate Twilio:

   ```go
   import (
       "github.com/twilio/twilio-go"
   )

   func (s *SMSService) SendSMS(ctx context.Context, to, message string) error {
       client := twilio.NewClient(s.accountSid, s.authToken)

       _, err := client.SendMessage(
           context.Background(),
           s.from,    // From: Twilio phone number
           to,        // To: User phone number
           message,   // SMS body
       )

       return err
   }
   ```

---

## Webhooks

### Webhook Handler Features

Located in `internal/webhook/handler.go`:

1. **Outgoing Webhooks**
   - HTTP POST with JSON payload
   - HMAC-SHA256 signature verification
   - Exponential backoff retry (max 60s)
   - Configurable timeout (default 30s)

2. **Incoming Webhooks**
   - Midtrans payment webhooks
   - Hotelbeds booking webhooks
   - Signature verification

### Sending Webhooks

```go
import "github.com/ekonugroho98/be-bookingkuy/internal/webhook"

// Initialize webhook handler
webhookHandler := webhook.NewHandler("your-webhook-secret")

// Send booking notification to external system
bookingData := map[string]interface{}{
    "booking_id":        "booking-123",
    "booking_reference": "BKG-ABC123",
    "status":            "CONFIRMED",
}

err := webhookHandler.SendBookingWebhook(
    ctx,
    "https://your-system.com/webhooks/booking",
    "booking.created",
    bookingData,
)
```

### Webhook Retry Logic

Automatic retry with exponential backoff:
- Attempt 1: Immediate
- Attempt 2: Wait 1s
- Attempt 3: Wait 2s
- Attempt 4: Wait 4s
- Attempt 5: Wait 8s
- Max wait: 60 seconds

### Receiving Webhooks

**Register webhook endpoints in `cmd/api/main.go`:**

```go
// Initialize webhook handler
webhookHandler := webhook.NewHandler(config.GetString("webhook.secret"))

// Register webhook routes
mux.HandleFunc("POST /api/v1/webhooks/midtrans", webhookHandler.HandleMidtransWebhook)
mux.HandleFunc("POST /api/v1/webhooks/hotelbeds", webhookHandler.HandleHotelbedsWebhook)
```

**Webhook Payload Format:**

```json
{
  "event": "booking.created",
  "timestamp": "2025-01-15T10:30:00Z",
  "data": {
    "booking_id": "booking-123",
    "booking_reference": "BKG-ABC123",
    "status": "CONFIRMED"
  }
}
```

**Webhook Headers:**
```
Content-Type: application/json
X-Webhook-Signature: sha256=...
X-Webhook-Timestamp: 1705320600
User-Agent: Bookingkuy-Webhook/1.0
```

**Verifying Webhook Signatures:**

```go
// Verify signature on receiving end
func VerifyWebhook(payload []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))

    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

---

## Queue-Based Async Processing

### RabbitMQ Integration

Notifications are sent asynchronously via RabbitMQ for better performance:

```go
// If RabbitMQ is available, notifications are queued
if queueClient != nil && queueClient.IsConnected() {
    // Publish to email queue
    queueClient.Publish(ctx, queue.QueueEmail, message)

    // Publish to SMS queue
    queueClient.Publish(ctx, queue.QueueSMS, message)
}

// Fallback to synchronous sending if queue unavailable
```

### Queue Names

- `email` - Email notifications
- `sms` - SMS notifications
- `booking_sync` - Booking synchronization webhooks
- `payment_sync` - Payment synchronization webhooks

### Worker Configuration

Worker processes queues in background (see `cmd/worker/main.go`):

```go
// Worker automatically processes:
// - Email queue → SendGrid API
// - SMS queue → Twilio API
// - Webhook queues → External HTTP endpoints
```

---

## Event-Driven Notifications

### Notification Events

The system subscribes to these events:

| Event | Handler | Notification |
|-------|---------|--------------|
| `booking.created` | HandleBookingCreated | Booking confirmation email |
| `booking.paid` | HandleBookingPaid | Payment confirmation email |
| `booking.confirmed` | HandleBookingConfirmed | Final confirmation with voucher |
| `booking.cancelled` | HandleBookingCancelled | Cancellation notice email |

### Registering Event Handlers

```go
import (
    "github.com/ekonugroho98/be-bookingkuy/internal/notification"
    "github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
)

// Register notification handlers
notification.RegisterEventHandlers(eventBus, notificationService)
```

### Publishing Events

```go
// After creating booking
eventBus.Publish(ctx, eventbus.EventBookingCreated, map[string]interface{}{
    "booking_id":        booking.ID,
    "booking_reference": booking.BookingReference,
    "user_id":           booking.UserID,
    "total_amount":      booking.TotalAmount,
    "currency":          booking.Currency,
})

// Email will be sent automatically to user
```

---

## Testing Notifications

### Development Mode

In development mode, emails are logged but not sent:

```bash
# .env
APP_ENV=development
SENDGRID_API_KEY=  # Leave empty to skip sending
```

**Expected output:**
```
[WARN] SendGrid client not configured, skipping email to user@example.com
[INFO] Sending booking confirmation to user@example.com
[INFO] Booking confirmation queued for user@example.com
```

### Testing SendGrid Integration

1. **Use SendGrid Sandbox Mode**
   ```go
   sendgridClient := sendgrid.NewClient(apiKey, false) // false = sandbox
   ```

2. **Check SendGrid Dashboard**
   - Go to https://app.sendgrid.com/email_activity
   - View sent emails and status

3. **Test Email Templates**
   ```go
   // Test email sending directly
   err := emailService.SendBookingConfirmationEmail(
       context.Background(),
       "your-test-email@example.com",
       "Test User",
       testBookingData,
   )
   ```

### Testing Webhooks

Use ngrok to test webhooks locally:

```bash
# Install ngrok
brew install ngrok

# Start your API server
go run ./cmd/api

# In another terminal, start ngrok
ngrok http 8080

# Use the ngrok URL for testing
# Example: https://abc123.ngrok.io/api/v1/webhooks/midtrans
```

---

## Troubleshooting

### Emails Not Sending

**Symptoms:** No emails received, no errors

**Solutions:**
1. Check SendGrid API key is valid
2. Verify sender email is verified in SendGrid
3. Check recipient email is correct
4. Review SendGrid email activity dashboard
5. Enable debug logging: `LOG_LEVEL=debug`

### Webhooks Failing

**Symptoms:** Webhooks returning error status

**Solutions:**
1. Verify webhook URL is accessible from server
2. Check firewall allows outbound HTTPS
3. Verify webhook secret matches on both ends
4. Check webhook handler returns 200 OK
5. Review webhook payload size (< 6MB recommended)

### Queue Not Processing

**Symptoms:** Messages queued but not sent

**Solutions:**
1. Verify RabbitMQ is running: `docker ps | grep rabbitmq`
2. Check worker is running: `ps aux | grep worker`
3. Review worker logs for errors
4. Verify queue connection: `queueClient.IsConnected()`

---

## Configuration Checklist

### Required for Production

- [ ] SendGrid API key configured
- [ ] Sender email verified in SendGrid
- [ ] Webhook secret generated (use strong random string)
- [ ] RabbitMQ running (or synchronous fallback enabled)
- [ ] Email templates reviewed and branded
- [ ] Webhook endpoints registered in routing
- [ ] Notification event handlers registered

### Optional Enhancements

- [ ] Twilio SMS integration
- [ ] Custom email templates
- [ ] Additional webhook endpoints
- [ ] SMS notifications for critical events
- [ ] Push notifications (mobile app)
- [ ] Webhook retry queue for failed deliveries

---

## Best Practices

### Email Notifications

1. **Always use queue for sending** (async, better performance)
2. **Handle SendGrid API errors gracefully** (fallback to logging)
3. **Limit email frequency** (avoid spamming users)
4. **Use transactional emails for important events** (bookings, payments)
5. **Include unsubscribe link** for marketing emails

### Webhook Delivery

1. **Always verify signatures** (security)
2. **Implement retry logic** (handle temporary failures)
3. **Set reasonable timeouts** (30s default)
4. **Log all webhook deliveries** (audit trail)
5. **Use exponential backoff** (avoid overwhelming servers)

### Testing

1. **Test in sandbox mode first** (avoid accidental sends)
2. **Use test email accounts** (don't spam real users)
3. **Verify all template variables** (missing data causes errors)
4. **Test webhook signature verification** (security check)
5. **Load test notification system** (ensure scalability)

---

## Security Considerations

### SendGrid API Key

- Never commit API keys to version control
- Use environment variables or secrets manager
- Rotate API keys periodically
- Use least-privilege access (only "Mail Send" permission)

### Webhook Secrets

- Use strong random strings (32+ characters)
- Store securely (environment variables, vault)
- Different secret per environment (dev/staging/prod)
- Rotate if compromised

### Webhook Verification

- Always verify HMAC signatures
- Check timestamp headers (prevent replay attacks)
- Use HTTPS only (prevent man-in-the-middle)
- Rate limit webhook endpoints (prevent abuse)

---

**Last Updated:** 2025-12-28
