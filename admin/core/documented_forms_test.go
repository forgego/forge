package core

import (
	"encoding/json"
	"testing"
)

// Fieldset configuration retains grouping data, but form rendering does not use it yet.
func TestAdminFormFieldsetsGroupConfiguredFields(t *testing.T) {
	cfg := Config[struct{}]{Fieldsets: []Fieldset[struct{}]{NewFieldset[struct{}]("Details", "name", "description").WithCollapsed(true).WithDescription("Main fields")}}
	if len(cfg.Fieldsets) != 1 || cfg.Fieldsets[0].Name != "Details" || len(cfg.Fieldsets[0].Fields) != 2 || !cfg.Fieldsets[0].Collapsed || cfg.Fieldsets[0].Description != "Main fields" {
		t.Fatalf("fieldsets=%+v", cfg.Fieldsets)
	}
}

// Inline relation configuration reaches metadata consumed by the form.
func TestAdminInlineRelationExposesChildFields(t *testing.T) {
	cfg := Config[struct{}]{InlineRelations: map[string]InlineRelationConfig{"items": {Type: "one_to_many", RelatedModel: "Item", RelatedField: "order_id", Fields: []string{"sku", "quantity"}}}}
	metadata := buildRelationsMetadata(testSchema{}, &cfg)
	if len(metadata) != 1 || metadata[0].Inline == nil {
		t.Fatalf("inline metadata=%+v", metadata)
	}
	data, err := json.Marshal(metadata[0].Inline)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got["related_model"] != "Item" || got["type"] != "one_to_many" {
		t.Fatalf("inline relation=%s", data)
	}
	fields, ok := got["fields"].([]any)
	if !ok || len(fields) != 2 {
		t.Fatalf("fields=%v", got["fields"])
	}
}
