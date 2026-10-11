-- +goose Up
ALTER TABLE IF EXISTS "quests" RENAME COLUMN "image_url" TO "image_key";
ALTER TABLE IF EXISTS "quests" ALTER COLUMN "image_key" DROP NOT NULL;

-- +goose Down
ALTER TABLE IF EXISTS "quests" ALTER COLUMN "image_key" SET NOT NULL;
ALTER TABLE IF EXISTS "quests" RENAME COLUMN "image_key" TO "image_url";


