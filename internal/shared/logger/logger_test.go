package logger

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	// Set environment variable
	os.Setenv("LOG_LEVEL", "debug")
	defer os.Unsetenv("LOG_LEVEL")

	// Re-initialize logger
	Init()

	// Just verify it doesn't panic
	assert.True(t, true)
}

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()

	requestID := "req-12345"
	newCtx := WithRequestID(ctx, requestID)

	retrievedID := GetRequestID(newCtx)
	assert.Equal(t, requestID, retrievedID)
}

func TestWithRequestIDExistingContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "first-request")

	// Overwrite with new request ID
	newCtx := WithRequestID(ctx, "second-request")

	retrievedID := GetRequestID(newCtx)
	// Should have the new request ID (last one wins)
	assert.Equal(t, "second-request", retrievedID)
}

func TestGetRequestIDFromEmptyContext(t *testing.T) {
	ctx := context.Background()

	requestID := GetRequestID(ctx)
	assert.Empty(t, requestID)
}

func TestLogFunctions(t *testing.T) {
	// Just test that logging functions don't panic
	tests := []struct {
		name string
		fn   func()
	}{
		{"Info", func() { Info("test info") }},
		{"Infof", func() { Infof("test %s", "info") }},
		{"Error", func() { Error("test error") }},
		{"Errorf", func() { Errorf("test %s", "error") }},
		{"ErrorWithErr", func() { ErrorWithErr(assert.AnError, "test") }},
		{"Debug", func() { Debug("test debug") }},
		{"Debugf", func() { Debugf("test %s", "debug") }},
		{"Warn", func() { Warn("test warn") }},
		{"Warnf", func() { Warnf("test %s", "warn") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.fn)
		})
	}
}

func TestErrorWithErrNil(t *testing.T) {
	assert.NotPanics(t, func() {
		ErrorWithErr(nil, "operation failed")
	})
}

func TestWithCtxWithRequestID(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-req-456")

	log := WithCtx(ctx)
	// Just verify it returns a logger and doesn't panic
	assert.NotPanics(t, func() {
		log.Info().Msg("test")
	})
}

func TestWithCtxWithoutRequestID(t *testing.T) {
	ctx := context.Background()

	log := WithCtx(ctx)
	// Just verify it returns a logger and doesn't panic
	assert.NotPanics(t, func() {
		log.Info().Msg("test")
	})
}
