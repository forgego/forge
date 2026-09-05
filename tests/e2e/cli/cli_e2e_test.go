package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/tests/testhelpers"
)

// TestCLIMakemigrations tests the makemigrations CLI command
func TestCLIMakemigrations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workdir, cleanup := testhelpers.TempWorkdir(t, "forge-e2e-")
	defer cleanup()

	// Create a sample models directory
	modelsDir := filepath.Join(workdir, "models")
	os.MkdirAll(modelsDir, 0755)

	// Write sample model file
	sampleModel := `package models

import "github.com/forgego/forge/schema"

type User struct {
	schema.BaseSchema
}

func (User) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("username").Required().MaxLength(150).Build(),
		schema.String("email").Required().MaxLength(254).Unique().Build(),
	}
}

func (User) Relations() []schema.Relation {
	return []schema.Relation{}
}
`
	testhelpers.WriteFileString(t, filepath.Join(modelsDir, "user.go"), sampleModel)

	// Run makemigrations
	env := map[string]string{
		"MODELS_DIR": modelsDir,
	}
	args := []string{
		"makemigrations",
		"init",
		"--auto",
		"--models", modelsDir,
		"--path", filepath.Join(workdir, "migrations"),
	}
	stdout, stderr, err := testhelpers.RunCLI(ctx, workdir, env, args, 10*time.Second)

	t.Logf("stdout: %s", stdout)
	if err != nil {
		t.Logf("stderr: %s", stderr)
	}

	// Check if migration files were created
	migrationsDir := filepath.Join(workdir, "migrations")
	_, err = os.Stat(migrationsDir)
	assert.True(t, !os.IsNotExist(err) || err == nil, "migrations directory should exist or be creatable")
}

// TestCLIApplyMigration tests the apply migration CLI command
func TestCLIApplyMigration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" && os.Getenv("RUN_POSTGRES_TESTS") == "" {
		t.Skip("Postgres not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Start Postgres container
	opts := testhelpers.DefaultPostgresOpts()
	db, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer db.Close()

	workdir, cleanupWd := testhelpers.TempWorkdir(t, "forge-e2e-apply-")
	defer cleanupWd()

	// Create migrations directory with sample migration
	migrationsDir := filepath.Join(workdir, "migrations")
	os.MkdirAll(migrationsDir, 0755)

	migrationUp := `
	CREATE TABLE users (
		id BIGSERIAL PRIMARY KEY,
		username VARCHAR(150) NOT NULL UNIQUE,
		email VARCHAR(254) NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT NOW()
	);
	`
	testhelpers.WriteFileString(t, filepath.Join(migrationsDir, "0001_create_users.up.sql"), migrationUp)

	migrationDown := `DROP TABLE users;`
	testhelpers.WriteFileString(t, filepath.Join(migrationsDir, "0001_create_users.down.sql"), migrationDown)

	// Run apply
	env := map[string]string{
		"DATABASE_URL": dsn,
	}
	stdout, stderr, err := testhelpers.RunCLI(ctx, workdir, env, []string{"migrate", "up"}, 15*time.Second)

	t.Logf("stdout: %s", stdout)
	if err != nil {
		t.Logf("stderr: %s", stderr)
	}

	// Verify table was created
	testhelpers.AssertTableExists(ctx, t, db, "postgres", "users")
}

// TestCLIStatus tests the migration status command
func TestCLIStatus(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" && os.Getenv("RUN_POSTGRES_TESTS") == "" {
		t.Skip("Postgres not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.DefaultPostgresOpts()
	_, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()

	workdir, cleanupWd := testhelpers.TempWorkdir(t, "forge-e2e-status-")
	defer cleanupWd()

	env := map[string]string{
		"DATABASE_URL": dsn,
	}
	stdout, _, err := testhelpers.RunCLI(ctx, workdir, env, []string{"migrate", "status"}, 10*time.Second)

	// Status command should produce output
	assert.NotEmpty(t, stdout, "status output should not be empty")
}
