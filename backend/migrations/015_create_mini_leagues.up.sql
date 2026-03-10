CREATE TABLE IF NOT EXISTS mini_leagues (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(128) NOT NULL,
    creator_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invite_code  VARCHAR(16) NOT NULL UNIQUE,
    event_id     UUID REFERENCES events(id) ON DELETE SET NULL,
    max_members  INT NOT NULL DEFAULT 50,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mini_league_members (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    league_id  UUID NOT NULL REFERENCES mini_leagues(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(league_id, user_id)
);

CREATE INDEX idx_mini_leagues_creator_id ON mini_leagues(creator_id);
CREATE INDEX idx_mini_leagues_invite_code ON mini_leagues(invite_code);
CREATE INDEX idx_mini_league_members_league_id ON mini_league_members(league_id);
CREATE INDEX idx_mini_league_members_user_id ON mini_league_members(user_id);
