package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/jwt"
	"github.com/ekonugroho98/be-bookingkuy/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository is a mock for user.Repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, page, pageSize int) ([]*user.User, int, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*user.User), args.Get(1).(int), args.Error(2)
}

// MockAuthRepository is a mock for auth.Repository
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) StorePassword(ctx context.Context, userID, passwordHash string) error {
	args := m.Called(ctx, userID, passwordHash)
	return args.Error(0)
}

func (m *MockAuthRepository) GetPassword(ctx context.Context, email string) (string, string, error) {
	args := m.Called(ctx, email)
	return args.String(0), args.String(1), args.Error(2)
}

// MockEventBus is a mock for eventbus.EventBus
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, eventType string, payload map[string]interface{}) error {
	args := m.Called(ctx, eventType, payload)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(ctx context.Context, eventType string, handler eventbus.Handler) error {
	args := m.Called(ctx, eventType, handler)
	return args.Error(0)
}

func (m *MockEventBus) SubscribeAsync(ctx context.Context, eventType string, handler eventbus.Handler) error {
	args := m.Called(ctx, eventType, handler)
	return args.Error(0)
}

// Test helpers
func setupTestService(userRepo *MockUserRepository, authRepo *MockAuthRepository, eventBus *MockEventBus) *service {
	jwtManager := jwt.NewManager("test-secret")
	return NewService(userRepo, authRepo, eventBus, jwtManager).(*service)
}

func TestAuthService_Register_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEventBus := new(MockEventBus)
	service := setupTestService(mockUserRepo, mockAuthRepo, mockEventBus)

	ctx := context.Background()
	req := &RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	// Setup expectations
	mockUserRepo.On("GetByEmail", ctx, "john@example.com").Return(nil, errors.New("not found"))
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*user.User")).Return(nil)
	mockAuthRepo.On("StorePassword", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	mockEventBus.On("Publish", ctx, "user.created", mock.MatchedBy(func(payload map[string]interface{}) bool {
		return payload["email"] == "john@example.com" && payload["name"] == "John Doe"
	})).Return(nil)

	// Execute
	result, err := service.Register(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "John Doe", result.Name)
	assert.Equal(t, "john@example.com", result.Email)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockAuthRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEventBus := new(MockEventBus)
	service := setupTestService(mockUserRepo, mockAuthRepo, mockEventBus)

	ctx := context.Background()
	req := &RegisterRequest{
		Name:     "Jane Doe",
		Email:    "existing@example.com",
		Password: "password123",
	}

	existingUser := user.NewUser("Existing User", "existing@example.com")

	// Setup expectations
	mockUserRepo.On("GetByEmail", ctx, "existing@example.com").Return(existingUser, nil)

	// Execute
	result, err := service.Register(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "email already registered")

	mockUserRepo.AssertExpectations(t)
	// These should not be called
	mockAuthRepo.AssertNotCalled(t, "StorePassword")
	mockEventBus.AssertNotCalled(t, "Publish")
}

func TestAuthService_Register_DatabaseError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEventBus := new(MockEventBus)
	service := setupTestService(mockUserRepo, mockAuthRepo, mockEventBus)

	ctx := context.Background()
	req := &RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Setup expectations
	mockUserRepo.On("GetByEmail", ctx, "test@example.com").Return(nil, errors.New("not found"))
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*user.User")).Return(errors.New("database error"))

	// Execute
	result, err := service.Register(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create user")

	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEventBus := new(MockEventBus)
	service := setupTestService(mockUserRepo, mockAuthRepo, mockEventBus)

	ctx := context.Background()
	req := &LoginRequest{
		Email:    "john@example.com",
		Password: "password123",
	}

	testUser := user.NewUser("John Doe", "john@example.com")

	// Setup expectations
	mockUserRepo.On("GetByEmail", ctx, "john@example.com").Return(testUser, nil)
	mockAuthRepo.On("GetPassword", ctx, "john@example.com").Return(testUser.ID, "$2a$10$Kgc2jLp2tvjpIlbYQrpMOunBWdHXA8RWtXYsyqdLFHGYfxH1vDZYS", nil) // bcrypt hash of "password123"

	// Execute
	result, err := service.Login(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, testUser, result.User)

	mockUserRepo.AssertExpectations(t)
	mockAuthRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEventBus := new(MockEventBus)
	service := setupTestService(mockUserRepo, mockAuthRepo, mockEventBus)

	ctx := context.Background()
	req := &LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	// Setup expectations
	mockUserRepo.On("GetByEmail", ctx, "nonexistent@example.com").Return(nil, errors.New("not found"))

	// Execute
	result, err := service.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid email or password")

	mockUserRepo.AssertExpectations(t)
	mockAuthRepo.AssertNotCalled(t, "GetPassword")
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEventBus := new(MockEventBus)
	service := setupTestService(mockUserRepo, mockAuthRepo, mockEventBus)

	ctx := context.Background()
	req := &LoginRequest{
		Email:    "john@example.com",
		Password: "wrongpassword",
	}

	testUser := user.NewUser("John Doe", "john@example.com")

	// Setup expectations
	mockUserRepo.On("GetByEmail", ctx, "john@example.com").Return(testUser, nil)
	mockAuthRepo.On("GetPassword", ctx, "john@example.com").Return(testUser.ID, "$2a$10$Kgc2jLp2tvjpIlbYQrpMOunBWdHXA8RWtXYsyqdLFHGYfxH1vDZYS", nil) // bcrypt hash of "password123"

	// Execute
	result, err := service.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid email or password")

	mockUserRepo.AssertExpectations(t)
	mockAuthRepo.AssertExpectations(t)
}

func TestAuthService_Login_PasswordDatabaseError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEventBus := new(MockEventBus)
	service := setupTestService(mockUserRepo, mockAuthRepo, mockEventBus)

	ctx := context.Background()
	req := &LoginRequest{
		Email:    "john@example.com",
		Password: "password123",
	}

	testUser := user.NewUser("John Doe", "john@example.com")

	// Setup expectations
	mockUserRepo.On("GetByEmail", ctx, "john@example.com").Return(testUser, nil)
	mockAuthRepo.On("GetPassword", ctx, "john@example.com").Return("", "", errors.New("database error"))

	// Execute
	result, err := service.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid email or password")

	mockUserRepo.AssertExpectations(t)
	mockAuthRepo.AssertExpectations(t)
}
