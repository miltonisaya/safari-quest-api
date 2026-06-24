-- +goose Up
CREATE TABLE users (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    uuid               UUID NOT NULL DEFAULT gen_random_uuid(),
    first_name         VARCHAR(100) NOT NULL,
    middle_name        TEXT,
    last_name          VARCHAR(100) NOT NULL,
    email              VARCHAR(255) NOT NULL,
    password           TEXT NOT NULL,
    sex                VARCHAR(20) NOT NULL,
    mobile             VARCHAR(30) NOT NULL,
    address            TEXT NOT NULL,
    is_active          BOOLEAN NOT NULL DEFAULT TRUE,
    email_verified_at  TIMESTAMP WITH TIME ZONE,
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMP WITH TIME ZONE,
    CONSTRAINT users_uuid_unique  UNIQUE (uuid),
    CONSTRAINT users_email_unique UNIQUE (email)
);

-- +goose Down
DROP TABLE IF EXISTS users CASCADE;
