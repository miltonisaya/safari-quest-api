-- +goose Up
-- Fields the public catalog shows for each package.
ALTER TABLE packages
    ADD COLUMN location      VARCHAR(255),                        -- e.g. "Serengeti, Tanzania"
    ADD COLUMN currency      CHAR(3)  NOT NULL DEFAULT 'USD',     -- ISO 4217 code of price
    ADD COLUMN eco_certified BOOLEAN  NOT NULL DEFAULT FALSE,
    ADD CONSTRAINT chk_packages_currency CHECK (currency ~ '^[A-Z]{3}$');

CREATE INDEX idx_packages_start_date ON packages (start_date);

-- +goose Down
DROP INDEX IF EXISTS idx_packages_start_date;
ALTER TABLE packages
    DROP CONSTRAINT IF EXISTS chk_packages_currency,
    DROP COLUMN IF EXISTS eco_certified,
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS location;
