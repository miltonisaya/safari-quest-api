-- +goose Up
-- Staff accounts belonging to a tour operator company.
CREATE TABLE tour_operator_users
(
    tour_operator_id BIGINT NOT NULL REFERENCES tour_operators (id) ON DELETE CASCADE,
    user_id          BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    PRIMARY KEY (tour_operator_id, user_id)
);

CREATE INDEX idx_tour_operator_users_user_id ON tour_operator_users (user_id);

-- +goose Down
DROP TABLE IF EXISTS tour_operator_users;
