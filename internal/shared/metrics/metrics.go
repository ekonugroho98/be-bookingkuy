package metrics

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Metrics holds application metrics
type Metrics struct {
	// HTTP metrics
	TotalRequests      int64
	SuccessRequests    int64
	FailedRequests     int64
	ActiveRequests     int64
	AverageResponseTime time.Duration

	// Business metrics
	TotalBookings      int64
	SuccessfulBookings int64
	FailedBookings     int64
	TotalPayments      int64
	SuccessfulPayments int64
	FailedPayments     int64

	// Provider metrics
	ProviderCalls      map[string]int64
	ProviderErrors     map[string]int64

	mu sync.Mutex
}

var globalMetrics = &Metrics{
	ProviderCalls:  make(map[string]int64),
	ProviderErrors: make(map[string]int64),
}

// Middleware returns HTTP metrics middleware
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Increment active requests
		globalMetrics.mu.Lock()
		globalMetrics.ActiveRequests++
		globalMetrics.TotalRequests++
		globalMetrics.mu.Unlock()

		// Create response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// Call next handler
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start)

		// Update metrics
		globalMetrics.mu.Lock()
		globalMetrics.ActiveRequests--

		if rw.status >= 200 && rw.status < 400 {
			globalMetrics.SuccessRequests++
		} else {
			globalMetrics.FailedRequests++
		}

		// Update average response time (simple moving average)
		if globalMetrics.TotalRequests > 0 {
			avgNs := (globalMetrics.AverageResponseTime.Nanoseconds()*(globalMetrics.TotalRequests-1) + duration.Nanoseconds()) / globalMetrics.TotalRequests
			globalMetrics.AverageResponseTime = time.Duration(avgNs)
		}

		globalMetrics.mu.Unlock()

		// Log slow requests
		if duration > 1*time.Second {
			logger.Warnf("Slow request: %s %s took %v", r.Method, r.URL.Path, duration)
		}
	})
}

// RecordBooking records booking attempt
func RecordBooking(success bool) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.TotalBookings++
	if success {
		globalMetrics.SuccessfulBookings++
	} else {
		globalMetrics.FailedBookings++
	}
}

// RecordPayment records payment attempt
func RecordPayment(success bool) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.TotalPayments++
	if success {
		globalMetrics.SuccessfulPayments++
	} else {
		globalMetrics.FailedPayments++
	}
}

// RecordProviderCall records provider API call
func RecordProviderCall(provider string, success bool) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.ProviderCalls[provider]++
	if !success {
		globalMetrics.ProviderErrors[provider]++
	}
}

// GetMetrics returns current metrics
func GetMetrics() *Metrics {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	// Return copy to avoid race conditions
	return &Metrics{
		TotalRequests:      globalMetrics.TotalRequests,
		SuccessRequests:    globalMetrics.SuccessRequests,
		FailedRequests:     globalMetrics.FailedRequests,
		ActiveRequests:     globalMetrics.ActiveRequests,
		AverageResponseTime: globalMetrics.AverageResponseTime,
		TotalBookings:      globalMetrics.TotalBookings,
		SuccessfulBookings: globalMetrics.SuccessfulBookings,
		FailedBookings:     globalMetrics.FailedBookings,
		TotalPayments:      globalMetrics.TotalPayments,
		SuccessfulPayments: globalMetrics.SuccessfulPayments,
		FailedPayments:     globalMetrics.FailedPayments,
		ProviderCalls:      copyMap(globalMetrics.ProviderCalls),
		ProviderErrors:     copyMap(globalMetrics.ProviderErrors),
	}
}

func copyMap(m map[string]int64) map[string]int64 {
	copy := make(map[string]int64)
	for k, v := range m {
		copy[k] = v
	}
	return copy
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// ExposeHandler returns metrics endpoint handler
func ExposeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := GetMetrics()

		// Convert to JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Simple JSON serialization
		json := `{
  "total_requests": ` + strconv.FormatInt(metrics.TotalRequests, 10) + `,
  "success_requests": ` + strconv.FormatInt(metrics.SuccessRequests, 10) + `,
  "failed_requests": ` + strconv.FormatInt(metrics.FailedRequests, 10) + `,
  "active_requests": ` + strconv.FormatInt(metrics.ActiveRequests, 10) + `,
  "average_response_time_ms": ` + strconv.FormatInt(metrics.AverageResponseTime.Milliseconds(), 10) + `,
  "total_bookings": ` + strconv.FormatInt(metrics.TotalBookings, 10) + `,
  "successful_bookings": ` + strconv.FormatInt(metrics.SuccessfulBookings, 10) + `,
  "failed_bookings": ` + strconv.FormatInt(metrics.FailedBookings, 10) + `,
  "total_payments": ` + strconv.FormatInt(metrics.TotalPayments, 10) + `,
  "successful_payments": ` + strconv.FormatInt(metrics.SuccessfulPayments, 10) + `,
  "failed_payments": ` + strconv.FormatInt(metrics.FailedPayments, 10) + `
}`

		w.Write([]byte(json))
	}
}

// ResetMetrics resets all metrics (useful for testing)
func ResetMetrics() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics = &Metrics{
		ProviderCalls:  make(map[string]int64),
		ProviderErrors: make(map[string]int64),
	}
}
