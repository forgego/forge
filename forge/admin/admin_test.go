package admin

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "creates non-nil config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig[struct{}]()
			if config == nil {
				t.Error("NewConfig() returned nil")
			}
		})
	}
}

func TestDefaultSite(t *testing.T) {
	if DefaultSite == nil {
		t.Error("DefaultSite is nil")
	}
	if DefaultSite.Name != "default" {
		t.Errorf("DefaultSite.Name = %q, want %q", DefaultSite.Name, "default")
	}
}

func TestFilterTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"FilterTypeText", FilterTypeText, "text"},
		{"FilterTypeString", FilterTypeString, "text"},
		{"FilterTypeNumber", FilterTypeNumber, "number"},
		{"FilterTypeBoolean", FilterTypeBoolean, "boolean"},
		{"FilterTypeChoice", FilterTypeChoice, "choice"},
		{"FilterTypeDateRange", FilterTypeDateRange, "date_range"},
		{"FilterTypeRelation", FilterTypeRelation, "relation"},
		{"FilterTypeRelated", FilterTypeRelated, "relation"},
		{"FilterTypeObjectID", FilterTypeObjectID, "object_id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestInlineTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"InlineTypeOneToMany", InlineTypeOneToMany, "one_to_many"},
		{"InlineTypeOneToOne", InlineTypeOneToOne, "one_to_one"},
		{"InlineTypeManyToMany", InlineTypeManyToMany, "many_to_many"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestNewFieldset(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
	}{
		{
			name:   "empty fieldset",
			fields: []string{},
		},
		{
			name:   "single field",
			fields: []string{"name"},
		},
		{
			name:   "multiple fields",
			fields: []string{"name", "email", "created_at"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldset := NewFieldset[struct{}](tt.name, tt.fields...)
			if fieldset.Name != tt.name {
				t.Errorf("Fieldset.Name = %q, want %q", fieldset.Name, tt.name)
			}
			if len(fieldset.Fields) != len(tt.fields) {
				t.Errorf("Fieldset.Fields length = %d, want %d", len(fieldset.Fields), len(tt.fields))
			}
		})
	}
}