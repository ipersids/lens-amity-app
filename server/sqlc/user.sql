-- name: GetUserProfile :one
SELECT
  u.id,
  u.username_key,
  u.username_display,
  u.about,
  u.profile_visibility,
  u.created_at AS joined_at,
  a.bucket AS avatar_bucket,
  a.object_key AS avatar_object_key,
  (
    SELECT COUNT(p.id)
    FROM photos p
    WHERE p.owner_user_id = u.id
      AND p.status = 'ready'
      AND p.deleted_at IS NULL
  ) AS photo_count
FROM users u
LEFT JOIN user_avatars a ON u.id = a.user_id
WHERE u.username_key = sqlc.arg(username_key);

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
