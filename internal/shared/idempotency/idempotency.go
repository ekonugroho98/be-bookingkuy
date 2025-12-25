package idempotency

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Result stores the result of an operation
type Result struct {
	Data      interface{}
	Error     error
	Timestamp time.Time
}

// Manager manages idempotent operations
type Manager struct {
	cache map[string]*Result
	mu    sync.RWMutex
}

// New creates a new idempotency manager
func New() *Manager {
	return &Manager{
		cache: make(map[string]*Result),
	}
}

// Execute executes an operation idempotently
func (m *Manager) Execute(ctx context.Context, key string, fn func(ctx context.Context) (interface{}, error), ttl time.Duration) (interface{}, error) {
	// Check cache first
	m.mu.RLock()
	if result, exists := m.cache[key]; exists {
		m.mu.RUnlock()
		if time.Since(result.Timestamp) < ttl {
			logger.Infof("Returning cached result for key: %s", key)
			return result.Data, result.Error
		}
	}
	m.mu.RUnlock()

	// Execute the function
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if result, exists := m.cache[key]; exists {
		if time.Since(result.Timestamp) < ttl {
			return result.Data, result.Error
		}
	}

	logger.Infof("Executing operation for key: %s", key)
	data, err := fn(ctx)

	// Cache the result
	m.cache[key] = &Result{
		Data:      data,
		Error:     err,
		Timestamp: time.Now(),
	}

	// Clean up expired entries periodically
	go m.cleanup()

	return data, err
}

// cleanup removes expired entries from cache
func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, result := range m.cache {
		if now.Sub(result.Timestamp) > 1*time.Hour {
			delete(m.cache, key)
		}
	}
}

// Retry retries a function with exponential backoff
func Retry(ctx context.Context, maxAttempts int, fn func() error) error {
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err = fn(); err == nil {
			return nil
		}

		if attempt < maxAttempts {
			backoff := time.Duration(attempt) * time.Second
			logger.Infof("Attempt %d failed, retrying in %v: %v", attempt, backoff, err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, err)
}
