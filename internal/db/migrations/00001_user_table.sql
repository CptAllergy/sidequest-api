-- +goose Up
CREATE EXTENSION IF NOT EXISTS "citext";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS "users"
(
    "id"           UUID PRIMARY KEY       DEFAULT uuid_generate_v4(),
    "email"        citext UNIQUE NOT NULL,
    "username"     citext UNIQUE NOT NULL,
    "display_name" TEXT,
    "avatar_url"   TEXT,
    "bio"          TEXT,
    "is_verified"  BOOLEAN       NOT NULL DEFAULT FALSE,
    "verified_at"  TIMESTAMPTZ,
    "created_at"   TIMESTAMPTZ            DEFAULT CURRENT_TIMESTAMP,
    "updated_at"   TIMESTAMPTZ            DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "user_accounts"
(
    "id"               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "user_id"          UUID NOT NULL,
    "provider"         TEXT NOT NULL, -- e.g., "local", "google", "GitHub"
    "provider_user_id" TEXT,
    "password"         bytea,
    "created_at"       TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP,
    "updated_at"       TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user_accounts_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS "user_accounts";
DROP TABLE IF EXISTS "users";

DROP EXTENSION IF EXISTS "citext";
DROP EXTENSION IF EXISTS "uuid-ossp";

