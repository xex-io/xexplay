CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR(128) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    is_secret BOOLEAN DEFAULT false,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Seed default settings
INSERT INTO settings (key, value, description, is_secret) VALUES
('ODDS_API_KEY', '', 'The Odds API key for fetching sports data', true),
('ANTHROPIC_API_KEY', '', 'Anthropic API key for Claude AI question generation', true),
('AUTO_SPORTS_ENABLED', 'true', 'Enable/disable sports automation cron jobs', false)
ON CONFLICT (key) DO NOTHING;
