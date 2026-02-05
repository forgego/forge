package filter

import (
	"net/http"
	"net/url"
	"testing"
)

func TestParser_ParseQueryParams(t *testing.T) {
	parser := NewParser(NewSecurityConfig())

	req := &http.Request{
		URL: &url.URL{
			RawQuery: "username__contains=john&is_active=true",
		},
	}

	filters, err := parser.ParseQueryParams(req, nil)
	if err != nil {
		t.Fatalf("Failed to parse query params: %v", err)
	}

	if len(filters) != 2 {
		t.Errorf("Expected 2 filters, got %d", len(filters))
	}

	if filters["username__contains"] != "john" {
		t.Error("Username filter value incorrect")
	}
}

func TestParser_ParseParamName(t *testing.T) {
	parser := NewParser(NewSecurityConfig())

	fieldPath, lookup, err := parser.parseParamName("username__contains")
	if err != nil {
		t.Fatalf("Failed to parse param name: %v", err)
	}

	if fieldPath != "username" {
		t.Errorf("Expected field path 'username', got '%s'", fieldPath)
	}

	if lookup != "contains" {
		t.Errorf("Expected lookup 'contains', got '%s'", lookup)
	}

	// Test simple field name
	fieldPath, lookup, err = parser.parseParamName("username")
	if err != nil {
		t.Fatalf("Failed to parse simple param name: %v", err)
	}

	if fieldPath != "username" {
		t.Errorf("Expected field path 'username', got '%s'", fieldPath)
	}

	if lookup != "exact" {
		t.Errorf("Expected default lookup 'exact', got '%s'", lookup)
	}
}

func TestParser_ParseValue(t *testing.T) {
	parser := NewParser(NewSecurityConfig())

	// Test string value
	value, err := parser.parseValue("john", "contains")
	if err != nil {
		t.Fatalf("Failed to parse string value: %v", err)
	}
	if value != "john" {
		t.Errorf("Expected 'john', got '%v'", value)
	}

	// Test IN value
	value, err = parser.parseValue("active,pending", "in")
	if err != nil {
		t.Fatalf("Failed to parse IN value: %v", err)
	}
	values, ok := value.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", value)
	}
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}

	// Test numeric value
	value, err = parser.parseValue("18", "gt")
	if err != nil {
		t.Fatalf("Failed to parse numeric value: %v", err)
	}
	if value == nil {
		t.Error("Numeric value is nil")
	}
}
