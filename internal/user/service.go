package user

import (
	"context"
	"errors"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Service defines the interface for user business logic
type Service interface {
	GetProfile(ctx context.Context, userID string) (*User, error)
	UpdateProfile(ctx context.Context, userID string, req *UpdateUserRequest) (*User, error)
}

type service struct {
	repo      Repository
	eventBus   eventbus.EventBus
}

// NewService creates a new user service
func NewService(repo Repository, eb eventbus.EventBus) Service {
	return &service{
		repo:    repo,
		eventBus: eb,
	}
}

func (s *service) GetProfile(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get user profile")
		return nil, err
	}

	return user, nil
}

func (s *service) UpdateProfile(ctx context.Context, userID string, req *UpdateUserRequest) (*User, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	// Get existing user
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get user for update")
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}

	// Save to database
	if err := s.repo.Update(ctx, user); err != nil {
		logger.ErrorWithErr(err, "Failed to update user")
		return nil, err
	}

	// Publish event
	if err := s.eventBus.Publish(ctx, eventbus.EventUserUpdated, map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.Name,
	}); err != nil {
		// Log error but don't fail the operation
		logger.ErrorWithErr(err, "Failed to publish user.updated event")
	}

	logger.Infof("User profile updated: %s", user.ID)
	return user, nil
}
