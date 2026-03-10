CREATE TABLE IF NOT EXISTS referrals (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    referrer_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    referred_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status         VARCHAR(32) NOT NULL DEFAULT 'signed_up' CHECK (status IN ('signed_up', 'first_session')),
    reward_granted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(referred_id)
);

CREATE INDEX idx_referrals_referrer_id ON referrals(referrer_id);
CREATE INDEX idx_referrals_referred_id ON referrals(referred_id);
