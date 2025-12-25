package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the database operations for admin service
type Repository interface {
	// Admin operations
	CreateAdmin(ctx context.Context, admin *Admin) error
	GetAdminByEmail(ctx context.Context, email string) (*Admin, error)
	GetAdminByID(ctx context.Context, id string) (*Admin, error)
	ListAdmins(ctx context.Context, limit, offset int) ([]*Admin, int, error)
	UpdateAdmin(ctx context.Context, id string, updates map[string]interface{}) error
	UpdateAdminLastLogin(ctx context.Context, id string) error
	DeleteAdmin(ctx context.Context, id string) error

	// Audit log operations
	CreateAuditLog(ctx context.Context, log *AuditLog) error
	ListAuditLogs(ctx context.Context, adminID string, entityType string, entityID string, limit, offset int) ([]*AuditLog, int, error)

	// Statistics
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
	GetBookingStats(ctx context.Context, startDate, endDate time.Time) (*BookingStats, error)
	GetRevenueStats(ctx context.Context, startDate, endDate time.Time) (*RevenueStats, error)
	GetUserStats(ctx context.Context, startDate, endDate time.Time) (*UserStats, error)
	GetProviderStats(ctx context.Context) ([]*ProviderStats, error)
}

type repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new admin repository
func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

// CreateAdmin creates a new admin user
func (r *repository) CreateAdmin(ctx context.Context, admin *Admin) error {
	admin.ID = uuid.New().String()
	admin.CreatedAt = time.Now()
	admin.UpdatedAt = time.Now()

	query := `
		INSERT INTO admin_users (id, email, password_hash, role, first_name, last_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		admin.ID,
		admin.Email,
		admin.PasswordHash,
		admin.Role,
		admin.FirstName,
		admin.LastName,
		admin.IsActive,
		admin.CreatedAt,
		admin.UpdatedAt,
	).Scan(&admin.ID, &admin.CreatedAt, &admin.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}

	return nil
}

// GetAdminByEmail retrieves an admin by email
func (r *repository) GetAdminByEmail(ctx context.Context, email string) (*Admin, error) {
	var admin Admin
	var lastLoginAt *time.Time

	query := `
		SELECT id, email, password_hash, role, first_name, last_name, is_active, last_login_at, created_at, updated_at
		FROM admin_users
		WHERE email = $1
	`

	err := r.db.QueryRow(ctx, query, email).Scan(
		&admin.ID,
		&admin.Email,
		&admin.PasswordHash,
		&admin.Role,
		&admin.FirstName,
		&admin.LastName,
		&admin.IsActive,
		&lastLoginAt,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("admin not found")
		}
		return nil, fmt.Errorf("failed to get admin by email: %w", err)
	}

	admin.LastLoginAt = lastLoginAt
	return &admin, nil
}

// GetAdminByID retrieves an admin by ID
func (r *repository) GetAdminByID(ctx context.Context, id string) (*Admin, error) {
	var admin Admin
	var lastLoginAt *time.Time

	query := `
		SELECT id, email, password_hash, role, first_name, last_name, is_active, last_login_at, created_at, updated_at
		FROM admin_users
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&admin.ID,
		&admin.Email,
		&admin.PasswordHash,
		&admin.Role,
		&admin.FirstName,
		&admin.LastName,
		&admin.IsActive,
		&lastLoginAt,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("admin not found")
		}
		return nil, fmt.Errorf("failed to get admin by id: %w", err)
	}

	admin.LastLoginAt = lastLoginAt
	return &admin, nil
}

// ListAdmins retrieves a list of admins with pagination
func (r *repository) ListAdmins(ctx context.Context, limit, offset int) ([]*Admin, int, error) {
	query := `
		SELECT id, email, password_hash, role, first_name, last_name, is_active, last_login_at, created_at, updated_at
		FROM admin_users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list admins: %w", err)
	}
	defer rows.Close()

	var admins []*Admin
	for rows.Next() {
		var admin Admin
		var lastLoginAt *time.Time

		err := rows.Scan(
			&admin.ID,
			&admin.Email,
			&admin.PasswordHash,
			&admin.Role,
			&admin.FirstName,
			&admin.LastName,
			&admin.IsActive,
			&lastLoginAt,
			&admin.CreatedAt,
			&admin.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan admin row: %w", err)
		}

		admin.LastLoginAt = lastLoginAt
		admins = append(admins, &admin)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM admin_users`
	err = r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get admin count: %w", err)
	}

	return admins, total, nil
}

