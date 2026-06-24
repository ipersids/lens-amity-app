-- +goose Up
CREATE TYPE photo_processing_status AS ENUM ('pending', 'ready', 'failed');

CREATE TABLE IF NOT EXISTS photos (
  id UUID PRIMARY KEY,
  owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  bucket VARCHAR(63) NOT NULL,
  object_key TEXT NOT NULL UNIQUE,
  photo_date_utc DATE NOT NULL,
  content_type TEXT NOT NULL,
  size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
  status photo_processing_status NOT NULL DEFAULT 'pending',
  upload_expires_at TIMESTAMPTZ NOT NULL,
  uploaded_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

-- 1 day = 1 photo enforcement
CREATE UNIQUE INDEX IF NOT EXISTS photos_owner_photo_date_active_unique
ON photos (owner_user_id, photo_date_utc)
WHERE deleted_at IS NULL
  AND status IN ('pending', 'ready');

-- cursor-based Pagination
CREATE INDEX IF NOT EXISTS photos_owner_photo_date_ready_idx
ON photos (owner_user_id, photo_date_utc DESC, id DESC)
WHERE deleted_at IS NULL
  AND status = 'ready';

-- +goose Down
DROP TABLE IF EXISTS photos;
DROP TYPE photo_processing_status;
