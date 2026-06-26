-- +goose Up
CREATE TYPE photo_processing_status AS ENUM ('pending', 'ready', 'failed');

CREATE TABLE IF NOT EXISTS photos (
  id UUID PRIMARY KEY,
  owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  bucket VARCHAR(63) NOT NULL,
  object_key TEXT NOT NULL UNIQUE,
  local_date DATE NOT NULL,
  status photo_processing_status NOT NULL DEFAULT 'pending',
  upload_expires_at TIMESTAMPTZ NOT NULL,
  uploaded_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

-- 1 day = 1 photo enforcement
CREATE UNIQUE INDEX IF NOT EXISTS photos_owner_photo_date_active_unique
ON photos (owner_user_id, local_date)
WHERE deleted_at IS NULL
  AND status IN ('pending', 'ready');

-- cursor-based Pagination
CREATE INDEX IF NOT EXISTS photos_owner_photo_date_ready_idx
ON photos (owner_user_id, local_date DESC, id DESC)
WHERE deleted_at IS NULL
  AND status = 'ready';

-- +goose Down
DROP TABLE IF EXISTS photos;
DROP TYPE photo_processing_status;
