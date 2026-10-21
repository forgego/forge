package parse

import (
	"strings"
	"testing"
)

func TestDestructiveOperations(t *testing.T) {
	cases := []struct {
		sql  string
		want string
	}{
		{`DROP TABLE IF EXISTS widgets;`, "DropTable widgets"},
		{`ALTER TABLE widgets DROP COLUMN name;`, "DropColumn widgets.name"},
		{`DELETE FROM widgets WHERE id = 1;`, "DELETE FROM"},
		{`TRUNCATE TABLE widgets;`, "TRUNCATE TABLE"},
		{`DROP SCHEMA old CASCADE;`, "DROP SCHEMA"},
		{`CREATE TABLE widgets (name TEXT DEFAULT 'DROP TABLE');`, ""},
		{`-- DROP TABLE widgets;
CREATE TABLE widgets (id INTEGER);`, ""},
	}
	for _, tc := range cases {
		t.Run(strings.ReplaceAll(tc.sql, " ", "_"), func(t *testing.T) {
			operations, err := DestructiveOperations(tc.sql)
			if err != nil {
				t.Fatal(err)
			}
			if tc.want == "" {
				if len(operations) != 0 {
					t.Fatalf("unexpected destructive operations: %v", operations)
				}
				return
			}
			if len(operations) != 1 || operations[0] != tc.want {
				t.Fatalf("got %v, want %q", operations, tc.want)
			}
		})
	}
}
