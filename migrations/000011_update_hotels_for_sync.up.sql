-- Update hotels table to support multi-provider sync
-- Migration: 000011

-- Add missing columns for sync support
ALTER TABLE hotels
ADD COLUMN IF NOT EXISTS provider_code VARCHAR(100) DEFAULT 'hotelbeds',
ADD COLUMN IF NOT EXISTS provider_hotel_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS description TEXT,
ADD COLUMN IF NOT EXISTS star_rating DECIMAL(3,1),
ADD COLUMN IF NOT EXISTS country_code VARCHAR(2),
ADD COLUMN IF NOT EXISTS address TEXT,
ADD COLUMN IF NOT EXISTS postal_code VARCHAR(20),
ADD COLUMN IF NOT EXISTS images JSONB,
ADD COLUMN IF NOT EXISTS amenities JSONB,
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- Populate provider_hotel_id from existing hotelbeds_id
UPDATE hotels
SET provider_hotel_id = hotelbeds_id,
    provider_code = 'hotelbeds'
WHERE provider_hotel_id IS NULL AND hotelbeds_id IS NOT NULL;

-- Add unique constraint on provider_code + provider_hotel_id
ALTER TABLE hotels
ADD CONSTRAINT hotels_provider_unique UNIQUE (provider_code, provider_hotel_id);

-- Create trigger for updated_at
CREATE TRIGGER update_hotels_updated_at BEFORE UPDATE ON hotels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add indexes for sync queries
CREATE INDEX IF NOT EXISTS idx_hotels_provider_hotel_id ON hotels(provider_hotel_id);
CREATE INDEX IF NOT EXISTS idx_hotels_country_code ON hotels(country_code);
CREATE INDEX IF NOT EXISTS idx_hotels_deleted_at ON hotels(deleted_at) WHERE deleted_at IS NOT NULL;
