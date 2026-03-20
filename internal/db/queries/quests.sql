-- name: ListQuests :many
SELECT *
FROM quests;

-- name: GetQuest :one
SELECT *
FROM quests
WHERE id = $1;

-- name: CreateQuest :one
INSERT INTO quests (user_id, title, description, type, status, image_url)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;