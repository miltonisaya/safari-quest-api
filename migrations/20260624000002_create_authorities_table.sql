-- +goose Up
CREATE TABLE authorities (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    uuid        UUID NOT NULL DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    code        VARCHAR(100) NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMP WITH TIME ZONE,
    CONSTRAINT authorities_uuid_unique UNIQUE (uuid),
    CONSTRAINT authorities_name_unique UNIQUE (name),
    CONSTRAINT authorities_code_unique UNIQUE (code)
);

-- +goose Down
DROP TABLE IF EXISTS authorities;
