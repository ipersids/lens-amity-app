-- name: CreatePendingPhotoRecord :exec
INSERT INTO photos (
  id, owner_user_id, bucket, object_key,
  local_date, upload_expires_at,
  status
) VALUES (
  sqlc.arg(id),
  sqlc.arg(owner_user_id),
  sqlc.arg(bucket),
  sqlc.arg(object_key),
  sqlc.arg(local_date),
  sqlc.arg(upload_expires_at),
  'pending'
);

-- name: MarkPhotoReady :one
UPDATE photos
  SET status = 'ready',
      uploaded_at = now()
WHERE id = sqlc.arg(id)
  AND owner_user_id = sqlc.arg(owner_user_id)
  AND status = 'pending'
RETURNING id;

-- name: GetPhotoRecord :one
SELECT id, owner_user_id, bucket,
  object_key, upload_expires_at, status
FROM photos
WHERE id = sqlc.arg(id)
  AND owner_user_id = sqlc.arg(owner_user_id);
