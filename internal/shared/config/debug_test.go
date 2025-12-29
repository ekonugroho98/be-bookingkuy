package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebugEnv(t *testing.T) {
	// Set env var
	os.Setenv("BOOKINGKUY_JWT_SECRET", "test-debug-secret")
	fmt.Printf("Set BOOKINGKUY_JWT_SECRET to: %s\n", os.Getenv("BOOKINGKUY_JWT_SECRET"))

	cfg, err := Load()
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Config: %+v\n", cfg)

	assert.NoError(t, err)
}
