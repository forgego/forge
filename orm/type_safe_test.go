package orm

import (
	"testing"
)

// TestUnifiedAPI tests that the same methods work with both strings and FieldExpression
func TestUnifiedAPI(t *testing.T) {
	// Test that FieldExpression implements FieldPath
	var _ FieldPath = NewField[string]("name", "users")

	// Test that OrderField implements OrderFieldSpec
	var _ OrderFieldSpec = OrderField{Field: "name", Ascending: true}

	// Test that OrderFieldExpr implements OrderFieldSpec
	fieldExpr := NewField[string]("name", "users")
	orderExpr := fieldExpr.Asc()
	var _ OrderFieldSpec = orderExpr

	// Test that RelationExpression implements RelationPath
	var _ RelationPath = NewRelationExpression("author")

	// Test ExtractPathFromAny with string
	if path := ExtractPathFromAny("name"); path != "name" {
		t.Errorf("ExtractPathFromAny(string) = %v, want %v", path, "name")
	}

	// Test ExtractPathFromAny with FieldExpression
	field := NewField[string]("email", "users")
	if path := ExtractPathFromAny(field); path != "email" {
		t.Errorf("ExtractPathFromAny(FieldExpression) = %v, want %v", path, "email")
	}

	// Test extractOrderFieldPath with OrderField
	orderField := OrderField{Field: "name", Ascending: true}
	if path := extractOrderFieldPath(orderField); path != "name" {
		t.Errorf("extractOrderFieldPath(OrderField) = %v, want %v", path, "name")
	}

	// Test extractOrderFieldPath with OrderFieldExpr
	if path := extractOrderFieldPath(orderExpr); path != "name" {
		t.Errorf("extractOrderFieldPath(OrderFieldExpr) = %v, want %v", path, "name")
	}

	// Test extractOrderFieldAscending
	if asc := extractOrderFieldAscending(orderField); !asc {
		t.Errorf("extractOrderFieldAscending(OrderField) = %v, want %v", asc, true)
	}
	if asc := extractOrderFieldAscending(orderExpr); !asc {
		t.Errorf("extractOrderFieldAscending(OrderFieldExpr) = %v, want %v", asc, true)
	}

	// Test extractRelationPathFromAny with string
	if path := extractRelationPathFromAny("author"); path != "author" {
		t.Errorf("extractRelationPathFromAny(string) = %v, want %v", path, "author")
	}

	// Test extractRelationPathFromAny with RelationExpression
	relExpr := NewRelationExpression("author")
	if path := extractRelationPathFromAny(relExpr); path != "author" {
		t.Errorf("extractRelationPathFromAny(RelationExpression) = %v, want %v", path, "author")
	}
}

// TestAscDescMethods tests the .Asc() and .Desc() methods on FieldExpression
func TestAscDescMethods(t *testing.T) {
	field := NewField[string]("name", "users")

	// Test Asc()
	ascOrder := field.Asc()
	if ascOrder.GetFieldPath() != "name" {
		t.Errorf("Asc().GetFieldPath() = %v, want %v", ascOrder.GetFieldPath(), "name")
	}
	if !ascOrder.IsAscending() {
		t.Error("Asc().IsAscending() = false, want true")
	}

	// Test Desc()
	descOrder := field.Desc()
	if descOrder.GetFieldPath() != "name" {
		t.Errorf("Desc().GetFieldPath() = %v, want %v", descOrder.GetFieldPath(), "name")
	}
	if descOrder.IsAscending() {
		t.Error("Desc().IsAscending() = true, want false")
	}
}
