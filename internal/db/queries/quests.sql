-- name: ListQuests :many
SELECT *
FROM quests;

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
INSERT INTO quests (user_id, title, description, type, status, image_url)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: CreateQuestEntry :one
INSERT INTO quest_entries (quest_id, user_id, type, content)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListQuestEntries :many
SELECT *
FROM quest_entries
WHERE quest_id = $1;