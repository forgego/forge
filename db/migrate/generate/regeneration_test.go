package generate_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/db/migrate/generate"
)

// Regression models for regeneration stability. Each exercises a column type,
// default, relation or constraint whose recorded DDL previously failed to
// round-trip, so makemigrations proposed a new migration for an unchanged model.
const functionalModels = `package models

import "github.com/forgego/forge/schema"

type Author struct {
	schema.BaseSchema
}

func (Author) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.Unique()),
		schema.Float64Field("score", schema.Default(1.5)),
		schema.Float32Field("ratio"),
		schema.Int32Field("visits", schema.Default(0)),
		schema.StringField("status", schema.Default("Active")),
		schema.StringField("motto", schema.Default("read, then write it's")),
		schema.Int32Field("rank", schema.DBDefault("(1 + 2)")),
		schema.DecimalField("balance", schema.MaxDigits(12), schema.DecimalPlaces(2)),
		schema.DateField("born_on"),
		schema.DateTimeField("seen_at"),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Author) Meta() schema.Meta {
	return schema.Meta{TableName: "authors"}
}

type Book struct {
	schema.BaseSchema
}

func (Book) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("author_id", schema.Required()),
		schema.Int64Field("editor_id"),
		schema.Int64Field("reviewer_id"),
		schema.Int32Field("pages", schema.Default(1)),
		schema.StringField("isbn"),
	}
}

func (Book) Meta() schema.Meta {
	return schema.Meta{
		TableName: "books",
		Indexes:   []schema.Index{{Name: "books_isbn_idx", Fields: []string{"isbn"}}},
		Constraints: []schema.Constraint{
			{Name: "books_pages_positive", Type: "CHECK", Condition: "pages > 0"},
			{Name: "books_isbn_key", Type: "UNIQUE", Fields: []string{"isbn"}},
		},
	}
}

func (Book) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("author_id", "Author", schema.OnDelete(schema.CascadeCASCADE)),
		schema.ForeignKeyField("editor_id", "Author", schema.OnDelete(schema.CascadeSET_NULL)),
		schema.OneToOneField("reviewer_id", "Author", schema.OnDelete(schema.CascadePROTECT)),
	}
}
`

// fluentModels uses the builder-chain spelling the parser also reads.
const fluentModels = `package models

import "github.com/forgego/forge/schema"

type User struct {
	schema.BaseSchema
}

func (User) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("username").Required().MaxLength(150).Unique().Build(),
		schema.Float64("score").Build(),
		schema.Int32("visits").Build(),
		schema.DateTime("created_at").Build(),
	}
}

func (User) Meta() schema.Meta {
	return schema.Meta{TableName: "users"}
}

type Post struct {
	schema.BaseSchema
}

func (Post) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.Int64("user_id").Required().Build(),
	}
}

func (Post) Meta() schema.Meta {
	return schema.Meta{TableName: "posts"}
}

func (Post) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("user_id", "User").OnDelete("CASCADE").Build(),
	}
}
`

func migrationFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names
}

func generateMigration(t *testing.T, modelsDir, migrationsDir string, driver core.Driver, name string) {
	t.Helper()
	gen, err := generate.NewMigrationGeneratorForDriver(modelsDir, migrationsDir, driver)
	if err != nil {
		t.Fatal(err)
	}
	if err := gen.GenerateMigrations(name); err != nil {
		t.Fatalf("generate %s: %v", name, err)
	}
}

