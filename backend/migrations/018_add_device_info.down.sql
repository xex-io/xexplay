DROP INDEX IF EXISTS idx_users_last_ip;
DROP INDEX IF EXISTS idx_users_device_id;

ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;
ALTER TABLE users DROP COLUMN IF EXISTS last_ip;
ALTER TABLE users DROP COLUMN IF EXISTS device_id;
