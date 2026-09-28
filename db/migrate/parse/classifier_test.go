package parse

import "testing"

// TestClassifyDropColumnByStructure checks that only ALTER TABLE .. DROP
// COLUMN is a drop: a literal that contains "DROP COLUMN" must not turn an
// ADD COLUMN or ADD CONSTRAINT into one, or it would be missing from state.
func TestClassifyDropColumnByStructure(t *testing.T) {
	c := NewClassifier()
	cases := []struct {
		stmt string
		want StatementKind
	}{
		{`ALTER TABLE authors DROP COLUMN "note";`, StmtDropColumn},
		{`alter table "authors" drop column note;`, StmtDropColumn},
		{`ALTER TABLE IF EXISTS authors DROP COLUMN IF EXISTS note;`, StmtDropColumn},
		{`ALTER TABLE authors ADD COLUMN "note" TEXT DEFAULT 'drop column';`, StmtAlterTable},
		{`ALTER TABLE authors ADD CONSTRAINT c CHECK (note <> 'DROP COLUMN');`, StmtAlterTable},
	}
	for _, tc := range cases {
		if got := c.Classify(tc.stmt); got != tc.want {
			t.Errorf("Classify(%q) = %v, want %v", tc.stmt, got, tc.want)
		}
	}
}
