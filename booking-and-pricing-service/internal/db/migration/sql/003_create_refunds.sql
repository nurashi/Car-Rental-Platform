-- 003_create_refunds.sql
CREATE TABLE IF NOT EXISTS refunds (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id    UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    amount        NUMERIC(12, 2) NOT NULL,
    reason        TEXT,
    status        TEXT NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT refunds_status_check CHECK (
        status IN ('pending', 'processed', 'failed')
    )
);

CREATE INDEX IF NOT EXISTS idx_refunds_booking_id ON refunds (booking_id);
