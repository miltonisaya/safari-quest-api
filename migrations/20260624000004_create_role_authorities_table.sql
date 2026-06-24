-- +goose Up
CREATE TABLE role_authorities (
    role_id      BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    authority_id BIGINT NOT NULL REFERENCES authorities(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, authority_id)
);

-- +goose Down
DROP TABLE IF EXISTS role_authorities;
