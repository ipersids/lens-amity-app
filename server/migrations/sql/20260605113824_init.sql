-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id UUID DEFAULT uuidv4() PRIMARY KEY,
    username_key TEXT NOT NULL UNIQUE,
    username_display TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE users;
