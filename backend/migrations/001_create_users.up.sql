CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    xex_user_id     UUID UNIQUE NOT NULL,
    display_name    VARCHAR(100) NOT NULL,
    email           VARCHAR(255),
    avatar_url      TEXT,
    role            VARCHAR(20) DEFAULT 'user',
    referral_code   VARCHAR(20) UNIQUE NOT NULL,
    referred_by     UUID REFERENCES users(id),
    language        VARCHAR(5) DEFAULT 'en',
    total_points    INTEGER DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_xex_user_id ON users(xex_user_id);
CREATE INDEX idx_users_referral_code ON users(referral_code);
