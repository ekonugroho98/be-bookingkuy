package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of user.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// MockEventBus is a mock implementation of eventbus.EventBus
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, eventType string, data map[string]interface{}) error {
	args := m.Called(ctx, eventType, data)
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

// TestNewService tests creating a new user service
func TestNewService(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	require.NotNil(t, service)
}

// TestService_GetProfile_Success tests successful user profile retrieval
func TestService_GetProfile_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	userID := "user-123"
	expectedUser := &User{
		ID:            userID,
		Name:          "John Doe",
		Email:         "john@example.com",
		EmailVerified: true,
		Phone:         "+628123456789",
		Role:          UserRoleUser,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

	// Execute
	user, err := service.GetProfile(ctx, userID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Name, user.Name)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.Equal(t, expectedUser.Phone, user.Phone)
	assert.Equal(t, expectedUser.Role, user.Role)

	mockRepo.AssertExpectations(t)
}

// TestService_GetProfile_EmptyUserID tests error when user ID is empty
func TestService_GetProfile_EmptyUserID(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()

	// Execute
	user, err := service.GetProfile(ctx, "")

	// Assertions
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user ID is required")
}

// TestService_GetProfile_UserNotFound tests error when user not found
func TestService_GetProfile_UserNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	userID := "non-existent-user"

	// Setup expectations
	mockRepo.On("GetByID", ctx, userID).Return(nil, errors.New("user not found"))

	// Execute
	user, err := service.GetProfile(ctx, userID)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, user)

	mockRepo.AssertExpectations(t)
}

// TestService_GetProfile_DatabaseError tests error when database fails
func TestService_GetProfile_DatabaseError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	userID := "user-123"

	// Setup expectations
	mockRepo.On("GetByID", ctx, userID).Return(nil, errors.New("database connection error"))

	// Execute
	user, err := service.GetProfile(ctx, userID)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, user)

	mockRepo.AssertExpectations(t)
}

// TestService_UpdateProfile_UpdateName tests updating user name
func TestService_UpdateProfile_UpdateName(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	userID := "user-123"
	existingUser := &User{
		ID:            userID,
		Name:          "John Doe",
		Email:         "john@example.com",
		EmailVerified: true,
		Phone:         "+628123456789",
		Role:          UserRoleUser,
	}
	newName := "Jane Doe"

	req := &UpdateUserRequest{
		Name: &newName,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(u *User) bool {
		return u.Name == newName && u.ID == userID
	})).Return(nil)
	mockEB.On("Publish", ctx, "user.updated", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Execute
	updatedUser, err := service.UpdateProfile(ctx, userID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, updatedUser)
	assert.Equal(t, newName, updatedUser.Name)
	assert.Equal(t, existingUser.Email, updatedUser.Email)
	assert.Equal(t, existingUser.Phone, updatedUser.Phone)

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestService_UpdateProfile_UpdatePhone tests updating user phone
func TestService_UpdateProfile_UpdatePhone(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	userID := "user-123"
	existingUser := &User{
		ID:            userID,
		Name:          "John Doe",
		Email:         "john@example.com",
		EmailVerified: true,
		Phone:         "+628123456789",
		Role:          UserRoleUser,
	}
	newPhone := "+628987654321"

	req := &UpdateUserRequest{
		Phone: &newPhone,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(u *User) bool {
		return u.Phone == newPhone && u.ID == userID
	})).Return(nil)
	mockEB.On("Publish", ctx, "user.updated", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Execute
	updatedUser, err := service.UpdateProfile(ctx, userID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, updatedUser)
	assert.Equal(t, newPhone, updatedUser.Phone)
	assert.Equal(t, existingUser.Name, updatedUser.Name)

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestService_UpdateProfile_UpdateAll tests updating both name and phone
func TestService_UpdateProfile_UpdateAll(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	userID := "user-123"
	existingUser := &User{
		ID:            userID,
		Name:          "John Doe",
		Email:         "john@example.com",
		EmailVerified: true,
		Phone:         "+628123456789",
		Role:          UserRoleUser,
	}
	newName := "Jane Doe"
	newPhone := "+628987654321"

	req := &UpdateUserRequest{
		Name:  &newName,
		Phone: &newPhone,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(u *User) bool {
		return u.Name == newName && u.Phone == newPhone && u.ID == userID
	})).Return(nil)
	mockEB.On("Publish", ctx, "user.updated", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Execute
	updatedUser, err := service.UpdateProfile(ctx, userID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, updatedUser)
	assert.Equal(t, newName, updatedUser.Name)
	assert.Equal(t, newPhone, updatedUser.Phone)

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestService_UpdateProfile_EmptyUserID tests error when user ID is empty
func TestService_UpdateProfile_EmptyUserID(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	newName := "Jane Doe"

	req := &UpdateUserRequest{
		Name: &newName,
	}

	// Execute
	user, err := service.UpdateProfile(ctx, "", req)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user ID is required")
}

// TestService_UpdateProfile_UserNotFound tests error when user not found
func TestService_UpdateProfile_UserNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	userID := "non-existent-user"
	newName := "Jane Doe"

	req := &UpdateUserRequest{
		Name: &newName,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, userID).Return(nil, errors.New("user not found"))

	// Execute
	user, err := service.UpdateProfile(ctx, userID, req)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, user)

	mockRepo.AssertExpectations(t)
}

// TestService_UpdateProfile_UpdateError tests error when update fails
func TestService_UpdateProfile_UpdateError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	userID := "user-123"
	existingUser := &User{
		ID:            userID,
		Name:          "John Doe",
		Email:         "john@example.com",
		EmailVerified: true,
		Phone:         "+628123456789",
		Role:          UserRoleUser,
	}
	newName := "Jane Doe"

	req := &UpdateUserRequest{
		Name: &newName,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*user.User")).Return(errors.New("database error"))

	// Execute
	user, err := service.UpdateProfile(ctx, userID, req)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, user)

	mockRepo.AssertExpectations(t)
}

