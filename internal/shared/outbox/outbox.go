package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Event represents an outbox event
type Event struct {
	ID        string                 `json:"id"`
	Aggregate string                 `json:"aggregate"`
	AggregateID string               `json:"aggregate_id"`
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
	Published bool                   `json:"published"`
}

// Publisher publishes events
type Publisher interface {
	Publish(ctx context.Context, event *Event) error
}

// Outbox implements the outbox pattern
type Outbox struct {
	publisher Publisher
	events    []*Event
	mu        sync.Mutex
}

// New creates a new outbox
func New(publisher Publisher) *Outbox {
	return &Outbox{
		publisher: publisher,
		events:    make([]*Event, 0),
	}
}

// Add adds an event to the outbox
func (o *Outbox) Add(ctx context.Context, aggregate, aggregateID, eventType string, payload map[string]interface{}) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	event := &Event{
		ID:         fmt.Sprintf("%s-%s-%d", aggregate, aggregateID, time.Now().UnixNano()),
		Aggregate:  aggregate,
		AggregateID: aggregateID,
		EventType:  eventType,
		Payload:    payload,
		CreatedAt:  time.Now(),
		Published:  false,
	}

	o.events = append(o.events, event)
	logger.Infof("Event added to outbox: %s for aggregate %s:%s", eventType, aggregate, aggregateID)

	return nil
}

// Publish publishes all unpublished events
func (o *Outbox) Publish(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	publishedCount := 0
	for _, event := range o.events {
		if event.Published {
			continue
		}

		if err := o.publisher.Publish(ctx, event); err != nil {
			logger.Errorf("Failed to publish event %s: %v", event.ID, err)
			// Continue trying to publish other events
			continue
		}

		event.Published = true
		publishedCount++
	}

	logger.Infof("Published %d events from outbox", publishedCount)

	// Clean up published events
	o.cleanup()

	return nil
}

// cleanup removes published events from memory
func (o *Outbox) cleanup() {
	var unpublishedEvents []*Event
	for _, event := range o.events {
		if !event.Published {
			unpublishedEvents = append(unpublishedEvents, event)
		}
	}
	o.events = unpublishedEvents
}

// GetEvents returns all events (for debugging/monitoring)
func (o *Outbox) GetEvents() []*Event {
	o.mu.Lock()
	defer o.mu.Unlock()

	eventsCopy := make([]*Event, len(o.events))
	copy(eventsCopy, o.events)
	return eventsCopy
}

// MockPublisher is a mock publisher for testing
type MockPublisher struct{}

// Publish publishes an event (mock implementation)
func (m *MockPublisher) Publish(ctx context.Context, event *Event) error {
	data, _ := json.Marshal(event)
	logger.Infof("Publishing event: %s", string(data))
	return nil
}
