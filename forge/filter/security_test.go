package filter

import (
	"testing"
)

func TestSecurityConfig_ValidateComplexity(t *testing.T) {
	config := NewSecurityConfig()
	config.MaxConditions = 5
	config.MaxORBranches = 3
	config.MaxJoinDepth = 2

	// Valid filter
	validNode := NewAndNode(
		NewFieldNode("username", "contains", "john"),
		NewFieldNode("email", "exact", "test@example.com"),
	)

	if err := config.ValidateComplexity(validNode); err != nil {
		t.Errorf("Valid filter should not error: %v", err)
	}

	// Too many conditions
	manyConditions := NewAndNode()
	for i := 0; i < 10; i++ {
		manyConditions.AddChild(NewFieldNode("field", "exact", i))
	}

	if err := config.ValidateComplexity(manyConditions); err == nil {
		t.Error("Filter with too many conditions should error")
	}
}

func TestSecurityConfig_ValidateCost(t *testing.T) {
	config := NewSecurityConfig()
	config.CostThreshold = 50

	// Valid cost
	if err := config.ValidateCost(30); err != nil {
		t.Errorf("Valid cost should not error: %v", err)
	}

	// Cost exceeds threshold
	if err := config.ValidateCost(100); err == nil {
		t.Error("Cost exceeding threshold should error")
	}
}

func TestMaskParameters(t *testing.T) {
	params := map[string]interface{}{
		"username": "john_doe",
		"email":    "test@example.com",
		"id":       123,
	}

	masked := MaskParameters(params)

	if masked["username"] == params["username"] {
		t.Error("Username should be masked")
	}

	if masked["id"] == params["id"] {
		t.Error("ID should be masked")
	}
}
