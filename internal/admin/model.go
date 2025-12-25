package admin

import (
	"time"
)

// Admin represents an admin user in the system
type Admin struct {
	ID           string     `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Role         AdminRole  `json:"role" db:"role"`
	FirstName    string     `json:"first_name" db:"first_name"`
	LastName     string     `json:"last_name" db:"last_name"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// AdminRole represents the role of an admin user
type AdminRole string

const (
	RoleSuperAdmin AdminRole = "super_admin" // Full access to everything
	RoleAdmin      AdminRole = "admin"       // Full access except user management
	RoleModerator  AdminRole = "moderator"   // Can manage reviews and bookings only
	RoleSupport    AdminRole = "support"     // Read-only access for customer support
)

// Permission represents a specific permission
type Permission string

const (
	// User permissions
	PermissionUserRead   Permission = "users:read"
	PermissionUserWrite  Permission = "users:write"
	PermissionUserDelete Permission = "users:delete"

	// Booking permissions
	PermissionBookingRead   Permission = "bookings:read"
	PermissionBookingWrite  Permission = "bookings:write"
	PermissionBookingDelete Permission = "bookings:delete"

	// Provider permissions
	PermissionProviderRead  Permission = "providers:read"
	PermissionProviderWrite Permission = "providers:write"

	// Review permissions
	PermissionReviewRead   Permission = "reviews:read"
	PermissionReviewWrite  Permission = "reviews:write"
	PermissionReviewDelete Permission = "reviews:delete"

	// Analytics permissions
	PermissionAnalyticsRead Permission = "analytics:read"

	// Configuration permissions
	PermissionConfigRead  Permission = "config:read"
	PermissionConfigWrite Permission = "config:write"

	// Admin management permissions
	PermissionAdminRead  Permission = "admins:read"
	PermissionAdminWrite Permission = "admins:write"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[AdminRole][]Permission{
	RoleSuperAdmin: {
		// All permissions
		PermissionUserRead, PermissionUserWrite, PermissionUserDelete,
		PermissionBookingRead, PermissionBookingWrite, PermissionBookingDelete,
		PermissionProviderRead, PermissionProviderWrite,
		PermissionReviewRead, PermissionReviewWrite, PermissionReviewDelete,
		PermissionAnalyticsRead,
		PermissionConfigRead, PermissionConfigWrite,
		PermissionAdminRead, PermissionAdminWrite,
	},
	RoleAdmin: {
		// Almost all permissions except admin management
		PermissionUserRead, PermissionUserWrite, PermissionUserDelete,
		PermissionBookingRead, PermissionBookingWrite, PermissionBookingDelete,
		PermissionProviderRead, PermissionProviderWrite,
		PermissionReviewRead, PermissionReviewWrite, PermissionReviewDelete,
		PermissionAnalyticsRead,
		PermissionConfigRead, PermissionConfigWrite,
	},
	RoleModerator: {
		// Limited to reviews and bookings
		PermissionReviewRead, PermissionReviewWrite, PermissionReviewDelete,
		PermissionBookingRead, PermissionBookingWrite,
	},
	RoleSupport: {
		// Read-only for support
		PermissionUserRead,
		PermissionBookingRead,
		PermissionReviewRead,
	},
}

// HasPermission checks if an admin role has a specific permission
func (r AdminRole) HasPermission(permission Permission) bool {
	permissions, exists := RolePermissions[r]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// AuditLog represents an audit trail entry for admin actions
type AuditLog struct {
	ID           string                 `json:"id" db:"id"`
	AdminID      string                 `json:"admin_id" db:"admin_id"`
	AdminEmail   string                 `json:"admin_email" db:"admin_email"` // Denormalized for easier querying
	Action       string                 `json:"action" db:"action"`           // e.g., "user.updated", "booking.deleted"
	EntityType   string                 `json:"entity_type" db:"entity_type"` // e.g., "user", "booking", "provider"
	EntityID     string                 `json:"entity_id" db:"entity_id"`
	OldValues    map[string]interface{} `json:"old_values,omitempty" db:"old_values"`
	NewValues    map[string]interface{} `json:"new_values,omitempty" db:"new_values"`
	IPAddress    string                 `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    string                 `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// LoginRequest represents an admin login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	Token     string `json:"token"`
	Admin     Admin  `json:"admin"`
	ExpiresAt int64  `json:"expires_at"` // Unix timestamp
}

// CreateAdminRequest represents a request to create a new admin
type CreateAdminRequest struct {
	Email     string    `json:"email" validate:"required,email"`
	Password  string    `json:"password" validate:"required,min=8"`
	Role      AdminRole `json:"role" validate:"required"`
	FirstName string    `json:"first_name" validate:"required"`
	LastName  string    `json:"last_name" validate:"required"`
}

// UpdateAdminRequest represents a request to update an admin
type UpdateAdminRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=2"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=2"`
	Role      *AdminRole `json:"role,omitempty" validate:"omitempty"`
	IsActive  *bool    `json:"is_active,omitempty"`
}

// UpdateUserRequest represents a request to update a regular user
type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=2"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=2"`
	Phone     *string `json:"phone,omitempty" validate:"omitempty,e164"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// UpdateBookingRequest represents a request to update a booking
type UpdateBookingRequest struct {
	Status        *string `json:"status,omitempty" validate:"omitempty,oneof=pending confirmed paid cancelled failed completed"`
	AdminNotes    *string `json:"admin_notes,omitempty" validate:"omitempty,max=1000"`
	CancellationReason *string `json:"cancellation_reason,omitempty" validate:"omitempty,max=500"`
}

// ProviderConfig represents provider configuration
type ProviderConfig struct {
	ProviderCode   string `json:"provider_code" validate:"required"`
	APIKey         string `json:"api_key,omitempty"`
	APISecret      string `json:"api_secret,omitempty"`
	BaseURL        string `json:"base_url,omitempty"`
	IsActive       bool   `json:"is_active"`
	RateLimitRPS   int    `json:"rate_limit_rps,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}

// UpdateProviderRequest represents a request to update provider config
type UpdateProviderRequest struct {
	APIKey         *string `json:"api_key,omitempty"`
	APISecret      *string `json:"api_secret,omitempty"`
	BaseURL        *string `json:"base_url,omitempty"`
	IsActive       *bool   `json:"is_active,omitempty"`
	RateLimitRPS   *int    `json:"rate_limit_rps,omitempty" validate:"omitempty,min=1,max=100"`
	TimeoutSeconds *int    `json:"timeout_seconds,omitempty" validate:"omitempty,min=1,max=60"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalUsers      int64 `json:"total_users"`
	TotalBookings   int64 `json:"total_bookings"`
	TotalRevenue    int64 `json:"total_revenue"` // In cents/rupiah
	TodayUsers      int64 `json:"today_users"`
	TodayBookings   int64 `json:"today_bookings"`
	TodayRevenue    int64 `json:"today_revenue"`
	ActiveProviders int   `json:"active_providers"`
}

// BookingStats represents booking statistics
type BookingStats struct {
	TotalBookings    int64            `json:"total_bookings"`
	ConfirmedBookings int64           `json:"confirmed_bookings"`
	PendingBookings  int64            `json:"pending_bookings"`
	CancelledBookings int64           `json:"cancelled_bookings"`
	CompletedBookings int64           `json:"completed_bookings"`
	FailedBookings   int64            `json:"failed_bookings"`
	ByStatus         map[string]int64 `json:"by_status"`
	ByDate           map[string]int64 `json:"by_date"` // Key: YYYY-MM-DD
}

// RevenueStats represents revenue statistics
type RevenueStats struct {
	TotalRevenue      int64            `json:"total_revenue"`
	ThisMonthRevenue  int64            `json:"this_month_revenue"`
	LastMonthRevenue  int64            `json:"last_month_revenue"`
	ByPaymentMethod   map[string]int64 `json:"by_payment_method"`
	ByProvider        map[string]int64 `json:"by_provider"`
	ByDate            map[string]int64 `json:"by_date"` // Key: YYYY-MM-DD
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers       int64            `json:"total_users"`
	ActiveUsers      int64            `json:"active_users"`
	ThisMonthUsers   int64            `json:"this_month_users"`
	LastMonthUsers   int64            `json:"last_month_users"`
	ByDate           map[string]int64 `json:"by_date"` // Key: YYYY-MM-DD
	TopBookers       []UserBookingStat `json:"top_bookers"`
}

// UserBookingStat represents user booking statistics
type UserBookingStat struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	TotalBookings int   `json:"total_bookings"`
	TotalSpent   int64  `json:"total_spent"` // In cents/rupiah
}

// ProviderStats represents provider statistics
type ProviderStats struct {
	ProviderCode      string  `json:"provider_code"`
	TotalRequests     int64   `json:"total_requests"`
	SuccessfulRequests int64  `json:"successful_requests"`
	FailedRequests    int64   `json:"failed_requests"`
	SuccessRate       float64 `json:"success_rate"`
	AverageResponseTimeMS int64 `json:"average_response_time_ms"`
	IsActive          bool    `json:"is_active"`
	LastUsedAt        *time.Time `json:"last_used_at"`
}

// FeatureFlag represents a feature flag
type FeatureFlag struct {
	Key         string      `json:"key" validate:"required"`
	Description string      `json:"description"`
	Enabled     bool        `json:"enabled"`
	RolloutPercentage int   `json:"rollout_percentage"` // 0-100
	UserWhitelist []string  `json:"user_whitelist,omitempty"` // User IDs
	Config      interface{} `json:"config,omitempty"`
}

// Config represents system configuration
type Config struct {
	Key         string      `json:"key" validate:"required"`
	Value       interface{} `json:"value"`
	Description string      `json:"description"`
	Type        string      `json:"type"` // string, number, boolean, json
	UpdatedAt   time.Time   `json:"updated_at"`
}
