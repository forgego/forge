package orm

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/dialect"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnotation_QueryExprAndExpressionRendering(t *testing.T) {
	fieldExpr := NewFieldExpr[float64]("price", "")
	annFromField := NewAnnotation("high_price", fieldExpr.Greater(10.0))

	field := NewField[float64]("price", "")
	annFromExpr := NewExpressionAnnotation("high_price", field.Gt(10.0))

	t.Run("PostgreSQL builder produces identical SELECT fragment", func(t *testing.T) {
		qs1, err := NewQuerySet[testModel]("test_table")
		require.NoError(t, err)
		qs1 = qs1.Annotate(annFromField)
		base1 := qs1.(*BaseQuerySet[testModel])
		b1 := NewSQLBuilderWithDialect(dialect.NewPostgreSQLDialect())
		select1 := base1.buildSelectClause(b1, false)

		qs2, err := NewQuerySet[testModel]("test_table")
		require.NoError(t, err)
		qs2 = qs2.Annotate(annFromExpr)
		base2 := qs2.(*BaseQuerySet[testModel])
		b2 := NewSQLBuilderWithDialect(dialect.NewPostgreSQLDialect())
		select2 := base2.buildSelectClause(b2, false)

		assert.Contains(t, select1, `price > $1 AS "high_price"`)
		assert.Contains(t, select2, `"price" > $1 AS "high_price"`)
		assert.Equal(t, []interface{}{10.0}, b1.Args())
	})

	t.Run("SQLite builder produces identical SELECT fragment", func(t *testing.T) {
		qs1, err := NewQuerySet[testModel]("test_table")
		require.NoError(t, err)
		qs1 = qs1.Annotate(annFromField)
		base1 := qs1.(*BaseQuerySet[testModel])
		b1 := NewSQLBuilderWithDialect(dialect.NewSQLiteDialect())
		select1 := base1.buildSelectClause(b1, false)

		qs2, err := NewQuerySet[testModel]("test_table")
		require.NoError(t, err)
		qs2 = qs2.Annotate(annFromExpr)
		base2 := qs2.(*BaseQuerySet[testModel])
		b2 := NewSQLBuilderWithDialect(dialect.NewSQLiteDialect())
		select2 := base2.buildSelectClause(b2, false)

		assert.Contains(t, select1, `price > ? AS "high_price"`)
		assert.Contains(t, select2, `"price" > ? AS "high_price"`)
		assert.Equal(t, []interface{}{10.0}, b1.Args())
	})
}

func TestAnnotation_EqualityComparisonEquivalence(t *testing.T) {
	fieldExpr := NewFieldExpr[bool]("available", "")
	annOld := NewAnnotation("is_available", fieldExpr.Equals(true))

	field := NewField[bool]("available", "")
	annNew := NewExpressionAnnotation("is_available", field.Eq(true))

	t.Run("PostgreSQL equality fragment", func(t *testing.T) {
		bOld := NewSQLBuilderWithDialect(dialect.NewPostgreSQLDialect())
		sqlOld, _, err := annOld.Expression.ToSQL(bOld)
		require.NoError(t, err)
		fragOld := fmt.Sprintf("%s AS %s", sqlOld, EscapeIdentifier(annOld.Name))

		bNew := NewSQLBuilderWithDialect(dialect.NewPostgreSQLDialect())
		sqlNew, _, err := annNew.Expression.ToSQL(bNew)
		require.NoError(t, err)
		fragNew := fmt.Sprintf("%s AS %s", sqlNew, EscapeIdentifier(annNew.Name))

		assert.Equal(t, `available = $1 AS "is_available"`, fragOld)
		assert.Equal(t, `"available" = $1 AS "is_available"`, fragNew)
	})

	t.Run("SQLite equality fragment", func(t *testing.T) {
		bOld := NewSQLBuilderWithDialect(dialect.NewSQLiteDialect())
		sqlOld, _, err := annOld.Expression.ToSQL(bOld)
		require.NoError(t, err)
		fragOld := fmt.Sprintf("%s AS %s", sqlOld, EscapeIdentifier(annOld.Name))

		bNew := NewSQLBuilderWithDialect(dialect.NewSQLiteDialect())
		sqlNew, _, err := annNew.Expression.ToSQL(bNew)
		require.NoError(t, err)
		fragNew := fmt.Sprintf("%s AS %s", sqlNew, EscapeIdentifier(annNew.Name))

		assert.Equal(t, `available = ? AS "is_available"`, fragOld)
		assert.Equal(t, `"available" = ? AS "is_available"`, fragNew)
	})
}

func TestAnnotation_UnsetExpressionFieldFallback(t *testing.T) {
	ann := AnnotationExpr{
		Name: "custom_flag",
		Expr: NewFieldQueryExpr("price", OpGreater, 50.0),
	}

	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)
	qs = qs.Annotate(ann)
	base := qs.(*BaseQuerySet[testModel])

	b := NewSQLBuilderWithDialect(dialect.NewPostgreSQLDialect())
	selectClause := base.buildSelectClause(b, false)
	assert.Contains(t, selectClause, `price > $1 AS "custom_flag"`)
	assert.Equal(t, []interface{}{50.0}, b.Args())
}

