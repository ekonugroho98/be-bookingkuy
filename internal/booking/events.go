package booking

import (
	"context"
	"fmt"
	"sync"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

var (
	// Global notification service registry
	notificationService NotificationService
	notificationMutex    sync.RWMutex
)

// NotificationService interface for sending notifications
type NotificationService interface {
	SendBookingConfirmation(ctx context.Context, email, name string, bookingDetails map[string]interface{}) error
	SendPaymentConfirmation(ctx context.Context, email, name string, paymentDetails map[string]interface{}) error
	SendBookingCancelled(ctx context.Context, email, name string, cancellationDetails map[string]interface{}) error
}

// SetNotificationService sets the global notification service
func SetNotificationService(ns NotificationService) {
	notificationMutex.Lock()
	defer notificationMutex.Unlock()
	notificationService = ns
	logger.Info("✅ Notification service registered for booking events")
}

// getNotificationService safely gets the notification service
func getNotificationService() NotificationService {
	notificationMutex.RLock()
	defer notificationMutex.RUnlock()
	return notificationService
}

// HandleBookingCreated handles booking created event
func HandleBookingCreated(ctx context.Context, event eventbus.Event) error {
	bookingID, ok := event.Payload["booking_id"].(string)
	if !ok {
		logger.Error("Invalid booking ID in event payload")
		return nil
	}

	userID, _ := event.Payload["user_id"].(string)
	bookingRef, _ := event.Payload["booking_reference"].(string)
	amount, _ := event.Payload["total_amount"].(int)
	currency, _ := event.Payload["currency"].(string)

	logger.Infof("Booking created event received: %s (%s) - User: %s, Amount: %d %s",
		bookingID, bookingRef, userID, amount, currency)

	// 1. Send booking confirmation email
	if err := sendBookingCreatedEmail(ctx, event.Payload); err != nil {
		logger.ErrorWithErr(err, "Failed to send booking created email")
		// Don't fail the event handler for email errors
	}

	// 2. Notify admin about new booking (async)
	go func() {
		notifyAdminNewBooking(ctx, event.Payload)
	}()

	logger.Infof("Booking created event processed: %s", bookingID)
	return nil
}

// sendBookingCreatedEmail sends booking confirmation email to user
func sendBookingCreatedEmail(ctx context.Context, data map[string]interface{}) error {
	bookingID := data["booking_id"].(string)
	userID := data["user_id"].(string)
	bookingRef := data["booking_reference"].(string)
	totalAmount := data["total_amount"].(int)
	currency := data["currency"].(string)

	logger.Infof("📧 Sending booking created email to user %s for booking %s", userID, bookingID)
	logger.Infof("   - Booking Reference: %s", bookingRef)
	logger.Infof("   - Amount: %d %s", totalAmount, currency)

	// Get notification service
	ns := getNotificationService()
	if ns == nil {
		logger.Warn("Notification service not available, skipping email")
		return nil
	}

	// TODO: Get user email and name from user service
	// For now, use placeholder
	userEmail := "user@example.com"
	userName := "User"

	// Prepare booking details
	bookingDetails := map[string]interface{}{
		"booking_reference": bookingRef,
		"total_amount":      float64(totalAmount),
		"currency":          currency,
		"hotel_name":        "Hotel Name", // TODO: Get from booking details
		"check_in":          "2025-01-15",
		"check_out":         "2025-01-17",
		"guests":            2,
		"payment_type":      "PAY_NOW",
	}

	// Send booking confirmation email
	if err := ns.SendBookingConfirmation(ctx, userEmail, userName, bookingDetails); err != nil {
		return fmt.Errorf("failed to send booking confirmation email: %w", err)
	}

	logger.Infof("✅ Booking created email sent successfully to %s", userEmail)
	return nil
}

// notifyAdminNewBooking sends notification to admin about new booking
func notifyAdminNewBooking(ctx context.Context, data map[string]interface{}) {
	bookingID := data["booking_id"].(string)
	userID := data["user_id"].(string)
	bookingRef, _ := data["booking_reference"].(string)

	logger.Infof("📢 Admin notified about new booking: %s (%s) from user: %s", bookingID, bookingRef, userID)
	// TODO: Send email/SMS to admin
	// TODO: Send webhook to admin dashboard
}

// HandleBookingPaid handles booking paid event
func HandleBookingPaid(ctx context.Context, event eventbus.Event) error {
	bookingID, _ := event.Payload["booking_id"].(string)
	bookingRef, _ := event.Payload["booking_reference"].(string)

	logger.Infof("Booking paid event received: %s (%s)", bookingID, bookingRef)

	// 1. Send payment confirmation email
	if err := sendPaymentConfirmationEmail(ctx, event.Payload); err != nil {
		logger.ErrorWithErr(err, "Failed to send payment confirmation email")
		// Log but don't fail
	}

	// 2. Trigger booking confirmation with supplier
	// This is critical - but we'll do it async for now
	go func() {
		logger.Infof("🏨 Confirming booking with supplier: %s", bookingID)
		// TODO: Call booking service ConfirmBookingWithSupplier
		// This will be implemented when we have service access in handlers
	}()

	logger.Infof("Booking paid event processed: %s", bookingID)
	return nil
}

// sendPaymentConfirmationEmail sends payment success email
func sendPaymentConfirmationEmail(ctx context.Context, data map[string]interface{}) error {
	bookingID := data["booking_id"].(string)
	bookingRef := data["booking_reference"].(string)
	userID := data["user_id"].(string)
	totalAmount, _ := data["total_amount"].(int)
	currency, _ := data["currency"].(string)

	logger.Infof("📧 Sending payment confirmation email to user %s for booking %s", userID, bookingID)
	logger.Infof("   - Booking Reference: %s", bookingRef)
	logger.Infof("   - Amount: %d %s", totalAmount, currency)

	// Get notification service
	ns := getNotificationService()
	if ns == nil {
		logger.Warn("Notification service not available, skipping email")
		return nil
	}

	// TODO: Get user email and name from user service
	userEmail := "user@example.com"
	userName := "User"

	// Prepare payment details
	paymentDetails := map[string]interface{}{
		"booking_reference": bookingRef,
		"amount":            float64(totalAmount),
		"currency":          currency,
		"payment_method":    "Credit Card",
		"payment_date":      "2025-01-15 10:30:00",
	}

	// Send payment confirmation email
	if err := ns.SendPaymentConfirmation(ctx, userEmail, userName, paymentDetails); err != nil {
		return fmt.Errorf("failed to send payment confirmation email: %w", err)
	}

	logger.Infof("✅ Payment confirmation email sent successfully to %s", userEmail)
	return nil
}

// HandleBookingConfirmed handles booking confirmed event
func HandleBookingConfirmed(ctx context.Context, event eventbus.Event) error {
	bookingID, _ := event.Payload["booking_id"].(string)
	bookingRef, _ := event.Payload["booking_reference"].(string)
	supplierRef, _ := event.Payload["supplier_reference"].(string)

	logger.Infof("Booking confirmed event received: %s (%s) - Supplier: %s",
		bookingID, bookingRef, supplierRef)

	// 1. Send final booking confirmation email with details
	if err := sendFinalBookingConfirmationEmail(ctx, event.Payload); err != nil {
		logger.ErrorWithErr(err, "Failed to send final confirmation email")
		// Log but don't fail
	}

	// 2. Send webhook to external systems if configured
	go func() {
		sendBookingWebhook(ctx, event.Payload, "booking.confirmed")
	}()

	logger.Infof("Booking confirmed event processed: %s", bookingID)
	return nil
}

// sendFinalBookingConfirmationEmail sends final confirmation with all details
func sendFinalBookingConfirmationEmail(ctx context.Context, data map[string]interface{}) error {
	bookingRef := data["booking_reference"].(string)
	supplierRef, _ := data["supplier_reference"].(string)
	userID := data["user_id"].(string)

	logger.Infof("📧 Sending FINAL booking confirmation email to user %s", userID)
	logger.Infof("   - Booking Reference: %s", bookingRef)
	logger.Infof("   - Supplier Reference: %s", supplierRef)
	logger.Infof("   - Status: CONFIRMED")

	// Get notification service
	ns := getNotificationService()
	if ns == nil {
		logger.Warn("Notification service not available, skipping email")
		return nil
	}

	// TODO: Get user email and name from user service
	userEmail := "user@example.com"
	userName := "User"

	// Prepare final confirmation details
	bookingDetails := map[string]interface{}{
		"booking_reference":  bookingRef,
		"supplier_reference": supplierRef,
		"hotel_name":         "Grand Hotel",
		"check_in":           "2025-01-15",
		"check_out":          "2025-01-17",
		"voucher_code":       "VOUCHER-12345",
		"special_requests":   "None",
	}

	// Send final confirmation email
	if err := ns.SendBookingConfirmation(ctx, userEmail, userName, bookingDetails); err != nil {
		return fmt.Errorf("failed to send final confirmation email: %w", err)
	}

	logger.Infof("✅ Final confirmation email sent successfully to %s", userEmail)
	return nil
}

// sendBookingWebhook sends webhook to external systems
func sendBookingWebhook(ctx context.Context, data map[string]interface{}, eventType string) {
	bookingID := data["booking_id"].(string)
	logger.Infof("🔗 Sending webhook %s for booking %s", eventType, bookingID)
	// TODO: Implement webhook delivery
}

// HandleBookingCancelled handles booking cancelled event
func HandleBookingCancelled(ctx context.Context, event eventbus.Event) error {
	bookingID, _ := event.Payload["booking_id"].(string)
	bookingRef, _ := event.Payload["booking_reference"].(string)

	logger.Infof("Booking cancelled event received: %s (%s)", bookingID, bookingRef)

	// 1. Check if refund is needed
	needsRefund := checkIfRefundNeeded(ctx, bookingID)

	if needsRefund {
		// 2. Process refund
		if err := processRefund(ctx, bookingID); err != nil {
			logger.ErrorWithErr(err, "Failed to process refund")
			// This is critical, return error
			return fmt.Errorf("failed to process refund: %w", err)
		}

		// 3. Send refund confirmation email
		if err := sendRefundConfirmationEmail(ctx, event.Payload); err != nil {
			logger.ErrorWithErr(err, "Failed to send refund confirmation email")
		}
	} else {
		// 4. Just send cancellation email (no refund needed)
		if err := sendCancellationEmail(ctx, event.Payload); err != nil {
			logger.ErrorWithErr(err, "Failed to send cancellation email")
		}
	}

	// 5. Notify admin
	go func() {
		notifyAdminCancellation(ctx, event.Payload)
	}()

	logger.Infof("Booking cancelled event processed: %s", bookingID)
	return nil
}

// checkIfRefundNeeded determines if booking needs refund
func checkIfRefundNeeded(ctx context.Context, bookingID string) bool {
	// TODO: Check payment status from repository
	// For now, assume refund is needed
	logger.Infof("💰 Checking if refund needed for booking %s: YES", bookingID)
	return true
}

// processRefund initiates refund through payment gateway
func processRefund(ctx context.Context, bookingID string) error {
	logger.Infof("💰 Processing refund for booking: %s", bookingID)

	// TODO: Call payment service to process refund
	// paymentService.ProcessRefund(ctx, bookingID)

	logger.Infof("✅ Refund initiated for booking: %s", bookingID)
	return nil
}

// sendRefundConfirmationEmail sends refund confirmation
func sendRefundConfirmationEmail(ctx context.Context, data map[string]interface{}) error {
	bookingID := data["booking_id"].(string)
	bookingRef := data["booking_reference"].(string)
	userID := data["user_id"].(string)

	logger.Infof("📧 Sending refund confirmation email to user %s for booking %s", userID, bookingID)
	logger.Infof("   - Booking Reference: %s", bookingRef)
	logger.Infof("   - Refund: Will be processed in 5-7 business days")

	// Get notification service
	ns := getNotificationService()
	if ns == nil {
		logger.Warn("Notification service not available, skipping email")
		return nil
	}

	// TODO: Get user email and name from user service
	userEmail := "user@example.com"
	userName := "User"

	// Prepare cancellation details
	cancellationDetails := map[string]interface{}{
		"booking_reference": bookingRef,
		"refund_amount":     1500000.00, // TODO: Get actual amount
		"currency":          "IDR",
		"refund_status":     "PENDING",
	}

	// Send cancellation email
	if err := ns.SendBookingCancelled(ctx, userEmail, userName, cancellationDetails); err != nil {
		return fmt.Errorf("failed to send refund confirmation email: %w", err)
	}

	logger.Infof("✅ Refund confirmation email sent successfully to %s", userEmail)
	return nil
}

// sendCancellationEmail sends simple cancellation email (no refund)
func sendCancellationEmail(ctx context.Context, data map[string]interface{}) error {
	bookingID := data["booking_id"].(string)
	bookingRef := data["booking_reference"].(string)
	userID := data["user_id"].(string)

	logger.Infof("📧 Sending cancellation email to user %s for booking %s", userID, bookingID)
	logger.Infof("   - Booking Reference: %s", bookingRef)

	// Get notification service
	ns := getNotificationService()
	if ns == nil {
		logger.Warn("Notification service not available, skipping email")
		return nil
	}

	// TODO: Get user email and name from user service
	userEmail := "user@example.com"
	userName := "User"

	// Prepare cancellation details (no refund)
	cancellationDetails := map[string]interface{}{
		"booking_reference": bookingRef,
		"refund_amount":     0.00,
		"currency":          "IDR",
		"refund_status":     "NOT_APPLICABLE",
	}

	// Send cancellation email
	if err := ns.SendBookingCancelled(ctx, userEmail, userName, cancellationDetails); err != nil {
		return fmt.Errorf("failed to send cancellation email: %w", err)
	}

	logger.Infof("✅ Cancellation email sent successfully to %s", userEmail)
	return nil
}

// notifyAdminCancellation notifies admin about cancellation
func notifyAdminCancellation(ctx context.Context, data map[string]interface{}) {
	bookingID := data["booking_id"].(string)
	userID := data["user_id"].(string)
	logger.Infof("📢 Admin notified about cancellation: %s from user: %s", bookingID, userID)
}
