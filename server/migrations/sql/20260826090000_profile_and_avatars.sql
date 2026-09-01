-- +goose Up
ALTER TABLE users
  ADD COLUMN about TEXT,
  ADD COLUMN profile_visibility TEXT NOT NULL DEFAULT 'public',
  ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD CONSTRAINT users_about_length
    CHECK (about IS NULL OR length(about) <= 300),
  ADD CONSTRAINT users_profile_visibility_allowed
    CHECK (profile_visibility IN ('public', 'private'));

CREATE TABLE user_avatars (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  bucket VARCHAR(63) NOT NULL,
  object_key TEXT NOT NULL UNIQUE,
  content_type TEXT NOT NULL,
  uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS user_avatars;

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_profile_visibility_allowed,
  DROP CONSTRAINT IF EXISTS users_about_length,
  DROP COLUMN IF EXISTS updated_at,
  DROP COLUMN IF EXISTS profile_visibility,
  DROP COLUMN IF EXISTS about;
