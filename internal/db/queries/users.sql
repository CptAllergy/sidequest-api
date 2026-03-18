-- name: ListUsers :many
SELECT * FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (username, password, email)
VALUES ($1, $2, $3)
RETURNING *;