package orm

import (
	"bytes"
	"context"
	"log/slog"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModelCategory defines a category for relation metadata resolution test.
type TestModelCategory struct {
	schema.BaseSchema
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (TestModelCategory) Meta() schema.Meta {
	return schema.Meta{TableName: "test_model_categories"}
}

func (TestModelCategory) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

// TestModelItemWithCategory has a ForeignKeyField("Category", "Category") with a category_id field.
type TestModelItemWithCategory struct {
	schema.BaseSchema
	ID         int64              `db:"id"`
	CategoryID int64              `db:"category_id"`
	Category   *TestModelCategory `db:"category"`
}

func (TestModelItemWithCategory) Meta() schema.Meta {
	return schema.Meta{TableName: "test_model_items"}
}

func (TestModelItemWithCategory) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("category_id"),
	}
}

func (TestModelItemWithCategory) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("Category", "Category"),
	}
}

// TestModelAuthor defines an author for relation metadata resolution test.
type TestModelAuthor struct {
	schema.BaseSchema
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (TestModelAuthor) Meta() schema.Meta {
	return schema.Meta{TableName: "test_model_authors"}
}

func (TestModelAuthor) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

// TestModelBookWithAuthor has a Go field AuthorID with DB column author_id.
type TestModelBookWithAuthor struct {
	schema.BaseSchema
	ID       int64            `db:"id"`
	AuthorID int64            `db:"author_id"`
	Author   *TestModelAuthor `db:"author"`
}

func (TestModelBookWithAuthor) Meta() schema.Meta {
	return schema.Meta{TableName: "test_model_books"}
}

func (TestModelBookWithAuthor) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("author_id"),
	}
}

func (TestModelBookWithAuthor) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("Author", "TestModelAuthor"),
	}
}

func TestBuildModelSchema_ResolvesFKColumnAndStructField(t *testing.T) {
	itemSchema, err := BuildModelSchema(TestModelItemWithCategory{})
	require.NoError(t, err)
	require.Len(t, itemSchema.Relations, 1)
	assert.Equal(t, "category_id", itemSchema.Relations[0].FKColumn)
	assert.Equal(t, "CategoryID", itemSchema.Relations[0].FKStructField)

	bookSchema, err := BuildModelSchema(TestModelBookWithAuthor{})
	require.NoError(t, err)
	require.Len(t, bookSchema.Relations, 1)
	assert.Equal(t, "author_id", bookSchema.Relations[0].FKColumn)
	assert.Equal(t, "AuthorID", bookSchema.Relations[0].FKStructField)
}

// legacyFKColumnFor replicates the matching rules fkColumnFor used prior to schema build resolution.
func legacyFKColumnFor(schema *ModelSchema, rel *RelationInfo) string {
	if schema == nil || rel == nil {
		return ""
	}

	for _, f := range schema.Fields {
		if strings.EqualFold(f.DBColumn, rel.Name) || strings.EqualFold(f.Name, rel.Name) {
			return f.DBColumn
		}
		if f.StructFieldName == rel.Name {
			return f.DBColumn
		}
		if f.StructFieldName == rel.Name+"ID" {
			return f.DBColumn
		}
		if strings.EqualFold(f.DBColumn, rel.Name+"_id") {
			return f.DBColumn
		}
		if strings.HasSuffix(f.Name, "ID") && strings.HasPrefix(strings.ToLower(f.Name), strings.ToLower(rel.Name)) {
			return f.DBColumn
		}
	}

	guess := strings.ToLower(rel.Name) + "_id"
	if f := schema.GetField(guess); f != nil {
		return f.DBColumn
	}

	if strings.HasSuffix(strings.ToLower(rel.Name), "_id") {
		if f := schema.GetField(strings.ToLower(rel.Name)); f != nil {
			return f.DBColumn
		}
	}

	guess = strings.ToLower(rel.TargetModel) + "_id"
	if f := schema.GetField(guess); f != nil {
		return f.DBColumn
	}

	return ""
}

