package testhelpers

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestDBGateContract(t *testing.T) {
	// Test testDatabaseURL branches
	t.Run("testDatabaseURL", func(t *testing.T) {
		const defaultURL = "postgres://default:5432/db"
		const customURL = "postgres://custom:5432/db"

		t.Run("uses custom URL when set", func(t *testing.T) {
			t.Setenv("FORGE_TEST_DATABASE_URL", customURL)
			if got := testDatabaseURL(defaultURL); got != customURL {
				t.Errorf("expected %q, got %q", customURL, got)
			}
		})

		t.Run("falls back to default URL when unset", func(t *testing.T) {
			t.Setenv("FORGE_TEST_DATABASE_URL", "")
			if got := testDatabaseURL(defaultURL); got != defaultURL {
				t.Errorf("expected %q, got %q", defaultURL, got)
			}
		})
	})

	// Test derived DSN keeps query parameters from FORGE_TEST_DATABASE_URL
	t.Run("derived DSN keeps query parameters", func(t *testing.T) {
		const customURL = "postgres://u:p@remotehost:5432/customdb?sslmode=require&connect_timeout=5"
		t.Setenv("FORGE_TEST_DATABASE_URL", customURL)

		opts := LocalPostgresOpts("gate_test")
		dsn := opts.DSN()

		if !strings.Contains(dsn, "sslmode=require") {
			t.Errorf("expected dsn to contain sslmode=require, got %q", dsn)
		}
		if !strings.Contains(dsn, "connect_timeout=5") {
			t.Errorf("expected dsn to contain connect_timeout=5, got %q", dsn)
		}
		if !strings.Contains(dsn, "remotehost:5432") {
			t.Errorf("expected dsn to contain remotehost:5432, got %q", dsn)
		}
		if strings.Contains(dsn, "/customdb?") {
			t.Errorf("expected customdb to be swapped for test database name, got %q", dsn)
		}
		if !strings.Contains(dsn, "test_gate_test_") {
			t.Errorf("expected test database name in dsn, got %q", dsn)
		}
	})

	t.Run("derived DSN unchanged without env var", func(t *testing.T) {
		t.Setenv("FORGE_TEST_DATABASE_URL", "")

		opts := LocalPostgresOpts("gate_test_default")
		dsn := opts.DSN()

		if !strings.Contains(dsn, "sslmode=disable") {
			t.Errorf("expected dsn to contain sslmode=disable, got %q", dsn)
		}
		if strings.Contains(dsn, "sslmode=require") {
			t.Errorf("expected dsn not to contain sslmode=require, got %q", dsn)
		}
	})

	// Test env parsing and pure decision functions
	t.Run("requireDB and dbUnavailableAction", func(t *testing.T) {
		t.Run("required when FORGE_REQUIRE_DB=1", func(t *testing.T) {
			t.Setenv("FORGE_REQUIRE_DB", "1")
			if !requireDB() {
				t.Errorf("expected requireDB() to be true")
			}
			if got := dbUnavailableAction(); got != "fatal" {
				t.Errorf("expected dbUnavailableAction() = 'fatal', got %q", got)
			}
		})

		t.Run("not required when FORGE_REQUIRE_DB is unset", func(t *testing.T) {
			t.Setenv("FORGE_REQUIRE_DB", "")
			if requireDB() {
				t.Errorf("expected requireDB() to be false")
			}
			if got := dbUnavailableAction(); got != "skip" {
				t.Errorf("expected dbUnavailableAction() = 'skip', got %q", got)
			}
		})

		t.Run("not required when FORGE_REQUIRE_DB is other value", func(t *testing.T) {
			t.Setenv("FORGE_REQUIRE_DB", "0")
			if requireDB() {
				t.Errorf("expected requireDB() to be false")
			}
			if got := dbUnavailableAction(); got != "skip" {
				t.Errorf("expected dbUnavailableAction() = 'skip', got %q", got)
			}
		})
	})
}

