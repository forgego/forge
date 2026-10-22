package schema

import "testing"

func TestDBCollationOptionSetsFieldCollation(t *testing.T) {
	field := StringField("title", DBCollation("en_US.utf8"))
	if field.DBCollation != "en_US.utf8" {
		t.Fatalf("collation = %q", field.DBCollation)
	}
}

func TestDBCommentOptionSetsFieldComment(t *testing.T) {
	field := StringField("title", DBComment("Public headline"))
	if field.DBComment != "Public headline" {
		t.Fatalf("comment = %q", field.DBComment)
	}
}
