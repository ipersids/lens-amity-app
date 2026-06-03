-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY username;

-- name: CreateUser :one
INSERT INTO users (
  id, username, password_hash
) VALUES (
  $1, $2, $3
)
RETURNING id, username;

-- name: UpdateUser :one
UPDATE users
  set username = $2
WHERE id = $1
RETURNING id, username;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