func TestBuildModelSchema_EquivalenceWithLegacyFKColumnFor(t *testing.T) {
	models := []struct {
		name     string
		instance schema.Schema
	}{
		{name: "User", instance: User{}},
		{name: "Post", instance: Post{}},
		{name: "TestCompany", instance: TestCompany{}},
		{name: "TestCustomer", instance: TestCustomer{}},
		{name: "TestOrder", instance: TestOrder{}},
		{name: "Tag", instance: Tag{}},
		{name: "Article", instance: Article{}},
		{name: "Category", instance: Category{}},
		{name: "Product", instance: Product{}},
	}

	for _, tc := range models {
		t.Run(tc.name, func(t *testing.T) {
			ms, err := BuildModelSchema(tc.instance)
			require.NoError(t, err)

			for _, rel := range ms.Relations {
				legacyCol := legacyFKColumnFor(ms, &rel)
				assert.Equal(t, legacyCol, rel.FKColumn, "FKColumn for relation %s on %s must match legacy fkColumnFor", rel.Name, tc.name)
				assert.Equal(t, legacyCol, fkColumnFor(ms, &rel), "fkColumnFor for relation %s on %s must match legacy fkColumnFor", rel.Name, tc.name)
			}
		})
	}
}

type TestModelUnmatchedTarget struct {
	schema.BaseSchema
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (TestModelUnmatchedTarget) Meta() schema.Meta {
	return schema.Meta{TableName: "unmatched_targets"}
}

func (TestModelUnmatchedTarget) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

type TestModelUnmatchedSource struct {
	schema.BaseSchema
	ID    int64                     `db:"id"`
	Title string                    `db:"title"`
	Rel   *TestModelUnmatchedTarget `db:"rel"`
}

func (TestModelUnmatchedSource) Meta() schema.Meta {
	return schema.Meta{TableName: "unmatched_sources"}
}

func (TestModelUnmatchedSource) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title"),
	}
}

func (TestModelUnmatchedSource) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("UnmatchedRel", "TestModelUnmatchedTarget"),
	}
}

func TestSelectRelated_UnmatchedFK_ReturnsRowsAndLogsWarningOnce(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "unmatched_fk_test.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)
	defer database.Close()

	_, err = database.Exec(`
		CREATE TABLE unmatched_sources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL
		);
		CREATE TABLE unmatched_targets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);
		INSERT INTO unmatched_sources (id, title) VALUES (1, 'Unmatched Row');
	`)
	require.NoError(t, err)

	_, err = GetModelSchema[TestModelUnmatchedTarget]()
	require.NoError(t, err)
	_, err = GetModelSchema[TestModelUnmatchedSource]()
	require.NoError(t, err)

	var logBuf bytes.Buffer
	handler := slog.NewTextHandler(&logBuf, nil)
	origLogger := slog.Default()
	slog.SetDefault(slog.New(handler))
	t.Cleanup(func() {
		slog.SetDefault(origLogger)
	})

	qs, err := NewQuerySet[TestModelUnmatchedSource]("unmatched_sources")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	ctx := context.Background()

	// First query: join is skipped, rows are returned, warning logged.
	rows1, err := qs.SelectRelated("UnmatchedRel").All(ctx)
	require.NoError(t, err)
	require.Len(t, rows1, 1)
	assert.Equal(t, "Unmatched Row", rows1[0].Title)

	// Second query: join is skipped, rows are returned, no duplicate warning.
	rows2, err := qs.SelectRelated("UnmatchedRel").All(ctx)
	require.NoError(t, err)
	require.Len(t, rows2, 1)
	assert.Equal(t, "Unmatched Row", rows2[0].Title)

	logOutput := logBuf.String()
	warnCount := strings.Count(logOutput, "select_related skipped: no foreign key column")
	assert.Equal(t, 1, warnCount, "expected exactly one warning for two queries, got log: %s", logOutput)
	assert.Contains(t, logOutput, "model=TestModelUnmatchedSource")
	assert.Contains(t, logOutput, "relation=UnmatchedRel")
	assert.Contains(t, logOutput, "expected_column=unmatched_rel_id")
}

