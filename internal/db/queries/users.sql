-- name: ListUsers :many
SELECT *
FROM users;

-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (email, username)
VALUES ($1, $2)
RETURNING *;