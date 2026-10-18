-- +goose Up
ALTER TABLE avatar_uploads
  DROP CONSTRAINT avatar_uploads_status_check,
  ADD COLUMN superseded_at TIMESTAMPTZ,
  ADD COLUMN deleted_at TIMESTAMPTZ,
  ADD CONSTRAINT avatar_uploads_status_check
    CHECK (status IN ('pending', 'processing', 'ready', 'failed', 'expired', 'superseded', 'deleted'));

-- +goose Down
ALTER TABLE avatar_uploads
  DROP CONSTRAINT avatar_uploads_status_check,
  DROP COLUMN IF EXISTS deleted_at,
  DROP COLUMN IF EXISTS superseded_at,
  ADD CONSTRAINT avatar_uploads_status_check
    CHECK (status IN ('pending', 'processing', 'ready', 'failed', 'expired'));
