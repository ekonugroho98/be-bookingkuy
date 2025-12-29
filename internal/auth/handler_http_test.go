package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/jwt"
	"github.com/ekonugroho98/be-bookingkuy/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestAuthHandlerHTTP_Register_Success tests successful user registration via HTTP
func TestAuthHandlerHTTP_Register_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEB := new(MockEventBus)
	jwtManager := jwt.NewManager("test-secret")

	service := NewService(mockUserRepo, mockAuthRepo, mockEB, jwtManager)
	handler := NewHandler(service)

	// Create request body
	reqBody := RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Mock: Email not taken yet
	mockUserRepo.On("GetByEmail", mockContext, "john@example.com").Return(nil, nil)
	// Mock: User created successfully
	mockUserRepo.On("Create", mockContext, mock.AnythingOfType("*user.User")).Return(nil)
	// Mock: Password stored
	mockAuthRepo.On("StorePassword", mockContext, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	// Mock: Event published (eventbus.EventUserCreated = "user.created")
	mockEB.On("Publish", mockContext, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Register(rr, req)

	// Assertions
	require.Equal(t, http.StatusCreated, rr.Code)

	var respBody map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Equal(t, "User registered successfully", respBody["message"])
	assert.NotNil(t, respBody["user"])

	userData := respBody["user"].(map[string]interface{})
	assert.Equal(t, "John Doe", userData["name"])
	assert.Equal(t, "john@example.com", userData["email"])
	assert.NotNil(t, userData["id"])

	mockUserRepo.AssertExpectations(t)
	mockAuthRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestAuthHandlerHTTP_Register_DuplicateEmail tests registration with duplicate email
func TestAuthHandlerHTTP_Register_DuplicateEmail(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEB := new(MockEventBus)
	jwtManager := jwt.NewManager("test-secret")

	service := NewService(mockUserRepo, mockAuthRepo, mockEB, jwtManager)
	handler := NewHandler(service)

	// Create request body
	reqBody := RegisterRequest{
		Name:     "John Doe",
		Email:    "existing@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Mock: User already exists
	existingUser := &user.User{
		ID:    "user-123",
		Email: "existing@example.com",
	}
	mockUserRepo.On("GetByEmail", mockContext, "existing@example.com").Return(existingUser, nil)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Register(rr, req)

	// Assertions
	require.Equal(t, http.StatusConflict, rr.Code)

	var respBody map[string]string
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Contains(t, respBody["error"], "already registered")

	mockUserRepo.AssertExpectations(t)
}

// TestAuthHandlerHTTP_Register_InvalidJSON tests registration with invalid JSON
func TestAuthHandlerHTTP_Register_InvalidJSON(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEB := new(MockEventBus)
	jwtManager := jwt.NewManager("test-secret")

	service := NewService(mockUserRepo, mockAuthRepo, mockEB, jwtManager)
	handler := NewHandler(service)

	// Create HTTP request with invalid JSON
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Register(rr, req)

	// Assertions
	require.Equal(t, http.StatusBadRequest, rr.Code)

	var respBody map[string]string
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Contains(t, respBody["error"], "Invalid request body")
}

// TestAuthHandlerHTTP_Login_Success tests successful login via HTTP
func TestAuthHandlerHTTP_Login_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEB := new(MockEventBus)
	jwtManager := jwt.NewManager("test-secret")

	service := NewService(mockUserRepo, mockAuthRepo, mockEB, jwtManager)
	handler := NewHandler(service)

	// Create request body
	reqBody := LoginRequest{
		Email:    "john@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Mock: User exists
	userFromDB := &user.User{
		ID:    "user-123",
		Name:  "John Doe",
		Email: "john@example.com",
	}
	mockUserRepo.On("GetByEmail", mockContext, "john@example.com").Return(userFromDB, nil)

	// Mock: Password exists and is correct
	mockAuthRepo.On("GetPassword", mockContext, "john@example.com").Return(
		"user-123",
		"$2a$10$lKZzSm1G6eoRLCLD71FT3eq63egelVOg0HJUAsUS3MVZpH8/UeN1G", // bcrypt of "password123"
		nil,
	)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Login(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code)

	var respBody map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.NotNil(t, respBody["token"])
	assert.NotNil(t, respBody["user"])

	userData := respBody["user"].(map[string]interface{})
	assert.Equal(t, "John Doe", userData["name"])
	assert.Equal(t, "john@example.com", userData["email"])
	assert.Equal(t, "user-123", userData["id"])

	token := respBody["token"].(string)
	assert.NotEmpty(t, token)

	mockUserRepo.AssertExpectations(t)
	mockAuthRepo.AssertExpectations(t)
}

// TestAuthHandlerHTTP_Login_InvalidCredentials tests login with invalid credentials
func TestAuthHandlerHTTP_Login_InvalidCredentials(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEB := new(MockEventBus)
	jwtManager := jwt.NewManager("test-secret")

	service := NewService(mockUserRepo, mockAuthRepo, mockEB, jwtManager)
	handler := NewHandler(service)

	// Create request body with wrong password
	reqBody := LoginRequest{
		Email:    "john@example.com",
		Password: "wrongpassword",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Mock: User exists
	userFromDB := &user.User{
		ID:    "user-123",
		Name:  "John Doe",
		Email: "john@example.com",
	}
	mockUserRepo.On("GetByEmail", mockContext, "john@example.com").Return(userFromDB, nil)

	// Mock: Password exists (but will fail bcrypt comparison)
	mockAuthRepo.On("GetPassword", mockContext, "john@example.com").Return(
		"user-123",
		"$2a$10$lKZzSm1G6eoRLCLD71FT3eq63egelVOg0HJUAsUS3MVZpH8/UeN1G", // bcrypt of "password123"
		nil,
	)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Login(rr, req)

	// Assertions
	require.Equal(t, http.StatusUnauthorized, rr.Code)

	var respBody map[string]string
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Contains(t, respBody["error"], "Invalid email or password")

	mockUserRepo.AssertExpectations(t)
	mockAuthRepo.AssertExpectations(t)
}

// TestAuthHandlerHTTP_Login_UserNotFound tests login with non-existent user
func TestAuthHandlerHTTP_Login_UserNotFound(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEB := new(MockEventBus)
	jwtManager := jwt.NewManager("test-secret")

	service := NewService(mockUserRepo, mockAuthRepo, mockEB, jwtManager)
	handler := NewHandler(service)

	// Create request body
	reqBody := LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Mock: User not found
	mockUserRepo.On("GetByEmail", mockContext, "nonexistent@example.com").Return(nil, nil)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Login(rr, req)

	// Assertions
	require.Equal(t, http.StatusUnauthorized, rr.Code)

	var respBody map[string]string
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Contains(t, respBody["error"], "Invalid email or password")

	mockUserRepo.AssertExpectations(t)
}

// TestAuthHandlerHTTP_Login_InvalidJSON tests login with invalid JSON
func TestAuthHandlerHTTP_Login_InvalidJSON(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepository)
	mockAuthRepo := new(MockAuthRepository)
	mockEB := new(MockEventBus)
	jwtManager := jwt.NewManager("test-secret")

	service := NewService(mockUserRepo, mockAuthRepo, mockEB, jwtManager)
	handler := NewHandler(service)

	// Create HTTP request with invalid JSON
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Login(rr, req)

	// Assertions
	require.Equal(t, http.StatusBadRequest, rr.Code)

	var respBody map[string]string
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Contains(t, respBody["error"], "Invalid request body")
}

// mockContext is a reusable context for tests
var mockContext = context.Background()
