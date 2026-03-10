CREATE TABLE IF NOT EXISTS reward_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_type VARCHAR(20) NOT NULL CHECK (period_type IN ('daily', 'weekly', 'tournament')),
    rank_from INTEGER NOT NULL,
    rank_to INTEGER NOT NULL,
    reward_type VARCHAR(20) NOT NULL CHECK (reward_type IN ('token', 'bonus_skip', 'bonus_answer', 'badge')),
    amount DECIMAL(18, 8) NOT NULL DEFAULT 0,
    description JSONB,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reward_distributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reward_config_id UUID REFERENCES reward_configs(id) ON DELETE SET NULL,
    period_type VARCHAR(20) NOT NULL,
    period_key VARCHAR(20) NOT NULL,
    reward_type VARCHAR(20) NOT NULL,
    amount DECIMAL(18, 8) NOT NULL DEFAULT 0,
    rank INTEGER,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'claimed', 'credited', 'expired')),
    claimed_at TIMESTAMPTZ,
    credited_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_reward_dist_user ON reward_distributions(user_id);
CREATE INDEX idx_reward_dist_status ON reward_distributions(user_id, status);
CREATE INDEX idx_reward_dist_period ON reward_distributions(period_type, period_key);