// UpdateAdmin updates an admin user
func (r *repository) UpdateAdmin(ctx context.Context, id string, updates map[string]interface{}) error {
	// Build dynamic UPDATE query
	setClause := ""
	args := []interface{}{}
	argPos := 1

	for key, value := range updates {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", key, argPos)
		args = append(args, value)
		argPos++
	}

	if setClause == "" {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE admin_users
		SET %s, updated_at = CURRENT_TIMESTAMP
		WHERE id = $%d
	`, setClause, argPos)

	args = append(args, id)

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update admin: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("admin not found")
	}

	return nil
}

// UpdateAdminLastLogin updates the last login timestamp
func (r *repository) UpdateAdminLastLogin(ctx context.Context, id string) error {
	query := `
		UPDATE admin_users
		SET last_login_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("admin not found")
	}

	return nil
}

// DeleteAdmin soft deletes an admin user
func (r *repository) DeleteAdmin(ctx context.Context, id string) error {
	query := `DELETE FROM admin_users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("admin not found")
	}

	return nil
}

// CreateAuditLog creates a new audit log entry
func (r *repository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	log.ID = uuid.New().String()
	log.CreatedAt = time.Now()

	query := `
		INSERT INTO admin_audit_log (id, admin_id, admin_email, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(ctx, query,
		log.ID,
		log.AdminID,
		log.AdminEmail,
		log.Action,
		log.EntityType,
		log.EntityID,
		log.OldValues,
		log.NewValues,
		log.IPAddress,
		log.UserAgent,
		log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// ListAuditLogs retrieves audit logs with optional filters
func (r *repository) ListAuditLogs(ctx context.Context, adminID, entityType, entityID string, limit, offset int) ([]*AuditLog, int, error) {
	query := `
		SELECT id, admin_id, admin_email, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at
		FROM admin_audit_log
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if adminID != "" {
		query += fmt.Sprintf(" AND admin_id = $%d", argPos)
		args = append(args, adminID)
		argPos++
	}

	if entityType != "" {
		query += fmt.Sprintf(" AND entity_type = $%d", argPos)
		args = append(args, entityType)
		argPos++
	}

	if entityID != "" {
		query += fmt.Sprintf(" AND entity_id = $%d", argPos)
		args = append(args, entityID)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		var log AuditLog

		err := rows.Scan(
			&log.ID,
			&log.AdminID,
			&log.AdminEmail,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.OldValues,
			&log.NewValues,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log row: %w", err)
		}

		logs = append(logs, &log)
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM admin_audit_log WHERE 1=1`
	countArgs := []interface{}{}
	argPos = 1

	if adminID != "" {
		countQuery += fmt.Sprintf(" AND admin_id = $%d", argPos)
		countArgs = append(countArgs, adminID)
		argPos++
	}

	if entityType != "" {
		countQuery += fmt.Sprintf(" AND entity_type = $%d", argPos)
		countArgs = append(countArgs, entityType)
		argPos++
	}

	if entityID != "" {
		countQuery += fmt.Sprintf(" AND entity_id = $%d", argPos)
		countArgs = append(countArgs, entityID)
	}

	var total int
	err = r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get audit log count: %w", err)
	}

	return logs, total, nil
}

// GetDashboardStats retrieves dashboard statistics
func (r *repository) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	var stats DashboardStats

	// Get total users
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&stats.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get total users: %w", err)
	}

	// Get total bookings
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM bookings`).Scan(&stats.TotalBookings)
	if err != nil {
		return nil, fmt.Errorf("failed to get total bookings: %w", err)
	}

	// Get total revenue (successful payments)
	err = r.db.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'success'`).Scan(&stats.TotalRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get total revenue: %w", err)
	}

	// Get today's stats
	today := time.Now().Format("2006-01-02")

	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM users
		WHERE DATE(created_at) = $1 AND deleted_at IS NULL
	`, today).Scan(&stats.TodayUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get today users: %w", err)
	}

	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM bookings
		WHERE DATE(created_at) = $1
	`, today).Scan(&stats.TodayBookings)
	if err != nil {
		return nil, fmt.Errorf("failed to get today bookings: %w", err)
	}

	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM payments
		WHERE DATE(created_at) = $1 AND status = 'success'
	`, today).Scan(&stats.TodayRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get today revenue: %w", err)
	}

	// Get active providers count
	err = r.db.QueryRow(ctx, `SELECT COUNT(DISTINCT provider_code) FROM hotels WHERE is_active = true`).Scan(&stats.ActiveProviders)
	if err != nil {
		return nil, fmt.Errorf("failed to get active providers: %w", err)
	}

	return &stats, nil
}

// GetBookingStats retrieves booking statistics
func (r *repository) GetBookingStats(ctx context.Context, startDate, endDate time.Time) (*BookingStats, error) {
	var stats BookingStats

	// Get total and status breakdown
	rows, err := r.db.Query(ctx, `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'confirmed') as confirmed,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled,
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) FILTER (WHERE status = 'failed') as failed
		FROM bookings
		WHERE created_at >= $1 AND created_at <= $2
	`, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking stats: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(
			&stats.TotalBookings,
			&stats.ConfirmedBookings,
			&stats.PendingBookings,
			&stats.CancelledBookings,
			&stats.CompletedBookings,
			&stats.FailedBookings,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking stats: %w", err)
		}
	}

	// Get by status
	stats.ByStatus = make(map[string]int64)
	stats.ByStatus["pending"] = stats.PendingBookings
	stats.ByStatus["confirmed"] = stats.ConfirmedBookings
	stats.ByStatus["cancelled"] = stats.CancelledBookings
	stats.ByStatus["completed"] = stats.CompletedBookings
	stats.ByStatus["failed"] = stats.FailedBookings

	// Get by date
	dateRows, err := r.db.Query(ctx, `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM bookings
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY DATE(created_at)
		ORDER BY date
	`, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookings by date: %w", err)
	}
	defer dateRows.Close()

	stats.ByDate = make(map[string]int64)
	for dateRows.Next() {
		var dateStr string
		var count int64
		err = dateRows.Scan(&dateStr, &count)
		if err != nil {
			continue
		}
		stats.ByDate[dateStr] = count
	}

	return &stats, nil
}

// GetRevenueStats retrieves revenue statistics
func (r *repository) GetRevenueStats(ctx context.Context, startDate, endDate time.Time) (*RevenueStats, error) {
	var stats RevenueStats

	// Get total revenue
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM payments
		WHERE status = 'success' AND created_at >= $1 AND created_at <= $2
	`, startDate, endDate).Scan(&stats.TotalRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get total revenue: %w", err)
	}

	// Get this month revenue
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM payments
		WHERE status = 'success' AND created_at >= $1
	`, firstOfMonth).Scan(&stats.ThisMonthRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get this month revenue: %w", err)
	}

	// Get last month revenue
	lastMonth := firstOfMonth.AddDate(0, -1, 0)
	firstOfLastMonth := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.UTC)

	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM payments
		WHERE status = 'success' AND created_at >= $1 AND created_at < $2
	`, firstOfLastMonth, firstOfMonth).Scan(&stats.LastMonthRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get last month revenue: %w", err)
	}

	// Get by payment method
	stats.ByPaymentMethod = make(map[string]int64)
	pmRows, err := r.db.Query(ctx, `
		SELECT method, COALESCE(SUM(amount), 0)
		FROM payments
		WHERE status = 'success' AND created_at >= $1 AND created_at <= $2
		GROUP BY method
	`, startDate, endDate)
	if err == nil {
		defer pmRows.Close()
		for pmRows.Next() {
			var method string
			var amount int64
			pmRows.Scan(&method, &amount)
			stats.ByPaymentMethod[method] = amount
		}
	}

	// Get by provider
	stats.ByProvider = make(map[string]int64)
	provRows, err := r.db.Query(ctx, `
		SELECT b.provider_code, COALESCE(SUM(p.amount), 0)
		FROM bookings b
		JOIN payments p ON p.booking_id = b.id
		WHERE p.status = 'success' AND p.created_at >= $1 AND p.created_at <= $2
		GROUP BY b.provider_code
	`, startDate, endDate)
	if err == nil {
		defer provRows.Close()
		for provRows.Next() {
			var provider string
			var amount int64
			provRows.Scan(&provider, &amount)
			stats.ByProvider[provider] = amount
		}
	}

	// Get by date
	stats.ByDate = make(map[string]int64)
	dateRows, err := r.db.Query(ctx, `
		SELECT DATE(created_at) as date, COALESCE(SUM(amount), 0)
		FROM payments
		WHERE status = 'success' AND created_at >= $1 AND created_at <= $2
		GROUP BY DATE(created_at)
		ORDER BY date
	`, startDate, endDate)
	if err == nil {
		defer dateRows.Close()
		for dateRows.Next() {
			var dateStr string
			var amount int64
			dateRows.Scan(&dateStr, &amount)
			stats.ByDate[dateStr] = amount
		}
	}

	return &stats, nil
}

