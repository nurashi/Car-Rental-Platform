-- 001_create_bookings.sql
CREATE TABLE IF NOT EXISTS bookings (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL,
    vehicle_id           UUID NOT NULL,
    status               TEXT NOT NULL DEFAULT 'pending',
    start_date           TIMESTAMPTZ NOT NULL,
    end_date             TIMESTAMPTZ NOT NULL,
    total_price          NUMERIC(12, 2) NOT NULL DEFAULT 0,
    currency             TEXT NOT NULL DEFAULT 'USD',
    pickup_location_id   UUID,
    dropoff_location_id  UUID,
    notes                TEXT,
    cancellation_reason  TEXT,
    payment_status       TEXT NOT NULL DEFAULT 'unpaid',
    payment_ref          TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT bookings_status_check CHECK (
        status IN ('pending', 'confirmed', 'active', 'completed', 'cancelled')
    ),
    CONSTRAINT bookings_payment_status_check CHECK (
        payment_status IN ('unpaid', 'paid', 'refunded', 'partially_refunded')
    ),
    CONSTRAINT bookings_dates_check CHECK (end_date > start_date)
);

CREATE INDEX IF NOT EXISTS idx_bookings_user_id    ON bookings (user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_vehicle_id ON bookings (vehicle_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status     ON bookings (status);
CREATE INDEX IF NOT EXISTS idx_bookings_dates      ON bookings (vehicle_id, start_date, end_date);
