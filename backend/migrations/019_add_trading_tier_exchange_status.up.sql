-- Add trading tier and exchange account status fields to users table.
-- trading_tier tracks the user's Exchange trading activity level for VIP card access.
-- exchange_status tracks the linked Exchange account's standing for reward claim verification.

ALTER TABLE users ADD COLUMN IF NOT EXISTS trading_tier VARCHAR(20) DEFAULT '' NOT NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS exchange_status VARCHAR(20) DEFAULT '' NOT NULL;

-- Index for querying active traders (e.g., for VIP card eligibility checks).
CREATE INDEX IF NOT EXISTS idx_users_trading_tier ON users (trading_tier) WHERE trading_tier != '';

-- Index for filtering by exchange status (e.g., finding users with active accounts).
CREATE INDEX IF NOT EXISTS idx_users_exchange_status ON users (exchange_status) WHERE exchange_status != '';
