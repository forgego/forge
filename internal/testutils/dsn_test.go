package testutils

import (
	"net/url"
	"testing"

	"github.com/lib/pq"
)

func TestPostgresURL_KeepsAwkwardPasswordIntact(t *testing.T) {
	const password = `p@ss word/'\"?#%`
	dsn := postgresURL("db.internal", 6543, "app user", password, "my db")

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse %q: %v", dsn, err)
	}
	got, _ := u.User.Password()
	if got != password {
		t.Fatalf("password = %q, want %q", got, password)
	}
	if u.User.Username() != "app user" || u.Host != "db.internal:6543" || u.Path != "/my db" {
		t.Fatalf("unexpected URL components: %#v", u)
	}

	if u.Query().Get("sslmode") != "disable" {
		t.Fatalf("sslmode = %q, want disable", u.Query().Get("sslmode"))
	}

	// lib/pq, the driver the tests open, accepts the URL.
	if _, err := pq.NewConnector(dsn); err != nil {
		t.Fatalf("pq.NewConnector: %v", err)
	}
}
