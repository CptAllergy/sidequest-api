-- name: ListQuests :many
SELECT * FROM quests;

-- name: GetQuest :one
SELECT * FROM quests
WHERE id = $1;

-- name: CreateQuest :one
INSERT INTO quests (name, description, reward)
VALUES ($1, $2, $3)
RETURNING *;