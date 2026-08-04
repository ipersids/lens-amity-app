-- +goose Up
CREATE TYPE upload_kind AS ENUM ('photo', 'avatar');

CREATE TYPE upload_status AS ENUM (
  'pending',
  'ready',
  'failed',
  'expired',
  'deleted'
);

CREATE TABLE IF NOT EXISTS uploaded_files (
  id UUID PRIMARY KEY,
  owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  kind upload_kind NOT NULL,
  status upload_status NOT NULL DEFAULT 'pending',
  failure_reason TEXT,

  bucket VARCHAR(63) NOT NULL,
  object_key_original TEXT NOT NULL UNIQUE,
  object_key_processed JSONB NOT NULL DEFAULT '{}'::jsonb,

  content_type TEXT NOT NULL,
  size BIGINT NOT NULL,
  width INTEGER,
  height INTEGER,
  photo_date DATE,
  title TEXT,
  description TEXT,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  processed_at TIMESTAMPTZ,
  expires_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ,

  CONSTRAINT uploaded_files_object_key_original_not_empty
    CHECK (length(object_key_original) > 0),
  CONSTRAINT uploaded_files_allowed_content_type
    CHECK (
      content_type IN (
        'image/jpeg',
        'image/png',
        'image/webp',
        'image/heic',
        'image/heif'
      )
    ),
  CONSTRAINT uploaded_files_size_positive
    CHECK (size > 0),
  CONSTRAINT uploaded_files_dimensions_positive
    CHECK (
      (width IS NULL AND height IS NULL)
      OR (width > 0 AND height > 0)
    ),
  CONSTRAINT uploaded_files_photo_requires_date
    CHECK (
      (kind = 'photo' AND photo_date IS NOT NULL)
      OR (kind = 'avatar' AND photo_date IS NULL)
    ),
  CONSTRAINT uploaded_files_avatar_has_no_photo_text
    CHECK (
      kind = 'photo'
      OR (title IS NULL AND description IS NULL)
    ),
  CONSTRAINT uploaded_files_title_length
    CHECK (title IS NULL OR length(title) <= 120),
  CONSTRAINT uploaded_files_description_length
    CHECK (description IS NULL OR length(description) <= 2000)
);

CREATE UNIQUE INDEX IF NOT EXISTS uploaded_files_owner_photo_date_active_unique
ON uploaded_files (owner_user_id, photo_date)
WHERE kind = 'photo'
  AND deleted_at IS NULL
  AND status IN ('pending', 'ready');

CREATE INDEX IF NOT EXISTS uploaded_files_expirable_idx
ON uploaded_files (expires_at)
WHERE status = 'pending';

ALTER TABLE users
ADD COLUMN IF NOT EXISTS avatar_file_id UUID;

ALTER TABLE users
ADD CONSTRAINT users_avatar_file_id_fkey
FOREIGN KEY (avatar_file_id)
REFERENCES uploaded_files(id)
ON DELETE SET NULL;

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_avatar_file_id_fkey;
ALTER TABLE users DROP COLUMN IF EXISTS avatar_file_id;

DROP TABLE IF EXISTS uploaded_files;
DROP TYPE IF EXISTS upload_status;
DROP TYPE IF EXISTS upload_kind;
