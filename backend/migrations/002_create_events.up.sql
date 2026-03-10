CREATE TABLE events (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name               JSONB NOT NULL,
    slug               VARCHAR(100) UNIQUE NOT NULL,
    description        JSONB,
    start_date         DATE NOT NULL,
    end_date           DATE NOT NULL,
    is_active          BOOLEAN DEFAULT FALSE,
    scoring_multiplier DECIMAL(3,2) DEFAULT 1.00,
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    updated_at         TIMESTAMPTZ DEFAULT NOW()
);
