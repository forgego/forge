package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/migrate"
	"github.com/forgego/forge/tests/testhelpers"
)

// TestGeneration_CreateTable tests CreateTable change detection
func TestGeneration_CreateTable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "127.0.0.1",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "generation_create_table_")
	defer cleanupTemp()
	modelsDir := filepath.Join(tempDir, "models")
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create a simple User model
	modelContent := `package models

import "github.com/forgego/forge/schema"

type User struct {
	schema.BaseSchema
}

func (User) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("email").Required().MaxLength(255).Unique().Build(),
		schema.String("username").Required().MaxLength(150).Build(),
	}
}

func (User) Meta() schema.Meta {
	return schema.Meta{
		TableName: "users",
	}
}

func (User) Relations() []schema.Relation {
	return []schema.Relation{}
}
`
	modelFile := filepath.Join(modelsDir, "user.go")
	require.NoError(t, os.WriteFile(modelFile, []byte(modelContent), 0644))

	// Generate migration
	gen, err := migrate.NewGenerator(modelsDir, migrationsDir)
	require.NoError(t, err)

	err = gen.GenerateMigrations("create_users")
	require.NoError(t, err)

	// Verify migration files were created
	upFile := filepath.Join(migrationsDir, "000001_create_users.up.sql")
	downFile := filepath.Join(migrationsDir, "000001_create_users.down.sql")
	require.FileExists(t, upFile)
	require.FileExists(t, downFile)

	// Read and verify up migration contains CREATE TABLE
	upContent, err := os.ReadFile(upFile)
	require.NoError(t, err)
	assert.Contains(t, string(upContent), "CREATE TABLE")
	assert.Contains(t, string(upContent), "users")

	// Read and verify down migration contains DROP TABLE
	downContent, err := os.ReadFile(downFile)
	require.NoError(t, err)
	assert.Contains(t, string(downContent), "DROP TABLE")

	// Apply migration and verify table exists
	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Verify table exists
	testhelpers.AssertTableExists(ctx, t, postgresDB, "postgres", "users")
	testhelpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "users", "id")
	testhelpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "users", "email")
	testhelpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "users", "username")
}

// TestGeneration_AddColumn tests AddColumn change detection
func TestGeneration_AddColumn(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano()),
	}
	postgresDB, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "generation_add_column_")
	defer cleanupTemp()
	modelsDir := filepath.Join(tempDir, "models")
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	// Step 1: Create initial model with basic fields
	modelV1 := `package models

import "github.com/forgego/forge/schema"

type User struct {
	schema.BaseSchema
}

func (User) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("email").Required().MaxLength(255).Unique().Build(),
	}
}

func (User) Meta() schema.Meta {
	return schema.Meta{
		TableName: "users",
	}
}

func (User) Relations() []schema.Relation {
	return []schema.Relation{}
}
`
	modelFile := filepath.Join(modelsDir, "user.go")
	require.NoError(t, os.WriteFile(modelFile, []byte(modelV1), 0644))

	gen, err := migrate.NewGenerator(modelsDir, migrationsDir)
	require.NoError(t, err)

	err = gen.GenerateMigrations("create_users")
	require.NoError(t, err)

	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Step 2: Add a new field (username)
	modelV2 := `package models

import "github.com/forgego/forge/schema"

type User struct {
	schema.BaseSchema
}

func (User) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("email").Required().MaxLength(255).Unique().Build(),
		schema.String("username").Required().MaxLength(150).Build(),
	}
}

func (User) Meta() schema.Meta {
	return schema.Meta{
		TableName: "users",
	}
}

func (User) Relations() []schema.Relation {
	return []schema.Relation{}
}
`
	require.NoError(t, os.WriteFile(modelFile, []byte(modelV2), 0644))

	// Create new generator to reload state
	gen2, err := migrate.NewGenerator(modelsDir, migrationsDir)
	require.NoError(t, err)

	err = gen2.GenerateMigrations("add_username_to_users")
	require.NoError(t, err)

	// Verify second migration was created
	migrationFiles, err := testhelpers.GetMigrationFiles(migrationsDir)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(migrationFiles), 2, "should have at least 2 migration files")

	// Apply second migration
	runner2, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner2.Close()

	err = runner2.Migrate(ctx)
	require.NoError(t, err)

	// Verify new column exists
	testhelpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "users", "username")
}
