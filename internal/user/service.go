package user

import (
	"context"
)

// Service defines the interface for user business logic
type Service interface {
	GetProfile(ctx context.Context, userID string) (*User, error)
	UpdateProfile(ctx context.Context, user *User) error
}

type service struct {
	repo Repository
}

// NewService creates a new user service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) GetProfile(ctx context.Context, userID string) (*User, error) {
	// TODO: Implement business logic
	return s.repo.GetByID(ctx, userID)
}

func (s *service) UpdateProfile(ctx context.Context, user *User) error {
	// TODO: Implement business logic
	return s.repo.Update(ctx, user)
}
