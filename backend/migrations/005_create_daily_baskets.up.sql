CREATE TABLE daily_baskets (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    basket_date  DATE NOT NULL,
    event_id     UUID NOT NULL REFERENCES events(id),
    is_published BOOLEAN DEFAULT FALSE,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_daily_baskets_date_event ON daily_baskets(basket_date, event_id);

CREATE TABLE daily_basket_cards (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    basket_id UUID NOT NULL REFERENCES daily_baskets(id),
    card_id   UUID NOT NULL REFERENCES cards(id),
    position  INTEGER NOT NULL,
    UNIQUE(basket_id, card_id),
    UNIQUE(basket_id, position)
);
