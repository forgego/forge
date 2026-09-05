package filter

import (
	"net/http"
	"net/url"
	"testing"
)

func TestParser_ParseQueryParams(t *testing.T) {
	// Use AllowAllSecurity for backward compatibility in tests
	parser := NewParser(WithAllowAllSecurity())

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
	parser := NewParser(WithAllowAllSecurity())

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
	parser := NewParser(WithAllowAllSecurity())

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

// TestParser_SecurityValidation tests security validation
func TestParser_SecurityValidation(t *testing.T) {
	t.Run("nil security denies access", func(t *testing.T) {
		parser := &Parser{security: nil}
		err := parser.validateFieldAccess("username", nil)
		if err == nil {
			t.Error("Expected error when security is nil, got nil")
		}
	})

	t.Run("default security denies all fields", func(t *testing.T) {
		parser := NewParser() // Uses DefaultSecurityConfig by default
		err := parser.validateFieldAccess("username", nil)
		if err == nil {
			t.Error("Expected error with default security config, got nil")
		}
	})

	t.Run("AllowAllSecurity allows all fields", func(t *testing.T) {
		parser := NewParser(WithAllowAllSecurity())
		err := parser.validateFieldAccess("username", nil)
		if err != nil {
			t.Errorf("Expected no error with AllowAllSecurity, got: %v", err)
		}
	})

	t.Run("specific allowed fields", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.AllowedFields["User"] = []string{"username", "email"}
		parser := NewParser(WithSecurity(config))
		
		// Allowed field
		err := parser.validateFieldAccess("username", nil)
		if err != nil {
			t.Errorf("Expected username to be allowed, got: %v", err)
		}
		
		// Another allowed field
		err = parser.validateFieldAccess("email", nil)
		if err != nil {
			t.Errorf("Expected email to be allowed, got: %v", err)
		}
		
		// Non-allowed field
		err = parser.validateFieldAccess("password", nil)
		if err == nil {
			t.Error("Expected password to be denied, got nil")
		}
	})

	t.Run("wildcard allowed fields", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.AllowedFields["User"] = []string{"*"}
		parser := NewParser(WithSecurity(config))
		
		err := parser.validateFieldAccess("any_field", nil)
		if err != nil {
			t.Errorf("Expected any_field to be allowed with wildcard, got: %v", err)
		}
	})
}

// TestParser_MaxDepthEnforcement tests max depth enforcement
func TestParser_MaxDepthEnforcement(t *testing.T) {
	t.Run("default config has max depth 3", func(t *testing.T) {
		config := DefaultSecurityConfig()
		if config.MaxJoinDepth != 3 {
			t.Errorf("Expected MaxJoinDepth 3, got %d", config.MaxJoinDepth)
		}
	})

	t.Run("AllowAllSecurity has higher max depth", func(t *testing.T) {
		config := AllowAllSecurity()
		if config.MaxJoinDepth != 10 {
			t.Errorf("Expected MaxJoinDepth 10, got %d", config.MaxJoinDepth)
		}
	})
}
