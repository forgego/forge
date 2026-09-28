package filter

import (
	"strings"
	"testing"

	"github.com/forgego/forge/orm"
)

func renderASTSQL[T any](t *testing.T, fs *FilterSet[T]) string {
	t.Helper()
	ast := fs.GetAST()
	if ast == nil {
		return ""
	}
	expr, err := fs.astToExpression(ast)
	if err != nil {
		t.Fatalf("failed to convert AST to expression: %v", err)
	}
	builder := orm.NewSQLBuilder()
	sql, _, err := expr.ToSQL(builder)
	if err != nil {
		t.Fatalf("failed to convert expression to SQL: %v", err)
	}
	return sql
}

// 1. OrGroup with two Where predicates -> root is an OR node with 2 leaf children.
func TestOrGroup_TwoWherePredicates(t *testing.T) {
	fs, err := NewFilterSet[MockModel]()
	if err != nil {
		t.Fatalf("Failed to create FilterSet: %v", err)
	}

	fs.OrGroup(func(q *QueryBuilder[MockModel]) {
		q.Where("username").Equals("alice")
		q.Where("email").Equals("bob@example.com")
	})

	ast := fs.GetAST()
	if ast == nil {
		t.Fatal("expected non-nil AST")
	}

	if ast.Op != OpOr {
		t.Fatalf("expected root Op to be %q (or), got %q", OpOr, ast.Op)
	}
	if len(ast.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(ast.Children))
	}
	if !ast.Children[0].IsLeaf() || ast.Children[0].Field != "username" {
		t.Errorf("expected child 0 to be leaf with field 'username', got %+v", ast.Children[0])
	}
	if !ast.Children[1].IsLeaf() || ast.Children[1].Field != "email" {
		t.Errorf("expected child 1 to be leaf with field 'email', got %+v", ast.Children[1])
	}

	sql := renderASTSQL(t, fs)
	if !strings.Contains(sql, " OR ") {
		t.Errorf("expected rendered SQL to contain ' OR ', got: %s", sql)
	}
}

// 2. AndGroup with two predicates -> AND node with 2 children.
func TestAndGroup_TwoPredicates(t *testing.T) {
	fs, err := NewFilterSet[MockModel]()
	if err != nil {
		t.Fatalf("Failed to create FilterSet: %v", err)
	}

	fs.AndGroup(func(q *QueryBuilder[MockModel]) {
		q.Where("username").Equals("alice")
		q.Where("email").Equals("bob@example.com")
	})

	ast := fs.GetAST()
	if ast == nil {
		t.Fatal("expected non-nil AST")
	}

	if ast.Op != OpAnd {
		t.Fatalf("expected root Op to be %q (and), got %q", OpAnd, ast.Op)
	}
	if len(ast.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(ast.Children))
	}
	if !ast.Children[0].IsLeaf() || ast.Children[0].Field != "username" {
		t.Errorf("expected child 0 to be leaf with field 'username', got %+v", ast.Children[0])
	}
	if !ast.Children[1].IsLeaf() || ast.Children[1].Field != "email" {
		t.Errorf("expected child 1 to be leaf with field 'email', got %+v", ast.Children[1])
	}

	sql := renderASTSQL(t, fs)
	if !strings.Contains(sql, " AND ") {
		t.Errorf("expected rendered SQL to contain ' AND ', got: %s", sql)
	}
}

// 3. Root predicate + OrGroup: fs.Where("id").Equals(0); fs.OrGroup(username=alice, email=bob)
// -> AND(id=0, OR(username=alice, email=bob)).
func TestRootPredicate_Plus_OrGroup(t *testing.T) {
	fs, err := NewFilterSet[MockModel]()
	if err != nil {
		t.Fatalf("Failed to create FilterSet: %v", err)
	}

	fs.Where("id").Equals(int64(0))
	fs.OrGroup(func(q *QueryBuilder[MockModel]) {
		q.Where("username").Equals("alice")
		q.Where("email").Equals("bob@example.com")
	})

	ast := fs.GetAST()
	if ast == nil {
		t.Fatal("expected non-nil AST")
	}

	if ast.Op != OpAnd {
		t.Fatalf("expected root Op to be %q (and), got %q", OpAnd, ast.Op)
	}
	if len(ast.Children) != 2 {
		t.Fatalf("expected 2 children at root, got %d", len(ast.Children))
	}

	// Child 0: id=0 leaf
	if !ast.Children[0].IsLeaf() || ast.Children[0].Field != "id" {
		t.Errorf("expected root child 0 to be leaf with field 'id', got %+v", ast.Children[0])
	}

	// Child 1: OR(username=alice, email=bob)
	orNode := ast.Children[1]
	if orNode.Op != OpOr {
		t.Fatalf("expected root child 1 Op to be %q (or), got %q", OpOr, orNode.Op)
	}
	if len(orNode.Children) != 2 {
		t.Fatalf("expected 2 children in OR group, got %d", len(orNode.Children))
	}
	if !orNode.Children[0].IsLeaf() || orNode.Children[0].Field != "username" {
		t.Errorf("expected OR child 0 to be leaf 'username', got %+v", orNode.Children[0])
	}
	if !orNode.Children[1].IsLeaf() || orNode.Children[1].Field != "email" {
		t.Errorf("expected OR child 1 to be leaf 'email', got %+v", orNode.Children[1])
	}

	sql := renderASTSQL(t, fs)
	if !strings.Contains(sql, " OR ") || !strings.Contains(sql, " AND ") {
		t.Errorf("expected rendered SQL to contain both ' OR ' and ' AND ', got: %s", sql)
	}
}

