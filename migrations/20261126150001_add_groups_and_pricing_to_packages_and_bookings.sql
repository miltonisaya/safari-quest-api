-- +goose Up
-- How a package is sold: several groups (units) run on the same dates, each
-- with `pax` seats. Groups can be booked shared (seats, strangers fill the rest)
-- and/or private (one booking takes a whole group). Each kind is priced per
-- person (charged for at least a minimum number of travellers) or per group.
ALTER TABLE packages
    ADD COLUMN units                  INT           NOT NULL DEFAULT 1,
    ADD COLUMN booking_mode           VARCHAR(10)   NOT NULL DEFAULT 'shared',
    ADD COLUMN shared_pricing         VARCHAR(12)   NOT NULL DEFAULT 'per_person',
    ADD COLUMN shared_min_travellers  INT           NOT NULL DEFAULT 1,
    ADD COLUMN private_pricing        VARCHAR(12),
    ADD COLUMN private_price          NUMERIC(12,2),
    ADD COLUMN private_min_travellers INT           NOT NULL DEFAULT 1,
    ADD COLUMN max_private_units      INT,
    ADD CONSTRAINT chk_packages_units CHECK (units BETWEEN 1 AND 50),
    ADD CONSTRAINT chk_packages_booking_mode CHECK (booking_mode IN ('shared', 'private', 'both')),
    ADD CONSTRAINT chk_packages_shared_pricing CHECK (shared_pricing IN ('per_person', 'per_group')),
    ADD CONSTRAINT chk_packages_private_pricing CHECK (private_pricing IS NULL OR private_pricing IN ('per_person', 'per_group')),
    ADD CONSTRAINT chk_packages_private_price CHECK (private_price IS NULL OR private_price >= 0),
    ADD CONSTRAINT chk_packages_min_travellers CHECK (shared_min_travellers >= 1 AND private_min_travellers >= 1),
    ADD CONSTRAINT chk_packages_max_private_units CHECK (max_private_units IS NULL OR max_private_units BETWEEN 1 AND units);

-- Each booking sits in one group and keeps a copy of how it was priced, so
-- later price changes never alter it. Existing bookings become shared bookings
-- in group 1, charged per person as before.
ALTER TABLE bookings
    ADD COLUMN booking_type       VARCHAR(10) NOT NULL DEFAULT 'shared',
    ADD COLUMN unit_number        INT         NOT NULL DEFAULT 1,
    ADD COLUMN pricing            VARCHAR(12) NOT NULL DEFAULT 'per_person',
    ADD COLUMN min_travellers     INT         NOT NULL DEFAULT 1,
    ADD COLUMN charged_travellers INT         NOT NULL DEFAULT 1,
    ADD CONSTRAINT chk_bookings_booking_type CHECK (booking_type IN ('shared', 'private')),
    ADD CONSTRAINT chk_bookings_unit_number CHECK (unit_number >= 1),
    ADD CONSTRAINT chk_bookings_pricing CHECK (pricing IN ('per_person', 'per_group'));

UPDATE bookings SET charged_travellers = pax;

CREATE INDEX idx_bookings_package_unit ON bookings (package_id, unit_number);

-- +goose Down
DROP INDEX IF EXISTS idx_bookings_package_unit;
ALTER TABLE bookings
    DROP CONSTRAINT IF EXISTS chk_bookings_pricing,
    DROP CONSTRAINT IF EXISTS chk_bookings_unit_number,
    DROP CONSTRAINT IF EXISTS chk_bookings_booking_type,
    DROP COLUMN IF EXISTS charged_travellers,
    DROP COLUMN IF EXISTS min_travellers,
    DROP COLUMN IF EXISTS pricing,
    DROP COLUMN IF EXISTS unit_number,
    DROP COLUMN IF EXISTS booking_type;
ALTER TABLE packages
    DROP CONSTRAINT IF EXISTS chk_packages_max_private_units,
    DROP CONSTRAINT IF EXISTS chk_packages_min_travellers,
    DROP CONSTRAINT IF EXISTS chk_packages_private_price,
    DROP CONSTRAINT IF EXISTS chk_packages_private_pricing,
    DROP CONSTRAINT IF EXISTS chk_packages_shared_pricing,
    DROP CONSTRAINT IF EXISTS chk_packages_booking_mode,
    DROP CONSTRAINT IF EXISTS chk_packages_units,
    DROP COLUMN IF EXISTS max_private_units,
    DROP COLUMN IF EXISTS private_min_travellers,
    DROP COLUMN IF EXISTS private_price,
    DROP COLUMN IF EXISTS private_pricing,
    DROP COLUMN IF EXISTS shared_min_travellers,
    DROP COLUMN IF EXISTS shared_pricing,
    DROP COLUMN IF EXISTS booking_mode,
    DROP COLUMN IF EXISTS units;
