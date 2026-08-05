-- name: CreatePendingPhotoUploadRecord :exec
INSERT INTO photos (
  id,
  owner_user_id,
  bucket,
  object_key_original,
  content_type,
  size,
  photo_date,
  title,
  description,
  expires_at
) VALUES (
  sqlc.arg(id),
  sqlc.arg(owner_user_id),
  sqlc.arg(bucket),
  sqlc.arg(object_key_original),
  sqlc.arg(content_type),
  sqlc.arg(size),
  sqlc.arg(photo_date),
  sqlc.arg(title),
  sqlc.arg(description),
  sqlc.arg(expires_at)
);

-- name: MarkPhotoUploadRecordProcessing :one
UPDATE photos
  SET status = 'processing',
      uploaded_at = now()
WHERE id = sqlc.arg(id)
  AND owner_user_id = sqlc.arg(owner_user_id)
  AND status = 'pending'
RETURNING
  id,
  owner_user_id,
  bucket,
  object_key_original,
  content_type,
  size,
  photo_date,
  status;

-- name: MarkProcessedPhotoUploadRecordCompleted :one
UPDATE photos
  SET status = 'ready',
      processed_at = now(),
      object_key_processed = sqlc.arg(object_key_processed),
      width = sqlc.arg(width),
      height = sqlc.arg(height)
WHERE id = sqlc.arg(id)
  AND owner_user_id = sqlc.arg(owner_user_id)
  AND status = 'processing'
RETURNING id;

-- name: MarkProcessedPhotoUploadRecordFailed :one
UPDATE photos
  SET status = 'failed',
      failure_reason = sqlc.arg(failure_reason)
WHERE id = sqlc.arg(id)
  AND owner_user_id = sqlc.arg(owner_user_id)
  AND status = 'processing'
RETURNING id;

-- -- name: MarkPhotoUploadRecordDeleted :one
-- UPDATE photos
--   SET status = 'deleted',
--       deleted_at = now()
-- WHERE id = sqlc.arg(id)
--   AND owner_user_id = sqlc.arg(owner_user_id)
--   AND status != 'deleted'
-- RETURNING id;

-- -- name: LockExpiredPendingPhotoUploads :many
-- SELECT id, bucket, object_key_original
-- FROM photos
-- WHERE status = 'pending'
--   AND expires_at < now()
-- ORDER BY expires_at
-- LIMIT sqlc.arg(limit_count)
-- FOR UPDATE SKIP LOCKED;

-- -- name: ExpirePhotoUploadsByIDs :many
-- UPDATE photos
-- SET status = 'expired',
--     failure_reason = 'unfinished_upload'
-- WHERE id = ANY(sqlc.arg(ids)::uuid[])
--   AND status = 'pending'
-- RETURNING id, bucket, object_key_original;
