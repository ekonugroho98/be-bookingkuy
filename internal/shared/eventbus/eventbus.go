package eventbus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event
type Event struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	Timestamp   time.Time              `json:"timestamp"`
	CorrelationID string               `json:"correlation_id,omitempty"`
}

// Handler is a function that handles events
type Handler func(ctx context.Context, event Event) error

// EventBus defines the interface for event bus
type EventBus interface {
	Publish(ctx context.Context, eventType string, payload map[string]interface{}) error
	Subscribe(ctx context.Context, eventType string, handler Handler) error
	SubscribeAsync(ctx context.Context, eventType string, handler Handler) error
}

// memoryEventBus implements in-memory event bus
type memoryEventBus struct {
	mu            sync.RWMutex
	subscribers   map[string][]Handler
	asyncSubscribers map[string][]Handler
}

// New creates a new event bus instance
func New() EventBus {
	return &memoryEventBus{
		subscribers:     make(map[string][]Handler),
		asyncSubscribers: make(map[string][]Handler),
	}
}

// Publish publishes an event to all subscribers
func (b *memoryEventBus) Publish(ctx context.Context, eventType string, payload map[string]interface{}) error {
	event := Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	// Get correlation ID from context if exists
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if cid, ok := correlationID.(string); ok {
			event.CorrelationID = cid
		}
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	// Publish to sync subscribers
	if handlers, exists := b.subscribers[eventType]; exists {
		for _, handler := range handlers {
			// Execute handler
			if err := handler(ctx, event); err != nil {
				return fmt.Errorf("handler error for event %s: %w", eventType, err)
			}
		}
	}

	// Publish to async subscribers (fire and forget)
	if handlers, exists := b.asyncSubscribers[eventType]; exists {
		for _, handler := range handlers {
			go func(h Handler) {
				// Create new context with timeout for async handler
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				// Execute handler asynchronously
				if err := h(ctx, event); err != nil {
					// Log error but don't block
					// TODO: Add proper error logging
					fmt.Printf("async handler error for event %s: %v\n", eventType, err)
				}
			}(handler)
		}
	}

	return nil
}

// Subscribe subscribes to an event type synchronously
func (b *memoryEventBus) Subscribe(ctx context.Context, eventType string, handler Handler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.subscribers[eventType] == nil {
		b.subscribers[eventType] = []Handler{}
	}

	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
	return nil
}

// SubscribeAsync subscribes to an event type asynchronously
func (b *memoryEventBus) SubscribeAsync(ctx context.Context, eventType string, handler Handler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.asyncSubscribers[eventType] == nil {
		b.asyncSubscribers[eventType] = []Handler{}
	}

	b.asyncSubscribers[eventType] = append(b.asyncSubscribers[eventType], handler)
	return nil
}

// Event type constants
const (
	// User events
	EventUserCreated     = "user.created"
	EventUserUpdated     = "user.updated"

	// Booking events
	EventBookingCreated  = "booking.created"
	EventBookingPaid     = "booking.paid"
	EventBookingConfirmed = "booking.confirmed"
	EventBookingCancelled = "booking.cancelled"

	// Payment events
	EventPaymentSuccess  = "payment.success"
	EventPaymentFailed   = "payment.failed"
	EventPaymentRefunded = "payment.refunded"

	// Notification events
	EventNotificationSent = "notification.sent"
	EventNotificationFailed = "notification.failed"
)
