package review

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Service defines the business logic interface for review service
type Service interface {
	// Review CRUD
	CreateReview(ctx context.Context, userID string, req *CreateReviewRequest) (*CreateReviewResponse, error)
	GetReviewByID(ctx context.Context, reviewID, userID string) (*Review, error)
	GetReviewsByHotelID(ctx context.Context, hotelID string, status ReviewStatus, page, pageSize int, sortBy string) (*ReviewListResponse, error)
	GetMyReviews(ctx context.Context, userID string, page, pageSize int) (*ReviewListResponse, error)
	UpdateReview(ctx context.Context, reviewID, userID string, req *UpdateReviewRequest) (*Review, error)
	DeleteReview(ctx context.Context, reviewID, userID string) error

	// Review moderation
	GetPendingReviews(ctx context.Context, page, pageSize int) (*ReviewListResponse, error)
	GetFlaggedReviews(ctx context.Context, page, pageSize int) (*ReviewListResponse, error)
	ModerateReview(ctx context.Context, adminID, reviewID string, req *ModerateReviewRequest) error

	// Helpful votes
	ToggleHelpful(ctx context.Context, reviewID, userID string) (bool, error)

	// Flags
	FlagReview(ctx context.Context, reviewID, userID string, req *FlagReviewRequest) error

	// Hotel responses
	AddHotelResponse(ctx context.Context, reviewID string, req *AddHotelResponseRequest) error
	UpdateHotelResponse(ctx context.Context, reviewID string, req *AddHotelResponseRequest) error
	DeleteHotelResponse(ctx context.Context, reviewID string) error

	// Statistics
	GetHotelStats(ctx context.Context, hotelID string) (*ReviewStats, error)
	GetModerationStats(ctx context.Context) (*ModerationStats, error)
	GetAnalytics(ctx context.Context) (*ReviewAnalytics, error)
}

type service struct {
	repo Repository
}

// NewService creates a new review service
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateReview creates a new review
func (s *service) CreateReview(ctx context.Context, userID string, req *CreateReviewRequest) (*CreateReviewResponse, error) {
	// Validate rating
	if req.OverallRating < 1 || req.OverallRating > 5 {
		return nil, errors.New("overall rating must be between 1 and 5")
	}

	// Validate comment length
	commentLen := len(strings.TrimSpace(req.Comment))
	if commentLen < 50 {
		return nil, errors.New("comment must be at least 50 characters")
	}
	if commentLen > 2000 {
		return nil, errors.New("comment must not exceed 2000 characters")
	}

	// Validate photos
	if len(req.Photos) > 5 {
		return nil, errors.New("maximum 5 photos allowed")
	}

	// Check for moderation keywords
	keywords, _ := s.repo.GetAllModerationKeywords(ctx)
	flagged, flagReason := s.checkForProhibitedContent(req.Comment, keywords)

	// Create review
	review := &Review{
		UserID:            userID,
		HotelID:           "hotel-id-from-booking", // TODO: Get from booking
		BookingID:         req.BookingID,
		OverallRating:     req.OverallRating,
		CleanlinessRating: req.CleanlinessRating,
		ServiceRating:     req.ServiceRating,
		LocationRating:    req.LocationRating,
		ValueRating:       req.ValueRating,
		FacilityRating:    req.FacilityRating,
		Title:             req.Title,
		Comment:           req.Comment,
		Photos:            req.Photos,
	}

	// Auto-flag if prohibited content found
	if flagged {
		review.Status = StatusFlagged
	}

	err := s.repo.CreateReview(ctx, review)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to create review")
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	response := &CreateReviewResponse{
		Review:  review,
		Status:  "pending",
		Message: "Your review has been submitted and is awaiting moderation",
	}

	if flagged {
		response.Status = "flagged"
		response.Flagged = true
		response.FlagReason = &flagReason
		response.Message = "Your review has been flagged for moderation"
	}

	logger.Infof("Review created: %s for hotel %s by user %s", review.ID, review.HotelID, userID)

	return response, nil
}