// TestService_UpdateProfile_EventPublishError tests that update succeeds even if event publishing fails
func TestService_UpdateProfile_EventPublishError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	userID := "user-123"
	existingUser := &User{
		ID:            userID,
		Name:          "John Doe",
		Email:         "john@example.com",
		EmailVerified: true,
		Phone:         "+628123456789",
		Role:          UserRoleUser,
	}
	newName := "Jane Doe"

	req := &UpdateUserRequest{
		Name: &newName,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*user.User")).Return(nil)
	mockEB.On("Publish", ctx, "user.updated", mock.AnythingOfType("map[string]interface {}")).Return(errors.New("event bus error"))

	// Execute
	user, err := service.UpdateProfile(ctx, userID, req)

	// Assertions - Should succeed despite event publishing failure
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, newName, user.Name)

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestService_UpdateProfile_NoFields tests update with no fields to update
func TestService_UpdateProfile_NoFields(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	userID := "user-123"
	existingUser := &User{
		ID:            userID,
		Name:          "John Doe",
		Email:         "john@example.com",
		EmailVerified: true,
		Phone:         "+628123456789",
		Role:          UserRoleUser,
	}

	req := &UpdateUserRequest{
		// No fields to update
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(u *User) bool {
		// Should still update with same values
		return u.Name == existingUser.Name && u.Phone == existingUser.Phone
	})).Return(nil)
	mockEB.On("Publish", ctx, "user.updated", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Execute
	updatedUser, err := service.UpdateProfile(ctx, userID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, updatedUser)
	assert.Equal(t, existingUser.Name, updatedUser.Name)

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestNewUser tests creating a new user
func TestNewUser(t *testing.T) {
	name := "John Doe"
	email := "john@example.com"

	user := NewUser(name, email)

	require.NotNil(t, user)
	assert.NotEmpty(t, user.ID)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, email, user.Email)
	assert.False(t, user.EmailVerified)
	assert.Equal(t, UserRoleUser, user.Role)
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
}

// TestUserRole_Constants tests user role constants
func TestUserRole_Constants(t *testing.T) {
	assert.Equal(t, UserRole("ADMIN"), UserRoleAdmin)
	assert.Equal(t, UserRole("USER"), UserRoleUser)
}

// TestUser_Structure tests user structure fields
func TestUser_Structure(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:            "user-123",
		Name:          "John Doe",
		Email:         "john@example.com",
		EmailVerified: true,
		Phone:         "+628123456789",
		Role:          UserRoleAdmin,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john@example.com", user.Email)
	assert.True(t, user.EmailVerified)
	assert.Equal(t, "+628123456789", user.Phone)
	assert.Equal(t, UserRoleAdmin, user.Role)
}

// TestUpdateUserRequest_Structure tests update request structure
func TestUpdateUserRequest_Structure(t *testing.T) {
	newName := "Jane Doe"
	newPhone := "+628987654321"

	req := &UpdateUserRequest{
		Name:  &newName,
		Phone: &newPhone,
	}

	assert.NotNil(t, req.Name)
	assert.NotNil(t, req.Phone)
	assert.Equal(t, newName, *req.Name)
	assert.Equal(t, newPhone, *req.Phone)
}

// TestService_GetProfile_MultipleUsers tests retrieving multiple different users
func TestService_GetProfile_MultipleUsers(t *testing.T) {
	ctx := context.Background()

	users := []*User{
		{
			ID:            "user-1",
			Name:          "User One",
			Email:         "user1@example.com",
			EmailVerified: true,
			Role:          UserRoleUser,
		},
		{
			ID:            "user-2",
			Name:          "User Two",
			Email:         "user2@example.com",
			EmailVerified: false,
			Role:          UserRoleAdmin,
		},
	}

	for _, expectedUser := range users {
		mockRepo := new(MockRepository)
		mockEB := new(MockEventBus)
		service := NewService(mockRepo, mockEB)

		mockRepo.On("GetByID", ctx, expectedUser.ID).Return(expectedUser, nil)

		user, err := service.GetProfile(ctx, expectedUser.ID)

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Name, user.Name)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Equal(t, expectedUser.EmailVerified, user.EmailVerified)
		assert.Equal(t, expectedUser.Role, user.Role)

		mockRepo.AssertExpectations(t)
	}
}

// TestService_UpdateProfile_NilFields tests update with nil fields (no changes)
func TestService_UpdateProfile_NilFields(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	userID := "user-123"
	existingUser := &User{
		ID:            userID,
		Name:          "John Doe",
		Email:         "john@example.com",
		EmailVerified: true,
		Phone:         "+628123456789",
		Role:          UserRoleUser,
	}

	req := &UpdateUserRequest{
		Name:  nil,
		Phone: nil,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(u *User) bool {
		// Values should remain unchanged
		return u.Name == existingUser.Name && u.Phone == existingUser.Phone
	})).Return(nil)
	mockEB.On("Publish", ctx, "user.updated", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Execute
	updatedUser, err := service.UpdateProfile(ctx, userID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, updatedUser)
	assert.Equal(t, existingUser.Name, updatedUser.Name)
	assert.Equal(t, existingUser.Phone, updatedUser.Phone)

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}