// Author and Book define mutually related many-to-many models.
type Author struct {
	schema.BaseSchema
	ID    int64   `db:"id"`
	Name  string  `db:"name"`
	Books []*Book `db:"books"`
}

func (Author) Meta() schema.Meta {
	return schema.Meta{TableName: "authors"}
}

func (Author) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

func (Author) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ManyToManyField("Books", "Book", schema.Through("author_books")),
	}
}

type Book struct {
	schema.BaseSchema
	ID      int64     `db:"id"`
	Title   string    `db:"title"`
	Authors []*Author `db:"authors"`
}

func (Book) Meta() schema.Meta {
	return schema.Meta{TableName: "books"}
}

func (Book) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title"),
	}
}

func (Book) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ManyToManyField("Authors", "Author", schema.Through("author_books")),
	}
}

func TestBuildModelSchema_MutuallyRelatedManyToMany(t *testing.T) {
	// Isolate and reset registry state and schema cache for test models.
	schemaMu.Lock()
	delete(schemaCache, reflect.TypeOf(Author{}))
	delete(schemaCache, reflect.TypeOf(Book{}))
	schemaMu.Unlock()

	schemaNameMu.Lock()
	delete(schemaNameRegistry, "Author")
	delete(schemaNameRegistry, "Book")
	schemaNameMu.Unlock()

	t.Cleanup(func() {
		schemaMu.Lock()
		delete(schemaCache, reflect.TypeOf(Author{}))
		delete(schemaCache, reflect.TypeOf(Book{}))
		schemaMu.Unlock()

		schemaNameMu.Lock()
		delete(schemaNameRegistry, "Author")
		delete(schemaNameRegistry, "Book")
		schemaNameMu.Unlock()
	})

	RegisterModelType("Author", reflect.TypeOf(Author{}))
	RegisterModelType("Book", reflect.TypeOf(Book{}))

	authorSchema, err := GetModelSchema[Author]()
	require.NoError(t, err)
	require.NotNil(t, authorSchema)

	bookSchema, err := GetModelSchema[Book]()
	require.NoError(t, err)
	require.NotNil(t, bookSchema)

	require.Len(t, authorSchema.Relations, 1)
	assert.Equal(t, "author_id", authorSchema.Relations[0].ThroughSourceColumn)
	assert.Equal(t, "book_id", authorSchema.Relations[0].ThroughTargetColumn)

	require.Len(t, bookSchema.Relations, 1)
	assert.Equal(t, "book_id", bookSchema.Relations[0].ThroughSourceColumn)
	assert.Equal(t, "author_id", bookSchema.Relations[0].ThroughTargetColumn)

	// Also verify direct BuildModelSchema calls on uncached instances.
	schemaMu.Lock()
	delete(schemaCache, reflect.TypeOf(Author{}))
	delete(schemaCache, reflect.TypeOf(Book{}))
	schemaMu.Unlock()

	authorMS, err := BuildModelSchema(Author{})
	require.NoError(t, err)
	require.NotNil(t, authorMS)
	require.Len(t, authorMS.Relations, 1)
	assert.Equal(t, "author_id", authorMS.Relations[0].ThroughSourceColumn)
	assert.Equal(t, "book_id", authorMS.Relations[0].ThroughTargetColumn)

	bookMS, err := BuildModelSchema(Book{})
	require.NoError(t, err)
	require.NotNil(t, bookMS)
	require.Len(t, bookMS.Relations, 1)
	assert.Equal(t, "book_id", bookMS.Relations[0].ThroughSourceColumn)
	assert.Equal(t, "author_id", bookMS.Relations[0].ThroughTargetColumn)
}
