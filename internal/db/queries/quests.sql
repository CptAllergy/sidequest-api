-- name: ListQuestsByUserId :many
SELECT *
FROM quests
WHERE user_id = $1;

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

-- name: ListQuestEntries :many
SELECT *
FROM quest_entries
WHERE quest_id = $1;