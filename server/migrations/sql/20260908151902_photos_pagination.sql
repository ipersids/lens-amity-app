-- +goose Up
CREATE INDEX IF NOT EXISTS photos_owner_ready_date_idx
ON photos (owner_user_id, photo_date DESC, id DESC)
WHERE status = 'ready'
  AND deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS photos_owner_ready_date_idx;
