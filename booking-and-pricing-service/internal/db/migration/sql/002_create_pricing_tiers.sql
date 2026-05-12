-- 002_create_pricing_tiers.sql
CREATE TABLE IF NOT EXISTS pricing_tiers (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 TEXT NOT NULL,
    vehicle_category     TEXT NOT NULL,
    base_daily_rate      NUMERIC(10, 2) NOT NULL,
    seasonal_multiplier  NUMERIC(5, 4) NOT NULL DEFAULT 1.0,
    season_start         TEXT,          -- 'MM-DD' e.g. '06-01'
    season_end           TEXT,          -- 'MM-DD' e.g. '08-31'
    is_active            BOOLEAN NOT NULL DEFAULT TRUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pricing_tiers_category  ON pricing_tiers (vehicle_category);
CREATE INDEX IF NOT EXISTS idx_pricing_tiers_is_active ON pricing_tiers (is_active);

-- Seed default tiers
INSERT INTO pricing_tiers (name, vehicle_category, base_daily_rate, seasonal_multiplier, season_start, season_end)
VALUES
    ('Economy Standard',   'economy',   29.99, 1.00, NULL, NULL),
    ('Economy Summer',     'economy',   39.99, 1.30, '06-01', '08-31'),
    ('Compact Standard',   'compact',   44.99, 1.00, NULL, NULL),
    ('Compact Summer',     'compact',   57.99, 1.25, '06-01', '08-31'),
    ('SUV Standard',       'suv',       74.99, 1.00, NULL, NULL),
    ('SUV Summer',         'suv',       94.99, 1.25, '06-01', '08-31'),
    ('Luxury Standard',    'luxury',   149.99, 1.00, NULL, NULL),
    ('Luxury Peak',        'luxury',   189.99, 1.30, '12-20', '01-05')
ON CONFLICT DO NOTHING;