// 4. OrGroup with a single predicate -> OR node with 1 child.
// NewOrNode produces an OR node with 1 child. The predicate must NOT leak
// into root as a bare AND sibling outside the group.
func TestOrGroup_SinglePredicate(t *testing.T) {
	fs, err := NewFilterSet[MockModel]()
	if err != nil {
		t.Fatalf("Failed to create FilterSet: %v", err)
	}

	fs.OrGroup(func(q *QueryBuilder[MockModel]) {
		q.Where("username").Equals("alice")
	})

	ast := fs.GetAST()
	if ast == nil {
		t.Fatal("expected non-nil AST")
	}

	// NewOrNode creates an OpOr node wrapping the single child
	if ast.Op != OpOr {
		t.Fatalf("expected root Op to be %q (or), got %q", OpOr, ast.Op)
	}
	if len(ast.Children) != 1 {
		t.Fatalf("expected 1 child in OR node, got %d", len(ast.Children))
	}
	if !ast.Children[0].IsLeaf() || ast.Children[0].Field != "username" {
		t.Errorf("expected child 0 to be leaf 'username', got %+v", ast.Children[0])
	}
}

// 6. OrFilter inside AndGroup: q.Where("a").Equals(1); q.OrFilter("b").Equals(2); q.Where("c").Equals(3)
// -> AND( OR(a=1, b=2), c=3 ).
func TestOrFilter_Inside_AndGroup(t *testing.T) {
	fs, err := NewFilterSet[MockModel]()
	if err != nil {
		t.Fatalf("Failed to create FilterSet: %v", err)
	}

	fs.AndGroup(func(q *QueryBuilder[MockModel]) {
		q.Where("username").Equals("alice")
		q.OrFilter("email").Equals("bob@example.com")
		q.Where("is_active").Equals(true)
	})

	ast := fs.GetAST()
	if ast == nil {
		t.Fatal("expected non-nil AST")
	}

	if ast.Op != OpAnd {
		t.Fatalf("expected root Op to be %q (and), got %q", OpAnd, ast.Op)
	}
	if len(ast.Children) != 2 {
		t.Fatalf("expected 2 children in root AND node, got %d", len(ast.Children))
	}

	// Child 0: OR(username=alice, email=bob)
	orNode := ast.Children[0]
	if orNode.Op != OpOr {
		t.Fatalf("expected child 0 to be %q (or), got %q", OpOr, orNode.Op)
	}
	if len(orNode.Children) != 2 {
		t.Fatalf("expected 2 children in OR group, got %d", len(orNode.Children))
	}
	if !orNode.Children[0].IsLeaf() || orNode.Children[0].Field != "username" {
		t.Errorf("expected OR child 0 to be leaf 'username', got %+v", orNode.Children[0])
	}
	if !orNode.Children[1].IsLeaf() || orNode.Children[1].Field != "email" {
		t.Errorf("expected OR child 1 to be leaf 'email', got %+v", orNode.Children[1])
	}

	// Child 1: is_active=true leaf
	if !ast.Children[1].IsLeaf() || ast.Children[1].Field != "is_active" {
		t.Errorf("expected child 1 to be leaf 'is_active', got %+v", ast.Children[1])
	}

	sql := renderASTSQL(t, fs)
	if !strings.Contains(sql, " OR ") || !strings.Contains(sql, " AND ") {
		t.Errorf("expected rendered SQL to contain both ' OR ' and ' AND ', got: %s", sql)
	}
}

// 7. Empty group callback -> fs.ast unchanged (nil stays nil).
func TestEmptyGroupCallback(t *testing.T) {
	fs, err := NewFilterSet[MockModel]()
	if err != nil {
		t.Fatalf("Failed to create FilterSet: %v", err)
	}

	fs.OrGroup(func(q *QueryBuilder[MockModel]) {})
	if fs.GetAST() != nil {
		t.Fatalf("expected nil AST after empty OrGroup, got %+v", fs.GetAST())
	}

	fs.AndGroup(func(q *QueryBuilder[MockModel]) {})
	if fs.GetAST() != nil {
		t.Fatalf("expected nil AST after empty AndGroup, got %+v", fs.GetAST())
	}

	// With pre-existing AST, empty group does not modify it
	fs.Where("username").Equals("alice")
	prevAST := fs.GetAST()
	if prevAST == nil {
		t.Fatal("expected non-nil AST after Where")
	}

	fs.OrGroup(func(q *QueryBuilder[MockModel]) {})
	if fs.GetAST() != prevAST {
		t.Fatalf("expected AST to remain unchanged, got %+v", fs.GetAST())
	}
}

