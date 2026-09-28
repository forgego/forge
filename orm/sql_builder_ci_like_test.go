package orm

import (
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dialectOnlyWrapper wraps an underlying dialect.Dialect and implements ONLY
// the dialect.Dialect interface methods explicitly. It embeds nothing that provides
// CaseInsensitiveLike and does not implement dialect.CaseInsensitiveLiker.
type dialectOnlyWrapper struct {
	inner dialect.Dialect
}

var _ dialect.Dialect = (*dialectOnlyWrapper)(nil)

func (d *dialectOnlyWrapper) Name() string                    { return d.inner.Name() }
func (d *dialectOnlyWrapper) Placeholder(position int) string { return d.inner.Placeholder(position) }
func (d *dialectOnlyWrapper) BuildPlaceholders(n int) string  { return d.inner.BuildPlaceholders(n) }
func (d *dialectOnlyWrapper) QuoteIdentifier(name string) string {
	return d.inner.QuoteIdentifier(name)
}
func (d *dialectOnlyWrapper) QuoteString(s string) string      { return d.inner.QuoteString(s) }
func (d *dialectOnlyWrapper) AutoIncrementType() string        { return d.inner.AutoIncrementType() }
func (d *dialectOnlyWrapper) SupportsReturning() bool          { return d.inner.SupportsReturning() }
func (d *dialectOnlyWrapper) CurrentTime() string              { return d.inner.CurrentTime() }
func (d *dialectOnlyWrapper) CurrentTimestamp() string         { return d.inner.CurrentTimestamp() }
func (d *dialectOnlyWrapper) BooleanLiteral(value bool) string { return d.inner.BooleanLiteral(value) }
func (d *dialectOnlyWrapper) LimitOffset(limit, offset int) string {
	return d.inner.LimitOffset(limit, offset)
}
func (d *dialectOnlyWrapper) LikeEscape() string          { return d.inner.LikeEscape() }
func (d *dialectOnlyWrapper) ConcatOperator() string      { return d.inner.ConcatOperator() }
func (d *dialectOnlyWrapper) CreateTableOptions() string  { return d.inner.CreateTableOptions() }
func (d *dialectOnlyWrapper) OnConflictDoNothing() string { return d.inner.OnConflictDoNothing() }
func (d *dialectOnlyWrapper) OnConflictDoUpdate(col string, up []string) string {
	return d.inner.OnConflictDoUpdate(col, up)
}

func TestSQLBuilder_CaseInsensitiveLike_CustomDialect(t *testing.T) {
	sqliteDialect := dialect.NewSQLiteDialect()
	custom := &dialectOnlyWrapper{inner: sqliteDialect}

	// Verify that custom dialect does NOT implement dialect.CaseInsensitiveLiker
	_, ok := any(custom).(dialect.CaseInsensitiveLiker)
	require.False(t, ok, "custom dialect must not implement CaseInsensitiveLiker")

	builder := NewSQLBuilderWithDialect(custom)

	// Direct method check on builder
	got := builder.CaseInsensitiveLike("\"name\"", "?")
	assert.Equal(t, "LOWER(\"name\") LIKE LOWER(?)", got)

	// icontains-style filter using Where
	expr := Where("name", OpIContains, "Alice")
	sql, args, err := expr.ToSQL(builder)
	require.NoError(t, err)
	assert.Equal(t, "LOWER(\"name\") LIKE LOWER(?)", sql)
	assert.Equal(t, []interface{}{"%Alice%"}, args)

	// icontains-style filter using F().IContains
	builder.Reset()
	exprField := F("name").IContains("Alice")
	sqlField, argsField, errField := exprField.ToSQL(builder)
	require.NoError(t, errField)
	assert.Equal(t, "LOWER(\"name\") LIKE LOWER(?)", sqlField)
	assert.Equal(t, []interface{}{"%Alice%"}, argsField)

	// Additional icontains/ILIKE filter variants with custom dialect
	builder.Reset()
	exprExact := Where("name", OpIExact, "Alice")
	sqlExact, _, errExact := exprExact.ToSQL(builder)
	require.NoError(t, errExact)
	assert.Equal(t, "LOWER(\"name\") LIKE LOWER(?)", sqlExact)

	builder.Reset()
	exprStarts := Where("name", OpIStartsWith, "Alice")
	sqlStarts, _, errStarts := exprStarts.ToSQL(builder)
	require.NoError(t, errStarts)
	assert.Equal(t, "LOWER(\"name\") LIKE LOWER(?)", sqlStarts)

	builder.Reset()
	exprEnds := Where("name", OpIEndsWith, "Alice")
	sqlEnds, _, errEnds := exprEnds.ToSQL(builder)
	require.NoError(t, errEnds)
	assert.Equal(t, "LOWER(\"name\") LIKE LOWER(?)", sqlEnds)
}

func TestSQLBuilder_CaseInsensitiveLike_PostgreSQLDialect(t *testing.T) {
	pgDialect := dialect.NewPostgreSQLDialect()

	// Verify PostgreSQL dialect implements CaseInsensitiveLiker
	ciLiker, ok := any(pgDialect).(dialect.CaseInsensitiveLiker)
	require.True(t, ok, "PostgreSQL dialect must implement CaseInsensitiveLiker")
	assert.Equal(t, "\"name\" ILIKE $1", ciLiker.CaseInsensitiveLike("\"name\"", "$1"))

	builder := NewSQLBuilderWithDialect(pgDialect)

	// Direct method check on builder
	got := builder.CaseInsensitiveLike("\"name\"", "$1")
	assert.Equal(t, "\"name\" ILIKE $1", got)

	// icontains-style filter using Where
	expr := Where("name", OpIContains, "Alice")
	sql, args, err := expr.ToSQL(builder)
	require.NoError(t, err)
	assert.Equal(t, "\"name\" ILIKE $1", sql)
	assert.Equal(t, []interface{}{"%Alice%"}, args)

	// icontains-style filter using F().IContains
	builder.Reset()
	exprField := F("name").IContains("Alice")
	sqlField, argsField, errField := exprField.ToSQL(builder)
	require.NoError(t, errField)
	assert.Equal(t, "\"name\" ILIKE $1", sqlField)
	assert.Equal(t, []interface{}{"%Alice%"}, argsField)

	// Additional icontains/ILIKE filter variants with PostgreSQL dialect
	builder.Reset()
	exprExact := Where("name", OpIExact, "Alice")
	sqlExact, _, errExact := exprExact.ToSQL(builder)
	require.NoError(t, errExact)
	assert.Equal(t, "\"name\" ILIKE $1", sqlExact)

	builder.Reset()
	exprStarts := Where("name", OpIStartsWith, "Alice")
	sqlStarts, _, errStarts := exprStarts.ToSQL(builder)
	require.NoError(t, errStarts)
	assert.Equal(t, "\"name\" ILIKE $1", sqlStarts)

	builder.Reset()
	exprEnds := Where("name", OpIEndsWith, "Alice")
	sqlEnds, _, errEnds := exprEnds.ToSQL(builder)
	require.NoError(t, errEnds)
	assert.Equal(t, "\"name\" ILIKE $1", sqlEnds)
}

func TestSQLBuilder_CaseInsensitiveLike_QuerySet(t *testing.T) {
	sqliteDialect := dialect.NewSQLiteDialect()
	custom := &dialectOnlyWrapper{inner: sqliteDialect}
	pgDialect := dialect.NewPostgreSQLDialect()

	customDB := &db.DB{Driver: "custom"}
	customDB.SetDialect(custom)

	pgDB := &db.DB{Driver: "postgres"}
	pgDB.SetDialect(pgDialect)

	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	// QuerySet with custom dialect (no CaseInsensitiveLiker) produces LOWER(...) LIKE LOWER(...)
	baseQSCustom := qs.SetDB(customDB).(*BaseQuerySet[testModel])
	filteredCustom := baseQSCustom.Filter(Where("name", OpIContains, "Alice")).(*BaseQuerySet[testModel])
	sqlCustom, _, err := filteredCustom.buildSQL()
	require.NoError(t, err)
	assert.Contains(t, sqlCustom, "LOWER(\"name\") LIKE LOWER(?)")

	// QuerySet with PostgreSQL dialect produces ILIKE
	baseQSPG := qs.SetDB(pgDB).(*BaseQuerySet[testModel])
	filteredPG := baseQSPG.Filter(Where("name", OpIContains, "Alice")).(*BaseQuerySet[testModel])
	sqlPG, _, err := filteredPG.buildSQL()
	require.NoError(t, err)
	assert.Contains(t, sqlPG, "\"name\" ILIKE $1")
}
