package filters

import (
	"testing"

	"github.com/forgego/forge/filter"
)

func TestDocumentedFilterConstructors(t *testing.T) {
	type product struct{}
	var objects = []filter.Filter[product]{
		NewCharFilter[product]("name"),
		NewNumberFilter[product]("price"),
		NewBooleanFilter[product]("available"),
		NewChoiceFilter[product]("status", []Choice{{Value: "active", Label: "Active"}}),
		NewDateFilter[product]("created_at"),
	}
	for _, object := range objects {
		if object.GetFieldPath() == "" {
			t.Fatal("filter constructor lost its field path")
		}
	}
	if value, err := objects[1].Parse("12.5"); err != nil || value != 12.5 {
		t.Fatalf("number parsing: value=%v err=%v", value, err)
	}
	if _, err := objects[3].Parse("missing"); err == nil {
		t.Fatal("choice filter accepted an unlisted value")
	}
}
