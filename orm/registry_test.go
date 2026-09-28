package orm

import "testing"

func TestRegisterAggregateAndBuildAggregate(t *testing.T) {
	RegisterAggregate("TotalPrice", "SUM", func(field string) Aggregate {
		return Aggregate{Field: field}
	})

	aggregate, ok := BuildAggregate(" totalprice ", "price")
	if !ok {
		t.Fatal("expected custom aggregate to be registered")
	}
	if aggregate.Name != "totalprice" {
		t.Fatalf("expected normalized name totalprice, got %q", aggregate.Name)
	}
	if aggregate.Field != "price" {
		t.Fatalf("expected field price, got %q", aggregate.Field)
	}
	if aggregate.Func != "SUM" {
		t.Fatalf("expected func SUM, got %q", aggregate.Func)
	}
}

func TestRegisterAnnotationAndBuildAnnotation(t *testing.T) {
	RegisterAnnotation("HasDiscount", func(args ...interface{}) AnnotationExpr {
		if len(args) != 1 {
			t.Fatalf("expected one arg, got %d", len(args))
		}
		threshold, ok := args[0].(int)
		if !ok {
			t.Fatalf("expected int threshold arg, got %T", args[0])
		}
		return AnnotationExpr{
			Expr: NewFieldQueryExpr("discount_percent", OpGreaterOrEqual, threshold),
		}
	})

	annotation, ok := BuildAnnotation(" hasdiscount ", 20)
	if !ok {
		t.Fatal("expected custom annotation to be registered")
	}
	if annotation.Name != "hasdiscount" {
		t.Fatalf("expected normalized name hasdiscount, got %q", annotation.Name)
	}
	sql, args, _ := annotation.Expr.ToSQL(1)
	if sql != "discount_percent >= $1" {
		t.Fatalf("unexpected SQL: %s", sql)
	}
	if len(args) != 1 || args[0] != 20 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestRegisterQueryExprAndBuildQueryExpr(t *testing.T) {
	RegisterQueryExpr("ActiveOnly", func(args ...interface{}) QueryExpr {
		if len(args) != 1 {
			t.Fatalf("expected one arg, got %d", len(args))
		}
		field, ok := args[0].(string)
		if !ok {
			t.Fatalf("expected string field arg, got %T", args[0])
		}
		return NewFieldQueryExpr(field, OpEquals, true)
	})

	expr, ok := BuildQueryExpr(" activeonly ", "is_active")
	if !ok {
		t.Fatal("expected custom query expression to be registered")
	}
	sql, args, _ := expr.ToSQL(1)
	if sql != "is_active = $1" {
		t.Fatalf("unexpected SQL: %s", sql)
	}
	if len(args) != 1 || args[0] != true {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestRegistryBuilders_ReturnFalseForUnknownNames(t *testing.T) {
	if _, ok := BuildAggregate("not_registered", "price"); ok {
		t.Fatal("expected aggregate builder lookup to fail for unknown name")
	}
	if _, ok := BuildAnnotation("not_registered"); ok {
		t.Fatal("expected annotation builder lookup to fail for unknown name")
	}
	if _, ok := BuildQueryExpr("not_registered"); ok {
		t.Fatal("expected query expression builder lookup to fail for unknown name")
	}
}
