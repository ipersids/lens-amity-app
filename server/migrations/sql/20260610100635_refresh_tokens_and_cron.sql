-- +goose Up
CREATE TYPE token_revoked_reason AS ENUM ('refresh', 'logout', 'replayed');

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    grace_period_until TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    revoked_reason token_revoked_reason
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);

CREATE EXTENSION IF NOT EXISTS pg_cron;

SELECT cron.schedule(
  'vacuum_refresh_tokens',
  '0 3 * * *',
  $$ DELETE FROM refresh_tokens WHERE expires_at < now() OR (revoked = true AND created_at < now() - INTERVAL '3 days') $$
);

SELECT cron.schedule(
  'delete-job-run-details',
  '0 12 * * *',
  $$ DELETE FROM cron.job_run_details WHERE end_time < now() - interval '7 days' $$
);

-- +goose Down
SELECT cron.unschedule('vacuum_refresh_tokens');
SELECT cron.unschedule('delete-job-run-details');
DROP TABLE refresh_tokens;
DROP TYPE token_revoked_reason;
