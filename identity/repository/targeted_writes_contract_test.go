package repository

import (
	"context"
	"testing"
	"time"
)

func TestUserMutationsWriteOnlyTargetedColumns(t *testing.T) {
	r := mutationDB(t, "")
	var mutations UserMutations = r
	ctx := context.Background()
	if err := mutations.UpdateFields(ctx, 1, map[string]any{"first_name": "Alice"}); err != nil {
		t.Fatal(err)
	}
	var first, password string
	var locked bool
	if err := r.db.QueryRowContext(ctx, `SELECT first_name, password, is_locked FROM users WHERE id = 1`).Scan(&first, &password, &locked); err != nil {
		t.Fatal(err)
	}
	if first != "Alice" || password != "old" || !locked {
		t.Fatalf("first=%q password=%q locked=%v", first, password, locked)
	}
}

func TestSQLTokenAndUserWritesCommitTogether(t *testing.T) {
	r := mutationDB(t, ", CHECK (password <> 'blocked')")
	ctx := context.Background()
	now := time.Now()
	if _, err := r.db.ExecContext(ctx, `INSERT INTO password_reset_tokens (id, user_id, expires_at) VALUES (2, 1, $1)`, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := r.CommitPasswordReset(ctx, 2, 1, "blocked", now); err == nil {
		t.Fatal("invalid user write succeeded")
	}
	var tokenCount int
	var password string
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM password_reset_tokens WHERE id = 2`).Scan(&tokenCount); err != nil {
		t.Fatal(err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT password FROM users WHERE id = 1`).Scan(&password); err != nil {
		t.Fatal(err)
	}
	if tokenCount != 1 || password != "old" {
		t.Fatalf("rollback token=%d password=%q", tokenCount, password)
	}
	if err := r.CommitPasswordReset(ctx, 2, 1, "new", now); err != nil {
		t.Fatal(err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM password_reset_tokens WHERE id = 2`).Scan(&tokenCount); err != nil {
		t.Fatal(err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT password FROM users WHERE id = 1`).Scan(&password); err != nil {
		t.Fatal(err)
	}
	if tokenCount != 0 || password != "new" {
		t.Fatalf("commit token=%d password=%q", tokenCount, password)
	}
}
