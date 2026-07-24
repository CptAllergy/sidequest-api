-- +goose Up
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE IF NOT EXISTS "users"
(
    "id"           TEXT PRIMARY KEY,
    "username"     citext UNIQUE NOT NULL,
    "display_name" TEXT          NOT NULL,
    "avatar_url"   TEXT,
    "bio"          TEXT,
    "created_at"   TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at"   TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS "users";

DROP EXTENSION IF EXISTS "citext";

