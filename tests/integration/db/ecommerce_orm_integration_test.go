package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	"github.com/forgego/forge/tests/helpers"
	"github.com/forgego/forge/tests/testhelpers"
)

type Author struct {
	schema.BaseSchema
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}

func (Author) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
		schema.StringField("email", schema.Required()),
		schema.TimeField("created_at", schema.Optional()),
	}
}

func (Author) Meta() schema.Meta {
	return schema.Meta{
		TableName: "authors",
	}
}

func (Author) Relations() []schema.Relation {
	return nil
}

type Book struct {
	schema.BaseSchema
	ID        int64     `db:"id"`
	Title     string    `db:"title"`
	Price     float64   `db:"price"`
	AuthorID  int64     `db:"author_id"`
	CreatedAt time.Time `db:"created_at"`
	Author    *Author   `db:"-"`
}

func (Book) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required()),
		schema.Float64Field("price", schema.Required()),
		schema.Int64Field("author_id", schema.Required()),
		schema.TimeField("created_at", schema.Optional()),
	}
}

func (Book) Meta() schema.Meta {
	return schema.Meta{
		TableName: "books",
	}
}

func (Book) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("Author", "Author", schema.OnDelete(schema.CascadeCASCADE)),
	}
}

func TestEcommerceCatalogMigrations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := testhelpers.DefaultPostgresOptsWithTest(t.Name())
	opts.User = "postgres"
	opts.Password = "123"

	postgresDB, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer func() { _ = cleanup() }()
	defer postgresDB.Close()

	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "ecommerce_catalog_")
	defer cleanupTemp()
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	modelsDir, err := ecommerceCatalogModelsDir()
	require.NoError(t, err)

	err = helpers.CreateMigrationFromModels(t, modelsDir, migrationsDir, "catalog_schema")
	require.NoError(t, err)

	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	err = helpers.ApplyMigrationSequence(ctx, t, database, migrationsDir)
	require.NoError(t, err)

	helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "categories")
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "categories", "slug")
}

func TestORMCRUDWithRelations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := testhelpers.DefaultPostgresOptsWithTest(t.Name())
	opts.User = "postgres"
	opts.Password = "123"

	postgresDB, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer func() { _ = cleanup() }()
	defer postgresDB.Close()

	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, `
		CREATE TABLE authors (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, `
		CREATE TABLE books (
			id BIGSERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			price NUMERIC NOT NULL,
			author_id BIGINT NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)

	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	authorManager, err := orm.NewManager[Author]("")
	require.NoError(t, err)
	authorManager.SetDB(database)

	bookManager, err := orm.NewManager[Book]("")
	require.NoError(t, err)
	bookManager.SetDB(database)

	author := &Author{Name: "Ada Lovelace", Email: "ada@example.com"}
	require.NoError(t, authorManager.Create(ctx, author))

	bookA := &Book{Title: "Analytics", Price: 9.99, AuthorID: author.ID}
	bookB := &Book{Title: "Compiler", Price: 19.99, AuthorID: author.ID}
	require.NoError(t, bookManager.Create(ctx, bookA))
	require.NoError(t, bookManager.Create(ctx, bookB))

	bookA.Title = "Analytics Revised"
	require.NoError(t, bookManager.Update(ctx, bookA))

	fa, err := bookManager.FieldAccessor()
	require.NoError(t, err)

	priceField := orm.FieldFor[Book, float64](fa, "price")
	titleField := orm.FieldFor[Book, string](fa, "title")

	filtered, err := bookManager.Filter(priceField.Gt(10.0))
	require.NoError(t, err)

	filtered = filtered.OrderBy(orm.Desc("price"))
	results, err := filtered.All(ctx)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "Compiler", results[0].Title)

	ordered, err := bookManager.Filter(titleField.Contains("a"))
	require.NoError(t, err)

	ordered = ordered.OrderBy(orm.Asc("title"))
	orderedResults, err := ordered.All(ctx)
	require.NoError(t, err)
	require.Len(t, orderedResults, 2)
	require.Equal(t, "Analytics Revised", orderedResults[0].Title)

	related, err := bookManager.Filter(titleField.Eq("Compiler"))
	require.NoError(t, err)

	related = related.SelectRelated("Author")
	relatedResults, err := related.All(ctx)
	require.NoError(t, err)
	require.Len(t, relatedResults, 1)
	require.NotNil(t, relatedResults[0].Author)
	require.Equal(t, "Ada Lovelace", relatedResults[0].Author.Name)

	require.NoError(t, bookManager.Delete(ctx, bookA))

	remaining, err := bookManager.All(ctx)
	require.NoError(t, err)
	require.Len(t, remaining, 1)

	require.NoError(t, authorManager.Delete(ctx, author))

	count, err := bookManager.Filter(priceField.Gt(0)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)
}

func ecommerceCatalogModelsDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	searchDir := wd
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(searchDir, "examples", "ecommerce", "app", "catalog")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			break
		}
		searchDir = parent
	}

	return "", fmt.Errorf("unable to locate examples/ecommerce/app/catalog from %s", wd)
}
