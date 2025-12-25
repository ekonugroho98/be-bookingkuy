package review

import (
	"time"
)

// ReviewStatus represents the moderation status of a review
type ReviewStatus string

const (
	StatusPending  ReviewStatus = "pending"
	StatusApproved ReviewStatus = "approved"
	StatusRejected ReviewStatus = "rejected"
	StatusFlagged  ReviewStatus = "flagged"
)

// FlagReason represents the reason for flagging a review
type FlagReason string

const (
	FlagInappropriate FlagReason = "inappropriate"
	FlagFake          FlagReason = "fake"
	FlagSpam          FlagReason = "spam"
	FlagMisleading    FlagReason = "misleading"
	FlagOther         FlagReason = "other"
)

// ModerationKeywordCategory represents the category of moderation keyword
type ModerationKeywordCategory string

const (
	CategoryProfanity     ModerationKeywordCategory = "profanity"
	CategorySpam          ModerationKeywordCategory = "spam"
	CategoryInappropriate ModerationKeywordCategory = "inappropriate"
)

// Review represents a hotel review
type Review struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	HotelID   string `json:"hotel_id"`
	BookingID *string `json:"booking_id,omitempty"`

	// Ratings (1-5 scale)
	OverallRating     int8   `json:"overall_rating"`
	CleanlinessRating *int8  `json:"cleanliness_rating,omitempty"`
	ServiceRating     *int8  `json:"service_rating,omitempty"`
	LocationRating    *int8  `json:"location_rating,omitempty"`
	ValueRating       *int8  `json:"value_rating,omitempty"`
	FacilityRating    *int8  `json:"facility_rating,omitempty"`

	// Review content
	Title   *string  `json:"title,omitempty"`
	Comment string   `json:"comment"`
	Photos  []string `json:"photos,omitempty"` // Array of photo URLs

	// Engagement
	HelpfulCount int `json:"helpful_count"`

	// Moderation
	Status         ReviewStatus `json:"status"`
	ModerationNote *string      `json:"moderation_note,omitempty"`
	ModeratedBy    *string      `json:"moderated_by,omitempty"`
	ModeratedAt    *time.Time   `json:"moderated_at,omitempty"`

	// Hotel response
	HotelResponse   *string    `json:"hotel_response,omitempty"`
	HotelResponseAt *time.Time `json:"hotel_response_at,omitempty"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	// Expanded fields (for API responses)
	UserEmail *string `json:"user_email,omitempty"`
	UserName  *string `json:"user_name,omitempty"`
	HotelName *string `json:"hotel_name,omitempty"`
	IsHelpful *bool   `json:"is_helpful,omitempty"` // For current user
	CanEdit   *bool   `json:"can_edit,omitempty"`   // Whether user can edit this review
}

// ReviewStats represents review statistics for a hotel
type ReviewStats struct {
	HotelID            string             `json:"hotel_id"`
	ReviewCount        int                `json:"review_count"`
	OverallRating      float64            `json:"overall_rating"`
	CleanlinessRating  *float64           `json:"cleanliness_rating,omitempty"`
	ServiceRating      *float64           `json:"service_rating,omitempty"`
	LocationRating     *float64           `json:"location_rating,omitempty"`
	ValueRating        *float64           `json:"value_rating,omitempty"`
	FacilityRating     *float64           `json:"facility_rating,omitempty"`
	RatingDistribution map[int]int        `json:"rating_distribution"` // 1-5 stars
}

// ReviewFlag represents a flag on a review
type ReviewFlag struct {
	ID       string     `json:"id"`
	ReviewID string     `json:"review_id"`
	UserID   *string    `json:"user_id,omitempty"`
	Reason   FlagReason `json:"reason"`
	Note     *string    `json:"note,omitempty"`
	CreatedAt time.Time  `json:"created_at"`

	// Expanded fields
	UserEmail *string `json:"user_email,omitempty"`
	UserName  *string `json:"user_name,omitempty"`
}

// HelpfulVote represents a "helpful" vote on a review
type HelpfulVote struct {
	ID        string    `json:"id"`
	ReviewID  string    `json:"review_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ModerationKeyword represents a word/phrase that triggers moderation
type ModerationKeyword struct {
	ID        int                       `json:"id"`
	Keyword   string                    `json:"keyword"`
	Category  ModerationKeywordCategory `json:"category"`
	IsActive  bool                      `json:"is_active"`
	CreatedAt time.Time                 `json:"created_at"`
}

// CreateReviewRequest represents a request to create a review
type CreateReviewRequest struct {
	BookingID         *string  `json:"booking_id,omitempty"`
	OverallRating     int8     `json:"overall_rating" validate:"required,min=1,max=5"`
	CleanlinessRating *int8    `json:"cleanliness_rating,omitempty" validate:"omitempty,min=1,max=5"`
	ServiceRating     *int8    `json:"service_rating,omitempty" validate:"omitempty,min=1,max=5"`
	LocationRating    *int8    `json:"location_rating,omitempty" validate:"omitempty,min=1,max=5"`
	ValueRating       *int8    `json:"value_rating,omitempty" validate:"omitempty,min=1,max=5"`
	FacilityRating    *int8    `json:"facility_rating,omitempty" validate:"omitempty,min=1,max=5"`
	Title             *string  `json:"title,omitempty" validate:"omitempty,max=255"`
	Comment           string   `json:"comment" validate:"required,min=50,max=2000"`
	Photos            []string `json:"photos,omitempty" validate:"omitempty,max=5"`
}

