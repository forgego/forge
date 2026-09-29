package orm

import (
	"testing"
	"time"

	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInsertSQL_WritesExplicitZeroValues(t *testing.T) {
	// #291: zeros on fields with a schema Default are written, so a column
	// default such as available DEFAULT true cannot replace them. The
	// optional email has no Default, so it stays NULL.
	instance := testModel{
		Name: "Widget",
	}

	sql, values, columns, err := BuildInsertSQLForPK(instance, "test_table", "id")
	require.NoError(t, err)

	assert.Equal(t, `INSERT INTO "test_table" ("name", "price", "available") VALUES ($1, $2, $3) RETURNING "id"`, sql)
	assert.Equal(t, []interface{}{"Widget", 0.0, false}, values)
	assert.Equal(t, []string{"name", "price", "available"}, columns)
}

func TestBuildInsertSQL_RequiredFieldIncludedEvenWhenZeroValue(t *testing.T) {
	instance := testModel{
		Name: "",
	}

	_, values, columns, err := BuildInsertSQLForPK(instance, "test_table", "id")
	require.NoError(t, err)

	assert.Equal(t, "name", columns[0])
	assert.Equal(t, "", values[0])
}

func TestBuildInsertSQL_MatchesDefaultForPK(t *testing.T) {
	instance := testModel{
		Name: "Widget",
	}

	sql1, values1, columns1, err1 := BuildInsertSQL(instance, "test_table")
	require.NoError(t, err1)

	sql2, values2, columns2, err2 := BuildInsertSQLForPK(instance, "test_table", "id")
	require.NoError(t, err2)

	assert.Equal(t, sql2, sql1)
	assert.Equal(t, values2, values1)
	assert.Equal(t, columns2, columns1)
}

func TestBuildBulkInsertSQL_ConsistentColumns(t *testing.T) {
	instances := []interface{}{
		testModel{Name: "A"},
		testModel{Name: "B", Available: true},
	}

	sql, values, columns, err := BuildBulkInsertSQLForPK(instances, "test_table", "id")
	require.NoError(t, err)

	assert.Equal(t, `INSERT INTO "test_table" ("name", "price", "available") VALUES ($1, $2, $3), ($4, $5, $6) RETURNING "id"`, sql)
	assert.Equal(t, []interface{}{"A", 0.0, false, "B", 0.0, true}, values)
	assert.Equal(t, []string{"name", "price", "available"}, columns)
}

func TestBuildBulkInsertSQL_RejectsInconsistentColumns(t *testing.T) {
	note := "set"
	instances := []interface{}{
		insertRulesModel{Name: "A"},
		insertRulesModel{Name: "B", Note: &note},
	}

	_, _, _, err := BuildBulkInsertSQLForPK(instances, "insert_rules", "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires consistent columns")
}

func TestBuildBulkInsertSQL_MatchesDefaultForPK(t *testing.T) {
	instances := []interface{}{
		testModel{Name: "A"},
		testModel{Name: "B"},
	}

	sql1, values1, columns1, err1 := BuildBulkInsertSQL(instances, "test_table")
	require.NoError(t, err1)

	sql2, values2, columns2, err2 := BuildBulkInsertSQLForPK(instances, "test_table", "id")
	require.NoError(t, err2)

	assert.Equal(t, sql2, sql1)
	assert.Equal(t, values2, values1)
	assert.Equal(t, columns2, columns1)
}

func TestBuildUpdateSQL_QuotesIdentifiers(t *testing.T) {
	instance := testModel{
		ID:   42,
		Name: "Widget",
	}

	sql, values, err := BuildUpdateSQL(instance, "order", "id")
	require.NoError(t, err)

	assert.Contains(t, sql, `UPDATE "order" SET`)
	assert.Contains(t, sql, `"name" = $1`)
	assert.Contains(t, sql, `WHERE "id" = $5`)
	assert.Equal(t, []interface{}{"Widget", "", float64(0), false, int64(42)}, values)
}

func TestBuildDeleteSQL_QuotesIdentifiers(t *testing.T) {
	sql, values := BuildDeleteSQL("order", "id", int64(42))
	assert.Equal(t, `DELETE FROM "order" WHERE "id" = $1`, sql)
	assert.Equal(t, []interface{}{int64(42)}, values)
}

type insertRulesModel struct {
	schema.BaseSchema
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Active    bool      `db:"active"`
	Rank      int64     `db:"rank"`
	Label     string    `db:"label"`
	InStock   bool      `db:"in_stock"`
	Note      *string   `db:"note"`
	Tags      []byte    `db:"tags"`
	SeenAt    time.Time `db:"seen_at"`
	ParentID  int64     `db:"parent_id"`
	OwnerID   int64     `db:"owner"`
	SKU       string    `db:"sku"`
	Total     float64   `db:"total"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt string    `db:"updated_at"`
}

func (insertRulesModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
		schema.BoolField("active", schema.Default(true)),
		schema.Int64Field("rank", schema.Default(5)),
		schema.StringField("label", schema.Default("x")),
		schema.BoolField("in_stock", schema.DBDefault("true")),
		schema.StringField("note", schema.Default("n")),
		schema.JSONField("tags"),
		schema.TimeField("seen_at"),
		schema.Int64Field("parent_id"),
		schema.Int64Field("owner"),
		schema.StringField("sku", schema.Unique()),
		schema.Float64Field("total", schema.GeneratedColumn("rank * 2", true)),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (insertRulesModel) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("parent_id", "insertRulesModel"),
		schema.ForeignKeyField("owner", "User"),
	}
}

func (insertRulesModel) Meta() schema.Meta { return schema.Meta{TableName: "insert_rules"} }

func TestBuildInsertSQL_ZeroValueColumnRules(t *testing.T) {
	_, values, columns, err := BuildInsertSQLForPK(insertRulesModel{Name: "a"}, "insert_rules", "id")
	require.NoError(t, err)
	// Written: required name and the scalar zeros of fields with a schema
	// Default. Omitted: the auto PK, DBDefault, nil pointer and slice, zero
	// time, zero foreign keys, zero unique optional, generated column and
	// zero AutoNow/AutoNowAdd timestamps.
	assert.Equal(t, []string{"name", "active", "rank", "label"}, columns)
	assert.Equal(t, []interface{}{"a", false, int64(0), ""}, values)

	note := ""
	seen := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	_, values, columns, err = BuildInsertSQLForPK(insertRulesModel{
		Name: "a", InStock: true, Note: &note, SeenAt: seen, ParentID: 7, SKU: "s",
		CreatedAt: seen, UpdatedAt: "2026-01-02T03:04:05Z",
	}, "insert_rules", "id")
	require.NoError(t, err)
	assert.Equal(t, []string{"name", "active", "rank", "label", "in_stock", "note", "seen_at", "parent_id", "sku", "created_at", "updated_at"}, columns)
	assert.Equal(t, &note, values[5])
}

func TestBuildUpdateSQL_SkipsGeneratedAndZeroAutoNowAdd(t *testing.T) {
	sql, _, err := BuildUpdateSQL(&insertRulesModel{ID: 1, Name: "a"}, "insert_rules", "id")
	require.NoError(t, err)
	assert.NotContains(t, sql, `"total"`)
	assert.NotContains(t, sql, `"created_at"`)
	assert.Contains(t, sql, `"updated_at"`)

	sql, _, err = BuildUpdateSQL(&insertRulesModel{ID: 1, Name: "a", CreatedAt: time.Now()}, "insert_rules", "id")
	require.NoError(t, err)
	assert.Contains(t, sql, `"created_at"`)
}

func TestApplyDefaults_SetsZeroFieldsWithDefault(t *testing.T) {
	m := &insertRulesModel{Label: "kept"}
	require.NoError(t, ApplyDefaults(m))
	assert.True(t, m.Active)
	assert.Equal(t, int64(5), m.Rank)
	assert.Equal(t, "kept", m.Label)
	require.NotNil(t, m.Note)
	assert.Equal(t, "n", *m.Note)
	assert.False(t, m.InStock, "DBDefault is the database's, not applied in Go")
}

func TestTouchAutoNow_SetsSupportedTypes(t *testing.T) {
	now := time.Date(2026, 9, 28, 10, 0, 0, 123, time.UTC)
	m := &insertRulesModel{}
	touchAutoNow(m, now)
	assert.Equal(t, now.Format(time.RFC3339Nano), m.UpdatedAt)
	assert.True(t, m.CreatedAt.IsZero(), "AutoNowAdd is not refreshed")
}

type stringDefaultModel struct {
	schema.BaseSchema
	Balance string  `db:"balance"`
	Code    *string `db:"code"`
	Flag    int     `db:"flag"`
}

func (stringDefaultModel) Fields() []schema.Field {
	return []schema.Field{
		schema.DecimalField("balance", schema.MaxDigits(12), schema.DecimalPlaces(2), schema.Default(0)),
		schema.StringField("code", schema.Default(65)),
		schema.Int32Field("flag", schema.Default(true)),
	}
}

func TestApplyDefaults_FormatsNumbersForStringFields(t *testing.T) {
	m := &stringDefaultModel{}
	err := ApplyDefaults(m)
	require.Error(t, err, "a bool default cannot fill an int field")
	assert.Contains(t, err.Error(), "flag")
	assert.Equal(t, "0", m.Balance)
	require.NotNil(t, m.Code)
	assert.Equal(t, "65", *m.Code)
}

type pointerDefaultModel struct {
	schema.BaseSchema
	Note  *string `db:"note"`
	Level *int64  `db:"level"`
}

var sharedNoteDefault = "shared"

func (pointerDefaultModel) Fields() []schema.Field {
	return []schema.Field{
		schema.StringField("note", schema.Default(&sharedNoteDefault)),
		schema.Int64Field("level", schema.Default(func() *int64 { v := int64(3); return &v })),
	}
}

func TestApplyDefaults_AcceptsPointerDefaultsForPointerFields(t *testing.T) {
	m := &pointerDefaultModel{}
	require.NoError(t, ApplyDefaults(m))
	require.NotNil(t, m.Note)
	assert.Equal(t, "shared", *m.Note)
	assert.NotSame(t, &sharedNoteDefault, m.Note, "instances must not share the default's pointee")
	require.NotNil(t, m.Level)
	assert.Equal(t, int64(3), *m.Level)
}

type nullableZeroModel struct {
	schema.BaseSchema
	ID       int64  `db:"id"`
	ReviewID int64  `db:"review_id"`
	IP       string `db:"ip_address"`
	Ref      string `db:"ref"`
	RefDef   string `db:"ref_def"`
	Amount   string `db:"amount"`
	Payload  string `db:"payload"`
	Title    string `db:"title"`
}

func (nullableZeroModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("review_id", schema.Required()),
		schema.StringField("ip_address"),
		schema.UUIDField("ref"),
		schema.UUIDField("ref_def", schema.Default("00000000-0000-0000-0000-000000000000")),
		schema.DecimalField("amount", schema.Default(0)),
		schema.JSONField("payload", schema.Default("{}")),
		schema.StringField("title", schema.Default("t")),
	}
}

func (nullableZeroModel) Meta() schema.Meta { return schema.Meta{TableName: "nullable_zero"} }

// Review of #291: optional fields without a Default stay NULL (so blank rows
// don't collide on a multi-column unique constraint), and "" is never written
// to a non-text column such as UUID, decimal or JSON.
func TestBuildInsertSQL_OptionalZeroWithoutDefaultStaysNull(t *testing.T) {
	_, values, columns, err := BuildInsertSQLForPK(nullableZeroModel{ReviewID: 1}, "nullable_zero", "id")
	require.NoError(t, err)
	assert.Equal(t, []string{"review_id", "title"}, columns)
	assert.Equal(t, []interface{}{int64(1), ""}, values)
}

type textDBTypeModel struct {
	schema.BaseSchema
	Code string `db:"code"`
	Ref  string `db:"ref"`
}

func (textDBTypeModel) Fields() []schema.Field {
	return []schema.Field{
		schema.StringField("code", schema.DBType("CITEXT"), schema.Default("x")),
		schema.StringField("ref", schema.DBType("inet"), schema.Default("127.0.0.1")),
	}
}

// An explicit "" is written to a text-like custom DBType, but not to a
// non-text one where it is not a valid value.
func TestBuildInsertSQL_TextCustomDBTypeWritesEmptyString(t *testing.T) {
	_, values, columns, err := BuildInsertSQL(textDBTypeModel{}, "text_dbtype")
	require.NoError(t, err)
	assert.Equal(t, []string{"code"}, columns)
	assert.Equal(t, []interface{}{""}, values)
}
