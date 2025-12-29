package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/jwt"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestUserHandler_GetProfile_Success tests successful profile retrieval via HTTP
func TestUserHandler_GetProfile_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	jwtManager := jwt.NewManager("test-secret")

	service := NewService(mockRepo, mockEB)
	handler := NewHandler(service)

	// Create authenticated request
	userID := "user-123"
	email := "john@example.com"
	token, _ := jwtManager.GenerateToken(userID, email)

	// Mock: User exists
	expectedUser := &User{
		ID:            userID,
		Name:          "John Doe",
		Email:         email,
		EmailVerified: true,
		Phone:         "+628123456789",
		Role:          UserRoleUser,
	}
	mockRepo.On("GetByID", mock.AnythingOfType("*context.valueCtx"), userID).Return(expectedUser, nil)

	// Create request with auth middleware
	req := httptest.NewRequest("GET", "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Add user context (simulating middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetProfile(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code)

	var respBody User
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Equal(t, userID, respBody.ID)
	assert.Equal(t, "John Doe", respBody.Name)
	assert.Equal(t, email, respBody.Email)
	assert.Equal(t, "+628123456789", respBody.Phone)

	mockRepo.AssertExpectations(t)
}

// TestUserHandler_GetProfile_UserNotFound tests profile retrieval with non-existent user
func TestUserHandler_GetProfile_UserNotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	jwtManager := jwt.NewManager("test-secret")

	service := NewService(mockRepo, mockEB)
	handler := NewHandler(service)

	// Create authenticated request
	userID := "non-existent-user"
	email := "john@example.com"
	token, _ := jwtManager.GenerateToken(userID, email)

	// Mock: User not found
	mockRepo.On("GetByID", mock.AnythingOfType("*context.valueCtx"), userID).Return(nil, ErrUserNotFound)

	// Create request with auth middleware
	req := httptest.NewRequest("GET", "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Add user context (simulating middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetProfile(rr, req)

	// Assertions
	require.Equal(t, http.StatusNotFound, rr.Code)

	var respBody map[string]string
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Contains(t, respBody["error"], "User not found")

	mockRepo.AssertExpectations(t)
}

