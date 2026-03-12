-- Sports table
CREATE TABLE IF NOT EXISTS sports (
    key VARCHAR(64) PRIMARY KEY,
    group_name VARCHAR(64) NOT NULL,
    title VARCHAR(128) NOT NULL,
    is_active BOOLEAN DEFAULT true
);

-- Seed popular sports
INSERT INTO sports (key, group_name, title) VALUES
('soccer_epl', 'Soccer', 'English Premier League'),
('soccer_uefa_champs_league', 'Soccer', 'UEFA Champions League'),
('soccer_spain_la_liga', 'Soccer', 'La Liga'),
('soccer_germany_bundesliga', 'Soccer', 'Bundesliga'),
('soccer_italy_serie_a', 'Soccer', 'Serie A'),
('soccer_france_ligue_one', 'Soccer', 'Ligue 1'),
('basketball_nba', 'Basketball', 'NBA'),
('basketball_euroleague', 'Basketball', 'EuroLeague'),
('americanfootball_nfl', 'American Football', 'NFL'),
('baseball_mlb', 'Baseball', 'MLB'),
('icehockey_nhl', 'Ice Hockey', 'NHL'),
('tennis_atp_french_open', 'Tennis', 'ATP French Open'),
('mma_mixed_martial_arts', 'MMA', 'UFC/MMA'),
('cricket_ipl', 'Cricket', 'IPL')
ON CONFLICT (key) DO NOTHING;

-- Extend matches table
ALTER TABLE matches ADD COLUMN IF NOT EXISTS external_id VARCHAR(128) UNIQUE;
ALTER TABLE matches ADD COLUMN IF NOT EXISTS sport_key VARCHAR(64) REFERENCES sports(key);
ALTER TABLE matches ADD COLUMN IF NOT EXISTS source VARCHAR(16) DEFAULT 'manual';

-- Extend cards table
ALTER TABLE cards ADD COLUMN IF NOT EXISTS source VARCHAR(16) DEFAULT 'manual';
ALTER TABLE cards ADD COLUMN IF NOT EXISTS ai_prompt_data JSONB;
ALTER TABLE cards ADD COLUMN IF NOT EXISTS resolution_criteria TEXT;

-- Extend events table
ALTER TABLE events ADD COLUMN IF NOT EXISTS source VARCHAR(16) DEFAULT 'manual';
ALTER TABLE events ADD COLUMN IF NOT EXISTS sport_key VARCHAR(64) REFERENCES sports(key);

-- Automation log for tracking auto-generated content
CREATE TABLE IF NOT EXISTS automation_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_name VARCHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL, -- 'success', 'error'
    details JSONB,
    items_processed INT DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_automation_logs_job_name ON automation_logs(job_name);
CREATE INDEX IF NOT EXISTS idx_automation_logs_created_at ON automation_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_matches_external_id ON matches(external_id) WHERE external_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_matches_sport_key ON matches(sport_key) WHERE sport_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_matches_source ON matches(source);
CREATE INDEX IF NOT EXISTS idx_cards_source ON cards(source);
CREATE INDEX IF NOT EXISTS idx_events_sport_key ON events(sport_key) WHERE sport_key IS NOT NULL;
