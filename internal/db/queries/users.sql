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

-- name: CreateUserAccount :one
INSERT INTO user_accounts (user_id, provider, provider_user_id, password)
VALUES ($1, $2, $3, $4)
RETURNING *;