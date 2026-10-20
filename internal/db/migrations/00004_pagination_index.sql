-- +goose Up
CREATE INDEX IF NOT EXISTS quests_user_updated_id_idx ON quests (user_id, updated_at, id);
CREATE INDEX IF NOT EXISTS quest_entries_quest_created_id_idx ON quest_entries (quest_id, created_at, id);

-- +goose Down
DROP INDEX IF EXISTS quest_entries_quest_created_id_idx;
DROP INDEX IF EXISTS quests_user_updated_id_idx;


