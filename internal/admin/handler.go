package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

type Handler struct {
	service    Service
	jwtSecret  string
}

func NewHandler(service Service, jwtSecret string) *Handler {
	return &Handler{
		service:   service,
		jwtSecret: jwtSecret,
	}
}

// AuthMiddleware validates JWT token and extracts admin info
func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check Bearer format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Validate token
		adminID, role, err := ValidateToken(token, h.jwtSecret)
		if err != nil {
			logger.ErrorWithErr(err, "Token validation failed")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add admin info to context
		ctx := r.Context()
		ctx = contextWithAdminInfo(ctx, adminID, role)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

// permissionMiddleware checks if admin has required permission
func (h *Handler) permissionMiddleware(permission Permission) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			adminID, role, err := extractAdminInfo(r.Context())
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check permission
			if !role.HasPermission(permission) {
				logger.Warnf("Admin %s attempted to access %s without permission", adminID, permission)
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}

// Helper function to write JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.ErrorWithErr(err, "Failed to encode JSON response")
	}
}

// Helper function to write error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// Helper function to get IP address
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// Helper function to parse pagination parameters
func parsePagination(r *http.Request) (limit, offset int, err error) {
	// Default values
	limit = 20
	offset = 0

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			return 0, 0, fmt.Errorf("invalid limit parameter")
		}
	}

	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return 0, 0, fmt.Errorf("invalid offset parameter")
		}
	}

	return limit, offset, nil
}

// Handler: POST /api/v1/admin/login
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get IP and user agent
	ipAddress := getIPAddress(r)
	userAgent := r.Header.Get("User-Agent")

	// Call service
	resp, err := h.service.Login(r.Context(), &req, ipAddress, userAgent)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Handler: POST /api/v1/admin/logout
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	adminID, _, err := extractAdminInfo(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	_ = h.service.Logout(r.Context(), adminID)

	writeJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// Handler: GET /api/v1/admin/me
func (h *Handler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	adminID, _, err := extractAdminInfo(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	admin, err := h.service.GetMe(r.Context(), adminID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Admin not found")
		return
	}

	writeJSON(w, http.StatusOK, admin)
}

// Handler: GET /api/v1/admin/dashboard
func (h *Handler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	stats, err := h.service.GetDashboardStats(r.Context())
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get dashboard stats")
		writeError(w, http.StatusInternalServerError, "Failed to get dashboard stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// Handler: GET /api/v1/admin/admins
func (h *Handler) HandleListAdmins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	admins, total, err := h.service.ListAdmins(r.Context(), limit, offset)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to list admins")
		writeError(w, http.StatusInternalServerError, "Failed to list admins")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"admins": admins,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Handler: POST /api/v1/admin/admins
func (h *Handler) HandleCreateAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req CreateAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	adminID, _, err := extractAdminInfo(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ipAddress := getIPAddress(r)
	userAgent := r.Header.Get("User-Agent")

	admin, err := h.service.CreateAdmin(r.Context(), adminID, &req, ipAddress, userAgent)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, admin)
}

// Handler: GET /api/v1/admin/admins/:id
func (h *Handler) HandleGetAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/admins/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Admin ID required")
		return
	}

	admin, err := h.service.GetAdmin(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Admin not found")
		return
	}

	writeJSON(w, http.StatusOK, admin)
}

// Handler: PUT /api/v1/admin/admins/:id
func (h *Handler) HandleUpdateAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/admins/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Admin ID required")
		return
	}

	var req UpdateAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	adminID, _, err := extractAdminInfo(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ipAddress := getIPAddress(r)
	userAgent := r.Header.Get("User-Agent")

	err = h.service.UpdateAdmin(r.Context(), adminID, id, &req, ipAddress, userAgent)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Admin updated successfully"})
}

// Handler: DELETE /api/v1/admin/admins/:id
func (h *Handler) HandleDeleteAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/admins/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Admin ID required")
		return
	}

	adminID, _, err := extractAdminInfo(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ipAddress := getIPAddress(r)
	userAgent := r.Header.Get("User-Agent")

	err = h.service.DeleteAdmin(r.Context(), adminID, id, ipAddress, userAgent)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Admin deleted successfully"})
}

// Handler: GET /api/v1/admin/users
func (h *Handler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	users, total, err := h.service.ListUsers(r.Context(), limit, offset)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to list users")
		writeError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Handler: GET /api/v1/admin/users/:id
func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/users/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "User ID required")
		return
	}

	user, err := h.service.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// Handler: PUT /api/v1/admin/users/:id
func (h *Handler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/users/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "User ID required")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	adminID, _, err := extractAdminInfo(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ipAddress := getIPAddress(r)
	userAgent := r.Header.Get("User-Agent")

	err = h.service.UpdateUser(r.Context(), adminID, id, &req, ipAddress, userAgent)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "User updated successfully"})
}

