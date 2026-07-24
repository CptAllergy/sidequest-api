-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS "quests"
(
    "id"          UUID PRIMARY KEY         DEFAULT uuid_generate_v4(),
    "user_id"     TEXT NOT NULL,
    "title"       TEXT NOT NULL,
    "description" TEXT,
    "type"        TEXT NOT NULL,
    "status"      TEXT NOT NULL,
    "image_url"   TEXT NOT NULL,
    "created_at"  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_quests_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "quest_entries"
(
    "id"         UUID PRIMARY KEY         DEFAULT uuid_generate_v4(),
    "quest_id"   UUID  NOT NULL,
    "user_id"    TEXT  NOT NULL,
    "type"       TEXT  NOT NULL,
    "content"    JSONB NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_quest_entries_quest_id FOREIGN KEY (quest_id) REFERENCES quests (id) ON DELETE CASCADE,
    CONSTRAINT fk_quest_log_entries_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
-- +goose Down
DROP TABLE IF EXISTS "quest_entries";
DROP TABLE IF EXISTS "quests";

DROP EXTENSION IF EXISTS "uuid-ossp";


