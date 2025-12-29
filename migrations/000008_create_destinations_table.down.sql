-- Drop destinations table
DROP TRIGGER IF EXISTS update_destinations_updated_at ON destinations;
DROP INDEX IF EXISTS idx_destinations_name_gin;
DROP INDEX IF EXISTS idx_destinations_type;
DROP INDEX IF EXISTS idx_destinations_country_code;
DROP TABLE IF EXISTS destinations;
