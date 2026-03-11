CREATE TABLE IF NOT EXISTS notification_history (
    id UUID PRIMARY KEY,
    admin_user_id UUID NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    target_type VARCHAR(50) NOT NULL DEFAULT 'all',
    recipient_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notification_history_created_at ON notification_history (created_at DESC);
