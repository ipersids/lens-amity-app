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
  grace_period_until, revoked_reason
FROM refresh_tokens
WHERE id = $1 AND user_id = $2
FOR UPDATE;

-- name: CreateNewRefreshToken :exec
INSERT INTO refresh_tokens (
  id, user_id, expires_at
) VALUES (
  $1, $2, $3
);

-- name: RotateRefreshToken :one
UPDATE refresh_tokens
    SET revoked = true,
    revoked_at = now(),
    revoked_reason = $1,
    grace_period_until = $2
WHERE id = $3 AND user_id = $4
RETURNING id, revoked, revoked_reason, grace_period_until, revoked_at;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
    SET revoked = true,
    revoked_at = now(),
    revoked_reason = $1
WHERE id = $2 AND user_id = $3 AND revoked = false
RETURNING id;

-- name: RevokeAllUserTokens :many
UPDATE refresh_tokens
    SET revoked = true,
    revoked_at = now(),
    revoked_reason = $1
WHERE user_id = $2 AND revoked = false
RETURNING id;