// GetReviewByID retrieves a review by ID
func (s *service) GetReviewByID(ctx context.Context, reviewID, userID string) (*Review, error) {
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	// Check if user has voted helpful
	if userID != "" {
		hasVoted, _ := s.repo.HasVotedHelpful(ctx, reviewID, userID)
		review.IsHelpful = &hasVoted
	}

	// Check if user can edit
	if userID != "" {
		canEdit := review.IsEditable(userID)
		review.CanEdit = &canEdit
	}

	return review, nil
}

// GetReviewsByHotelID retrieves reviews for a hotel
func (s *service) GetReviewsByHotelID(ctx context.Context, hotelID string, status ReviewStatus, page, pageSize int, sortBy string) (*ReviewListResponse, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// Get approved reviews only if no status specified
	if status == "" {
		status = StatusApproved
	}

	reviews, total, err := s.repo.GetReviewsByHotelID(ctx, hotelID, status, pageSize, offset, sortBy)
	if err != nil {
		return nil, err
	}

	// Calculate average rating
	var avgRating float64
	if len(reviews) > 0 {
		sum := 0
		for _, r := range reviews {
			sum += int(r.OverallRating)
		}
		avgRating = float64(sum) / float64(len(reviews))
	}

	return &ReviewListResponse{
		Reviews:       reviews,
		TotalCount:    total,
		Page:          page,
		PageSize:      pageSize,
		HasNext:       offset+pageSize < total,
		AverageRating: avgRating,
	}, nil
}

// GetMyReviews retrieves reviews by current user
func (s *service) GetMyReviews(ctx context.Context, userID string, page, pageSize int) (*ReviewListResponse, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	reviews, total, err := s.repo.GetReviewsByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &ReviewListResponse{
		Reviews:    reviews,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		HasNext:    offset+pageSize < total,
	}, nil
}

// UpdateReview updates a review
func (s *service) UpdateReview(ctx context.Context, reviewID, userID string, req *UpdateReviewRequest) (*Review, error) {
	// Get existing review
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	// Check if user can edit this review
	if !review.IsEditable(userID) {
		return nil, errors.New("you can only edit your own reviews within 7 days of creation")
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.OverallRating != nil {
		if *req.OverallRating < 1 || *req.OverallRating > 5 {
			return nil, errors.New("overall rating must be between 1 and 5")
		}
		updates["overall_rating"] = *req.OverallRating
	}

	if req.CleanlinessRating != nil {
		updates["cleanliness_rating"] = *req.CleanlinessRating
	}

	if req.ServiceRating != nil {
		updates["service_rating"] = *req.ServiceRating
	}

	if req.LocationRating != nil {
		updates["location_rating"] = *req.LocationRating
	}

	if req.ValueRating != nil {
		updates["value_rating"] = *req.ValueRating
	}

	if req.FacilityRating != nil {
		updates["facility_rating"] = *req.FacilityRating
	}

	if req.Title != nil {
		updates["title"] = *req.Title
	}

	if req.Comment != nil {
		commentLen := len(strings.TrimSpace(*req.Comment))
		if commentLen < 50 {
			return nil, errors.New("comment must be at least 50 characters")
		}
		if commentLen > 2000 {
			return nil, errors.New("comment must not exceed 2000 characters")
		}
		updates["comment"] = *req.Comment
	}

	if req.Photos != nil {
		if len(req.Photos) > 5 {
			return nil, errors.New("maximum 5 photos allowed")
		}
		updates["photos"] = req.Photos
	}

	// Reset to pending for re-moderation
	updates["status"] = StatusPending

	// Update review
	err = s.repo.UpdateReview(ctx, reviewID, updates)
	if err != nil {
		return nil, err
	}

	// Get updated review
	updatedReview, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	logger.Infof("Review updated: %s by user %s", reviewID, userID)

	return updatedReview, nil
}

// DeleteReview deletes a review
func (s *service) DeleteReview(ctx context.Context, reviewID, userID string) error {
	// Get review
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return err
	}

	// Check ownership
	if review.UserID != userID {
		return errors.New("you can only delete your own reviews")
	}

	// Delete review
	err = s.repo.DeleteReview(ctx, reviewID)
	if err != nil {
		return err
	}

	logger.Infof("Review deleted: %s by user %s", reviewID, userID)

	return nil
}

