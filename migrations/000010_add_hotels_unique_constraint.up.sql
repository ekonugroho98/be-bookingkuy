-- Add unique constraint on hotels table for provider_code + provider_hotel_id
-- This allows us to use UPSERT for hotel sync
-- Migration: 000010

-- Add unique constraint
ALTER TABLE hotels
ADD CONSTRAINT hotels_provider_unique UNIQUE (provider_code, provider_hotel_id);
