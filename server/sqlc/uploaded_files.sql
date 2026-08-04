-- name: CreatePendingUploadRecord :exec
INSERT INTO uploaded_files (
  id,
  owner_user_id,
  kind,
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
  sqlc.arg(kind),
  sqlc.arg(bucket),
  sqlc.arg(object_key_original),
  sqlc.arg(content_type),
  sqlc.arg(size),
  sqlc.arg(photo_date),
  sqlc.arg(title),
  sqlc.arg(description),
  sqlc.arg(expires_at)
);

-- name: CompletePendingUploadRecord :one
UPDATE uploaded_files
  SET status = 'ready',
      uploaded_at = now(),
      processed_at = now(),
      object_key_processed = sqlc.arg(object_key_processed),
      content_type = sqlc.arg(content_type),
      size = sqlc.arg(size),
      width = sqlc.arg(width),
      height = sqlc.arg(height)
WHERE id = sqlc.arg(id)
  AND owner_user_id = sqlc.arg(owner_user_id)
  AND status = 'pending'
RETURNING id;

-- name: FailPendingUploadRecord :one
UPDATE uploaded_files
  SET status = 'failed',
      failure_reason = sqlc.arg(failure_reason)
WHERE id = sqlc.arg(id)
  AND owner_user_id = sqlc.arg(owner_user_id)
  AND status = 'pending'
RETURNING id;

-- name: ExpirePendingUploadRecords :many
UPDATE uploaded_files
  SET status = 'expired',
      failure_reason = 'unfinished_upload'
WHERE expires_at < now()
  AND status = 'pending'
RETURNING id, bucket, object_key_original;

-- name: MarkUploadRecordDeleted :one
UPDATE uploaded_files
  SET status = 'deleted',
      deleted_at = now()
WHERE id = sqlc.arg(id)
  AND owner_user_id = sqlc.arg(owner_user_id)
  AND status != 'deleted'
RETURNING id;

-- -- name: GetUploadRecord :one
-- SELECT
--   id,
--   owner_user_id,
--   kind,
--   status,
--   failure_reason,
--   bucket,
--   object_key_original,
--   object_key_processed,
--   content_type,
--   size,
--   width,
--   height,
--   photo_date,
--   title,
--   description,
--   created_at,
--   uploaded_at,
--   processed_at,
--   expires_at,
--   deleted_at
-- FROM uploaded_files
-- WHERE id = sqlc.arg(id)
--   AND owner_user_id = sqlc.arg(owner_user_id);
