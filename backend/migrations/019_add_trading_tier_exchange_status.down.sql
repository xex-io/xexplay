DROP INDEX IF EXISTS idx_users_exchange_status;
DROP INDEX IF EXISTS idx_users_trading_tier;
ALTER TABLE users DROP COLUMN IF EXISTS exchange_status;
ALTER TABLE users DROP COLUMN IF EXISTS trading_tier;
