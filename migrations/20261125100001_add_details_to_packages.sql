-- +goose Up
-- Traveler-facing details of a sale package. Lists are JSON arrays of strings.
ALTER TABLE packages
    ADD COLUMN highlights    JSONB        NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN inclusions    JSONB        NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN exclusions    JSONB        NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN difficulty    VARCHAR(20),
    ADD COLUMN meeting_point VARCHAR(255),
    ADD CONSTRAINT chk_packages_difficulty CHECK (difficulty IN ('easy', 'moderate', 'challenging'));

-- Attractions a package visits.
CREATE TABLE package_attractions
(
    package_id    BIGINT NOT NULL REFERENCES packages (id) ON DELETE CASCADE,
    attraction_id BIGINT NOT NULL REFERENCES attractions (id) ON DELETE CASCADE,
    PRIMARY KEY (package_id, attraction_id)
);

CREATE INDEX idx_package_attractions_attraction_id ON package_attractions (attraction_id);

-- +goose Down
DROP TABLE IF EXISTS package_attractions;
ALTER TABLE packages
    DROP CONSTRAINT IF EXISTS chk_packages_difficulty,
    DROP COLUMN IF EXISTS meeting_point,
    DROP COLUMN IF EXISTS difficulty,
    DROP COLUMN IF EXISTS exclusions,
    DROP COLUMN IF EXISTS inclusions,
    DROP COLUMN IF EXISTS highlights;
