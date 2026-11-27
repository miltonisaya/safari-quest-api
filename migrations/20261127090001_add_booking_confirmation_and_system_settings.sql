-- +goose Up
-- Administrator-managed system settings: one row per known key, typed and
-- range-checked in code (internal/setting).
CREATE TABLE system_settings (
    key         VARCHAR(64) PRIMARY KEY,
    value       TEXT        NOT NULL,
    updated_by  BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Booking confirmation: companies confirm or decline pending bookings, and
-- pending bookings not confirmed by their deadline expire. Declined and
-- expired are final and release the seats.
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS chk_bookings_status;
ALTER TABLE bookings
    ADD CONSTRAINT chk_bookings_status CHECK (status IN ('pending', 'confirmed', 'completed', 'cancelled', 'declined', 'expired')),
    ADD COLUMN decline_reason TEXT,
    ADD COLUMN declined_at    TIMESTAMPTZ,
    ADD COLUMN expires_at     TIMESTAMPTZ,
    ADD COLUMN expired_at     TIMESTAMPTZ;

-- Finds overdue pending bookings quickly.
CREATE INDEX idx_bookings_pending_expiry ON bookings (expires_at) WHERE status = 'pending' AND deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_bookings_pending_expiry;
UPDATE bookings SET status = 'cancelled', cancelled_at = COALESCE(cancelled_at, declined_at, expired_at, NOW())
WHERE status IN ('declined', 'expired');
ALTER TABLE bookings
    DROP COLUMN IF EXISTS expired_at,
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS declined_at,
    DROP COLUMN IF EXISTS decline_reason,
    DROP CONSTRAINT IF EXISTS chk_bookings_status;
ALTER TABLE bookings ADD CONSTRAINT chk_bookings_status CHECK (status IN ('pending', 'confirmed', 'completed', 'cancelled'));
DROP TABLE IF EXISTS system_settings;
