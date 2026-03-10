CREATE TABLE matches (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id     UUID NOT NULL REFERENCES events(id),
    home_team    VARCHAR(100) NOT NULL,
    away_team    VARCHAR(100) NOT NULL,
    kickoff_time TIMESTAMPTZ NOT NULL,
    status       VARCHAR(20) DEFAULT 'upcoming',
    home_score   INTEGER,
    away_score   INTEGER,
    result_data  JSONB,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_matches_event_id ON matches(event_id);
CREATE INDEX idx_matches_kickoff_time ON matches(kickoff_time);
CREATE INDEX idx_matches_status ON matches(status);
