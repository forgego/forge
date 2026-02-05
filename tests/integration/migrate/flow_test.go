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
	"github.com/forgego/forge/tests/infra/docker"
	"github.com/forgego/forge/tests/infra/filesystem"
)

// TestMigrationFlow_MigrateToVersion tests migrating to a specific version
func TestMigrationFlow_MigrateToVersion(t *testing.T) {
	opts := docker.PostgresOpts{
		UseDirect: true,
		Host:      "127.0.0.1",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_migrate_to_%d", time.Now().UnixNano()),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	postgresDB, dsn, cleanup, err := docker.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	tempDir, cleanupTemp := filesystem.TempDirInTests(t, "migration_migrate_to_")
	defer cleanupTemp()
	modelsDir := filepath.Join(tempDir, "models")
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	gen, err := migrate.NewGenerator(modelsDir, migrationsDir)
	require.NoError(t, err)

	// Create 5 migrations
	for i := 1; i <= 5; i++ {
		modelContent := fmt.Sprintf(`package models

import "github.com/forgego/forge/schema"

type Item%d struct {
	schema.BaseSchema
}

func (Item%d) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().MaxLength(255).Build(),
	}
}

func (Item%d) Meta() schema.Meta {
	return schema.Meta{
		TableName: "items_%d",
	}
}

func (Item%d) Relations() []schema.Relation {
	return []schema.Relation{}
}
`, i, i, i, i, i)

		modelFile := filepath.Join(modelsDir, fmt.Sprintf("item%d.go", i))
		require.NoError(t, os.WriteFile(modelFile, []byte(modelContent), 0644))

		err = gen.GenerateMigrations(fmt.Sprintf("create_items_%d", i))
		require.NoError(t, err)
	}

	// Migrate to version 3
	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	err = runner.MigrateTo(ctx, 3)
	require.NoError(t, err)

	// Verify we're at version 3
	version, dirty, err := runner.Version(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint(3), version)
	assert.False(t, dirty)

	// Verify tables 1, 2, 3 exist but 4, 5 don't
	for i := 1; i <= 5; i++ {
		var exists bool
		err = database.DB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)
		`, fmt.Sprintf("items_%d", i)).Scan(&exists)
		require.NoError(t, err)

		if i <= 3 {
			assert.True(t, exists, "items_%d table should exist", i)
		} else {
			assert.False(t, exists, "items_%d table should not exist", i)
		}
	}

	// Now migrate to version 5
	err = runner.MigrateTo(ctx, 5)
	require.NoError(t, err)

	version, _, err = runner.Version(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint(5), version)

	// Verify all tables exist now
	for i := 1; i <= 5; i++ {
		var exists bool
		err = database.DB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)
		`, fmt.Sprintf("items_%d", i)).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "items_%d table should exist after migrating to version 5", i)
	}
}

// TestMigrationFlow_NoChangesDetected tests that no migration is generated when there are no changes
func TestMigrationFlow_NoChangesDetected(t *testing.T) {
	opts := docker.PostgresOpts{
		UseDirect: true,
		Host:      "127.0.0.1",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_no_changes_%d", time.Now().UnixNano()),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	postgresDB, dsn, cleanup, err := docker.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	tempDir, cleanupTemp := filesystem.TempDirInTests(t, "migration_no_changes_")
	defer cleanupTemp()
	modelsDir := filepath.Join(tempDir, "models")
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create initial model
	modelContent := `package models

import "github.com/forgego/forge/schema"

type Author struct {
	schema.BaseSchema
}

func (Author) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().MaxLength(255).Build(),
	}
}

func (Author) Meta() schema.Meta {
	return schema.Meta{
		TableName: "authors",
	}
}

func (Author) Relations() []schema.Relation {
	return []schema.Relation{}
}
`
	modelFile := filepath.Join(modelsDir, "author.go")
	require.NoError(t, os.WriteFile(modelFile, []byte(modelContent), 0644))

	gen, err := migrate.NewGenerator(modelsDir, migrationsDir)
	require.NoError(t, err)

	// Generate first migration
	err = gen.GenerateMigrations("create_authors")
	require.NoError(t, err)

	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Count migration files before
	entries, err := os.ReadDir(migrationsDir)
	require.NoError(t, err)
	initialFileCount := len(entries)

	// Try to generate migration again with same model (should detect no changes)
	err = gen.GenerateMigrations("create_authors_again")
	// This should either return nil (no changes) or not create new files
	// The generator returns nil when no changes are detected
	if err == nil {
		// Check that no new files were created
		entries, err = os.ReadDir(migrationsDir)
		require.NoError(t, err)
		assert.Equal(t, initialFileCount, len(entries), "no new migration files should be created when there are no changes")
	}
}
