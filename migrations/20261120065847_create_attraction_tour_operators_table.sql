-- +goose Up
-- Which tour operators offer trips to which attractions.
CREATE TABLE attraction_tour_operators
(
    attraction_id    BIGINT NOT NULL REFERENCES attractions (id) ON DELETE CASCADE,
    tour_operator_id BIGINT NOT NULL REFERENCES tour_operators (id) ON DELETE CASCADE,
    PRIMARY KEY (attraction_id, tour_operator_id)
);

CREATE INDEX idx_attraction_tour_operators_tour_operator_id ON attraction_tour_operators (tour_operator_id);

-- +goose Down
DROP TABLE IF EXISTS attraction_tour_operators;
