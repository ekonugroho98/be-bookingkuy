package admin

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Service defines the business logic interface for admin service
type Service interface {
	// Authentication
	Login(ctx context.Context, req *LoginRequest, ipAddress, userAgent string) (*LoginResponse, error)
	Logout(ctx context.Context, adminID string) error
	GetMe(ctx context.Context, adminID string) (*Admin, error)

	// Admin management
	CreateAdmin(ctx context.Context, adminID string, req *CreateAdminRequest, ipAddress, userAgent string) (*Admin, error)
	ListAdmins(ctx context.Context, limit, offset int) ([]*Admin, int, error)
	GetAdmin(ctx context.Context, id string) (*Admin, error)
	UpdateAdmin(ctx context.Context, adminID, targetID string, req *UpdateAdminRequest, ipAddress, userAgent string) error
	DeleteAdmin(ctx context.Context, adminID, targetID string, ipAddress, userAgent string) error

	// User management
	ListUsers(ctx context.Context, limit, offset int) ([]*UserData, int, error)
	GetUser(ctx context.Context, id string) (*UserData, error)
	UpdateUser(ctx context.Context, adminID, userID string, req *UpdateUserRequest, ipAddress, userAgent string) error
	DeleteUser(ctx context.Context, adminID, userID string, ipAddress, userAgent string) error

	// Booking management
	ListBookings(ctx context.Context, status string, startDate, endDate *time.Time, limit, offset int) ([]*BookingData, int, error)
	GetBooking(ctx context.Context, id string) (*BookingData, error)
	UpdateBooking(ctx context.Context, adminID, bookingID string, req *UpdateBookingRequest, ipAddress, userAgent string) error

	// Provider management
	ListProviders(ctx context.Context) ([]*ProviderInfo, error)
	GetProvider(ctx context.Context, code string) (*ProviderInfo, error)
	UpdateProvider(ctx context.Context, adminID, providerCode string, req *UpdateProviderRequest, ipAddress, userAgent string) error

	// Analytics
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
	GetBookingStats(ctx context.Context, startDate, endDate time.Time) (*BookingStats, error)
	GetRevenueStats(ctx context.Context, startDate, endDate time.Time) (*RevenueStats, error)
	GetUserStats(ctx context.Context, startDate, endDate time.Time) (*UserStats, error)
	GetProviderStats(ctx context.Context) ([]*ProviderStats, error)

	// Audit logs
	ListAuditLogs(ctx context.Context, adminID, entityType, entityID string, limit, offset int) ([]*AuditLog, int, error)
}

// UserData represents user data from the users table
type UserData struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// BookingData represents booking data from the bookings table
type BookingData struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	UserEmail      string     `json:"user_email"`
	HotelID        string     `json:"hotel_id"`
	HotelName      string     `json:"hotel_name"`
	CheckIn        time.Time  `json:"check_in"`
	CheckOut       time.Time  `json:"check_out"`
	Status         string     `json:"status"`
	TotalAmount    int64      `json:"total_amount"`
	ProviderCode   string     `json:"provider_code"`
	AdminNotes     string     `json:"admin_notes"`
	CancellationReason string `json:"cancellation_reason"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ProviderInfo represents provider information
type ProviderInfo struct {
	ProviderCode   string `json:"provider_code"`
	Name           string `json:"name"`
	IsActive       bool   `json:"is_active"`
	BaseURL        string `json:"base_url"`
	RateLimitRPS   int    `json:"rate_limit_rps"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

type service struct {
	repo       Repository
	eventBus   eventbus.EventBus
	jwtSecret  string
	jwtExpiry  time.Duration
}