// GetUserStats retrieves user statistics
func (r *repository) GetUserStats(ctx context.Context, startDate, endDate time.Time) (*UserStats, error) {
	var stats UserStats

	// Get total users
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&stats.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get total users: %w", err)
	}

	// Get active users (users with bookings in last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT user_id)
		FROM bookings
		WHERE created_at >= $1
	`, thirtyDaysAgo).Scan(&stats.ActiveUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	// Get this month users
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM users
		WHERE created_at >= $1 AND deleted_at IS NULL
	`, firstOfMonth).Scan(&stats.ThisMonthUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get this month users: %w", err)
	}

	// Get last month users
	lastMonth := firstOfMonth.AddDate(0, -1, 0)
	firstOfLastMonth := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.UTC)

	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM users
		WHERE created_at >= $1 AND created_at < $2 AND deleted_at IS NULL
	`, firstOfLastMonth, firstOfMonth).Scan(&stats.LastMonthUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get last month users: %w", err)
	}

	// Get by date
	dateRows, err := r.db.Query(ctx, `
		SELECT DATE(created_at) as date, COUNT(*)
		FROM users
		WHERE created_at >= $1 AND created_at <= $2 AND deleted_at IS NULL
		GROUP BY DATE(created_at)
		ORDER BY date
	`, startDate, endDate)
	if err == nil {
		defer dateRows.Close()
		stats.ByDate = make(map[string]int64)
		for dateRows.Next() {
			var dateStr string
			var count int64
			dateRows.Scan(&dateStr, &count)
			stats.ByDate[dateStr] = count
		}
	}

	// Get top bookers
	topRows, err := r.db.Query(ctx, `
		SELECT
			u.id as user_id,
			u.email,
			u.first_name,
			u.last_name,
			COUNT(b.id) as total_bookings,
			COALESCE(SUM(p.amount), 0) as total_spent
		FROM users u
		LEFT JOIN bookings b ON b.user_id = u.id
		LEFT JOIN payments p ON p.booking_id = b.id AND p.status = 'success'
		WHERE u.deleted_at IS NULL
		GROUP BY u.id
		ORDER BY total_bookings DESC
		LIMIT 10
	`)
	if err == nil {
		defer topRows.Close()
		for topRows.Next() {
			var stat UserBookingStat
			topRows.Scan(
				&stat.UserID,
				&stat.Email,
				&stat.FirstName,
				&stat.LastName,
				&stat.TotalBookings,
				&stat.TotalSpent,
			)
			stats.TopBookers = append(stats.TopBookers, stat)
		}
	}

	return &stats, nil
}

// GetProviderStats retrieves provider statistics
func (r *repository) GetProviderStats(ctx context.Context) ([]*ProviderStats, error) {
	// This is a placeholder - in real implementation, you would have a provider_metrics table
	// For now, return empty stats
	return []*ProviderStats{}, nil
}
