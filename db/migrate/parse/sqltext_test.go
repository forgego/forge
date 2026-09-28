package parse

import "testing"

func TestDefaultExpression(t *testing.T) {
	for _, tc := range []struct {
		column string
		want   interface{}
		// raw is the expected verbatim DB default of a non-literal expression.
		raw string
	}{
		{column: ` TEXT DEFAULT 'read, then write it''s' NOT NULL`, want: "read, then write it's"},
		{column: ` TEXT DEFAULT 'NOT NULL'`, want: "NOT NULL"},
		{column: ` INTEGER DEFAULT (1 + 2)`, raw: "(1 + 2)"},
		{column: ` TIMESTAMP DEFAULT now() NOT NULL`, raw: "now()"},
		{column: ` REAL DEFAULT 1.500000`, want: 1.5},
		{column: ` TEXT DEFAULT '1'`, want: "1"},
		{column: ` INTEGER DEFAULT 0`, want: int64(0)},
		{column: ` INTEGER DEFAULT -1 NOT NULL`, want: int64(-1)},
		{column: ` BOOLEAN DEFAULT FALSE`, want: false},
		{column: ` JSONB NOT NULL DEFAULT '{}'::jsonb`, raw: "'{}'::jsonb"},
		{column: ` TEXT DEFAULT 'x' || 'y' UNIQUE`, raw: "'x' || 'y'"},
		{column: ` TEXT DEFAULT 'Mixed Case'::text NOT NULL`, raw: "'Mixed Case'::text"},
		{column: ` TEXT DEFAULT 'a'::character varying COLLATE "C"`, raw: "'a'::character varying"},
		{column: ` TEXT DEFAULT NULL`, raw: "NULL"},
		{column: ` TEXT DEFAULT 'it''s' CHECK (length(c) > 0)`, want: "it's"},
		{column: ` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL`, raw: "CURRENT_TIMESTAMP"},
		{column: ` TEXT DEFAULT ('x' || 'y') CONSTRAINT c_ok CHECK (c <> '')`, raw: "('x' || 'y')"},
	} {
		field, err := NewTableParser().parseColumnDefinition(`"c"` + tc.column)
		if err != nil {
			t.Fatalf("%s: %v", tc.column, err)
		}
		if field.Default != tc.want {
			t.Errorf("%s: default %#v, want %#v", tc.column, field.Default, tc.want)
		}
		raw, _ := field.Options[DBDefaultOption].(string)
		if raw != tc.raw {
			t.Errorf("%s: DB default %q, want %q", tc.column, raw, tc.raw)
		}
	}

	field, err := NewTableParser().parseColumnDefinition(`"c" BOOLEAN DEFAULT (1 IS NOT NULL) UNIQUE`)
	if err != nil {
		t.Fatal(err)
	}
	if field.Required || field.Options["unique"] != true {
		t.Errorf("an expression default sets constraints: %+v", field)
	}

	field, err = NewTableParser().parseColumnDefinition(`"c" TEXT DEFAULT 'NOT NULL UNIQUE'`)
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
