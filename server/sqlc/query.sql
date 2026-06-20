-- name: GetPublicUserProfile :one
SELECT username_key, username_display FROM users
WHERE username_key = $1;

-- name: CreateUser :one
INSERT INTO users (
  username_key, username_display, password_hash
) VALUES (
  $1, $2, $3
)
RETURNING username_key, username_display;

-- name: UpdateUser :one
UPDATE users
  SET username_key = $2,
  username_display = $3
WHERE username_key = $1
RETURNING username_key, username_display;

-- name: DeleteUser :exec
DELETE FROM users
WHERE username_key = $1;

-- name: GetFullUserDataByKey :one
SELECT id, username_key, username_display, password_hash FROM users
WHERE username_key = $1;
