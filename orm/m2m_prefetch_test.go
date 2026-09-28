package orm

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Category struct {
	schema.BaseSchema
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (Category) Meta() schema.Meta {
	return schema.Meta{TableName: "categories"}
}

func (Category) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

type Product struct {
	schema.BaseSchema
	ID         int64      `db:"id"`
	Name       string     `db:"name"`
	Categories []Category `db:"categories"`
}

func (Product) Meta() schema.Meta {
	return schema.Meta{TableName: "products"}
}

func (Product) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

func (Product) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ManyToManyField("Categories", "Category", schema.Through("product_categories")),
	}
}

func setupProductCategoryDB(t *testing.T) *db.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "m2m_prefetch_test.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)

	_, err = database.Exec(`
		CREATE TABLE products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);
		CREATE TABLE categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);
		CREATE TABLE product_categories (
			product_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			PRIMARY KEY (product_id, category_id)
		);
		INSERT INTO products (id, name) VALUES (1, 'Laptop');
		INSERT INTO categories (id, name) VALUES (1, 'Electronics');
		INSERT INTO categories (id, name) VALUES (2, 'Computers');
		INSERT INTO product_categories (product_id, category_id) VALUES (1, 1), (1, 2);
	`)
	require.NoError(t, err)

	_, err = GetModelSchema[Category]()
	require.NoError(t, err)
	_, err = GetModelSchema[Product]()
	require.NoError(t, err)

	return database
}

func TestManyToMany_Prefetch_SnakeCaseModelNameColumns(t *testing.T) {
	database := setupProductCategoryDB(t)
	ctx := context.Background()

	qs, err := NewQuerySet[Product]("products")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	products, err := qs.PrefetchRelated("Categories").All(ctx)
	require.NoError(t, err)
	require.Len(t, products, 1)

	assert.Equal(t, "Laptop", products[0].Name)
	require.Len(t, products[0].Categories, 2)

	catNames := []string{products[0].Categories[0].Name, products[0].Categories[1].Name}
	assert.Contains(t, catNames, "Electronics")
	assert.Contains(t, catNames, "Computers")
}

func TestThroughColumns(t *testing.T) {
	tests := []struct {
		name         string
		sourceSchema *ModelSchema
		targetSchema *ModelSchema
		sourceTable  string
		targetModel  string
		wantSource   string
		wantTarget   string
	}{
		{
			name:         "model names Product and Category",
			sourceSchema: &ModelSchema{ModelName: "Product"},
			targetSchema: &ModelSchema{ModelName: "Category"},
			sourceTable:  "products",
			targetModel:  "Category",
			wantSource:   "product_id",
			wantTarget:   "category_id",
		},
		{
			name:         "multi-word model names BlogPost and UserGroup",
			sourceSchema: &ModelSchema{ModelName: "BlogPost"},
			targetSchema: &ModelSchema{ModelName: "UserGroup"},
			sourceTable:  "blog_posts",
			targetModel:  "UserGroup",
			wantSource:   "blog_post_id",
			wantTarget:   "user_group_id",
		},
		{
			name:         "fallback when ModelName empty",
			sourceSchema: &ModelSchema{},
			targetSchema: &ModelSchema{TableName: "categories"},
			sourceTable:  "products",
			targetModel:  "Category",
			wantSource:   "product_id",
			wantTarget:   "category_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSource, gotTarget := throughColumns(tt.sourceSchema, tt.targetSchema, tt.sourceTable, tt.targetModel)
			assert.Equal(t, tt.wantSource, gotSource)
			assert.Equal(t, tt.wantTarget, gotTarget)
		})
	}
}
