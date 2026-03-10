CREATE TABLE IF NOT EXISTS abuse_flags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID         NOT NULL REFERENCES users(id),
    flag_type   VARCHAR(50)  NOT NULL,
    details     JSONB        NOT NULL DEFAULT '{}',
    status      VARCHAR(20)  NOT NULL DEFAULT 'pending',
    reviewed_by UUID         REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_abuse_flags_user_id ON abuse_flags(user_id);
CREATE INDEX idx_abuse_flags_status  ON abuse_flags(status);
