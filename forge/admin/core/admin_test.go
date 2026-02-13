package core

import (
	"context"
	"testing"
)

func TestNewAction(t *testing.T) {
	tests := []struct {
		name    string
		action  string
		label   string
	}{
		{
			name:   "basic action",
			action: "delete",
			label:  "Delete Selected",
		},
		{
			name:   "empty name",
			action: "",
			label:  "Empty Action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(ctx context.Context, instances []*struct{}) error { return nil }
			action := NewAction[struct{}](tt.action, tt.label, handler)
			if action.Name != tt.action {
				t.Errorf("Action.Name = %q, want %q", action.Name, tt.action)
			}
			if action.Label != tt.label {
				t.Errorf("Action.Label = %q, want %q", action.Label, tt.label)
			}
		})
	}
}

func TestAction_WithDescription(t *testing.T) {
	action := Action[struct{}]{Name: "test", Label: "Test"}
	result := action.WithDescription("Test description")
	if result.Description != "Test description" {
		t.Errorf("Action.Description = %q, want %q", result.Description, "Test description")
	}
}

func TestAction_WithPermissions(t *testing.T) {
	action := Action[struct{}]{Name: "test", Label: "Test"}
	result := action.WithPermissions("admin", "delete")
	if len(result.Permissions) != 2 {
		t.Errorf("Action.Permissions length = %d, want 2", len(result.Permissions))
	}
}

func TestAction_WithConfirmation(t *testing.T) {
	action := Action[struct{}]{Name: "test", Label: "Test"}
	result := action.WithConfirmation("Are you sure?")
	if result.Confirmation != "Are you sure?" {
		t.Errorf("Action.Confirmation = %q, want %q", result.Confirmation, "Are you sure?")
	}
}

func TestAction_WithDangerous(t *testing.T) {
	action := Action[struct{}]{Name: "test", Label: "Test"}
	result := action.WithDangerous(true)
	if !result.Dangerous {
		t.Error("Action.Dangerous should be true")
	}
}

func TestAction_WithIcon(t *testing.T) {
	action := Action[struct{}]{Name: "test", Label: "Test"}
	result := action.WithIcon("trash")
	if result.Icon != "trash" {
		t.Errorf("Action.Icon = %q, want %q", result.Icon, "trash")
	}
}

func TestAction_WithUIComponent(t *testing.T) {
	action := Action[struct{}]{Name: "test", Label: "Test"}
	result := action.WithUIComponent("ConfirmDialog")
	if result.UIComponent != "ConfirmDialog" {
		t.Errorf("Action.UIComponent = %q, want %q", result.UIComponent, "ConfirmDialog")
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

func TestFieldset_WithCollapsed(t *testing.T) {
	fieldset := Fieldset[struct{}]{Name: "test"}
	result := fieldset.WithCollapsed(true)
	if !result.Collapsed {
		t.Error("Fieldset.Collapsed should be true")
	}
}

func TestFieldset_WithDescription(t *testing.T) {
	fieldset := Fieldset[struct{}]{Name: "test"}
	result := fieldset.WithDescription("Test description")
	if result.Description != "Test description" {
		t.Errorf("Fieldset.Description = %q, want %q", result.Description, "Test description")
	}
}

func TestComputed(t *testing.T) {
	method := Computed("get_full_name")
	if method.Path() != "get_full_name" {
		t.Errorf("Method.Path() = %q, want %q", method.Path(), "get_full_name")
	}
}

func TestMethod_Path(t *testing.T) {
	method := Method("custom_method")
	if method.Path() != "custom_method" {
		t.Errorf("Method.Path() = %q, want %q", method.Path(), "custom_method")
	}
}

func TestRadioLayout_Constants(t *testing.T) {
	if RadioHorizontal != "horizontal" {
		t.Errorf("RadioHorizontal = %q, want %q", RadioHorizontal, "horizontal")
	}
	if RadioVertical != "vertical" {
		t.Errorf("RadioVertical = %q, want %q", RadioVertical, "vertical")
	}
}

func TestConfig_Defaults(t *testing.T) {
	config := Config[struct{}]{}
	// Test that a new config has zero values
	if config.ListPerPage != 0 {
		t.Errorf("Config.ListPerPage = %d, want 0", config.ListPerPage)
	}
	if config.PageType != "" {
		t.Errorf("Config.PageType = %q, want empty", config.PageType)
	}
}

func TestInlineRelationConfig_Fields(t *testing.T) {
	config := InlineRelationConfig{
		Type:         "one_to_many",
		Label:        "Items",
		Fields:       []string{"name", "quantity"},
		RelatedModel: "Item",
		RelatedField: "order_id",
	}
	if config.Type != "one_to_many" {
		t.Errorf("Type = %q, want %q", config.Type, "one_to_many")
	}
	if config.Label != "Items" {
		t.Errorf("Label = %q, want %q", config.Label, "Items")
	}
	if len(config.Fields) != 2 {
		t.Errorf("Fields length = %d, want 2", len(config.Fields))
	}
}

func TestInlineConfig_Fields(t *testing.T) {
	config := InlineConfig{
		ListDisplay: []string{"name", "email"},
	}
	if len(config.ListDisplay) != 2 {
		t.Errorf("ListDisplay length = %d, want 2", len(config.ListDisplay))
	}
}

func TestFilter_Type(t *testing.T) {
	filter := Filter[struct{}]{
		Name:    "status",
		Label:   "Status",
		Type:    "choice",
		Choices: []Choice{{Value: "active", Label: "Active"}},
	}
	if filter.Name != "status" {
		t.Errorf("Filter.Name = %q, want %q", filter.Name, "status")
	}
	if filter.Type != "choice" {
		t.Errorf("Filter.Type = %q, want %q", filter.Type, "choice")
	}
}

func TestChoice_Fields(t *testing.T) {
	choice := Choice{
		Value: "active",
		Label: "Active",
	}
	if choice.Value != "active" {
		t.Errorf("Choice.Value = %q, want %q", choice.Value, "active")
	}
	if choice.Label != "Active" {
		t.Errorf("Choice.Label = %q, want %q", choice.Label, "Active")
	}
}
