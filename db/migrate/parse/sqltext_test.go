package parse

import "testing"

func TestDefaultExpression(t *testing.T) {
	for _, tc := range []struct {
		column string
		want   interface{}
	}{
		{` TEXT DEFAULT 'read, then write it''s' NOT NULL`, "read, then write it's"},
		{` TEXT DEFAULT 'NOT NULL'`, "NOT NULL"},
		{` INTEGER DEFAULT (1 + 2)`, "(1 + 2)"},
		{` TIMESTAMP DEFAULT now() NOT NULL`, "now()"},
		{` REAL DEFAULT 1.500000`, 1.5},
		{` TEXT DEFAULT '1'`, "1"},
		{` INTEGER DEFAULT 0`, int64(0)},
	} {
		field, err := NewTableParser().parseColumnDefinition(`"c"` + tc.column)
		if err != nil {
			t.Fatalf("%s: %v", tc.column, err)
		}
		if field.Default != tc.want {
			t.Errorf("%s: default %#v, want %#v", tc.column, field.Default, tc.want)
		}
	}

	field, err := NewTableParser().parseColumnDefinition(`"c" TEXT DEFAULT 'NOT NULL UNIQUE'`)
	if err != nil {
		t.Fatal(err)
	}
	if field.Required || field.Options["unique"] != nil {
		t.Errorf("keywords inside a default literal set constraints: %+v", field)
	}
	if _, ok := defaultExpression(` TEXT NOT NULL`); ok {
		t.Error("found a default in a column without one")
	}
}

func TestSplitColumnDefinitionsKeepsQuotedCommas(t *testing.T) {
	got := splitColumnDefinitions(`"a" TEXT DEFAULT 'x, y', "b" TEXT DEFAULT ')'`)
	if len(got) != 2 || got[0] != `"a" TEXT DEFAULT 'x, y'` || got[1] != `"b" TEXT DEFAULT ')'` {
		t.Errorf("split %q", got)
	}
}