// GetPendingReviews retrieves pending reviews for moderation
func (s *service) GetPendingReviews(ctx context.Context, page, pageSize int) (*ReviewListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	reviews, total, err := s.repo.GetPendingReviews(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &ReviewListResponse{
		Reviews:    reviews,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		HasNext:    offset+pageSize < total,
	}, nil
}

// GetFlaggedReviews retrieves flagged reviews
func (s *service) GetFlaggedReviews(ctx context.Context, page, pageSize int) (*ReviewListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	reviews, total, err := s.repo.GetFlaggedReviews(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &ReviewListResponse{
		Reviews:    reviews,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		HasNext:    offset+pageSize < total,
	}, nil
}

// ModerateReview moderates a review
func (s *service) ModerateReview(ctx context.Context, adminID, reviewID string, req *ModerateReviewRequest) error {
	// Validate action
	if req.Action != "approve" && req.Action != "reject" && req.Action != "flag" {
		return errors.New("invalid action. must be approve, reject, or flag")
	}

	// Get review
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return err
	}

	// Check if already moderated
	if review.Status != StatusPending && review.Status != StatusFlagged {
		return fmt.Errorf("review has already been %s", review.Status)
	}

	// Perform action
	switch req.Action {
	case "approve":
		err = s.repo.ApproveReview(ctx, reviewID, adminID, req.ModerationNote)
	case "reject":
		err = s.repo.RejectReview(ctx, reviewID, adminID, req.ModerationNote)
	case "flag":
		err = s.repo.FlagReview(ctx, reviewID)
	}

	if err != nil {
		return err
	}

	logger.Infof("Review %s: %s by admin %s", req.Action, reviewID, adminID)

	return nil
}

// ToggleHelpful toggles helpful vote
func (s *service) ToggleHelpful(ctx context.Context, reviewID, userID string) (bool, error) {
	// Check if review exists
	_, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return false, err
	}

	// Check if already voted
	hasVoted, err := s.repo.HasVotedHelpful(ctx, reviewID, userID)
	if err != nil {
		return false, err
	}

	if hasVoted {
		// Remove vote
		err = s.repo.DeleteHelpfulVote(ctx, reviewID, userID)
		if err != nil {
			return false, err
		}

		err = s.repo.DecrementHelpfulCount(ctx, reviewID)
		if err != nil {
			return false, err
		}

		logger.Infof("Helpful vote removed: review %s by user %s", reviewID, userID)
		return false, nil
	}

	// Add vote
	vote := &HelpfulVote{
		ReviewID: reviewID,
		UserID:   userID,
	}

	err = s.repo.CreateHelpfulVote(ctx, vote)
	if err != nil {
		return false, err
	}

	err = s.repo.IncrementHelpfulCount(ctx, reviewID)
	if err != nil {
		return false, err
	}

	logger.Infof("Helpful vote added: review %s by user %s", reviewID, userID)
	return true, nil
}

// FlagReview flags a review
func (s *service) FlagReview(ctx context.Context, reviewID, userID string, req *FlagReviewRequest) error {
	// Check if review exists
	review, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return err
	}

	// Check if user is flagging their own review
	if review.UserID == userID {
		return errors.New("you cannot flag your own review")
	}

	// Create flag
	flag := &ReviewFlag{
		ReviewID: reviewID,
		UserID:   &userID,
		Reason:   req.Reason,
		Note:     req.Note,
	}

	err = s.repo.CreateFlag(ctx, flag)
	if err != nil {
		return err
	}

	// Flag the review if not already flagged
	if review.Status != StatusFlagged {
		err = s.repo.FlagReview(ctx, reviewID)
		if err != nil {
			logger.ErrorWithErr(err, "Failed to flag review")
		}
	}

	logger.Infof("Review flagged: %s by user %s for reason: %s", reviewID, userID, req.Reason)

	return nil
}

