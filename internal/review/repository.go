package review

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the database operations for review service
type Repository interface {
	// Review CRUD
	CreateReview(ctx context.Context, review *Review) error
	GetReviewByID(ctx context.Context, id string) (*Review, error)
	GetReviewsByHotelID(ctx context.Context, hotelID string, status ReviewStatus, limit, offset int, sortBy string) ([]*Review, int, error)
	GetReviewsByUserID(ctx context.Context, userID string, limit, offset int) ([]*Review, int, error)
	UpdateReview(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteReview(ctx context.Context, id string) error
	GetReviewByBookingID(ctx context.Context, bookingID string) (*Review, error)

	// Review moderation
	GetPendingReviews(ctx context.Context, limit, offset int) ([]*Review, int, error)
	GetFlaggedReviews(ctx context.Context, limit, offset int) ([]*Review, int, error)
	ApproveReview(ctx context.Context, reviewID, moderatorID string, note *string) error
	RejectReview(ctx context.Context, reviewID, moderatorID string, note *string) error
	FlagReview(ctx context.Context, reviewID string) error

	// Review statistics
	GetHotelReviewStats(ctx context.Context, hotelID string) (*ReviewStats, error)
	GetModerationStats(ctx context.Context) (*ModerationStats, error)
	GetReviewAnalytics(ctx context.Context) (*ReviewAnalytics, error)

	// Helpful votes
	CreateHelpfulVote(ctx context.Context, vote *HelpfulVote) error
	DeleteHelpfulVote(ctx context.Context, reviewID, userID string) error
	HasVotedHelpful(ctx context.Context, reviewID, userID string) (bool, error)
	IncrementHelpfulCount(ctx context.Context, reviewID string) error
	DecrementHelpfulCount(ctx context.Context, reviewID string) error

	// Flags
	CreateFlag(ctx context.Context, flag *ReviewFlag) error
	GetFlagsByReviewID(ctx context.Context, reviewID string) ([]*ReviewFlag, error)

	// Hotel responses
	UpdateHotelResponse(ctx context.Context, reviewID string, response string) error
	DeleteHotelResponse(ctx context.Context, reviewID string) error

	// Moderation keywords
	GetAllModerationKeywords(ctx context.Context) ([]*ModerationKeyword, error)
}

type repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new review repository
func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

// CreateReview creates a new review
func (r *repository) CreateReview(ctx context.Context, review *Review) error {
	review.ID = uuid.New().String()
	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()
	review.Status = StatusPending
	review.HelpfulCount = 0

	query := `
		INSERT INTO reviews (
			id, user_id, hotel_id, booking_id,
			overall_rating, cleanliness_rating, service_rating, location_rating, value_rating, facility_rating,
			title, comment, photos, helpful_count, status,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		review.ID, review.UserID, review.HotelID, review.BookingID,
		review.OverallRating, review.CleanlinessRating, review.ServiceRating,
		review.LocationRating, review.ValueRating, review.FacilityRating,
		review.Title, review.Comment, review.Photos, review.HelpfulCount, review.Status,
		review.CreatedAt, review.UpdatedAt,
	).Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}

	return nil
}

// GetReviewByID retrieves a review by ID
func (r *repository) GetReviewByID(ctx context.Context, id string) (*Review, error) {
	var review Review
	var deletedAt *time.Time

	query := `
		SELECT
			r.id, r.user_id, r.hotel_id, r.booking_id,
			r.overall_rating, r.cleanliness_rating, r.service_rating, r.location_rating, r.value_rating, r.facility_rating,
			r.title, r.comment, r.photos, r.helpful_count,
			r.status, r.moderation_note, r.moderated_by, r.moderated_at,
			r.hotel_response, r.hotel_response_at,
			r.created_at, r.updated_at, r.deleted_at,
			u.email as user_email, CONCAT(u.first_name, ' ', u.last_name) as user_name,
			h.name as hotel_name
		FROM reviews r
		LEFT JOIN users u ON r.user_id = u.id
		LEFT JOIN hotels h ON r.hotel_id = h.id
		WHERE r.id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&review.ID, &review.UserID, &review.HotelID, &review.BookingID,
		&review.OverallRating, &review.CleanlinessRating, &review.ServiceRating,
		&review.LocationRating, &review.ValueRating, &review.FacilityRating,
		&review.Title, &review.Comment, &review.Photos, &review.HelpfulCount,
		&review.Status, &review.ModerationNote, &review.ModeratedBy, &review.ModeratedAt,
		&review.HotelResponse, &review.HotelResponseAt,
		&review.CreatedAt, &review.UpdatedAt, &deletedAt,
		&review.UserEmail, &review.UserName, &review.HotelName,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("review not found")
		}
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	review.DeletedAt = deletedAt
	return &review, nil
}

