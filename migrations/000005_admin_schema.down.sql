-- Drop trigger
DROP TRIGGER IF EXISTS update_admin_users_updated_at ON admin_users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_admin_users_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_admin_audit_log_created_at;
DROP INDEX IF EXISTS idx_admin_audit_log_entity;
DROP INDEX IF EXISTS idx_admin_audit_log_action;
DROP INDEX IF EXISTS idx_admin_audit_log_admin_id;
DROP INDEX IF EXISTS idx_admin_users_created_at;
DROP INDEX IF EXISTS idx_admin_users_is_active;
DROP INDEX IF EXISTS idx_admin_users_role;
DROP INDEX IF EXISTS idx_admin_users_email;

-- Drop tables
DROP TABLE IF EXISTS admin_audit_log;
DROP TABLE IF EXISTS admin_users;
