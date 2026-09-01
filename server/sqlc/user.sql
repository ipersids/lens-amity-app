-- name: GetUserProfile :one
SELECT username_key, username_display FROM users
WHERE username_key = sqlc.arg(username_key);

-- name: CreateUser :one
INSERT INTO users (
  username_key, username_display, password_hash
) VALUES (
  sqlc.arg(username_key),
  sqlc.arg(username_display),
  sqlc.arg(password_hash)
)
RETURNING username_key, username_display;

-- name: GetUserDataForLogin :one
SELECT id, username_key, username_display, password_hash FROM users
WHERE username_key = sqlc.arg(username_key);
