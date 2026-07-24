-- name: ListUsers :many
SELECT *
FROM users;

-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByIdForShare :one
SELECT *
FROM users
WHERE id = $1
FOR SHARE;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (id, username, display_name)
VALUES ($1, $2, $3)
RETURNING *;