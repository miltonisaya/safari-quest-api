-- +goose Up
ALTER TABLE geographies
    ADD COLUMN parent_id BIGINT REFERENCES geographies (id) ON DELETE RESTRICT ON UPDATE CASCADE;

CREATE INDEX idx_geographies_parent_id ON geographies (parent_id);

-- +goose Down
DROP INDEX IF EXISTS idx_geographies_parent_id;
ALTER TABLE geographies DROP COLUMN IF EXISTS parent_id;
