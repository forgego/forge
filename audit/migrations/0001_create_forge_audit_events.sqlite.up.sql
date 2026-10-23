CREATE TABLE IF NOT EXISTS forge_audit_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    occurred_at TIMESTAMP NOT NULL,
    actor_id TEXT NOT NULL,
    actor_kind TEXT NOT NULL,
    operation TEXT NOT NULL,
    model TEXT NULL,
    object_id TEXT NULL,
    object_repr TEXT NULL,
    outcome TEXT NOT NULL,
    changes TEXT NULL,
    correlation_id TEXT NOT NULL,
    source TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS forge_audit_events_target_idx ON forge_audit_events (model, object_id, occurred_at);
