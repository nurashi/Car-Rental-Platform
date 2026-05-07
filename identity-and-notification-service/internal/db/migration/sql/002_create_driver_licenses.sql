CREATE TABLE IF NOT EXISTS driver_licenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    license_number VARCHAR(100) NOT NULL,
    issuing_country VARCHAR(100) NOT NULL,
    issue_date DATE NOT NULL,
    expiry_date DATE NOT NULL,
    front_image_url TEXT NOT NULL DEFAULT '',
    back_image_url TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, license_number)
);

CREATE INDEX idx_driver_licenses_user_id ON driver_licenses(user_id);
CREATE INDEX idx_driver_licenses_status ON driver_licenses(status);
