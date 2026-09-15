package docker

import "testing"

func TestLiteralPostgresOptsUsesDatabaseURL(t *testing.T) {
	t.Setenv("FORGE_TEST_DATABASE_URL", "postgres://forge_test:secret@dbhost:6543/base?sslmode=require&connect_timeout=5")
	t.Setenv("POSTGRES_USER", "ignored")
	opts := PostgresOpts{User: "postgres", Password: "123", Host: "localhost", Port: "5432", DBName: "caller_database", UseDirect: true}
	want := "postgres://forge_test:secret@dbhost:6543/caller_database?sslmode=require&connect_timeout=5"
	if got := opts.DSN(); got != want {
		t.Fatalf("DSN = %q, want %q", got, want)
	}
}
