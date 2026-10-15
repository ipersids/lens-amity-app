-- +goose Up
ALTER TABLE user_avatars
  ADD COLUMN size BIGINT NOT NULL,
  ADD CONSTRAINT user_avatars_size_positive CHECK (size > 0);

-- +goose Down
ALTER TABLE user_avatars
  DROP CONSTRAINT IF EXISTS user_avatars_size_positive,
  DROP COLUMN IF EXISTS size;
