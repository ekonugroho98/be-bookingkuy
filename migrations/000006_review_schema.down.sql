-- Rollback: Review Schema
-- Version: 006

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_reviews_updated_at ON reviews;
DROP TRIGGER IF EXISTS trigger_update_hotel_review_stats ON reviews;
DROP TRIGGER IF EXISTS trigger_update_hotel_review_stats_delete ON reviews;

-- Drop functions
DROP FUNCTION IF EXISTS update_reviews_updated_at();
DROP FUNCTION IF EXISTS update_hotel_review_stats();

-- Drop tables
DROP TABLE IF EXISTS review_helpful_votes CASCADE;
DROP TABLE IF EXISTS review_flags CASCADE;
DROP TABLE IF EXISTS reviews CASCADE;
DROP TABLE IF EXISTS moderation_keywords CASCADE;

-- Remove columns from hotels table
ALTER TABLE hotels DROP COLUMN IF EXISTS review_count;
ALTER TABLE hotels DROP COLUMN IF EXISTS overall_rating;
ALTER TABLE hotels DROP COLUMN IF EXISTS cleanliness_rating;
ALTER TABLE hotels DROP COLUMN IF EXISTS service_rating;
ALTER TABLE hotels DROP COLUMN IF EXISTS location_rating;
ALTER TABLE hotels DROP COLUMN IF EXISTS value_rating;
ALTER TABLE hotels DROP COLUMN IF EXISTS facility_rating;

-- Drop index for hotels
DROP INDEX IF EXISTS idx_hotels_overall_rating;