// TestRegenerationOfUnchangedModelsWritesNothing covers PostgreSQL and SQLite,
// whose foreign keys and constraints are declared inside CREATE TABLE and read
// back from there. The initial migration must also apply (see applyMigrations).
func TestRegenerationOfUnchangedModelsWritesNothing(t *testing.T) {
	for _, driver := range []core.Driver{core.DriverPostgreSQL, core.DriverSQLite} {
		for name, source := range map[string]string{"functional": functionalModels, "fluent": fluentModels} {
			t.Run(string(driver)+"/"+name, func(t *testing.T) {
				modelsDir := t.TempDir()
				migrationsDir := t.TempDir()
				if err := os.WriteFile(filepath.Join(modelsDir, "models.go"), []byte(source), 0o600); err != nil {
					t.Fatal(err)
				}
				generateMigration(t, modelsDir, migrationsDir, driver, "initial")
				initial := migrationFiles(t, migrationsDir)
				if len(initial) != 2 {
					t.Fatalf("initial generation wrote %v", initial)
				}
				for i := 0; i < 2; i++ {
					generateMigration(t, modelsDir, migrationsDir, driver, "again")
					if got := migrationFiles(t, migrationsDir); len(got) != 2 {
						extra, _ := os.ReadFile(filepath.Join(migrationsDir, got[len(got)-1]))
						t.Fatalf("regenerating unchanged models wrote %v:\n%s", got, extra)
					}
				}
				applyMigrations(t, driver, migrationsDir)
			})
		}
	}
}

func TestInitialMigrationCreatesDeclaredConstraintsAndForeignKeys(t *testing.T) {
	modelsDir := t.TempDir()
	migrationsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(modelsDir, "models.go"), []byte(functionalModels), 0o600); err != nil {
		t.Fatal(err)
	}
	generateMigration(t, modelsDir, migrationsDir, core.DriverPostgreSQL, "initial")
	up, err := os.ReadFile(filepath.Join(migrationsDir, "000001_initial.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`ADD CONSTRAINT fk_books_author_id FOREIGN KEY ("author_id") REFERENCES authors (id) ON DELETE CASCADE ON UPDATE NO ACTION;`,
		`ADD CONSTRAINT fk_books_editor_id FOREIGN KEY ("editor_id") REFERENCES authors (id) ON DELETE SET NULL ON UPDATE NO ACTION;`,
		`ADD CONSTRAINT fk_books_reviewer_id FOREIGN KEY ("reviewer_id") REFERENCES authors (id) ON DELETE RESTRICT ON UPDATE NO ACTION;`,
		`ALTER TABLE books ADD CONSTRAINT books_pages_positive CHECK (pages > 0);`,
		`ALTER TABLE books ADD CONSTRAINT books_isbn_key UNIQUE ("isbn");`,
		`CREATE INDEX IF NOT EXISTS books_isbn_idx ON books ("isbn");`,
	} {
		if !strings.Contains(string(up), want) {
			t.Errorf("initial migration lacks %s\n%s", want, up)
		}
	}
}

// TestChangedForeignKeyActionIsDetected keeps the stability fix from masking
// real relation changes.
func TestChangedForeignKeyActionIsDetected(t *testing.T) {
	modelsDir := t.TempDir()
	migrationsDir := t.TempDir()
	file := filepath.Join(modelsDir, "models.go")
	if err := os.WriteFile(file, []byte(functionalModels), 0o600); err != nil {
		t.Fatal(err)
	}
	generateMigration(t, modelsDir, migrationsDir, core.DriverPostgreSQL, "initial")
	changed := strings.Replace(functionalModels, `schema.OnDelete(schema.CascadeSET_NULL)`, `schema.OnDelete(schema.CascadeCASCADE)`, 1)
	changed = strings.Replace(changed, `schema.Float32Field("ratio")`, `schema.Float64Field("ratio")`, 1)
	if err := os.WriteFile(file, []byte(changed), 0o600); err != nil {
		t.Fatal(err)
	}
	generateMigration(t, modelsDir, migrationsDir, core.DriverPostgreSQL, "change")
	up, err := os.ReadFile(filepath.Join(migrationsDir, "000002_change.up.sql"))
	if err != nil {
		t.Fatalf("changed model produced no migration: %v", err)
	}
	for _, want := range []string{
		`ALTER TABLE books DROP CONSTRAINT IF EXISTS fk_books_editor_id;`,
		`FOREIGN KEY ("editor_id") REFERENCES authors (id) ON DELETE CASCADE`,
		`ALTER TABLE authors ALTER COLUMN "ratio" TYPE DOUBLE PRECISION`,
	} {
		if !strings.Contains(string(up), want) {
			t.Errorf("change migration lacks %s\n%s", want, up)
		}
	}
	for _, unwanted := range []string{"fk_books_author_id", "fk_books_reviewer_id", `"score"`, `"visits"`, "books_pages_positive"} {
		if strings.Contains(string(up), unwanted) {
			t.Errorf("change migration touches unchanged %s\n%s", unwanted, up)
		}
	}
}

