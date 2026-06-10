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

-- name: GetRefreshTokenForUpdate :one
SELECT id, user_id, revoked, expires_at,
  grace_period_until, replaced_by_access, replaced_by_refresh
FROM refresh_tokens
WHERE id = $1 AND user_id = $2
FOR UPDATE;

-- name: CreateNewRefreshToken :exec
INSERT INTO refresh_tokens (
  id, user_id, expires_at
) VALUES (
  $1, $2, $3
);

-- name: RotateRefreshToken :exec
UPDATE refresh_tokens
    SET revoked = true,
    grace_period_until = $2,
    replaced_by_access = $3,
    replaced_by_refresh = $4
WHERE id = $1;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
    SET revoked = true
WHERE id = $1 AND user_id = $2 AND revoked = false
RETURNING id;

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens
    SET revoked = true
WHERE user_id = $1 AND revoked = false
