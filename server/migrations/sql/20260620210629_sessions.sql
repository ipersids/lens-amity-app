-- +goose Up
CREATE TYPE session_revoked_reason AS ENUM ('renewed', 'logout', 'replayed');

CREATE TABLE sessions (
    token_hash BYTEA PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL,
    absolute_expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    revoked_reason session_revoked_reason,
    grace_period_until TIMESTAMPTZ,
    CONSTRAINT sessions_revocation_consistent CHECK (
        (revoked_at IS NULL AND revoked_reason IS NULL)
        OR (revoked_at IS NOT NULL AND revoked_reason IS NOT NULL)
    ),
    CONSTRAINT sessions_renewal_grace_consistent CHECK (
        (grace_period_until IS NULL AND revoked_reason IS DISTINCT FROM 'renewed')
        OR (
            grace_period_until IS NOT NULL
            AND revoked_reason IS NOT DISTINCT FROM 'renewed'
            AND grace_period_until > revoked_at
        )
    )
);

CREATE INDEX sessions_user_id_idx ON sessions(user_id);
CREATE INDEX sessions_absolute_expires_at_idx ON sessions(absolute_expires_at);
CREATE INDEX sessions_revoked_at_idx ON sessions(revoked_at) WHERE revoked_at IS NOT NULL;

CREATE EXTENSION IF NOT EXISTS pg_cron;

SELECT cron.schedule(
  'lensamity_cleanup_sessions',
  '0 3 * * *',
  $$
    DELETE FROM sessions
    WHERE absolute_expires_at <= now() - INTERVAL '3 days'
       OR revoked_at <= now() - INTERVAL '3 days'
  $$
);

-- +goose Down
SELECT cron.unschedule('lensamity_cleanup_sessions');
DROP TABLE sessions;
DROP TYPE session_revoked_reason;
