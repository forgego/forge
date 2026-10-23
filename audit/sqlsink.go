package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/forgego/forge/db"
)

// SQLSink stores events in forge_audit_events. Use the supplied transaction
// whenever audit and business data must commit or roll back together.
type SQLSink struct{ database *db.DB }

func NewSQLSink(database *db.DB) *SQLSink { return &SQLSink{database: database} }

// EnsureSchema creates the audit table and target lookup index for SQLite or PostgreSQL.
func (s *SQLSink) EnsureSchema(ctx context.Context) error {
	if s == nil || s.database == nil || s.database.DB == nil {
		return fmt.Errorf("audit database is nil")
	}
	id := "INTEGER PRIMARY KEY AUTOINCREMENT"
	if s.database.Driver == "postgres" {
		id = "BIGSERIAL PRIMARY KEY"
	}
	queries := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS forge_audit_events (
		id %s, occurred_at TIMESTAMP NOT NULL, actor_id TEXT NOT NULL, actor_kind TEXT NOT NULL,
		operation TEXT NOT NULL, model TEXT NULL, object_id TEXT NULL, object_repr TEXT NULL,
		outcome TEXT NOT NULL, changes TEXT NULL, correlation_id TEXT NOT NULL, source TEXT NOT NULL
	)`, id),
		`CREATE INDEX IF NOT EXISTS forge_audit_events_target_idx ON forge_audit_events (model, object_id, occurred_at)`,
	}
	for _, query := range queries {
		if _, err := s.database.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("ensure audit schema: %w", err)
		}
	}
	return nil
}

func (s *SQLSink) Record(ctx context.Context, tx DBTX, event Event) error {
	if s == nil || s.database == nil || s.database.DB == nil {
		return fmt.Errorf("audit database is nil")
	}
	var executor DBTX = s.database.DB
	if tx != nil {
		executor = tx
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}
	query := `INSERT INTO forge_audit_events (occurred_at, actor_id, actor_kind, operation, model, object_id, object_repr, outcome, changes, correlation_id, source) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	if _, err := executor.ExecContext(ctx, s.database.RebindPlaceholders(query), event.OccurredAt, event.ActorID, event.ActorKind, event.Operation, nullString(event.Model), nullString(event.ObjectID), nullString(event.ObjectRepr), event.Outcome, nullString(event.Changes), event.CorrelationID, event.Source); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}
	return nil
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