func TestDatabaseURLDerivedDSN(t *testing.T) {
	const customURL = "postgres://user:pass@dbhost:5432/origdb?sslmode=require&connect_timeout=5"
	t.Setenv("FORGE_TEST_DATABASE_URL", customURL)

	opts := LocalPostgresOpts("derived")
	dsn := opts.DSN()

	if !strings.Contains(dsn, "sslmode=require") {
		t.Errorf("expected dsn to contain sslmode=require, got %q", dsn)
	}
	if !strings.Contains(dsn, "connect_timeout=5") {
		t.Errorf("expected dsn to contain connect_timeout=5, got %q", dsn)
	}
	if strings.Contains(dsn, "/origdb?") {
		t.Errorf("expected origdb to be replaced with test database name, got %q", dsn)
	}
	if !strings.Contains(dsn, "test_derived_") {
		t.Errorf("expected test database name in dsn, got %q", dsn)
	}

	derived := DeriveDSN(opts)
	if derived != dsn {
		t.Errorf("expected DeriveDSN(opts) == opts.DSN(), got %q vs %q", derived, dsn)
	}
}

func TestPostgresPrecedenceContract(t *testing.T) {
	t.Setenv("FORGE_TEST_DATABASE_URL", "postgres://u1:p1@h1:6543/base?sslmode=require")
	t.Setenv("POSTGRES_USER", "other")

	t.Run("DefaultPostgresOptsWithTest", func(t *testing.T) {
		opts := DefaultPostgresOptsWithTest("X")
		if opts.User != "u1" {
			t.Errorf("expected User = %q, got %q", "u1", opts.User)
		}
		if opts.Password != "p1" {
			t.Errorf("expected Password = %q, got %q", "p1", opts.Password)
		}
		if opts.Host != "h1" {
			t.Errorf("expected Host = %q, got %q", "h1", opts.Host)
		}
		if opts.Port != "6543" {
			t.Errorf("expected Port = %q, got %q", "6543", opts.Port)
		}
		if !strings.Contains(opts.DBName, "x") {
			t.Errorf("expected DBName to contain per-test name 'x', got %q", opts.DBName)
		}
		dsn := opts.DSN()
		if !strings.Contains(dsn, "postgres://u1:p1@h1:6543/") {
			t.Errorf("expected DSN to contain postgres://u1:p1@h1:6543/, got %q", dsn)
		}
		if !strings.Contains(dsn, "sslmode=require") {
			t.Errorf("expected DSN to contain sslmode=require, got %q", dsn)
		}
		if strings.Contains(dsn, "other") {
			t.Errorf("expected DSN not to contain POSTGRES_USER 'other', got %q", dsn)
		}
		if !strings.Contains(dsn, "testdb_x_") {
			t.Errorf("expected DSN to contain per-test database name, got %q", dsn)
		}
	})

	t.Run("LocalPostgresOpts", func(t *testing.T) {
		opts := LocalPostgresOpts("X")
		if opts.User != "u1" {
			t.Errorf("expected User = %q, got %q", "u1", opts.User)
		}
		if opts.Password != "p1" {
			t.Errorf("expected Password = %q, got %q", "p1", opts.Password)
		}
		if opts.Host != "h1" {
			t.Errorf("expected Host = %q, got %q", "h1", opts.Host)
		}
		if opts.Port != "6543" {
			t.Errorf("expected Port = %q, got %q", "6543", opts.Port)
		}
		if !strings.Contains(opts.DBName, "x") {
			t.Errorf("expected DBName to contain per-test name 'x', got %q", opts.DBName)
		}
		dsn := opts.DSN()
		if !strings.Contains(dsn, "postgres://u1:p1@h1:6543/") {
			t.Errorf("expected DSN to contain postgres://u1:p1@h1:6543/, got %q", dsn)
		}
		if !strings.Contains(dsn, "sslmode=require") {
			t.Errorf("expected DSN to contain sslmode=require, got %q", dsn)
		}
		if strings.Contains(dsn, "other") {
			t.Errorf("expected DSN not to contain POSTGRES_USER 'other', got %q", dsn)
		}
		if !strings.Contains(dsn, "test_x_") {
			t.Errorf("expected DSN to contain per-test database name, got %q", dsn)
		}
	})

	t.Run("Direct-connection DSN builder", func(t *testing.T) {
		literalOpts := PostgresOpts{
			UseDirect: true,
			Host:      "127.0.0.1",
			Port:      "5432",
			User:      "postgres",
			Password:  "123",
			DBName:    "test_custom_x",
		}
		dsn := DirectPostgresDSN(literalOpts)
		if dsn != "postgres://u1:p1@h1:6543/test_custom_x?sslmode=require" {
			t.Errorf("expected DirectPostgresDSN to be %q, got %q",
				"postgres://u1:p1@h1:6543/test_custom_x?sslmode=require", dsn)
		}
		if strings.Contains(dsn, "other") {
			t.Errorf("expected DSN not to contain POSTGRES_USER 'other', got %q", dsn)
		}
	})
}

