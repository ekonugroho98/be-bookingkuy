-- Migration: Review Schema
-- Version: 006
-- Description: Add tables for hotel review system with ratings, moderation, and user feedback

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create reviews table
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hotel_id UUID NOT NULL REFERENCES hotels(id) ON DELETE CASCADE,
    booking_id UUID REFERENCES bookings(id) ON DELETE SET NULL,

    -- Ratings (1-5 scale)
    overall_rating INTEGER NOT NULL CHECK (overall_rating BETWEEN 1 AND 5),
    cleanliness_rating INTEGER CHECK (cleanliness_rating BETWEEN 1 AND 5),
    service_rating INTEGER CHECK (service_rating BETWEEN 1 AND 5),
    location_rating INTEGER CHECK (location_rating BETWEEN 1 AND 5),
    value_rating INTEGER CHECK (value_rating BETWEEN 1 AND 5),
    facility_rating INTEGER CHECK (facility_rating BETWEEN 1 AND 5),

    -- Review content
    title VARCHAR(255),
    comment TEXT NOT NULL,
    photos TEXT[], -- Array of photo URLs (S3/Cloudinary)

    -- Engagement
    helpful_count INTEGER DEFAULT 0,

    -- Moderation
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'flagged')),
    moderation_note TEXT,
    moderated_by UUID REFERENCES admin_users(id),
    moderated_at TIMESTAMP WITH TIME ZONE,

    -- Hotel response
    hotel_response TEXT,
    hotel_response_at TIMESTAMP WITH TIME ZONE,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for reviews
CREATE INDEX idx_reviews_hotel_id ON reviews(hotel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_reviews_user_id ON reviews(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_reviews_booking_id ON reviews(booking_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_reviews_status ON reviews(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_reviews_created_at ON reviews(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_reviews_overall_rating ON reviews(overall_rating) WHERE deleted_at IS NULL;

-- Ensure one review per booking (unless soft-deleted)
CREATE UNIQUE INDEX idx_reviews_booking_unique ON reviews(booking_id) WHERE deleted_at IS NULL AND booking_id IS NOT NULL;

-- Create review_flags table
CREATE TABLE IF NOT EXISTS review_flags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    review_id UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reason VARCHAR(100) CHECK (reason IN ('inappropriate', 'fake', 'spam', 'misleading', 'other')),
    note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for review_flags
CREATE INDEX idx_review_flags_review_id ON review_flags(review_id);
CREATE INDEX idx_review_flags_user_id ON review_flags(user_id);

-- Create review_helpful_votes table
CREATE TABLE IF NOT EXISTS review_helpful_votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    review_id UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(review_id, user_id)
);

-- Create indexes for review_helpful_votes
CREATE INDEX idx_review_helpful_votes_review_id ON review_helpful_votes(review_id);
CREATE INDEX idx_review_helpful_votes_user_id ON review_helpful_votes(user_id);

-- Add review statistics columns to hotels table
ALTER TABLE hotels ADD COLUMN IF NOT EXISTS review_count INTEGER DEFAULT 0;
ALTER TABLE hotels ADD COLUMN IF NOT EXISTS overall_rating DECIMAL(3,2);
ALTER TABLE hotels ADD COLUMN IF NOT EXISTS cleanliness_rating DECIMAL(3,2);
ALTER TABLE hotels ADD COLUMN IF NOT EXISTS service_rating DECIMAL(3,2);
ALTER TABLE hotels ADD COLUMN IF NOT EXISTS location_rating DECIMAL(3,2);
ALTER TABLE hotels ADD COLUMN IF NOT EXISTS value_rating DECIMAL(3,2);
ALTER TABLE hotels ADD COLUMN IF NOT EXISTS facility_rating DECIMAL(3,2);

-- Create index for sorting by rating
CREATE INDEX idx_hotels_overall_rating ON hotels(overall_rating DESC) WHERE overall_rating IS NOT NULL;

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_reviews_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_reviews_updated_at
BEFORE UPDATE ON reviews
FOR EACH ROW
EXECUTE FUNCTION update_reviews_updated_at();

-- Create function to update hotel review statistics
CREATE OR REPLACE FUNCTION update_hotel_review_stats()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE hotels
    SET
        review_count = (
            SELECT COUNT(*)
            FROM reviews
            WHERE reviews.hotel_id = NEW.hotel_id
            AND reviews.status = 'approved'
            AND reviews.deleted_at IS NULL
        ),
        overall_rating = (
            SELECT ROUND(AVG(overall_rating)::numeric, 2)
            FROM reviews
            WHERE reviews.hotel_id = NEW.hotel_id
            AND reviews.status = 'approved'
            AND reviews.deleted_at IS NULL
        ),
        cleanliness_rating = (
            SELECT ROUND(AVG(cleanliness_rating)::numeric, 2)
            FROM reviews
            WHERE reviews.hotel_id = NEW.hotel_id
            AND reviews.status = 'approved'
            AND reviews.deleted_at IS NULL
            AND cleanliness_rating IS NOT NULL
        ),
        service_rating = (
            SELECT ROUND(AVG(service_rating)::numeric, 2)
            FROM reviews
            WHERE reviews.hotel_id = NEW.hotel_id
            AND reviews.status = 'approved'
            AND reviews.deleted_at IS NULL
            AND service_rating IS NOT NULL
        ),
        location_rating = (
            SELECT ROUND(AVG(location_rating)::numeric, 2)
            FROM reviews
            WHERE reviews.hotel_id = NEW.hotel_id
            AND reviews.status = 'approved'
            AND reviews.deleted_at IS NULL
            AND location_rating IS NOT NULL
        ),
        value_rating = (
            SELECT ROUND(AVG(value_rating)::numeric, 2)
            FROM reviews
            WHERE reviews.hotel_id = NEW.hotel_id
            AND reviews.status = 'approved'
            AND reviews.deleted_at IS NULL
            AND value_rating IS NOT NULL
        ),
        facility_rating = (
            SELECT ROUND(AVG(facility_rating)::numeric, 2)
            FROM reviews
            WHERE reviews.hotel_id = NEW.hotel_id
            AND reviews.status = 'approved'
            AND reviews.deleted_at IS NULL
            AND facility_rating IS NOT NULL
        )
    WHERE id = NEW.hotel_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to auto-update hotel review stats when review is approved
CREATE TRIGGER trigger_update_hotel_review_stats
AFTER INSERT OR UPDATE ON reviews
FOR EACH ROW
WHEN (NEW.status = 'approved' AND (TG_OP = 'INSERT' OR OLD.status IS DISTINCT FROM NEW.status))
EXECUTE FUNCTION update_hotel_review_stats();

-- Create trigger to update hotel review stats when review is deleted
CREATE TRIGGER trigger_update_hotel_review_stats_delete
AFTER UPDATE ON reviews
FOR EACH ROW
WHEN (NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL AND OLD.status = 'approved')
EXECUTE FUNCTION update_hotel_review_stats();

-- Insert sample prohibited words for moderation (can be extended)
CREATE TABLE IF NOT EXISTS moderation_keywords (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(50) NOT NULL, -- profanity, spam, inappropriate
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert some basic moderation keywords
INSERT INTO moderation_keywords (keyword, category) VALUES
('shit', 'profanity'),
('fuck', 'profanity'),
('damn', 'profanity'),
('ass', 'profanity'),
('bastard', 'profanity'),
('bitch', 'profanity')
ON CONFLICT (keyword) DO NOTHING;
