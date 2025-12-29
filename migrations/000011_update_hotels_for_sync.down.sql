-- Rollback hotels table updates
-- Migration: 000011

-- Drop indexes
DROP INDEX IF EXISTS idx_hotels_deleted_at;
DROP INDEX IF EXISTS idx_hotels_country_code;
DROP INDEX IF EXISTS idx_hotels_provider_hotel_id;

-- Drop trigger
DROP TRIGGER IF EXISTS update_hotels_updated_at ON hotels;

-- Drop unique constraint
ALTER TABLE hotels DROP CONSTRAINT IF EXISTS hotels_provider_unique;

-- Drop columns (this will fail if there's data, which is good - prevents accidental data loss)
-- ALTER TABLE hotels DROP COLUMN IF EXISTS deleted_at;
-- ALTER TABLE hotels DROP COLUMN IF EXISTS updated_at;
-- ALTER TABLE hotels DROP COLUMN IF EXISTS amenities;
-- ALTER TABLE hotels DROP COLUMN IF EXISTS images;
-- ALTER TABLE hotels DROP COLUMN IF EXISTS postal_code;
-- ALTER TABLE hotels DROP COLUMN IF EXISTS address;
-- ALTER TABLE hotels DROP COLUMN IF EXISTS country_code;
-- ALTER TABLE hotels DROP COLUMN IF EXISTS star_rating;
-- ALTER TABLE hotels DROP COLUMN IF EXISTS description;
-- ALTER TABLE hotels DROP COLUMN IF EXISTS provider_hotel_id;
-- ALTER TABLE hotels DROP COLUMN IF EXISTS provider_code;