func TestPostgresOptsDSN_EscapedCredentialsAndIPv6(t *testing.T) {
	t.Setenv("FORGE_TEST_DATABASE_URL", "")
	opts := PostgresOpts{
		User:     "u@x",
		Password: "p:ss/word",
		Host:     "::1",
		Port:     "5432",
		DBName:   "testdb",
	}
	dsn := opts.DSN()
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("failed to parse DSN %q: %v", dsn, err)
	}
	if u.User == nil {
		t.Fatalf("expected User to be set in parsed DSN %q", dsn)
	}
	if got := u.User.Username(); got != "u@x" {
		t.Errorf("expected username %q, got %q", "u@x", got)
	}
	pass, _ := u.User.Password()
	if pass != "p:ss/word" {
		t.Errorf("expected password %q, got %q", "p:ss/word", pass)
	}
	if got := u.Hostname(); got != "::1" {
		t.Errorf("expected hostname %q, got %q", "::1", got)
	}
	if got := u.Port(); got != "5432" {
		t.Errorf("expected port %q, got %q", "5432", got)
	}
}

type mockTB struct {
	testing.TB
	failed   bool
	skipped  bool
	fatalMsg string
	skipMsg  string
}

func (m *mockTB) Helper() {}
func (m *mockTB) Fatalf(format string, args ...any) {
	m.failed = true
	m.fatalMsg = fmt.Sprintf(format, args...)
}
func (m *mockTB) Skipf(format string, args ...any) {
	m.skipped = true
	m.skipMsg = fmt.Sprintf(format, args...)
}

func TestRequirePostgresURL(t *testing.T) {
	t.Run("prefers FORGE_TEST_DATABASE_URL when both set", func(t *testing.T) {
		t.Setenv("FORGE_TEST_DATABASE_URL", "postgres://forge-test:5432/db")
		t.Setenv("DATABASE_URL", "postgres://fallback:5432/db")
		m := &mockTB{}
		url := RequirePostgresURL(m)
		if url != "postgres://forge-test:5432/db" {
			t.Errorf("expected %q, got %q", "postgres://forge-test:5432/db", url)
		}
		if m.failed || m.skipped {
			t.Errorf("expected not failed or skipped")
		}
	})

	t.Run("falls back to DATABASE_URL when FORGE_TEST_DATABASE_URL is unset", func(t *testing.T) {
		t.Setenv("FORGE_TEST_DATABASE_URL", "")
		t.Setenv("DATABASE_URL", "postgres://fallback:5432/db")
		m := &mockTB{}
		url := RequirePostgresURL(m)
		if url != "postgres://fallback:5432/db" {
			t.Errorf("expected %q, got %q", "postgres://fallback:5432/db", url)
		}
		if m.failed || m.skipped {
			t.Errorf("expected not failed or skipped")
		}
	})

	t.Run("fails when both unset and FORGE_REQUIRE_DB=1", func(t *testing.T) {
		t.Setenv("FORGE_TEST_DATABASE_URL", "")
		t.Setenv("DATABASE_URL", "")
		t.Setenv("FORGE_REQUIRE_DB", "1")
		m := &mockTB{}
		url := RequirePostgresURL(m)
		if url != "" {
			t.Errorf("expected empty url, got %q", url)
		}
		if !m.failed {
			t.Errorf("expected m.failed to be true")
		}
		if m.skipped {
			t.Errorf("expected m.skipped to be false")
		}
		if !strings.Contains(m.fatalMsg, "FORGE_REQUIRE_DB=1") {
			t.Errorf("expected fatal message to mention FORGE_REQUIRE_DB=1, got %q", m.fatalMsg)
		}
	})

	t.Run("skips when both unset and FORGE_REQUIRE_DB unset", func(t *testing.T) {
		t.Setenv("FORGE_TEST_DATABASE_URL", "")
		t.Setenv("DATABASE_URL", "")
		t.Setenv("FORGE_REQUIRE_DB", "")
		m := &mockTB{}
		url := RequirePostgresURL(m)
		if url != "" {
			t.Errorf("expected empty url, got %q", url)
		}
		if m.failed {
			t.Errorf("expected m.failed to be false")
		}
		if !m.skipped {
			t.Errorf("expected m.skipped to be true")
		}
		if !strings.Contains(m.skipMsg, "FORGE_REQUIRE_DB=1") {
			t.Errorf("expected skip message to mention FORGE_REQUIRE_DB=1, got %q", m.skipMsg)
		}
	})
}
