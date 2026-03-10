CREATE TABLE user_answers (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id    UUID NOT NULL REFERENCES user_sessions(id),
    card_id       UUID NOT NULL REFERENCES cards(id),
    user_id       UUID NOT NULL REFERENCES users(id),
    answer        BOOLEAN NOT NULL,
    points_earned INTEGER DEFAULT 0,
    is_correct    BOOLEAN,
    answered_at   TIMESTAMPTZ DEFAULT NOW(),
    resolved_at   TIMESTAMPTZ,
    UNIQUE(session_id, card_id)
);

CREATE INDEX idx_user_answers_user_id ON user_answers(user_id);
CREATE INDEX idx_user_answers_card_id ON user_answers(card_id);
CREATE INDEX idx_user_answers_is_correct ON user_answers(is_correct);
