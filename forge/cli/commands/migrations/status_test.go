package migrations

import (
	"bytes"
	"strings"
	"testing"

	"github.com/forgego/forge/db"
)

func TestRenderMigrationStatus_DirtyState(t *testing.T) {
	var out bytes.Buffer

	renderMigrationStatus(&out, &db.MigrationStatus{Version: 7, Dirty: true}, nil)

	text := out.String()
	if !strings.Contains(text, "Status: DIRTY") {
		t.Fatalf("expected dirty status output, got: %q", text)
	}
	if !strings.Contains(text, "[WARN] Database is in a dirty state") {
		t.Fatalf("expected dirty warning output, got: %q", text)
	}
	if !strings.Contains(text, "(Detailed status not available)") {
		t.Fatalf("expected fallback detailed status message, got: %q", text)
	}
}

func TestRenderMigrationStatus_DetailedOutput(t *testing.T) {
	var out bytes.Buffer

	renderMigrationStatus(&out,
		&db.MigrationStatus{Version: 2, Dirty: false},
		&db.DetailedMigrationStatus{
			Applied:    []string{"[000001] init", "[000002] add_users"},
			Pending:    []string{"[000003] add_orders"},
			OutOfOrder: []string{"[000000] legacy_fix"},
			Next:       "000003",
		},
	)

	text := out.String()
	checks := []string{
		"Status: OK",
		"Applied Migrations (2):",
		"[x] [000001] init",
		"Pending Migrations (1):",
		"[ ] [000003] add_orders",
		"[WARN] Out-of-Order Migrations (1):",
		"[!] [000000] legacy_fix (applied before current version)",
		"Next Migration: [000003]",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("expected output to contain %q, got: %q", check, text)
		}
	}
}

func TestRenderMigrationStatus_AlreadyLatest(t *testing.T) {
	var out bytes.Buffer

	renderMigrationStatus(&out,
		&db.MigrationStatus{Version: 5, Dirty: false},
		&db.DetailedMigrationStatus{Next: "Already at latest version"},
	)

	text := out.String()
	if !strings.Contains(text, "Next Migration: Already at latest version") {
		t.Fatalf("expected already-latest message, got: %q", text)
	}
}

func TestRenderMigrationStatus_NilStatus(t *testing.T) {
	var out bytes.Buffer

	renderMigrationStatus(&out, nil, nil)

	text := out.String()
	if !strings.Contains(text, "Migration status unavailable") {
		t.Fatalf("expected unavailable status warning, got: %q", text)
	}
	if !strings.Contains(text, "(Detailed status not available)") {
		t.Fatalf("expected fallback detailed status message, got: %q", text)
	}
}

func TestRenderMigrationFiles_Empty(t *testing.T) {
	var out bytes.Buffer

	renderMigrationFiles(&out, nil)

	text := out.String()
	if !strings.Contains(text, "Migration Files (0): none found") {
		t.Fatalf("expected empty migration files message, got: %q", text)
	}
}

func TestRenderMigrationFiles_SortsFileNames(t *testing.T) {
	var out bytes.Buffer

	renderMigrationFiles(&out, []string{
		`C:\migrations\000010_create_orders.up.sql`,
		`C:\migrations\000001_init.up.sql`,
		`C:\migrations\000002_add_users.up.sql`,
	})

	text := out.String()
	first := strings.Index(text, "000001_init.up.sql")
	second := strings.Index(text, "000002_add_users.up.sql")
	third := strings.Index(text, "000010_create_orders.up.sql")
	if first == -1 || second == -1 || third == -1 {
		t.Fatalf("expected all files in output, got: %q", text)
	}
	if !(first < second && second < third) {
		t.Fatalf("expected files to be sorted, got: %q", text)
	}
}
