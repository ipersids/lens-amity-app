-- +goose Up
CREATE TABLE users (
    uuid UUID PRIMARY KEY,
    username_key TEXT NOT NULL UNIQUE,
    username_display TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

-- +goose Down
DROP TABLE users;
