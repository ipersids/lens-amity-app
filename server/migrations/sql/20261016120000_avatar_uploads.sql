-- +goose Up
CREATE TABLE avatar_uploads (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  bucket VARCHAR(63) NOT NULL,
  object_key TEXT NOT NULL UNIQUE,
  content_type TEXT NOT NULL,
  size BIGINT NOT NULL CHECK (size > 0),
  status TEXT NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'processing', 'ready', 'failed', 'expired')),
  failure_reason TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ
);

CREATE INDEX avatar_uploads_user_id_created_at_idx
  ON avatar_uploads (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS avatar_uploads;