// TestUserHandler_UpdateProfile_NameOnly tests updating profile with name only
func TestUserHandler_UpdateProfile_NameOnly(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	jwtManager := jwt.NewManager("test-secret")

	service := NewService(mockRepo, mockEB)
	handler := NewHandler(service)

	// Create authenticated request
	userID := "user-123"
	email := "john@example.com"
	token, _ := jwtManager.GenerateToken(userID, email)

	newName := "Jane Doe"

	// Mock: Get existing user
	existingUser := &User{
		ID:            userID,
		Name:          "John Doe",
		Email:         email,
		EmailVerified: true,
		Phone:         "+628123456789",
		Role:          UserRoleUser,
	}
	mockRepo.On("GetByID", mock.AnythingOfType("*context.valueCtx"), userID).Return(existingUser, nil)

	// Mock: Update successful
	mockRepo.On("Update", mock.AnythingOfType("*context.valueCtx"), mock.MatchedBy(func(u *User) bool {
		return u.Name == newName && u.ID == userID
	})).Return(nil)

	// Mock: Event published
	mockEB.On("Publish", mock.AnythingOfType("*context.valueCtx"), "user.updated", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Create request body
	reqBody := UpdateUserRequest{
		Name: &newName,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Create request with auth middleware
	req := httptest.NewRequest("PUT", "/users/me", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Add user context (simulating middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.UpdateProfile(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code)

	var respBody map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Equal(t, "Profile updated successfully", respBody["message"])
	assert.NotNil(t, respBody["user"])

	userData := respBody["user"].(map[string]interface{})
	assert.Equal(t, newName, userData["name"])
	assert.Equal(t, email, userData["email"])

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestUserHandler_UpdateProfile_PhoneOnly tests updating profile with phone only
func TestUserHandler_UpdateProfile_PhoneOnly(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	jwtManager := jwt.NewManager("test-secret")

	service := NewService(mockRepo, mockEB)
	handler := NewHandler(service)

	// Create authenticated request
	userID := "user-123"
	email := "john@example.com"
	token, _ := jwtManager.GenerateToken(userID, email)

	newPhone := "+628987654321"

	// Mock: Get existing user
	existingUser := &User{
		ID:            userID,
		Name:          "John Doe",
		Email:         email,
		EmailVerified: true,
		Phone:         "+628123456789",
		Role:          UserRoleUser,
	}
	mockRepo.On("GetByID", mock.AnythingOfType("*context.valueCtx"), userID).Return(existingUser, nil)

	// Mock: Update successful
	mockRepo.On("Update", mock.AnythingOfType("*context.valueCtx"), mock.MatchedBy(func(u *User) bool {
		return u.Phone == newPhone && u.ID == userID
	})).Return(nil)

	// Mock: Event published
	mockEB.On("Publish", mock.AnythingOfType("*context.valueCtx"), "user.updated", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Create request body
	reqBody := UpdateUserRequest{
		Phone: &newPhone,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Create request with auth middleware
	req := httptest.NewRequest("PUT", "/users/me", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Add user context (simulating middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.UpdateProfile(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code)

	var respBody map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Equal(t, "Profile updated successfully", respBody["message"])
	assert.NotNil(t, respBody["user"])

	userData := respBody["user"].(map[string]interface{})
	assert.Equal(t, newPhone, userData["phone"])
	assert.Equal(t, "John Doe", userData["name"])

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestUserHandler_UpdateProfile_BothFields tests updating profile with both name and phone
func TestUserHandler_UpdateProfile_BothFields(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	jwtManager := jwt.NewManager("test-secret")

	service := NewService(mockRepo, mockEB)
	handler := NewHandler(service)

	// Create authenticated request
	userID := "user-123"
	email := "john@example.com"
	token, _ := jwtManager.GenerateToken(userID, email)

	newName := "Jane Smith"
	newPhone := "+628987654321"

	// Mock: Get existing user
	existingUser := &User{
		ID:            userID,
		Name:          "John Doe",
		Email:         email,
		EmailVerified: true,
		Phone:         "+628123456789",
		Role:          UserRoleUser,
	}
	mockRepo.On("GetByID", mock.AnythingOfType("*context.valueCtx"), userID).Return(existingUser, nil)

	// Mock: Update successful
	mockRepo.On("Update", mock.AnythingOfType("*context.valueCtx"), mock.MatchedBy(func(u *User) bool {
		return u.Name == newName && u.Phone == newPhone && u.ID == userID
	})).Return(nil)

	// Mock: Event published
	mockEB.On("Publish", mock.AnythingOfType("*context.valueCtx"), "user.updated", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Create request body
	reqBody := UpdateUserRequest{
		Name:  &newName,
		Phone: &newPhone,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Create request with auth middleware
	req := httptest.NewRequest("PUT", "/users/me", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Add user context (simulating middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.UpdateProfile(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code)

	var respBody map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Equal(t, "Profile updated successfully", respBody["message"])
	assert.NotNil(t, respBody["user"])

	userData := respBody["user"].(map[string]interface{})
	assert.Equal(t, newName, userData["name"])
	assert.Equal(t, newPhone, userData["phone"])

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestUserHandler_UpdateProfile_InvalidJSON tests updating profile with invalid JSON
func TestUserHandler_UpdateProfile_InvalidJSON(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	jwtManager := jwt.NewManager("test-secret")

	service := NewService(mockRepo, mockEB)
	handler := NewHandler(service)

	// Create authenticated request
	userID := "user-123"
	token, _ := jwtManager.GenerateToken(userID, "john@example.com")

	// Create request with invalid JSON
	req := httptest.NewRequest("PUT", "/users/me", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Add user context (simulating middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.UpdateProfile(rr, req)

	// Assertions
	require.Equal(t, http.StatusBadRequest, rr.Code)

	var respBody map[string]string
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Contains(t, respBody["error"], "Invalid request body")
}

// mockContext is a reusable context for tests
var mockContext = context.Background()
