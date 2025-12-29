-- Remove unique constraint from hotels table
-- Migration: 000010

ALTER TABLE hotels
DROP CONSTRAINT IF EXISTS hotels_provider_unique;
