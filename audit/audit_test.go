package audit

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	forgeDB "github.com/forgego/forge/db"
	_ "github.com/mattn/go-sqlite3"
)

func TestSQLSinkWritesWithCallerTransaction(t *testing.T) {
	ctx := context.Background()
	db, err := forgeDB.NewDBWithDriver("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	sink := NewSQLSink(db)
	if err := sink.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	event := Event{ActorID: "user-1", ActorKind: "user", Operation: "update", Outcome: OutcomeSuccess, CorrelationID: "req-1", Source: "admin"}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := sink.Record(ctx, tx, event); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM forge_audit_events").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rolled-back event count = %d", count)
	}
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := sink.Record(ctx, tx, event); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM forge_audit_events").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("committed event count = %d", count)
	}
}

type brokenSink struct{}

func (brokenSink) Record(context.Context, DBTX, Event) error { return errors.New("write failed") }

func TestPolicyFailureModes(t *testing.T) {
	ctx := context.Background()
	event := Event{CorrelationID: "correlation-42"}
	t.Run("closed returns error", func(t *testing.T) {
		if err := Record(ctx, brokenSink{}, nil, event, FailClosed, Options{}); err == nil {
			t.Fatal("expected write error")
		}
	})
	t.Run("open counts logs and continues", func(t *testing.T) {
		var logs bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logs, nil))
		counter := &Counter{}
		if err := Record(ctx, brokenSink{}, nil, event, FailOpen, Options{Logger: logger, Counter: counter}); err != nil {
			t.Fatal(err)
		}
		if counter.Count() != 1 {
			t.Fatalf("failure count = %d", counter.Count())
		}
		if !strings.Contains(logs.String(), event.CorrelationID) {
			t.Fatalf("log omitted correlation ID: %s", logs.String())
		}
	})
}