// NewService creates a new admin service
func NewService(repo Repository, eb eventbus.EventBus, jwtSecret string, jwtExpiry time.Duration) Service {
	return &service{
		repo:      repo,
		eventBus:  eb,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// Login authenticates an admin user
func (s *service) Login(ctx context.Context, req *LoginRequest, ipAddress, userAgent string) (*LoginResponse, error) {
	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	// Get admin by email
	admin, err := s.repo.GetAdminByEmail(ctx, req.Email)
	if err != nil {
		logger.ErrorWithErr(err, "Admin not found")
		return nil, errors.New("invalid email or password")
	}

	// Check if admin is active
	if !admin.IsActive {
		logger.Warnf("Inactive admin attempted login: %s", req.Email)
		return nil, errors.New("admin account is inactive")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password))
	if err != nil {
		logger.Warnf("Failed login attempt for admin: %s", req.Email)
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	expiresAt := time.Now().Add(s.jwtExpiry)
	token, err := s.generateToken(admin.ID, admin.Email, string(admin.Role), expiresAt)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to generate token")
		return nil, errors.New("failed to generate token")
	}

	// Update last login
	err = s.repo.UpdateAdminLastLogin(ctx, admin.ID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to update last login")
	}

	// Create audit log
	auditLog := &AuditLog{
		AdminID:    admin.ID,
		AdminEmail: admin.Email,
		Action:     "admin.login",
		EntityType: "admin",
		EntityID:   admin.ID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
	_ = s.repo.CreateAuditLog(ctx, auditLog)

	logger.Infof("Admin logged in: %s (%s)", admin.Email, admin.Role)

	return &LoginResponse{
		Token:     token,
		Admin:     *admin,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

// Logout handles admin logout
func (s *service) Logout(ctx context.Context, adminID string) error {
	// In a stateless JWT system, logout is handled client-side by discarding the token
	// We just log the action
	logger.Infof("Admin logged out: %s", adminID)
	return nil
}

// GetMe returns the current admin user
func (s *service) GetMe(ctx context.Context, adminID string) (*Admin, error) {
	admin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	return admin, nil
}

// CreateAdmin creates a new admin user
func (s *service) CreateAdmin(ctx context.Context, adminID string, req *CreateAdminRequest, ipAddress, userAgent string) (*Admin, error) {
	// Get requesting admin
	requestingAdmin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	// Check permission - only super_admin can create admins
	if !requestingAdmin.Role.HasPermission(PermissionAdminWrite) {
		return nil, errors.New("insufficient permissions")
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return nil, errors.New("email, password, first_name, and last_name are required")
	}

	// Check if admin already exists
	_, err = s.repo.GetAdminByEmail(ctx, req.Email)
	if err == nil {
		return nil, errors.New("admin with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to hash password")
		return nil, errors.New("failed to process password")
	}

	// Create admin
	admin := &Admin{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
	}

	err = s.repo.CreateAdmin(ctx, admin)
	if err != nil {
		return nil, err
	}

	// Create audit log
	auditLog := &AuditLog{
		AdminID:    adminID,
		AdminEmail: requestingAdmin.Email,
		Action:     "admin.created",
		EntityType: "admin",
		EntityID:   admin.ID,
		NewValues: map[string]interface{}{
			"email":   admin.Email,
			"role":    admin.Role,
			"name":    fmt.Sprintf("%s %s", admin.FirstName, admin.LastName),
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	_ = s.repo.CreateAuditLog(ctx, auditLog)

	// Publish event
	_ = s.eventBus.Publish(ctx, "admin.created", map[string]interface{}{
		"admin_id": admin.ID,
		"email":    admin.Email,
		"role":     admin.Role,
	})

	logger.Infof("Admin created: %s by %s", admin.Email, requestingAdmin.Email)

	return admin, nil
}

// ListAdmins returns a list of admins
func (s *service) ListAdmins(ctx context.Context, limit, offset int) ([]*Admin, int, error) {
	return s.repo.ListAdmins(ctx, limit, offset)
}

// GetAdmin returns an admin by ID
func (s *service) GetAdmin(ctx context.Context, id string) (*Admin, error) {
	return s.repo.GetAdminByID(ctx, id)
}

// UpdateAdmin updates an admin user
func (s *service) UpdateAdmin(ctx context.Context, adminID, targetID string, req *UpdateAdminRequest, ipAddress, userAgent string) error {
	// Get requesting admin
	requestingAdmin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return err
	}

	// Check permission
	if !requestingAdmin.Role.HasPermission(PermissionAdminWrite) {
		return errors.New("insufficient permissions")
	}

	// Get target admin to update
	targetAdmin, err := s.repo.GetAdminByID(ctx, targetID)
	if err != nil {
		return err
	}

	// Build old values for audit log
	oldValues := map[string]interface{}{
		"first_name": targetAdmin.FirstName,
		"last_name":  targetAdmin.LastName,
		"role":       targetAdmin.Role,
		"is_active":  targetAdmin.IsActive,
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	// Update admin
	err = s.repo.UpdateAdmin(ctx, targetID, updates)
	if err != nil {
		return err
	}

	// Create audit log
	auditLog := &AuditLog{
		AdminID:    adminID,
		AdminEmail: requestingAdmin.Email,
		Action:     "admin.updated",
		EntityType: "admin",
		EntityID:   targetID,
		OldValues:  oldValues,
		NewValues:  updates,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
	_ = s.repo.CreateAuditLog(ctx, auditLog)

	logger.Infof("Admin updated: %s by %s", targetAdmin.Email, requestingAdmin.Email)

	return nil
}

// DeleteAdmin deletes an admin user
func (s *service) DeleteAdmin(ctx context.Context, adminID, targetID string, ipAddress, userAgent string) error {
	// Get requesting admin
	requestingAdmin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return err
	}

	// Check permission
	if !requestingAdmin.Role.HasPermission(PermissionAdminWrite) {
		return errors.New("insufficient permissions")
	}

	// Get target admin
	targetAdmin, err := s.repo.GetAdminByID(ctx, targetID)
	if err != nil {
		return err
	}

	// Prevent self-deletion
	if adminID == targetID {
		return errors.New("cannot delete yourself")
	}

	// Delete admin
	err = s.repo.DeleteAdmin(ctx, targetID)
	if err != nil {
		return err
	}

	// Create audit log
	auditLog := &AuditLog{
		AdminID:    adminID,
		AdminEmail: requestingAdmin.Email,
		Action:     "admin.deleted",
		EntityType: "admin",
		EntityID:   targetID,
		OldValues: map[string]interface{}{
			"email": targetAdmin.Email,
			"role":  targetAdmin.Role,
			"name":  fmt.Sprintf("%s %s", targetAdmin.FirstName, targetAdmin.LastName),
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	_ = s.repo.CreateAuditLog(ctx, auditLog)

	logger.Infof("Admin deleted: %s by %s", targetAdmin.Email, requestingAdmin.Email)

	return nil
}

// ListUsers returns a list of users
func (s *service) ListUsers(ctx context.Context, limit, offset int) ([]*UserData, int, error) {
	// This is a simplified implementation - in real scenario, you'd have a UserRepository
	// For now, return empty data
	return []*UserData{}, 0, nil
}

// GetUser returns a user by ID
func (s *service) GetUser(ctx context.Context, id string) (*UserData, error) {
	// Placeholder - would query from users table
	return nil, errors.New("not implemented yet")
}

// UpdateUser updates a user
func (s *service) UpdateUser(ctx context.Context, adminID, userID string, req *UpdateUserRequest, ipAddress, userAgent string) error {
	// Get requesting admin
	requestingAdmin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return err
	}

	// Check permission
	if !requestingAdmin.Role.HasPermission(PermissionUserWrite) {
		return errors.New("insufficient permissions")
	}

	// Placeholder - would update user in users table
	logger.Infof("User update requested: %s by %s", userID, requestingAdmin.Email)

	return nil
}

// DeleteUser deletes a user
func (s *service) DeleteUser(ctx context.Context, adminID, userID string, ipAddress, userAgent string) error {
	// Get requesting admin
	requestingAdmin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return err
	}

	// Check permission
	if !requestingAdmin.Role.HasPermission(PermissionUserDelete) {
		return errors.New("insufficient permissions")
	}

	// Placeholder - would soft delete user
	logger.Infof("User deletion requested: %s by %s", userID, requestingAdmin.Email)

	return nil
}

// ListBookings returns a list of bookings with filters
func (s *service) ListBookings(ctx context.Context, status string, startDate, endDate *time.Time, limit, offset int) ([]*BookingData, int, error) {
	// Placeholder - would query from bookings table with filters
	return []*BookingData{}, 0, nil
}

// GetBooking returns a booking by ID
func (s *service) GetBooking(ctx context.Context, id string) (*BookingData, error) {
	// Placeholder
	return nil, errors.New("not implemented yet")
}

// UpdateBooking updates a booking
func (s *service) UpdateBooking(ctx context.Context, adminID, bookingID string, req *UpdateBookingRequest, ipAddress, userAgent string) error {
	// Get requesting admin
	requestingAdmin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return err
	}

	// Check permission
	if !requestingAdmin.Role.HasPermission(PermissionBookingWrite) {
		return errors.New("insufficient permissions")
	}

	// Placeholder - would update booking
	logger.Infof("Booking update requested: %s by %s", bookingID, requestingAdmin.Email)

	return nil
}

// ListProviders returns a list of providers
func (s *service) ListProviders(ctx context.Context) ([]*ProviderInfo, error) {
	// Placeholder - would return from providers table
	return []*ProviderInfo{}, nil
}

// GetProvider returns provider info
func (s *service) GetProvider(ctx context.Context, code string) (*ProviderInfo, error) {
	// Placeholder
	return nil, errors.New("not implemented yet")
}

// UpdateProvider updates provider configuration
func (s *service) UpdateProvider(ctx context.Context, adminID, providerCode string, req *UpdateProviderRequest, ipAddress, userAgent string) error {
	// Get requesting admin
	requestingAdmin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return err
	}

	// Check permission
	if !requestingAdmin.Role.HasPermission(PermissionProviderWrite) {
		return errors.New("insufficient permissions")
	}

	// Placeholder - would update provider config
	logger.Infof("Provider update requested: %s by %s", providerCode, requestingAdmin.Email)

	return nil
}

// GetDashboardStats returns dashboard statistics
func (s *service) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	return s.repo.GetDashboardStats(ctx)
}

// GetBookingStats returns booking statistics
func (s *service) GetBookingStats(ctx context.Context, startDate, endDate time.Time) (*BookingStats, error) {
	return s.repo.GetBookingStats(ctx, startDate, endDate)
}

// GetRevenueStats returns revenue statistics
func (s *service) GetRevenueStats(ctx context.Context, startDate, endDate time.Time) (*RevenueStats, error) {
	return s.repo.GetRevenueStats(ctx, startDate, endDate)
}

// GetUserStats returns user statistics
func (s *service) GetUserStats(ctx context.Context, startDate, endDate time.Time) (*UserStats, error) {
	return s.repo.GetUserStats(ctx, startDate, endDate)
}

// GetProviderStats returns provider statistics
func (s *service) GetProviderStats(ctx context.Context) ([]*ProviderStats, error) {
	return s.repo.GetProviderStats(ctx)
}

// ListAuditLogs returns audit logs with filters
func (s *service) ListAuditLogs(ctx context.Context, adminID, entityType, entityID string, limit, offset int) ([]*AuditLog, int, error) {
	return s.repo.ListAuditLogs(ctx, adminID, entityType, entityID, limit, offset)
}

// generateToken generates a JWT token for an admin
func (s *service) generateToken(adminID, email, role string, expiresAt time.Time) (string, error) {
	// Create token claims
	claims := jwt.MapClaims{
		"sub":  adminID,
		"email": email,
		"role":  role,
		"type":  "admin",
		"exp":   expiresAt.Unix(),
		"iat":   time.Now().Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the admin ID
func ValidateToken(tokenString, jwtSecret string) (string, string, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", "", err
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		adminID := claims["sub"].(string)
		role := claims["role"].(string)
		return adminID, role, nil
	}

	return "", "", errors.New("invalid token")
}

// HashAPIKey hashes an API key for storage
func HashAPIKey(apiKey string) string {
	hash := sha512.Sum512([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}
