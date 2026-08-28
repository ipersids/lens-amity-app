-- name: CreateSession :one
INSERT INTO sessions (
  token_hash, user_id, created_at, last_seen_at, absolute_expires_at
) VALUES (
  sqlc.arg(token_hash),
  sqlc.arg(user_id),
  sqlc.arg(created_at),
  sqlc.arg(last_seen_at),
  sqlc.arg(absolute_expires_at)
)
RETURNING token_hash;

-- name: GetSession :one
SELECT
  token_hash,
  user_id,
  created_at,
  last_seen_at,
  absolute_expires_at,
  revoked_at
FROM sessions
WHERE token_hash = sqlc.arg(token_hash);

-- name: UpdateSessionActivity :one
UPDATE sessions
  SET last_seen_at = sqlc.arg(last_seen_at)
WHERE token_hash = sqlc.arg(token_hash)
  AND revoked_at IS NULL
  AND absolute_expires_at > sqlc.arg(last_seen_at)
RETURNING token_hash, last_seen_at;

-- name: RevokeSession :one
UPDATE sessions
  SET revoked_at = sqlc.arg(revoked_at)
WHERE token_hash = sqlc.arg(token_hash)
  AND revoked_at IS NULL
RETURNING token_hash, revoked_at;

-- name: RevokeAllSessions :exec
UPDATE sessions
  SET revoked_at = sqlc.arg(revoked_at)
WHERE user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL;

-- name: GetSessionOwner :one
SELECT username_key, username_display FROM users
WHERE id = sqlc.arg(id);