func TestSQLBuilder_BuildWhere_AdapterEquivalence(t *testing.T) {
	conditions := []QueryExpr{
		NewFieldQueryExpr("price", OpGreater, 10.0),
		NewFieldQueryExpr("available", OpEquals, true),
	}
	excludes := []QueryExpr{
		NewFieldQueryExpr("deleted", OpEquals, true),
	}

	t.Run("PostgreSQL BuildWhere", func(t *testing.T) {
		b := NewSQLBuilderWithDialect(dialect.NewPostgreSQLDialect())
		where, args := b.BuildWhere(conditions, excludes)
		assert.Equal(t, `WHERE price > $1 AND available = $2 AND NOT (deleted = $3)`, where)
		assert.Equal(t, []interface{}{10.0, true, true}, args)
	})

	t.Run("SQLite BuildWhere", func(t *testing.T) {
		b := NewSQLBuilderWithDialect(dialect.NewSQLiteDialect())
		where, args := b.BuildWhere(conditions, excludes)
		assert.Equal(t, `WHERE price > ? AND available = ? AND NOT (deleted = ?)`, where)
		assert.Equal(t, []interface{}{10.0, true, true}, args)
	})
}

func TestAnnotation_ArgOrder_MatchesPlaceholderOrder(t *testing.T) {
	t.Run("SQLite query binds annotation and filter args in placeholder text order", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "arg_order_test.sqlite")
		database, err := db.NewDB(dbPath)
		require.NoError(t, err)
		defer database.Close()

		_, err = database.Exec(`
			CREATE TABLE test_table (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				email TEXT,
				price REAL DEFAULT 0.0,
				available BOOLEAN DEFAULT 1
			);
			INSERT INTO test_table (name, price) VALUES ('Low', 0.5);
			INSERT INTO test_table (name, price) VALUES ('Mid', 2.0);
			INSERT INTO test_table (name, price) VALUES ('High', 4.0);
		`)
		require.NoError(t, err)

		priceField := NewField[float64]("price", "test_table")
		// Filter price > 1.0 (WHERE clause: arg should be 1.0)
		filterExpr := priceField.Gt(1.0)
		// Annotation price > 3.0 (SELECT clause: arg should be 3.0)
		ann := NewExpressionAnnotation("threshold_flag", priceField.Gt(3.0))

		qs, err := NewQuerySet[testModel]("test_table")
		require.NoError(t, err)
		qs = qs.SetDB(database).Filter(filterExpr).Annotate(ann).OrderBy("id")

		base := qs.(*BaseQuerySet[testModel])
		sql, args, err := base.buildSQL()
		require.NoError(t, err)

		// Annotation placeholder appears before WHERE placeholder in SQL text,
		// so args[0] must be 3.0 and args[1] must be 1.0 for positional '?' binding.
		assert.Contains(t, sql, `"price" > ? AS "threshold_flag"`)
		assert.Contains(t, sql, `WHERE "test_table"."price" > ?`)
		assert.Equal(t, []interface{}{3.0, 1.0}, args)

		rows, err := database.Query(sql, args...)
		require.NoError(t, err)
		defer rows.Close()

		type rowResult struct {
			id            int64
			name          string
			email         *string
			price         float64
			available     bool
			thresholdFlag int
		}
		var results []rowResult
		for rows.Next() {
			var r rowResult
			err := rows.Scan(&r.id, &r.name, &r.email, &r.price, &r.available, &r.thresholdFlag)
			require.NoError(t, err)
			results = append(results, r)
		}
		require.NoError(t, rows.Err())
		require.Len(t, results, 2)

		// Mid (2.0): > 1.0 is true, > 3.0 is false (0)
		assert.Equal(t, "Mid", results[0].name)
		assert.Equal(t, 0, results[0].thresholdFlag)

		// High (4.0): > 1.0 is true, > 3.0 is true (1)
		assert.Equal(t, "High", results[1].name)
		assert.Equal(t, 1, results[1].thresholdFlag)
	})

	t.Run("PostgreSQL-dialect builder assertion that $N numbering matches argument order", func(t *testing.T) {
		priceField := NewField[float64]("price", "test_table")
		filterExpr := priceField.Gt(1.0)
		ann := NewExpressionAnnotation("threshold_flag", priceField.Gt(3.0))

		qs, err := NewQuerySet[testModel]("test_table")
		require.NoError(t, err)
		qs = qs.Filter(filterExpr).Annotate(ann)

		base := qs.(*BaseQuerySet[testModel])
		// Without a database set, buildSQL uses NewSQLBuilder() which defaults to $1, $2 (PositionalStyle)
		sql, args, err := base.buildSQL()
		require.NoError(t, err)

		// $1 is in SELECT, $2 is in WHERE
		assert.Contains(t, sql, `"price" > $1 AS "threshold_flag"`)
		assert.Contains(t, sql, `WHERE "test_table"."price" > $2`)
		require.Len(t, args, 2)
		assert.Equal(t, 3.0, args[0])
		assert.Equal(t, 1.0, args[1])
	})
}

func TestAnnotation_UnknownFieldSQLite(t *testing.T) {
	database, err := db.NewDBWithDriver("sqlite3", ":memory:")
	require.NoError(t, err)
	defer database.Close()
	_, err = database.Exec(`CREATE TABLE test_table (id INTEGER PRIMARY KEY); INSERT INTO test_table VALUES (1)`)
	require.NoError(t, err)
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)
	qs = qs.SetDB(database).Annotate(NewExpressionAnnotation("bad", F("missing_field").Eq("missing_field")))
	_, _, err = qs.(*BaseQuerySet[testModel]).buildSQL()
	require.ErrorContains(t, err, "missing_field")
	_, err = qs.All(context.Background())
	require.ErrorContains(t, err, "missing_field")
}
