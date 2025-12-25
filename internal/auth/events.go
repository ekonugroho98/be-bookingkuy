package auth

import (
	"context"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// HandleUserCreated handles user created event
func HandleUserCreated(ctx context.Context, event eventbus.Event) error {
	userID, ok := event.Payload["user_id"].(string)
	if !ok {
		logger.Error("Invalid user ID in event payload")
		return nil
	}

	email, _ := event.Payload["email"].(string)
	name, _ := event.Payload["name"].(string)

	logger.Infof("User created event received: %s (%s) - %s", userID, email, name)

	// TODO: Send verification email
	// This will be implemented in notification service (ticket #022)

	return nil
}
