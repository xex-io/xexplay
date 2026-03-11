CREATE TABLE IF NOT EXISTS prize_pools (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    total_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    currency VARCHAR(20) NOT NULL DEFAULT 'XEX',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS prize_pool_distributions (
    id UUID PRIMARY KEY,
    prize_pool_id UUID NOT NULL REFERENCES prize_pools(id),
    user_id UUID NOT NULL REFERENCES users(id),
    amount DOUBLE PRECISION NOT NULL,
    rank INT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    distributed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_prize_pools_status ON prize_pools (status);
CREATE INDEX idx_prize_pool_distributions_pool_id ON prize_pool_distributions (prize_pool_id);
CREATE INDEX idx_prize_pool_distributions_user_id ON prize_pool_distributions (user_id);