// GetReviewsByHotelID retrieves reviews for a hotel with pagination
func (r *repository) GetReviewsByHotelID(ctx context.Context, hotelID string, status ReviewStatus, limit, offset int, sortBy string) ([]*Review, int, error) {
	// Build query with status filter
	statusFilter := ""
	args := []interface{}{hotelID}
	argPos := 2

	if status != "" {
		statusFilter = fmt.Sprintf(" AND r.status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	// Determine sort order
	orderBy := "r.created_at DESC"
	switch sortBy {
	case "rating_asc":
		orderBy = "r.overall_rating ASC"
	case "rating_desc":
		orderBy = "r.overall_rating DESC"
	case "helpful":
		orderBy = "r.helpful_count DESC"
	case "oldest":
		orderBy = "r.created_at ASC"
	}

	query := fmt.Sprintf(`
		SELECT
			r.id, r.user_id, r.hotel_id, r.booking_id,
			r.overall_rating, r.cleanliness_rating, r.service_rating, r.location_rating, r.value_rating, r.facility_rating,
			r.title, r.comment, r.photos, r.helpful_count,
			r.status, r.moderation_note, r.moderated_by, r.moderated_at,
			r.hotel_response, r.hotel_response_at,
			r.created_at, r.updated_at, r.deleted_at,
			u.email as user_email, CONCAT(u.first_name, ' ', u.last_name) as user_name,
			h.name as hotel_name
		FROM reviews r
		LEFT JOIN users u ON r.user_id = u.id
		LEFT JOIN hotels h ON r.hotel_id = h.id
		WHERE r.hotel_id = $1 AND r.deleted_at IS NULL%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, statusFilter, orderBy, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*Review
	for rows.Next() {
		var review Review
		var deletedAt *time.Time

		err := rows.Scan(
			&review.ID, &review.UserID, &review.HotelID, &review.BookingID,
			&review.OverallRating, &review.CleanlinessRating, &review.ServiceRating,
			&review.LocationRating, &review.ValueRating, &review.FacilityRating,
			&review.Title, &review.Comment, &review.Photos, &review.HelpfulCount,
			&review.Status, &review.ModerationNote, &review.ModeratedBy, &review.ModeratedAt,
			&review.HotelResponse, &review.HotelResponseAt,
			&review.CreatedAt, &review.UpdatedAt, &deletedAt,
			&review.UserEmail, &review.UserName, &review.HotelName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan review: %w", err)
		}

		review.DeletedAt = deletedAt
		reviews = append(reviews, &review)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM reviews
		WHERE hotel_id = $1 AND deleted_at IS NULL
	`
	countArgs := []interface{}{hotelID}

	if status != "" {
		countQuery += " AND status = $2"
		countArgs = append(countArgs, status)
	}

	var total int
	err = r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get review count: %w", err)
	}

	return reviews, total, nil
}

// GetReviewsByUserID retrieves reviews by a user
func (r *repository) GetReviewsByUserID(ctx context.Context, userID string, limit, offset int) ([]*Review, int, error) {
	query := `
		SELECT
			r.id, r.user_id, r.hotel_id, r.booking_id,
			r.overall_rating, r.cleanliness_rating, r.service_rating, r.location_rating, r.value_rating, r.facility_rating,
			r.title, r.comment, r.photos, r.helpful_count,
			r.status, r.moderation_note, r.moderated_by, r.moderated_at,
			r.hotel_response, r.hotel_response_at,
			r.created_at, r.updated_at, r.deleted_at,
			h.name as hotel_name
		FROM reviews r
		LEFT JOIN hotels h ON r.hotel_id = h.id
		WHERE r.user_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*Review
	for rows.Next() {
		var review Review
		var deletedAt *time.Time

		err := rows.Scan(
			&review.ID, &review.UserID, &review.HotelID, &review.BookingID,
			&review.OverallRating, &review.CleanlinessRating, &review.ServiceRating,
			&review.LocationRating, &review.ValueRating, &review.FacilityRating,
			&review.Title, &review.Comment, &review.Photos, &review.HelpfulCount,
			&review.Status, &review.ModerationNote, &review.ModeratedBy, &review.ModeratedAt,
			&review.HotelResponse, &review.HotelResponseAt,
			&review.CreatedAt, &review.UpdatedAt, &deletedAt,
			&review.HotelName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan review: %w", err)
		}

		review.DeletedAt = deletedAt
		reviews = append(reviews, &review)
	}

	// Get total count
	var total int
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM reviews WHERE user_id = $1 AND deleted_at IS NULL`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get review count: %w", err)
	}

	return reviews, total, nil
}

// UpdateReview updates a review
func (r *repository) UpdateReview(ctx context.Context, id string, updates map[string]interface{}) error {
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
		UPDATE reviews
		SET %s, updated_at = CURRENT_TIMESTAMP
		WHERE id = $%d
	`, setClause, argPos)

	args = append(args, id)

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update review: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

// DeleteReview soft deletes a review
func (r *repository) DeleteReview(ctx context.Context, id string) error {
	query := `UPDATE reviews SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

// GetReviewByBookingID retrieves a review by booking ID
func (r *repository) GetReviewByBookingID(ctx context.Context, bookingID string) (*Review, error) {
	var review Review
	var deletedAt *time.Time

	query := `
		SELECT id, user_id, hotel_id, booking_id,
			overall_rating, cleanliness_rating, service_rating, location_rating, value_rating, facility_rating,
			title, comment, photos, helpful_count,
			status, moderation_note, moderated_by, moderated_at,
			hotel_response, hotel_response_at,
			created_at, updated_at, deleted_at
		FROM reviews
		WHERE booking_id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRow(ctx, query, bookingID).Scan(
		&review.ID, &review.UserID, &review.HotelID, &review.BookingID,
		&review.OverallRating, &review.CleanlinessRating, &review.ServiceRating,
		&review.LocationRating, &review.ValueRating, &review.FacilityRating,
		&review.Title, &review.Comment, &review.Photos, &review.HelpfulCount,
		&review.Status, &review.ModerationNote, &review.ModeratedBy, &review.ModeratedAt,
		&review.HotelResponse, &review.HotelResponseAt,
		&review.CreatedAt, &review.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No review found for this booking
		}
		return nil, fmt.Errorf("failed to get review by booking: %w", err)
	}

	review.DeletedAt = deletedAt
	return &review, nil
}

// GetPendingReviews retrieves pending reviews for moderation
func (r *repository) GetPendingReviews(ctx context.Context, limit, offset int) ([]*Review, int, error) {
	query := `
		SELECT
			r.id, r.user_id, r.hotel_id, r.booking_id,
			r.overall_rating, r.cleanliness_rating, r.service_rating, r.location_rating, r.value_rating, r.facility_rating,
			r.title, r.comment, r.photos, r.helpful_count,
			r.status, r.moderation_note, r.moderated_by, r.moderated_at,
			r.hotel_response, r.hotel_response_at,
			r.created_at, r.updated_at, r.deleted_at,
			u.email as user_email, CONCAT(u.first_name, ' ', u.last_name) as user_name,
			h.name as hotel_name
		FROM reviews r
		LEFT JOIN users u ON r.user_id = u.id
		LEFT JOIN hotels h ON r.hotel_id = h.id
		WHERE r.status = 'pending' AND r.deleted_at IS NULL
		ORDER BY r.created_at ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get pending reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*Review
	for rows.Next() {
		var review Review
		var deletedAt *time.Time

		err := rows.Scan(
			&review.ID, &review.UserID, &review.HotelID, &review.BookingID,
			&review.OverallRating, &review.CleanlinessRating, &review.ServiceRating,
			&review.LocationRating, &review.ValueRating, &review.FacilityRating,
			&review.Title, &review.Comment, &review.Photos, &review.HelpfulCount,
			&review.Status, &review.ModerationNote, &review.ModeratedBy, &review.ModeratedAt,
			&review.HotelResponse, &review.HotelResponseAt,
			&review.CreatedAt, &review.UpdatedAt, &deletedAt,
			&review.UserEmail, &review.UserName, &review.HotelName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan review: %w", err)
		}

		review.DeletedAt = deletedAt
		reviews = append(reviews, &review)
	}

	// Get total count
	var total int
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM reviews WHERE status = 'pending' AND deleted_at IS NULL`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get pending count: %w", err)
	}

	return reviews, total, nil
}

// GetFlaggedReviews retrieves flagged reviews
func (r *repository) GetFlaggedReviews(ctx context.Context, limit, offset int) ([]*Review, int, error) {
	query := `
		SELECT
			r.id, r.user_id, r.hotel_id, r.booking_id,
			r.overall_rating, r.cleanliness_rating, r.service_rating, r.location_rating, r.value_rating, r.facility_rating,
			r.title, r.comment, r.photos, r.helpful_count,
			r.status, r.moderation_note, r.moderated_by, r.moderated_at,
			r.hotel_response, r.hotel_response_at,
			r.created_at, r.updated_at, r.deleted_at,
			u.email as user_email, CONCAT(u.first_name, ' ', u.last_name) as user_name,
			h.name as hotel_name
		FROM reviews r
		LEFT JOIN users u ON r.user_id = u.id
		LEFT JOIN hotels h ON r.hotel_id = h.id
		WHERE r.status = 'flagged' AND r.deleted_at IS NULL
		ORDER BY r.created_at ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get flagged reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*Review
	for rows.Next() {
		var review Review
		var deletedAt *time.Time

		err := rows.Scan(
			&review.ID, &review.UserID, &review.HotelID, &review.BookingID,
			&review.OverallRating, &review.CleanlinessRating, &review.ServiceRating,
			&review.LocationRating, &review.ValueRating, &review.FacilityRating,
			&review.Title, &review.Comment, &review.Photos, &review.HelpfulCount,
			&review.Status, &review.ModerationNote, &review.ModeratedBy, &review.ModeratedAt,
			&review.HotelResponse, &review.HotelResponseAt,
			&review.CreatedAt, &review.UpdatedAt, &deletedAt,
			&review.UserEmail, &review.UserName, &review.HotelName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan review: %w", err)
		}

		review.DeletedAt = deletedAt
		reviews = append(reviews, &review)
	}

	// Get total count
	var total int
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM reviews WHERE status = 'flagged' AND deleted_at IS NULL`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get flagged count: %w", err)
	}

	return reviews, total, nil
}

// ApproveReview approves a review
func (r *repository) ApproveReview(ctx context.Context, reviewID, moderatorID string, note *string) error {
	query := `
		UPDATE reviews
		SET status = 'approved',
			moderated_by = $2,
			moderated_at = CURRENT_TIMESTAMP,
			moderation_note = $3,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, reviewID, moderatorID, note)
	if err != nil {
		return fmt.Errorf("failed to approve review: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

// RejectReview rejects a review
func (r *repository) RejectReview(ctx context.Context, reviewID, moderatorID string, note *string) error {
	query := `
		UPDATE reviews
		SET status = 'rejected',
			moderated_by = $2,
			moderated_at = CURRENT_TIMESTAMP,
			moderation_note = $3,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, reviewID, moderatorID, note)
	if err != nil {
		return fmt.Errorf("failed to reject review: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

// FlagReview flags a review
func (r *repository) FlagReview(ctx context.Context, reviewID string) error {
	query := `UPDATE reviews SET status = 'flagged', updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.Exec(ctx, query, reviewID)
	if err != nil {
		return fmt.Errorf("failed to flag review: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

// GetHotelReviewStats retrieves review statistics for a hotel
func (r *repository) GetHotelReviewStats(ctx context.Context, hotelID string) (*ReviewStats, error) {
	var stats ReviewStats
	stats.HotelID = hotelID
	stats.RatingDistribution = make(map[int]int)

	// Get overall stats
	query := `
		SELECT
			COUNT(*) as review_count,
			ROUND(AVG(overall_rating)::numeric, 2) as overall_rating,
			ROUND(AVG(cleanliness_rating)::numeric, 2) as cleanliness_rating,
			ROUND(AVG(service_rating)::numeric, 2) as service_rating,
			ROUND(AVG(location_rating)::numeric, 2) as location_rating,
			ROUND(AVG(value_rating)::numeric, 2) as value_rating,
			ROUND(AVG(facility_rating)::numeric, 2) as facility_rating
		FROM reviews
		WHERE hotel_id = $1 AND status = 'approved' AND deleted_at IS NULL
	`

	err := r.db.QueryRow(ctx, query, hotelID).Scan(
		&stats.ReviewCount,
		&stats.OverallRating,
		&stats.CleanlinessRating,
		&stats.ServiceRating,
		&stats.LocationRating,
		&stats.ValueRating,
		&stats.FacilityRating,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get hotel stats: %w", err)
	}

	// Get rating distribution
	distQuery := `
		SELECT overall_rating, COUNT(*) as count
		FROM reviews
		WHERE hotel_id = $1 AND status = 'approved' AND deleted_at IS NULL
		GROUP BY overall_rating
		ORDER BY overall_rating
	`

	rows, err := r.db.Query(ctx, distQuery, hotelID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rating, count int
			rows.Scan(&rating, &count)
			stats.RatingDistribution[rating] = count
		}
	}

	return &stats, nil
}

// GetModerationStats retrieves moderation statistics
func (r *repository) GetModerationStats(ctx context.Context) (*ModerationStats, error) {
	var stats ModerationStats

	query := `
		SELECT
			COUNT(*) FILTER (WHERE status = 'pending' AND deleted_at IS NULL) as pending,
			COUNT(*) FILTER (WHERE status = 'flagged' AND deleted_at IS NULL) as flagged,
			COUNT(*) FILTER (WHERE status = 'approved' AND deleted_at IS NULL) as approved,
			COUNT(*) FILTER (WHERE status = 'rejected' AND deleted_at IS NULL) as rejected,
			COUNT(*) FILTER (WHERE deleted_at IS NULL) as total
		FROM reviews
	`

	err := r.db.QueryRow(ctx, query).Scan(
		&stats.PendingReviews,
		&stats.FlaggedReviews,
		&stats.ApprovedReviews,
		&stats.RejectedReviews,
		&stats.TotalReviews,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get moderation stats: %w", err)
	}

	// Calculate approval rate
	if stats.TotalReviews > 0 {
		stats.ApprovalRate = float64(stats.ApprovedReviews) / float64(stats.TotalReviews) * 100
	}

	return &stats, nil
}

// GetReviewAnalytics retrieves review analytics
func (r *repository) GetReviewAnalytics(ctx context.Context) (*ReviewAnalytics, error) {
	var analytics ReviewAnalytics
	analytics.RatingDist = make(map[int]int64)
	analytics.StatusDist = make(map[string]int64)
	analytics.ReviewsByMonth = make(map[string]int64)

	// Get total reviews and average rating
	query := `
		SELECT
			COUNT(*) FILTER (WHERE deleted_at IS NULL) as total,
			ROUND(AVG(overall_rating)::numeric, 2) as avg_rating
		FROM reviews
	`
	err := r.db.QueryRow(ctx, query).Scan(&analytics.TotalReviews, &analytics.AverageRating)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	// Get rating distribution
	rows, err := r.db.Query(ctx, `
		SELECT overall_rating, COUNT(*) as count
		FROM reviews
		WHERE deleted_at IS NULL
		GROUP BY overall_rating
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rating int
			var count int64
			rows.Scan(&rating, &count)
			analytics.RatingDist[rating] = count
		}
	}

	// Get status distribution
	rows, err = r.db.Query(ctx, `
		SELECT status, COUNT(*) as count
		FROM reviews
		WHERE deleted_at IS NULL
		GROUP BY status
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int64
			rows.Scan(&status, &count)
			analytics.StatusDist[status] = count
		}
	}

	// Get reviews by month
	rows, err = r.db.Query(ctx, `
		SELECT TO_CHAR(created_at, 'YYYY-MM') as month, COUNT(*) as count
		FROM reviews
		WHERE deleted_at IS NULL
		GROUP BY TO_CHAR(created_at, 'YYYY-MM')
		ORDER BY month DESC
		LIMIT 12
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var month string
			var count int64
			rows.Scan(&month, &count)
			analytics.ReviewsByMonth[month] = count
		}
	}

	return &analytics, nil
}

// CreateHelpfulVote creates a helpful vote
func (r *repository) CreateHelpfulVote(ctx context.Context, vote *HelpfulVote) error {
	vote.ID = uuid.New().String()
	vote.CreatedAt = time.Now()

	query := `
		INSERT INTO review_helpful_votes (id, review_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query, vote.ID, vote.ReviewID, vote.UserID, vote.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create helpful vote: %w", err)
	}

	return nil
}

// DeleteHelpfulVote deletes a helpful vote
func (r *repository) DeleteHelpfulVote(ctx context.Context, reviewID, userID string) error {
	query := `DELETE FROM review_helpful_votes WHERE review_id = $1 AND user_id = $2`

	_, err := r.db.Exec(ctx, query, reviewID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete helpful vote: %w", err)
	}

	return nil
}

// HasVotedHelpful checks if user has voted helpful
func (r *repository) HasVotedHelpful(ctx context.Context, reviewID, userID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM review_helpful_votes WHERE review_id = $1 AND user_id = $2)`

	err := r.db.QueryRow(ctx, query, reviewID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check helpful vote: %w", err)
	}

	return exists, nil
}

// IncrementHelpfulCount increments helpful count
func (r *repository) IncrementHelpfulCount(ctx context.Context, reviewID string) error {
	query := `UPDATE reviews SET helpful_count = helpful_count + 1 WHERE id = $1`

	_, err := r.db.Exec(ctx, query, reviewID)
	if err != nil {
		return fmt.Errorf("failed to increment helpful count: %w", err)
	}

	return nil
}

// DecrementHelpfulCount decrements helpful count
func (r *repository) DecrementHelpfulCount(ctx context.Context, reviewID string) error {
	query := `UPDATE reviews SET helpful_count = GREATEST(0, helpful_count - 1) WHERE id = $1`

	_, err := r.db.Exec(ctx, query, reviewID)
	if err != nil {
		return fmt.Errorf("failed to decrement helpful count: %w", err)
	}

	return nil
}

// CreateFlag creates a flag
func (r *repository) CreateFlag(ctx context.Context, flag *ReviewFlag) error {
	flag.ID = uuid.New().String()
	flag.CreatedAt = time.Now()

	query := `
		INSERT INTO review_flags (id, review_id, user_id, reason, note, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query, flag.ID, flag.ReviewID, flag.UserID, flag.Reason, flag.Note, flag.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create flag: %w", err)
	}

	return nil
}

// GetFlagsByReviewID retrieves flags for a review
func (r *repository) GetFlagsByReviewID(ctx context.Context, reviewID string) ([]*ReviewFlag, error) {
	query := `
		SELECT
			rf.id, rf.review_id, rf.user_id, rf.reason, rf.note, rf.created_at,
			u.email as user_email, CONCAT(u.first_name, ' ', u.last_name) as user_name
		FROM review_flags rf
		LEFT JOIN users u ON rf.user_id = u.id
		WHERE rf.review_id = $1
		ORDER BY rf.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, reviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flags: %w", err)
	}
	defer rows.Close()

	var flags []*ReviewFlag
	for rows.Next() {
		var flag ReviewFlag

		err := rows.Scan(
			&flag.ID, &flag.ReviewID, &flag.UserID, &flag.Reason, &flag.Note, &flag.CreatedAt,
			&flag.UserEmail, &flag.UserName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flag: %w", err)
		}

		flags = append(flags, &flag)
	}

	return flags, nil
}

// UpdateHotelResponse updates hotel response
func (r *repository) UpdateHotelResponse(ctx context.Context, reviewID string, response string) error {
	query := `
		UPDATE reviews
		SET hotel_response = $2, hotel_response_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, reviewID, response)
	if err != nil {
		return fmt.Errorf("failed to update hotel response: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

// DeleteHotelResponse deletes hotel response
func (r *repository) DeleteHotelResponse(ctx context.Context, reviewID string) error {
	query := `UPDATE reviews SET hotel_response = NULL, hotel_response_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.Exec(ctx, query, reviewID)
	if err != nil {
		return fmt.Errorf("failed to delete hotel response: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

// GetAllModerationKeywords retrieves all moderation keywords
func (r *repository) GetAllModerationKeywords(ctx context.Context) ([]*ModerationKeyword, error) {
	query := `
		SELECT id, keyword, category, is_active, created_at
		FROM moderation_keywords
		WHERE is_active = true
		ORDER BY category, keyword
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get moderation keywords: %w", err)
	}
	defer rows.Close()

	var keywords []*ModerationKeyword
	for rows.Next() {
		var keyword ModerationKeyword
		err := rows.Scan(&keyword.ID, &keyword.Keyword, &keyword.Category, &keyword.IsActive, &keyword.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan keyword: %w", err)
		}
		keywords = append(keywords, &keyword)
	}

	return keywords, nil
}
