-- name: GetPublicUserProfile :one
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

-- name: UpdateUser :one
UPDATE users
  SET username_key = sqlc.arg(username_key),
  username_display = sqlc.arg(username_display)
WHERE username_key = sqlc.arg(current_username_key)
RETURNING username_key, username_display;

-- name: DeleteUser :exec
DELETE FROM users
WHERE username_key = sqlc.arg(username_key);

-- name: GetFullUserDataByKey :one
SELECT id, username_key, username_display, password_hash FROM users
WHERE username_key = sqlc.arg(username_key);