// Test Exact/Gt/Lt methods on FilterBuilder
func TestFilterBuilder_Exact_Gt_Lt(t *testing.T) {
	fs, err := NewFilterSet[MockModel]()
	if err != nil {
		t.Fatalf("Failed to create FilterSet: %v", err)
	}

	fs.OrGroup(func(q *QueryBuilder[MockModel]) {
		q.Where("username").Exact("alice")
		q.Where("id").Gt(int64(5))
		q.Where("id").Lt(int64(10))
	})

	ast := fs.GetAST()
	if ast == nil {
		t.Fatal("expected non-nil AST")
	}
	if ast.Op != OpOr {
		t.Fatalf("expected OpOr, got %v", ast.Op)
	}
	if len(ast.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(ast.Children))
	}
	if ast.Children[0].Lookup != "exact" || ast.Children[0].Value != "alice" {
		t.Errorf("expected exact lookup for child 0, got %+v", ast.Children[0])
	}
	if ast.Children[1].Lookup != "gt" || ast.Children[1].Value != int64(5) {
		t.Errorf("expected gt lookup for child 1, got %+v", ast.Children[1])
	}
	if ast.Children[2].Lookup != "lt" || ast.Children[2].Value != int64(10) {
		t.Errorf("expected lt lookup for child 2, got %+v", ast.Children[2])
	}
}

// Test Nested groups inside a group callback
func TestNestedGroups(t *testing.T) {
	fs, err := NewFilterSet[MockModel]()
	if err != nil {
		t.Fatalf("Failed to create FilterSet: %v", err)
	}

	// fs.OrGroup(a=1, AndGroup(b=2, c=3))
	fs.OrGroup(func(q *QueryBuilder[MockModel]) {
		q.Where("username").Exact("alice")
		q.AndGroup(func(q2 *QueryBuilder[MockModel]) {
			q2.Where("email").Exact("bob@example.com")
			q2.Where("is_active").Exact(true)
		})
	})

	ast := fs.GetAST()
	if ast == nil {
		t.Fatal("expected non-nil AST")
	}
	// Root should be OR
	if ast.Op != OpOr {
		t.Fatalf("expected root Op to be OpOr, got %v", ast.Op)
	}
	if len(ast.Children) != 2 {
		t.Fatalf("expected 2 children in OR group, got %d", len(ast.Children))
	}

	// Child 0 is username=alice
	if !ast.Children[0].IsLeaf() || ast.Children[0].Field != "username" {
		t.Errorf("expected child 0 to be username leaf, got %+v", ast.Children[0])
	}

	// Child 1 is AND group
	nestedAnd := ast.Children[1]
	if nestedAnd.Op != OpAnd {
		t.Fatalf("expected child 1 Op to be OpAnd, got %v", nestedAnd.Op)
	}
	if len(nestedAnd.Children) != 2 {
		t.Fatalf("expected 2 children in nested AND group, got %d", len(nestedAnd.Children))
	}
	if nestedAnd.Children[0].Field != "email" || nestedAnd.Children[1].Field != "is_active" {
		t.Errorf("expected nested AND children to be email and is_active, got %+v, %+v",
			nestedAnd.Children[0], nestedAnd.Children[1])
	}

	sql := renderASTSQL(t, fs)
	if !strings.Contains(sql, " OR ") || !strings.Contains(sql, " AND ") {
		t.Errorf("expected rendered SQL to contain both ' OR ' and ' AND ', got: %s", sql)
	}
}

// Test OrFilter when qb.nodes is empty
func TestOrFilter_AsFirstPredicate(t *testing.T) {
	fs, err := NewFilterSet[MockModel]()
	if err != nil {
		t.Fatalf("Failed to create FilterSet: %v", err)
	}

	fs.AndGroup(func(q *QueryBuilder[MockModel]) {
		q.OrFilter("username").Exact("alice")
		q.OrFilter("email").Exact("bob@example.com")
	})

	ast := fs.GetAST()
	if ast == nil {
		t.Fatal("expected non-nil AST")
	}
	// Root is AND with 1 child: OR(username, email)
	if ast.Op != OpAnd {
		t.Fatalf("expected OpAnd root, got %v", ast.Op)
	}
	if len(ast.Children) != 1 {
		t.Fatalf("expected 1 child in root, got %d", len(ast.Children))
	}
	orNode := ast.Children[0]
	if orNode.Op != OpOr {
		t.Fatalf("expected OpOr child, got %v", orNode.Op)
	}
	if len(orNode.Children) != 2 {
		t.Fatalf("expected 2 children in OR node, got %d", len(orNode.Children))
	}
}
