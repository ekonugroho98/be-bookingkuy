package auth

import (
	"context"
	"errors"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/jwt"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
	"github.com/ekonugroho98/be-bookingkuy/internal/user"
	"golang.org/x/crypto/bcrypt"
)

// Service defines interface for auth business logic
type Service interface {
	Register(ctx context.Context, req *RegisterRequest) (*user.User, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
}

type service struct {
	userRepo   user.Repository
	authRepo   Repository
	eventBus   eventbus.EventBus
	jwtManager *jwt.Manager
}

// NewService creates a new auth service
func NewService(userRepo user.Repository, authRepo Repository, eb eventbus.EventBus, jwtManager *jwt.Manager) Service {
	return &service{
		userRepo:   userRepo,
		authRepo:   authRepo,
		eventBus:   eb,
		jwtManager: jwtManager,
	}
}

func (s *service) Register(ctx context.Context, req *RegisterRequest) (*user.User, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to hash password")
		return nil, errors.New("failed to process password")
	}

	// Create user
	newUser := user.NewUser(req.Name, req.Email)

	// Store user
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		logger.ErrorWithErr(err, "Failed to create user")
		return nil, errors.New("failed to create user")
	}

	// Store password
	if err := s.authRepo.StorePassword(ctx, newUser.ID, string(hashedPassword)); err != nil {
		logger.ErrorWithErr(err, "Failed to store password")
		return nil, errors.New("failed to store password")
	}

	// Publish user created event
	if err := s.eventBus.Publish(ctx, eventbus.EventUserCreated, map[string]interface{}{
		"user_id": newUser.ID,
		"email":   newUser.Email,
		"name":    newUser.Name,
	}); err != nil {
		logger.ErrorWithErr(err, "Failed to publish user.created event")
	}

	logger.Infof("User registered: %s", newUser.ID)
	return newUser, nil
}

func (s *service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Get user by email
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil || existingUser == nil {
		return nil, errors.New("invalid email or password")
	}

	// Get password hash
	userID, passwordHash, err := s.authRepo.GetPassword(ctx, req.Email)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get password")
		return nil, errors.New("invalid email or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(existingUser.ID, existingUser.Email)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to generate token")
		return nil, errors.New("failed to generate token")
	}

	logger.Infof("User logged in: %s", userID)

	return &LoginResponse{
		Token: token,
		User:  existingUser,
	}, nil
}
