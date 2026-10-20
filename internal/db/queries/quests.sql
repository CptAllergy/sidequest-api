-- name: ListQuestsByUserIdFirst :many
SELECT *
FROM quests
WHERE user_id = @user_id
ORDER BY updated_at DESC, id DESC
LIMIT @lim;

-- name: ListQuestsByUserIdAfter :many
SELECT *
FROM quests
WHERE user_id = @user_id
  AND (updated_at, id) < (@after_updated_at::timestamptz, @after_id::uuid)
ORDER BY updated_at DESC, id DESC
LIMIT @lim;

-- name: GetQuest :one
SELECT *
FROM quests
WHERE id = $1;

-- name: GetQuestForShare :one
SELECT *
FROM quests
WHERE id = $1
FOR SHARE;

-- name: CreateQuest :one
INSERT INTO quests (id, user_id, title, description, type, status, image_key)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: CreateQuestEntry :one
INSERT INTO quest_entries (quest_id, user_id, type, content)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListQuestEntriesFirst :many
SELECT *
FROM quest_entries
WHERE quest_id = @quest_id
ORDER BY created_at DESC, id DESC
LIMIT @lim;

-- name: ListQuestEntriesAfter :many
SELECT *
FROM quest_entries
WHERE quest_id = @quest_id
    AND (created_at, id) < (@after_created_at::timestamptz, @after_id::uuid)
ORDER BY created_at DESC, id DESC
LIMIT @lim;

-- name: TouchQuest :exec
UPDATE quests SET updated_at = now() WHERE id = $1;