// UpdateReviewRequest represents a request to update a review
type UpdateReviewRequest struct {
	OverallRating     *int8    `json:"overall_rating,omitempty" validate:"omitempty,min=1,max=5"`
	CleanlinessRating *int8    `json:"cleanliness_rating,omitempty" validate:"omitempty,min=1,max=5"`
	ServiceRating     *int8    `json:"service_rating,omitempty" validate:"omitempty,min=1,max=5"`
	LocationRating    *int8    `json:"location_rating,omitempty" validate:"omitempty,min=1,max=5"`
	ValueRating       *int8    `json:"value_rating,omitempty" validate:"omitempty,min=1,max=5"`
	FacilityRating    *int8    `json:"facility_rating,omitempty" validate:"omitempty,min=1,max=5"`
	Title             *string  `json:"title,omitempty" validate:"omitempty,max=255"`
	Comment           *string  `json:"comment,omitempty" validate:"omitempty,min=50,max=2000"`
	Photos            []string `json:"photos,omitempty" validate:"omitempty,max=5"`
}

// ModerateReviewRequest represents a request to moderate a review
type ModerateReviewRequest struct {
	Action         string  `json:"action" validate:"required,oneof=approve reject flag"`
	ModerationNote *string `json:"moderation_note,omitempty" validate:"omitempty,max=1000"`
}

// AddHotelResponseRequest represents a request to add hotel response
type AddHotelResponseRequest struct {
	Response string `json:"response" validate:"required,min=10,max=2000"`
}

// FlagReviewRequest represents a request to flag a review
type FlagReviewRequest struct {
	Reason FlagReason `json:"reason" validate:"required"`
	Note   *string    `json:"note,omitempty" validate:"omitempty,max=500"`
}

// ReviewListResponse represents a paginated list of reviews
type ReviewListResponse struct {
	Reviews       []*Review `json:"reviews"`
	TotalCount    int       `json:"total_count"`
	Page          int       `json:"page"`
	PageSize      int       `json:"page_size"`
	HasNext       bool      `json:"has_next"`
	AverageRating float64   `json:"average_rating"`
}

// CreateReviewResponse represents the response after creating a review
type CreateReviewResponse struct {
	Review     *Review  `json:"review"`
	Status     string   `json:"status"` // pending, approved, auto_flagged
	Message    string   `json:"message,omitempty"`
	Flagged    bool     `json:"flagged,omitempty"`
	FlagReason *string  `json:"flag_reason,omitempty"`
}

// ModerationStats represents moderation statistics
type ModerationStats struct {
	PendingReviews  int     `json:"pending_reviews"`
	FlaggedReviews  int     `json:"flagged_reviews"`
	ApprovedReviews int     `json:"approved_reviews"`
	RejectedReviews int     `json:"rejected_reviews"`
	TotalReviews    int     `json:"total_reviews"`
	ApprovalRate    float64 `json:"approval_rate"`
}

// ReviewAnalytics represents review analytics
type ReviewAnalytics struct {
	TotalReviews   int64                 `json:"total_reviews"`
	AverageRating  float64               `json:"average_rating"`
	RatingDist     map[int]int64         `json:"rating_distribution"` // 1-5 stars
	StatusDist     map[string]int64      `json:"status_distribution"`
	ReviewsByMonth map[string]int64      `json:"reviews_by_month"`
	ReviewsByHotel []*HotelReviewStats   `json:"top_hotels,omitempty"`
}

// HotelReviewStats represents review stats for a single hotel
type HotelReviewStats struct {
	HotelID       string  `json:"hotel_id"`
	HotelName     string  `json:"hotel_name"`
	ReviewCount   int     `json:"review_count"`
	AverageRating float64 `json:"average_rating"`
}

// Helper functions

// IsEditable checks if a review can be edited by a user
func (r *Review) IsEditable(userID string) bool {
	// User can edit their own review within 7 days of creation
	if r.UserID != userID {
		return false
	}

	// Check if review is deleted
	if r.DeletedAt != nil {
		return false
	}

	// Check if within 7 days
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	return r.CreatedAt.After(sevenDaysAgo)
}

// IsApproved checks if review is approved
func (r *Review) IsApproved() bool {
	return r.Status == StatusApproved
}

// IsPending checks if review is pending moderation
func (r *Review) IsPending() bool {
	return r.Status == StatusPending
}

// IsFlagged checks if review is flagged
func (r *Review) IsFlagged() bool {
	return r.Status == StatusFlagged
}

// RequiresModeration checks if review contains prohibited content
func (r *Review) RequiresModeration() bool {
	return r.Status == StatusPending || r.Status == StatusFlagged
}

// GetAverageCategoryRating calculates the average of all category ratings
func (r *Review) GetAverageCategoryRating() float64 {
	if r.CleanlinessRating == nil && r.ServiceRating == nil &&
		r.LocationRating == nil && r.ValueRating == nil && r.FacilityRating == nil {
		return float64(r.OverallRating)
	}

	count := 0
	sum := 0

	if r.CleanlinessRating != nil {
		sum += int(*r.CleanlinessRating)
		count++
	}
	if r.ServiceRating != nil {
		sum += int(*r.ServiceRating)
		count++
	}
	if r.LocationRating != nil {
		sum += int(*r.LocationRating)
		count++
	}
	if r.ValueRating != nil {
		sum += int(*r.ValueRating)
		count++
	}
	if r.FacilityRating != nil {
		sum += int(*r.FacilityRating)
		count++
	}

	if count == 0 {
		return float64(r.OverallRating)
	}

	return float64(sum) / float64(count)
}
