ALTER TABLE users ADD COLUMN IF NOT EXISTS device_id VARCHAR(255) DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_ip VARCHAR(45) DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_users_device_id ON users(device_id) WHERE device_id != '';
CREATE INDEX IF NOT EXISTS idx_users_last_ip ON users(last_ip) WHERE last_ip != '';