// regenerateWith generates the initial migration for functionalModels, then
// one named name for source, and returns that migration's up and down SQL. A
// further regeneration must write nothing, and the migrations must apply up,
// down and up again (see applyMigrations).
func regenerateWith(t *testing.T, driver core.Driver, source, name string) (string, string) {
	t.Helper()
	return regenerateFrom(t, driver, functionalModels, source, name)
}

// regenerateFrom is regenerateWith for an initial migration generated from
// initial instead of functionalModels.
func regenerateFrom(t *testing.T, driver core.Driver, initial, source, name string) (string, string) {
	t.Helper()
	modelsDir := t.TempDir()
	migrationsDir := t.TempDir()
	file := filepath.Join(modelsDir, "models.go")
	if err := os.WriteFile(file, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}
	generateMigration(t, modelsDir, migrationsDir, driver, "initial")
	if err := os.WriteFile(file, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	generateMigration(t, modelsDir, migrationsDir, driver, name)
	up, err := os.ReadFile(filepath.Join(migrationsDir, "000002_"+name+".up.sql"))
	if err != nil {
		t.Fatalf("changed model produced no migration: %v", err)
	}
	down, err := os.ReadFile(filepath.Join(migrationsDir, "000002_"+name+".down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	generateMigration(t, modelsDir, migrationsDir, driver, "again")
	if got := migrationFiles(t, migrationsDir); len(got) != 4 {
		extra, _ := os.ReadFile(filepath.Join(migrationsDir, got[len(got)-1]))
		t.Fatalf("regenerating after %s wrote %v:\n%s", name, got, extra)
	}
	applyMigrations(t, driver, migrationsDir)
	return string(up), string(down)
}

func mustReplace(t *testing.T, source, old, replacement string) string {
	t.Helper()
	if !strings.Contains(source, old) {
		t.Fatalf("fixture lacks %q", old)
	}
	return strings.Replace(source, old, replacement, 1)
}

func assertContainsAll(t *testing.T, label, got string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Errorf("%s lacks %s\n%s", label, want, got)
		}
	}
}

// TestDroppedColumnIsReadBack covers ALTER TABLE .. DROP COLUMN, which used to
// be classified as an unparsed ALTER TABLE, so every later run dropped the
// column again.
func TestDroppedColumnIsReadBack(t *testing.T) {
	source := mustReplace(t, functionalModels, "\t\tschema.Float32Field(\"ratio\"),\n", "")
	up, down := regenerateWith(t, core.DriverPostgreSQL, source, "drop_ratio")
	assertContainsAll(t, "up", up, `ALTER TABLE authors DROP COLUMN IF EXISTS ratio;`)
	assertContainsAll(t, "down", down, `ALTER TABLE authors ADD COLUMN "ratio" REAL;`)
}

// TestDroppedForeignKeyIsRestoredOnDown covers a model that drops a relation
// but keeps its column: the down migration must re-add the recorded FK.
func TestDroppedForeignKeyIsRestoredOnDown(t *testing.T) {
	source := mustReplace(t, functionalModels,
		"\t\tschema.OneToOneField(\"reviewer_id\", \"Author\", schema.OnDelete(schema.CascadePROTECT)),\n", "")
	up, down := regenerateWith(t, core.DriverPostgreSQL, source, "drop_reviewer_fk")
	assertContainsAll(t, "up", up, `ALTER TABLE books DROP CONSTRAINT IF EXISTS fk_books_reviewer_id;`)
	assertContainsAll(t, "down", down,
		`ALTER TABLE books ADD CONSTRAINT fk_books_reviewer_id FOREIGN KEY ("reviewer_id") REFERENCES authors (id) ON DELETE RESTRICT ON UPDATE NO ACTION;`)
	for _, unwanted := range []string{"fk_books_author_id", "fk_books_editor_id", "DROP COLUMN"} {
		if strings.Contains(up, unwanted) {
			t.Errorf("up touches unchanged %s\n%s", unwanted, up)
		}
	}
}

// TestDroppedForeignKeyColumnIsRestoredBeforeItsForeignKey drops a relation
// with its column: the down migration must re-add the column before the FK.
func TestDroppedForeignKeyColumnIsRestoredBeforeItsForeignKey(t *testing.T) {
	source := mustReplace(t, functionalModels,
		"\t\tschema.OneToOneField(\"reviewer_id\", \"Author\", schema.OnDelete(schema.CascadePROTECT)),\n", "")
	source = mustReplace(t, source, "\t\tschema.Int64Field(\"reviewer_id\"),\n", "")
	up, down := regenerateWith(t, core.DriverPostgreSQL, source, "drop_reviewer")
	dropFK := strings.Index(up, "DROP CONSTRAINT IF EXISTS fk_books_reviewer_id")
	dropColumn := strings.Index(up, "DROP COLUMN IF EXISTS reviewer_id")
	if dropFK < 0 || dropColumn < 0 || dropFK > dropColumn {
		t.Errorf("up must drop the FK, then the column:\n%s", up)
	}
	addColumn := strings.Index(down, `ADD COLUMN "reviewer_id"`)
	addFK := strings.Index(down, "ADD CONSTRAINT fk_books_reviewer_id")
	if addColumn < 0 || addFK < 0 || addColumn > addFK {
		t.Errorf("down must add the column, then the FK:\n%s", down)
	}
}

// TestDroppedConstraintIsRestoredOnDown covers a model that drops a Meta
// constraint: the down migration must re-add the recorded definition.
func TestDroppedConstraintIsRestoredOnDown(t *testing.T) {
	source := mustReplace(t, functionalModels,
		"\t\t\t{Name: \"books_pages_positive\", Type: \"CHECK\", Condition: \"pages > 0\"},\n", "")
	up, down := regenerateWith(t, core.DriverPostgreSQL, source, "drop_pages_check")
	assertContainsAll(t, "up", up, `ALTER TABLE books DROP CONSTRAINT IF EXISTS books_pages_positive;`)
	assertContainsAll(t, "down", down, `ALTER TABLE books ADD CONSTRAINT books_pages_positive CHECK (pages > 0);`)
	if strings.Contains(up, "books_isbn_key") {
		t.Errorf("up touches the unchanged UNIQUE constraint:\n%s", up)
	}
}

// TestModifiedColumnIsReadBack covers ALTER TABLE .. ALTER COLUMN, which was
// not read back, so every later run modified the column again.
func TestModifiedColumnIsReadBack(t *testing.T) {
	source := mustReplace(t, functionalModels, `schema.Float32Field("ratio")`,
		`schema.Float64Field("ratio", schema.Required(), schema.Default(2.5))`)
	source = mustReplace(t, source, `schema.Int32Field("visits", schema.Default(0))`, `schema.Int32Field("visits")`)
	up, down := regenerateWith(t, core.DriverPostgreSQL, source, "modify_columns")
	assertContainsAll(t, "up", up,
		`ALTER TABLE authors ALTER COLUMN "ratio" TYPE DOUBLE PRECISION USING ("ratio"::DOUBLE PRECISION);`,
		`ALTER TABLE authors ALTER COLUMN "ratio" SET NOT NULL;`,
		`ALTER TABLE authors ALTER COLUMN "ratio" SET DEFAULT 2.500000;`,
		`ALTER TABLE authors ALTER COLUMN "visits" DROP DEFAULT;`)
	assertContainsAll(t, "down", down,
		`ALTER TABLE authors ALTER COLUMN "ratio" TYPE REAL USING ("ratio"::REAL);`,
		`ALTER TABLE authors ALTER COLUMN "ratio" DROP NOT NULL;`,
		`ALTER TABLE authors ALTER COLUMN "ratio" DROP DEFAULT;`,
		`ALTER TABLE authors ALTER COLUMN "visits" SET DEFAULT 0;`)
}

// TestChangedConstraintIsDetected covers a constraint that keeps its name but
// changes its CHECK condition or UNIQUE fields: it is dropped and re-added.
func TestChangedConstraintIsDetected(t *testing.T) {
	source := mustReplace(t, functionalModels, `Condition: "pages > 0"`, `Condition: "pages > 1"`)
	source = mustReplace(t, source, `Fields: []string{"isbn"}},`, `Fields: []string{"isbn", "pages"}},`)
	up, down := regenerateWith(t, core.DriverPostgreSQL, source, "change_constraints")
	assertContainsAll(t, "up", up,
		`ALTER TABLE books DROP CONSTRAINT IF EXISTS books_pages_positive;`,
		`ALTER TABLE books ADD CONSTRAINT books_pages_positive CHECK (pages > 1);`,
		`ALTER TABLE books DROP CONSTRAINT IF EXISTS books_isbn_key;`,
		`ALTER TABLE books ADD CONSTRAINT books_isbn_key UNIQUE ("isbn", "pages");`)
	assertContainsAll(t, "down", down,
		`ALTER TABLE books ADD CONSTRAINT books_pages_positive CHECK (pages > 0);`,
		`ALTER TABLE books ADD CONSTRAINT books_isbn_key UNIQUE ("isbn");`)
	if strings.Index(up, "DROP CONSTRAINT IF EXISTS books_pages_positive") > strings.Index(up, "ADD CONSTRAINT books_pages_positive") {
		t.Errorf("up must drop the old constraint before adding the new one:\n%s", up)
	}
}

// TestPostgresArrayColumnIsReadBack covers PostgreSQL array types, whose []
// suffix the parser dropped, so the column was modified on every run and the
// down migration cast it to the element type.
func TestPostgresArrayColumnIsReadBack(t *testing.T) {
	withTags := mustReplace(t, functionalModels, "\t\tschema.StringField(\"isbn\"),\n",
		"\t\tschema.StringField(\"isbn\"),\n\t\tschema.StringField(\"tags\", schema.DBType(\"TEXT[]\")),\n")
	up, _ := regenerateWith(t, core.DriverPostgreSQL, withTags, "add_tags")
	assertContainsAll(t, "up", up, `ALTER TABLE books ADD COLUMN "tags" TEXT[];`)

	changed := mustReplace(t, withTags, `schema.DBType("TEXT[]")`, `schema.DBType("VARCHAR(40)[]")`)
	up, down := regenerateFrom(t, core.DriverPostgreSQL, withTags, changed, "change_tags")
	assertContainsAll(t, "up", up, `ALTER TABLE books ALTER COLUMN "tags" TYPE VARCHAR(40)[] USING ("tags"::VARCHAR(40)[]);`)
	assertContainsAll(t, "down", down, `ALTER TABLE books ALTER COLUMN "tags" TYPE TEXT[] USING ("tags"::TEXT[]);`)
}

// TestSQLiteDeclaresNewTableConstraintsInCreateTable covers SQLite, which has
// no ALTER TABLE .. ADD CONSTRAINT: a new table's Meta constraints go inside
// CREATE TABLE, and are read back from there.
func TestSQLiteDeclaresNewTableConstraintsInCreateTable(t *testing.T) {
	modelsDir := t.TempDir()
	migrationsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(modelsDir, "models.go"), []byte(functionalModels), 0o600); err != nil {
		t.Fatal(err)
	}
	generateMigration(t, modelsDir, migrationsDir, core.DriverSQLite, "initial")
	up, err := os.ReadFile(filepath.Join(migrationsDir, "000001_initial.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	assertContainsAll(t, "up", string(up),
		`CONSTRAINT books_pages_positive CHECK (pages > 0)`,
		`CONSTRAINT books_isbn_key UNIQUE ("isbn")`)
	if strings.Contains(string(up), "ADD CONSTRAINT") {
		t.Errorf("SQLite migration uses ALTER TABLE .. ADD CONSTRAINT:\n%s", up)
	}
	generateMigration(t, modelsDir, migrationsDir, core.DriverSQLite, "again")
	if got := migrationFiles(t, migrationsDir); len(got) != 2 {
		extra, _ := os.ReadFile(filepath.Join(migrationsDir, got[len(got)-1]))
		t.Fatalf("regenerating unchanged models wrote %v:\n%s", got, extra)
	}
}

// TestDefaultsWithSpacesAreReadBack covers string defaults with spaces,
// commas and escaped quotes, and parenthesized expressions, which the parser
// truncated at the first space or comma, so every run re-set the default and
// the down migration restored a truncated one.
func TestDefaultsWithSpacesAreReadBack(t *testing.T) {
	source := mustReplace(t, functionalModels, `schema.Default("read, then write it's")`, `schema.Default("read, then write")`)
	source = mustReplace(t, source, "\t\tschema.StringField(\"isbn\"),\n",
		"\t\tschema.StringField(\"isbn\"),\n\t\tschema.StringField(\"blurb\", schema.Default(\"a b, 'c'\")),\n")
	up, down := regenerateWith(t, core.DriverPostgreSQL, source, "change_defaults")
	assertContainsAll(t, "up", up,
		`ALTER TABLE authors ALTER COLUMN "motto" SET DEFAULT 'read, then write';`,
		`ALTER TABLE books ADD COLUMN "blurb" TEXT DEFAULT 'a b, ''c''';`)
	assertContainsAll(t, "down", down, `ALTER TABLE authors ALTER COLUMN "motto" SET DEFAULT 'read, then write it''s';`)
	if strings.Contains(up, `"rank"`) {
		t.Errorf("up touches the unchanged expression default:\n%s", up)
	}
}

// TestSQLiteFunctionalModelsApply covers auto timestamps on SQLite, which
// rejects DEFAULT now(): the SQLite migration uses CURRENT_TIMESTAMP, applies
// to a real database, and reads back without churn.
func TestSQLiteFunctionalModelsApply(t *testing.T) {
	modelsDir := t.TempDir()
	migrationsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(modelsDir, "models.go"), []byte(functionalModels), 0o600); err != nil {
		t.Fatal(err)
	}
	generateMigration(t, modelsDir, migrationsDir, core.DriverSQLite, "initial")
	up, err := os.ReadFile(filepath.Join(migrationsDir, "000001_initial.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	assertContainsAll(t, "up", string(up),
		`"created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL`,
		`"updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP`)
	if strings.Contains(string(up), "now()") {
		t.Errorf("SQLite migration uses now():\n%s", up)
	}
	applyMigrations(t, core.DriverSQLite, migrationsDir)
	generateMigration(t, modelsDir, migrationsDir, core.DriverSQLite, "again")
	if got := migrationFiles(t, migrationsDir); len(got) != 2 {
		extra, _ := os.ReadFile(filepath.Join(migrationsDir, got[len(got)-1]))
		t.Fatalf("regenerating unchanged models wrote %v:\n%s", got, extra)
	}
}

// TestChangedStringDefaultCaseIsDetected covers a default that differs only in
// case: string literals must compare exactly.
func TestChangedStringDefaultCaseIsDetected(t *testing.T) {
	source := mustReplace(t, functionalModels, `schema.Default("Active")`, `schema.Default("active")`)
	up, down := regenerateWith(t, core.DriverPostgreSQL, source, "lower_status")
	assertContainsAll(t, "up", up, `ALTER TABLE authors ALTER COLUMN "status" SET DEFAULT 'active';`)
	assertContainsAll(t, "down", down, `ALTER TABLE authors ALTER COLUMN "status" SET DEFAULT 'Active';`)
}

// withBookColumns adds column declarations after the isbn column of
// functionalModels.
func withBookColumns(t *testing.T, columns ...string) string {
	t.Helper()
	isbn := "\t\tschema.StringField(\"isbn\"),\n"
	added := isbn
	for _, column := range columns {
		added += "\t\t" + column + ",\n"
	}
	return mustReplace(t, functionalModels, isbn, added)
}

// TestCastAndOperatorDBDefaultsAreReadBack covers DB defaults with a cast or
// an operator, which the parser cut at the first space and unquoted, so every
// run dropped the default and the down migration restored an invalid one.
func TestCastAndOperatorDBDefaultsAreReadBack(t *testing.T) {
	source := withBookColumns(t,
		`schema.StringField("meta", schema.DBType("JSONB"), schema.DBDefault("'{}'::jsonb"), schema.Required())`,
		`schema.StringField("joined", schema.DBDefault("'x' || 'y'"))`,
		`schema.StringField("label", schema.DBDefault("'Mixed Case'::text"), schema.Unique())`)
	columns := []string{
		`ALTER TABLE books ADD COLUMN "meta" JSONB NOT NULL DEFAULT '{}'::jsonb;`,
		`ALTER TABLE books ADD COLUMN "joined" TEXT DEFAULT 'x' || 'y';`,
		`ALTER TABLE books ADD COLUMN "label" TEXT DEFAULT 'Mixed Case'::text UNIQUE;`,
	}
	up, _ := regenerateWith(t, core.DriverPostgreSQL, source, "add_db_defaults")
	assertContainsAll(t, "up", up, columns...)

	// Dropping the columns restores their defaults whole on down.
	_, down := regenerateFrom(t, core.DriverPostgreSQL, source, functionalModels, "drop_db_defaults")
	assertContainsAll(t, "down", down, columns...)

	// SQLite requires an expression default to be parenthesized.
	source = withBookColumns(t, `schema.StringField("joined", schema.DBDefault("('x' || 'y')"), schema.Required())`)
	up, _ = regenerateWith(t, core.DriverSQLite, source, "add_db_defaults")
	assertContainsAll(t, "up", up, `ALTER TABLE books ADD COLUMN "joined" TEXT NOT NULL DEFAULT ('x' || 'y');`)
}

// TestChangedForeignKeyTargetIsRestoredOnDown covers a foreign key whose
// target table changes: the down migration used to re-add the old foreign key
// against the new target.
func TestChangedForeignKeyTargetIsRestoredOnDown(t *testing.T) {
	source := mustReplace(t, functionalModels, `schema.ForeignKeyField("editor_id", "Author", schema.OnDelete(schema.CascadeSET_NULL))`,
		`schema.ForeignKeyField("editor_id", "Book", schema.OnDelete(schema.CascadeSET_NULL))`)
	up, down := regenerateWith(t, core.DriverPostgreSQL, source, "retarget_editor")
	assertContainsAll(t, "up", up,
		`ALTER TABLE books DROP CONSTRAINT IF EXISTS fk_books_editor_id;`,
		`ADD CONSTRAINT fk_books_editor_id FOREIGN KEY ("editor_id") REFERENCES books (id) ON DELETE SET NULL`)
	assertContainsAll(t, "down", down,
		`ALTER TABLE books DROP CONSTRAINT IF EXISTS fk_books_editor_id;`,
		`ADD CONSTRAINT fk_books_editor_id FOREIGN KEY ("editor_id") REFERENCES authors (id) ON DELETE SET NULL`)
}

// TestStringDefaultWithParenthesesIsQuoted covers Default values that look
// like a function call, which were rendered unquoted and produced invalid
// SQL. Only DBDefault is an expression.
func TestStringDefaultWithParenthesesIsQuoted(t *testing.T) {
	source := withBookColumns(t, `schema.StringField("note", schema.Default("x(1)"))`, `schema.StringField("aside", schema.Default("(see below)"))`)
	for _, driver := range []core.Driver{core.DriverPostgreSQL, core.DriverSQLite} {
		up, _ := regenerateWith(t, driver, source, "add_notes")
		assertContainsAll(t, string(driver)+" up", up,
			`ALTER TABLE books ADD COLUMN "note" TEXT DEFAULT 'x(1)';`,
			`ALTER TABLE books ADD COLUMN "aside" TEXT DEFAULT '(see below)';`)
	}
}

// TestSQLiteColumnChangeFails covers SQLite, which cannot alter a column
// without rebuilding its table: makemigrations used to write only a comment,
// so the change was proposed again on every run. It now fails and writes
// nothing.
func TestSQLiteColumnChangeFails(t *testing.T) {
	modelsDir := t.TempDir()
	migrationsDir := t.TempDir()
	file := filepath.Join(modelsDir, "models.go")
	if err := os.WriteFile(file, []byte(functionalModels), 0o600); err != nil {
		t.Fatal(err)
	}
	generateMigration(t, modelsDir, migrationsDir, core.DriverSQLite, "initial")
	changed := mustReplace(t, functionalModels, `schema.Int32Field("visits", schema.Default(0))`, `schema.Int32Field("visits", schema.Default(5))`)
	if err := os.WriteFile(file, []byte(changed), 0o600); err != nil {
		t.Fatal(err)
	}
	gen, err := generate.NewMigrationGeneratorForDriver(modelsDir, migrationsDir, core.DriverSQLite)
	if err != nil {
		t.Fatal(err)
	}
	err = gen.GenerateMigrations("change_visits")
	if err == nil || !strings.Contains(err.Error(), "modify column authors.visits is not supported on SQLite") {
		t.Fatalf("SQLite column change: got error %v", err)
	}
	if got := migrationFiles(t, migrationsDir); len(got) != 2 {
		t.Fatalf("a failed generation wrote %v", got)
	}
}

// TestChangedDBDefaultIsModified covers PostgreSQL ALTER COLUMN for DB
// defaults, which the builder ignored: a changed expression was dropped, and
// a DB default added to an existing column wrote an empty migration.
func TestChangedDBDefaultIsModified(t *testing.T) {
	source := mustReplace(t, functionalModels, `schema.DBDefault("(1 + 2)")`, `schema.DBDefault("(2 + 3)")`)
	source = mustReplace(t, source, `schema.StringField("isbn")`, `schema.StringField("isbn", schema.DBDefault("'none'::text"))`)
	source = mustReplace(t, source, `schema.Int32Field("visits", schema.Default(0))`, `schema.Int32Field("visits", schema.DBDefault("(0 + 1)"))`)
	source = mustReplace(t, source, `schema.DateField("born_on")`, `schema.DateField("born_on", schema.AutoNowAdd())`)
	up, down := regenerateWith(t, core.DriverPostgreSQL, source, "change_db_defaults")
	assertContainsAll(t, "up", up,
		`ALTER TABLE authors ALTER COLUMN "rank" SET DEFAULT (2 + 3);`,
		`ALTER TABLE books ALTER COLUMN "isbn" SET DEFAULT 'none'::text;`,
		`ALTER TABLE authors ALTER COLUMN "visits" SET DEFAULT (0 + 1);`,
		`ALTER TABLE authors ALTER COLUMN "born_on" SET NOT NULL;`,
		`ALTER TABLE authors ALTER COLUMN "born_on" SET DEFAULT now();`)
	assertContainsAll(t, "down", down,
		`ALTER TABLE authors ALTER COLUMN "rank" SET DEFAULT (1 + 2);`,
		`ALTER TABLE books ALTER COLUMN "isbn" DROP DEFAULT;`,
		`ALTER TABLE authors ALTER COLUMN "visits" SET DEFAULT 0;`,
		`ALTER TABLE authors ALTER COLUMN "born_on" DROP NOT NULL;`,
		`ALTER TABLE authors ALTER COLUMN "born_on" DROP DEFAULT;`)
	if strings.Contains(up, "DROP DEFAULT") {
		t.Errorf("up drops a default:\n%s", up)
	}
}
