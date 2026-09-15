package testhelpers

import (
	"fmt"
	"strings"
	"testing"
)

func TestRedactDSN(t *testing.T) {
	for _, tc := range []struct{ name, dsn, password string }{
		{"URL", "postgres://user:private-secret@localhost/db?sslmode=require", "private-secret"},
		{"escaped URL", "postgresql://user:private%40secret@localhost/db", "private%40secret"},
		{"URL query", "postgres://user@localhost/db?password=private-secret", "private-secret"},
		{"keyword", "host=localhost password=private-secret dbname=test", "private-secret"},
		{"quoted keyword", "host=localhost password = 'private secret' dbname=test", "private secret"},
		{"escaped keyword", `host=localhost password=private\ secret dbname=test`, `private\ secret`},
		{"escaped quote", `password='private\'secret' host=localhost`, "secret"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := fmt.Errorf("connection retries exhausted. DSN: %s", redactDSN(tc.dsn))
			if strings.Contains(err.Error(), tc.password) {
				t.Fatal("error leaked password")
			}
			if !strings.Contains(err.Error(), "***") && !strings.Contains(err.Error(), "xxxxx") {
				t.Fatalf("missing redaction marker: %s", err)
			}
		})
	}
}
