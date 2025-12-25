package ratelimit

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientBucket
}

// clientBucket tracks rate limit for a client
type clientBucket struct {
	tokens  int
	lastRefill time.Time
}

// Config holds rate limiter configuration
type Config struct {
	RequestsPerSecond int
	BurstSize         int
	CleanupInterval   time.Duration
}

// New creates a new rate limiter
func New(config Config) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientBucket),
	}

	// Start cleanup goroutine
	go rl.cleanup(config.CleanupInterval)

	return rl
}

// Middleware returns rate limiting middleware
func (rl *RateLimiter) Middleware(requestsPerSecond int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client identifier (IP or user ID)
			clientID := getClientID(r)

			// Check rate limit
			if !rl.allow(clientID, requestsPerSecond) {
				logger.Warnf("Rate limit exceeded for client: %s", clientID)

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(requestsPerSecond))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Second).Unix(), 10))

				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "Rate limit exceeded"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// allow checks if request is allowed
func (rl *RateLimiter) allow(clientID string, maxRequests int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.clients[clientID]

	if !exists {
		// New client - create bucket with max tokens
		rl.clients[clientID] = &clientBucket{
			tokens:     maxRequests - 1,
			lastRefill: now,
		}
		return true
	}

	// Refill tokens based on time passed
	elapsed := now.Sub(bucket.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * maxRequests

	if tokensToAdd > 0 {
		bucket.tokens = min(maxRequests, bucket.tokens+tokensToAdd)
		bucket.lastRefill = now
	}

	// Check if tokens available
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// cleanup removes stale client entries
func (rl *RateLimiter) cleanup(interval time.Duration) {
	if interval == 0 {
		interval = 5 * time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()

		// Remove clients not seen in the last interval
		for clientID, bucket := range rl.clients {
			if now.Sub(bucket.lastRefill) > interval*2 {
				delete(rl.clients, clientID)
			}
		}

		rl.mu.Unlock()
		logger.Debugf("Rate limiter cleanup completed, active clients: %d", len(rl.clients))
	}
}

// getClientID extracts client identifier from request
func getClientID(r *http.Request) string {
	// Try to get user ID from context first
	if userID := r.Context().Value("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			return "user:" + uid
		}
	}

	// Fallback to IP address
	return "ip:" + r.RemoteAddr
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
