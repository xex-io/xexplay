CREATE TABLE cards (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id           UUID NOT NULL REFERENCES matches(id),
    question_text      JSONB NOT NULL,
    tier               VARCHAR(10) NOT NULL,
    high_answer_is_yes BOOLEAN,
    correct_answer     BOOLEAN,
    is_resolved        BOOLEAN DEFAULT FALSE,
    available_date     DATE NOT NULL,
    expires_at         TIMESTAMPTZ NOT NULL,
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    updated_at         TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_cards_match_id ON cards(match_id);
CREATE INDEX idx_cards_available_date ON cards(available_date);
CREATE INDEX idx_cards_tier ON cards(tier);
CREATE INDEX idx_cards_is_resolved ON cards(is_resolved);
