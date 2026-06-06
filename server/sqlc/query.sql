-- name: GetUser :one
SELECT uuid, username_key, username_display FROM users
WHERE uuid = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY username_display;

-- name: CreateUser :one
INSERT INTO users (
  uuid, username_key, username_display, password_hash
) VALUES (
  $1, $2, $3, $4
)
RETURNING uuid, username_key, username_display;

-- name: UpdateUser :one
UPDATE users
  set username_key = $2,
  username_display = $3
WHERE uuid = $1
RETURNING uuid, username_key, username_display;

-- name: DeleteUser :exec
DELETE FROM users
WHERE uuid = $1;
