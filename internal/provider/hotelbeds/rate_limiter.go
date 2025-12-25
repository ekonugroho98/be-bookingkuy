package hotelbeds

import (
	"context"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	tokens chan struct{}
}

// NewRateLimiter creates a new rate limiter
// requestsPerMinute is the maximum number of requests allowed per minute
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		tokens: make(chan struct{}, requestsPerMinute),
	}

	// Fill tokens initially
	for i := 0; i < requestsPerMinute; i++ {
		rl.tokens <- struct{}{}
	}

	// Start refill goroutine
	go rl.refill(requestsPerMinute)

	return rl
}

// Wait waits for a token to be available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// refill periodically adds tokens to the bucket
func (rl *RateLimiter) refill(requestsPerMinute int) {
	ticker := time.NewTicker(time.Minute / time.Duration(requestsPerMinute))
	defer ticker.Stop()

	for range ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Bucket is full, drop token
		}
	}
}