// Handler: DELETE /api/v1/admin/users/:id
func (h *Handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/users/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "User ID required")
		return
	}

	adminID, _, err := extractAdminInfo(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ipAddress := getIPAddress(r)
	userAgent := r.Header.Get("User-Agent")

	err = h.service.DeleteUser(r.Context(), adminID, id, ipAddress, userAgent)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// Handler: GET /api/v1/admin/bookings
func (h *Handler) HandleListBookings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Parse filters
	status := r.URL.Query().Get("status")
	var startDate, endDate *time.Time

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		t, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = &t
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		t, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = &t
		}
	}

	bookings, total, err := h.service.ListBookings(r.Context(), status, startDate, endDate, limit, offset)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to list bookings")
		writeError(w, http.StatusInternalServerError, "Failed to list bookings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"bookings": bookings,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// Handler: GET /api/v1/admin/bookings/:id
func (h *Handler) HandleGetBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/bookings/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Booking ID required")
		return
	}

	booking, err := h.service.GetBooking(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Booking not found")
		return
	}

	writeJSON(w, http.StatusOK, booking)
}

// Handler: PUT /api/v1/admin/bookings/:id
func (h *Handler) HandleUpdateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/bookings/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Booking ID required")
		return
	}

	var req UpdateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	adminID, _, err := extractAdminInfo(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ipAddress := getIPAddress(r)
	userAgent := r.Header.Get("User-Agent")

	err = h.service.UpdateBooking(r.Context(), adminID, id, &req, ipAddress, userAgent)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Booking updated successfully"})
}

// Handler: GET /api/v1/admin/bookings/stats
func (h *Handler) HandleBookingStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		t, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = t
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		t, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = t
		}
	}

	stats, err := h.service.GetBookingStats(r.Context(), startDate, endDate)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get booking stats")
		writeError(w, http.StatusInternalServerError, "Failed to get booking stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// Handler: GET /api/v1/admin/providers
func (h *Handler) HandleListProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	providers, err := h.service.ListProviders(r.Context())
	if err != nil {
		logger.ErrorWithErr(err, "Failed to list providers")
		writeError(w, http.StatusInternalServerError, "Failed to list providers")
		return
	}

	writeJSON(w, http.StatusOK, providers)
}

// Handler: GET /api/v1/admin/providers/:code
func (h *Handler) HandleGetProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	code := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/providers/")
	if code == "" {
		writeError(w, http.StatusBadRequest, "Provider code required")
		return
	}

	provider, err := h.service.GetProvider(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusNotFound, "Provider not found")
		return
	}

	writeJSON(w, http.StatusOK, provider)
}

// Handler: PUT /api/v1/admin/providers/:code
func (h *Handler) HandleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	code := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/providers/")
	if code == "" {
		writeError(w, http.StatusBadRequest, "Provider code required")
		return
	}

	var req UpdateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	adminID, _, err := extractAdminInfo(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ipAddress := getIPAddress(r)
	userAgent := r.Header.Get("User-Agent")

	err = h.service.UpdateProvider(r.Context(), adminID, code, &req, ipAddress, userAgent)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Provider updated successfully"})
}

// Handler: GET /api/v1/admin/analytics/revenue
func (h *Handler) HandleRevenueStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse date range (default to this month)
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := now

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		t, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = t
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		t, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = t
		}
	}

	stats, err := h.service.GetRevenueStats(r.Context(), startDate, endDate)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get revenue stats")
		writeError(w, http.StatusInternalServerError, "Failed to get revenue stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// Handler: GET /api/v1/admin/analytics/users
func (h *Handler) HandleUserStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse date range (default to this month)
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := now

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		t, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = t
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		t, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = t
		}
	}

	stats, err := h.service.GetUserStats(r.Context(), startDate, endDate)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get user stats")
		writeError(w, http.StatusInternalServerError, "Failed to get user stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// Handler: GET /api/v1/admin/analytics/providers
func (h *Handler) HandleProviderStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	stats, err := h.service.GetProviderStats(r.Context())
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get provider stats")
		writeError(w, http.StatusInternalServerError, "Failed to get provider stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// Handler: GET /api/v1/admin/audit-logs
func (h *Handler) HandleAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Parse filters
	adminID := r.URL.Query().Get("admin_id")
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")

	logs, total, err := h.service.ListAuditLogs(r.Context(), adminID, entityType, entityID, limit, offset)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to list audit logs")
		writeError(w, http.StatusInternalServerError, "Failed to list audit logs")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Context key for admin info
type contextKey string

const (
	adminIDKey contextKey = "adminID"
	roleKey    contextKey = "role"
)

// Helper functions for context
func contextWithAdminInfo(ctx context.Context, adminID, role string) context.Context {
	ctx = context.WithValue(ctx, adminIDKey, adminID)
	ctx = context.WithValue(ctx, roleKey, role)
	return ctx
}

func extractAdminInfo(ctx context.Context) (string, AdminRole, error) {
	adminID, ok := ctx.Value(adminIDKey).(string)
	if !ok {
		return "", "", fmt.Errorf("admin ID not found in context")
	}

	roleStr, ok := ctx.Value(roleKey).(string)
	if !ok {
		return "", "", fmt.Errorf("role not found in context")
	}

	return adminID, AdminRole(roleStr), nil
}
