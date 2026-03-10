CREATE TABLE user_sessions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id),
    basket_id     UUID NOT NULL REFERENCES daily_baskets(id),
    shuffle_order INTEGER[] NOT NULL,
    current_index INTEGER DEFAULT 0,
    answers_used  INTEGER DEFAULT 0,
    skips_used    INTEGER DEFAULT 0,
    bonus_answers INTEGER DEFAULT 0,
    bonus_skips   INTEGER DEFAULT 0,
    status        VARCHAR(20) DEFAULT 'active',
    started_at    TIMESTAMPTZ DEFAULT NOW(),
    completed_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_user_sessions_user_basket ON user_sessions(user_id, basket_id);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_status ON user_sessions(status);
