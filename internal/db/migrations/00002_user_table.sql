-- +goose Up
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE IF NOT EXISTS "users"
(
    "id"         UUID PRIMARY KEY         DEFAULT uuid_generate_v4(),
    "username"   VARCHAR(255)  NOT NULL,
    "email"      citext UNIQUE NOT NULL,
    "password"   bytea         NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "citext";
