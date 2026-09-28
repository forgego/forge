package orm

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/errors"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ItemWithCustomDBColumn struct {
	schema.BaseSchema
	ID    int64   `db:"id"`
	Name  string  `db:"name"`
	Price float64 `db:"unit_price"`
}

func (ItemWithCustomDBColumn) Meta() schema.Meta {
	return schema.Meta{TableName: "custom_column_items"}
}

func (ItemWithCustomDBColumn) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
		schema.FloatField("price", schema.DBColumn("unit_price")),
	}
}

type ProductWithCustomPK struct {
	schema.BaseSchema
	ProductID int64  `db:"product_id"`
	Title     string `db:"title"`
}

func (ProductWithCustomPK) Meta() schema.Meta {
	return schema.Meta{TableName: "custom_pk_products"}
}

func (ProductWithCustomPK) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("product_id", schema.Primary(), schema.AutoIncrement(), schema.DBColumn("product_id")),
		schema.StringField("title"),
	}
}

type StringPKModel struct {
	schema.BaseSchema
	Code string `db:"code"`
	Name string `db:"name"`
}

func (StringPKModel) Meta() schema.Meta {
	return schema.Meta{TableName: "string_pk_models"}
}

func (StringPKModel) Fields() []schema.Field {
	return []schema.Field{
		schema.StringField("code", schema.Primary()),
		schema.StringField("name"),
	}
}

func setupWriteColumnNamesDB(t *testing.T) *db.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "write_column_names_test.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)

	_, err = database.Exec(`
		CREATE TABLE custom_column_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			unit_price REAL NOT NULL
		);
		CREATE TABLE custom_pk_products (
			product_id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL
		);
		CREATE TABLE string_pk_models (
			code TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);
	`)
	require.NoError(t, err)

	_, err = GetModelSchema[ItemWithCustomDBColumn]()
	require.NoError(t, err)
	_, err = GetModelSchema[ProductWithCustomPK]()
	require.NoError(t, err)
	_, err = GetModelSchema[StringPKModel]()
	require.NoError(t, err)

	return database
}

func TestUpdateFields_CustomDBColumn(t *testing.T) {
	database := setupWriteColumnNamesDB(t)
	ctx := context.Background()

	_, err := database.Exec("INSERT INTO custom_column_items (id, name, unit_price) VALUES (1, 'Widget', 10.0)")
	require.NoError(t, err)

	mgr, err := NewManager[ItemWithCustomDBColumn]("custom_column_items")
	require.NoError(t, err)
	mgr.SetDB(database)

	err = mgr.UpdateFields(ctx, 1, UpdateMap{"price": 9.5})
	require.NoError(t, err)

	var gotPrice float64
	err = database.QueryRow("SELECT unit_price FROM custom_column_items WHERE id = 1").Scan(&gotPrice)
	require.NoError(t, err)
	assert.Equal(t, 9.5, gotPrice)
}

func TestQuerySetUpdate_KeyResolvedCaseInsensitively(t *testing.T) {
	database := setupWriteColumnNamesDB(t)
	ctx := context.Background()

	_, err := database.Exec("INSERT INTO custom_column_items (id, name, unit_price) VALUES (1, 'Widget', 10.0)")
	require.NoError(t, err)

	qs, err := NewQuerySet[ItemWithCustomDBColumn]("custom_column_items")
	require.NoError(t, err)
	qs = qs.SetDB(database).Filter(F("id").Eq(1))

	rows, err := qs.Update(ctx, UpdateMap{"PRICE": 12.0})
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	var gotPrice float64
	err = database.QueryRow("SELECT unit_price FROM custom_column_items WHERE id = 1").Scan(&gotPrice)
	require.NoError(t, err)
	assert.Equal(t, 12.0, gotPrice)
}

func TestQuerySetUpdate_UnknownField(t *testing.T) {
	database := setupWriteColumnNamesDB(t)
	ctx := context.Background()

	qs, err := NewQuerySet[ItemWithCustomDBColumn]("custom_column_items")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	_, err = qs.Update(ctx, UpdateMap{"nope": 123})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nope")
}

func TestCreate_CustomPrimaryKeyColumn(t *testing.T) {
	database := setupWriteColumnNamesDB(t)
	ctx := context.Background()

	mgr, err := NewManager[ProductWithCustomPK]("custom_pk_products")
	require.NoError(t, err)
	mgr.SetDB(database)

	product := &ProductWithCustomPK{
		Title: "Gadget",
	}

	err = mgr.Create(ctx, product)
	require.NoError(t, err)
	assert.Equal(t, int64(1), product.ProductID)

	var gotID int64
	var gotTitle string
	err = database.QueryRow("SELECT product_id, title FROM custom_pk_products WHERE product_id = ?", product.ProductID).Scan(&gotID, &gotTitle)
	require.NoError(t, err)
	assert.Equal(t, int64(1), gotID)
	assert.Equal(t, "Gadget", gotTitle)
}

func TestCreate_NonIntegerPrimaryKeyNotImplemented(t *testing.T) {
	database := setupWriteColumnNamesDB(t)
	ctx := context.Background()

	mgr, err := NewManager[StringPKModel]("string_pk_models")
	require.NoError(t, err)
	mgr.SetDB(database)

	item := &StringPKModel{
		Code: "SKU-1",
		Name: "Widget",
	}

	err = mgr.Create(ctx, item)
	require.Error(t, err)
	assert.True(t, errors.IsNotImplemented(err), "expected NotImplementedError, got %T: %v", err, err)
}
