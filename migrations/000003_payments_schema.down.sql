-- Rollback payments schema
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP INDEX IF EXISTS idx_payments_deleted_at;
DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_provider_reference;
DROP INDEX IF EXISTS idx_payments_booking_id;
DROP TABLE IF EXISTS payments;
