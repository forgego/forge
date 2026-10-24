-- Saved filter ASTs for API and admin query building.
CREATE TABLE forge_saved_filters (
    id TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    ast TEXT NOT NULL DEFAULT '',
    dialect_hint TEXT NOT NULL DEFAULT '',
    schema_version TEXT NOT NULL DEFAULT '',
    public BOOLEAN NOT NULL DEFAULT FALSE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    sample_count INTEGER NOT NULL DEFAULT 0,
    estimated_cost INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX forge_saved_filters_owner_idx ON forge_saved_filters (owner_id, public);

-- Idempotency keys and cached HTTP responses.
CREATE TABLE idempotency_keys (
    key TEXT PRIMARY KEY,
    response_json TEXT,
    expires_at TIMESTAMP NOT NULL
);
CREATE INDEX idempotency_keys_expires_at_idx ON idempotency_keys (expires_at);