// AddHotelResponse adds a hotel response to a review
func (s *service) AddHotelResponse(ctx context.Context, reviewID string, req *AddHotelResponseRequest) error {
	// Validate response length
	responseLen := len(strings.TrimSpace(req.Response))
	if responseLen < 10 {
		return errors.New("response must be at least 10 characters")
	}
	if responseLen > 2000 {
		return errors.New("response must not exceed 2000 characters")
	}

	// Check if review exists
	_, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return err
	}

	// Add response
	err = s.repo.UpdateHotelResponse(ctx, reviewID, req.Response)
	if err != nil {
		return err
	}

	logger.Infof("Hotel response added to review: %s", reviewID)

	return nil
}

// UpdateHotelResponse updates a hotel response
func (s *service) UpdateHotelResponse(ctx context.Context, reviewID string, req *AddHotelResponseRequest) error {
	// Validate response length
	responseLen := len(strings.TrimSpace(req.Response))
	if responseLen < 10 {
		return errors.New("response must be at least 10 characters")
	}
	if responseLen > 2000 {
		return errors.New("response must not exceed 2000 characters")
	}

	// Check if review exists
	_, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return err
	}

	// Update response
	err = s.repo.UpdateHotelResponse(ctx, reviewID, req.Response)
	if err != nil {
		return err
	}

	logger.Infof("Hotel response updated for review: %s", reviewID)

	return nil
}

// DeleteHotelResponse deletes a hotel response
func (s *service) DeleteHotelResponse(ctx context.Context, reviewID string) error {
	// Check if review exists
	_, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return err
	}

	// Delete response
	err = s.repo.DeleteHotelResponse(ctx, reviewID)
	if err != nil {
		return err
	}

	logger.Infof("Hotel response deleted for review: %s", reviewID)

	return nil
}

// GetHotelStats retrieves review statistics for a hotel
func (s *service) GetHotelStats(ctx context.Context, hotelID string) (*ReviewStats, error) {
	stats, err := s.repo.GetHotelReviewStats(ctx, hotelID)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetModerationStats retrieves moderation statistics
func (s *service) GetModerationStats(ctx context.Context) (*ModerationStats, error) {
	stats, err := s.repo.GetModerationStats(ctx)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetAnalytics retrieves review analytics
func (s *service) GetAnalytics(ctx context.Context) (*ReviewAnalytics, error) {
	analytics, err := s.repo.GetReviewAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	return analytics, nil
}

// Helper functions

// checkForProhibitedContent checks if text contains prohibited keywords
func (s *service) checkForProhibitedContent(text string, keywords []*ModerationKeyword) (bool, string) {
	if len(keywords) == 0 {
		return false, ""
	}

	// Convert to lowercase for comparison
	textLower := strings.ToLower(text)

	for _, keyword := range keywords {
		if strings.Contains(textLower, strings.ToLower(keyword.Keyword)) {
			return true, fmt.Sprintf("Contains prohibited content: %s", keyword.Category)
		}
	}

	return false, ""
}

// sanitizeComment removes excessive whitespace and special characters
func sanitizeComment(comment string) string {
	// Replace multiple spaces with single space
	words := strings.Fields(comment)
	return strings.Join(words, " ")
}

// containsProfanity checks if text contains common profanity patterns
func containsProfanity(text string) bool {
	// Simple check for repeated characters (common in attempts to bypass filters)
	textLower := strings.ToLower(text)

	// Check for 3+ consecutive same characters
	count := 1
	for i := 1; i < len(textLower); i++ {
		if textLower[i] == textLower[i-1] {
			count++
			if count >= 3 {
				return true
			}
		} else {
			count = 1
		}
	}

	// Check for excessive special characters
	specialCharCount := 0
	for _, r := range text {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && !unicode.IsSpace(r) {
			specialCharCount++
		}
	}

	// If more than 30% special characters, likely spam
	if len(text) > 0 && float64(specialCharCount)/float64(len(text)) > 0.3 {
		return true
	}

	return false
}

// validateRating validates a rating value
func validateRating(rating *int8) error {
	if rating == nil {
		return nil
	}

	if *rating < 1 || *rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}

	return nil
}

// canUserEditReview checks if a user can edit a review
func canUserEditReview(review *Review, userID string) bool {
	// Must be own review
	if review.UserID != userID {
		return false
	}

	// Must not be deleted
	if review.DeletedAt != nil {
		return false
	}

	// Must be within 7 days
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	return review.CreatedAt.After(sevenDaysAgo)
}
