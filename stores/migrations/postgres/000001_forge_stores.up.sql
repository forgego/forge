-- Framework-owned tables for the shared stores (server.stores: database).
-- Applied by `forge migrate up` and tracked in forge_framework_migrations,
-- separately from the application's own migrations.

-- scs sessions (forge_session cookie). expires_at is Unix milliseconds.
CREATE TABLE forge_sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expires_at BIGINT NOT NULL
);
CREATE INDEX forge_sessions_expires_at_idx ON forge_sessions (expires_at);

-- Fixed-window counters for API throttling and the admin login lockout.
-- reset_at is Unix milliseconds.
CREATE TABLE forge_rate_limits (
    bucket TEXT PRIMARY KEY,
    hits BIGINT NOT NULL,
    reset_at BIGINT NOT NULL
);
CREATE INDEX forge_rate_limits_reset_at_idx ON forge_rate_limits (reset_at);

-- Admin API bearer tokens. Only the SHA-256 hash of a token is stored.
-- expires_at is Unix milliseconds.
CREATE TABLE forge_admin_tokens (
    token_hash TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at BIGINT NOT NULL
);
CREATE INDEX forge_admin_tokens_expires_at_idx ON forge_admin_tokens (expires_at);

-- Admin saved list views, per user and model. name_key is the lower-cased
-- name, so saving a view under an existing name updates it.
CREATE TABLE forge_admin_saved_views (
    id TEXT PRIMARY KEY,
    user_key TEXT NOT NULL,
    model TEXT NOT NULL,
    name TEXT NOT NULL,
    name_key TEXT NOT NULL,
    filters TEXT NOT NULL,
    ordering TEXT NOT NULL,
    display TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT forge_admin_saved_views_name_key UNIQUE (user_key, model, name_key)
);

-- Admin change history.
CREATE TABLE forge_admin_log (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    action_time TIMESTAMPTZ NOT NULL,
    user_id TEXT NOT NULL,
    user_name TEXT NOT NULL,
    model_name TEXT NOT NULL,
    object_id TEXT NOT NULL,
    object_repr TEXT NOT NULL,
    action TEXT NOT NULL,
    change_stats TEXT NOT NULL
);
CREATE INDEX forge_admin_log_object_idx ON forge_admin_log (lower(model_name), object_id, id);
