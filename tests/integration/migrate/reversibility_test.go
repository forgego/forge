package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/tests/helpers"
	"github.com/forgego/forge/tests/testhelpers"
)

// TestMigrationUpDown tests that migrations can be applied and rolled back
func TestMigrationUpDown(t *testing.T) {
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

	// Create a temporary directory for migrations under tests/tmp (returns relative path)
	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "reversibility_")
	defer cleanupTemp()
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create database connection
	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	// Create migration files manually
	createMigration1Up := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(150) NOT NULL,
			email VARCHAR(254) NOT NULL UNIQUE
		);
	`
	createMigration1Down := `DROP TABLE IF EXISTS users;`

	migration1UpPath := filepath.Join(migrationsDir, "000001_create_users.up.sql")
	migration1DownPath := filepath.Join(migrationsDir, "000001_create_users.down.sql")

	require.NoError(t, os.WriteFile(migration1UpPath, []byte(createMigration1Up), 0644))
	require.NoError(t, os.WriteFile(migration1DownPath, []byte(createMigration1Down), 0644))

	// Create second migration
	createMigration2Up := `
		CREATE TABLE posts (
			id BIGSERIAL PRIMARY KEY,
			title VARCHAR(200) NOT NULL,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE
		);
	`
	createMigration2Down := `DROP TABLE IF EXISTS posts;`

	migration2UpPath := filepath.Join(migrationsDir, "000002_create_posts.up.sql")
	migration2DownPath := filepath.Join(migrationsDir, "000002_create_posts.down.sql")

	require.NoError(t, os.WriteFile(migration2UpPath, []byte(createMigration2Up), 0644))
	require.NoError(t, os.WriteFile(migration2DownPath, []byte(createMigration2Down), 0644))

	// Apply first migration
	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Verify tables exist
	helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "users")
	helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "posts")

	// Verify migration state using the existing runner
	helpers.AssertMigrationStateWithRunner(ctx, t, database, migrationsDir, 2, false, runner)

	// Rollback one migration
	err = runner.Rollback(ctx)
	require.NoError(t, err)

	// Verify posts table is gone
	var exists int
	err = postgresDB.QueryRowContext(ctx, `
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'posts' AND table_schema = 'public'
	`).Scan(&exists)
	require.Error(t, err, "posts table should not exist after rollback")

	// Verify users table still exists
	helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "users")

	// Verify migration state using the existing runner
	helpers.AssertMigrationStateWithRunner(ctx, t, database, migrationsDir, 1, false, runner)

	// Rollback again
	err = runner.Rollback(ctx)
	require.NoError(t, err)

	// Verify users table is gone
	err = postgresDB.QueryRowContext(ctx, `
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'users' AND table_schema = 'public'
	`).Scan(&exists)
	require.Error(t, err, "users table should not exist after rollback")

	// Verify migration state using the existing runner
	helpers.AssertMigrationStateWithRunner(ctx, t, database, migrationsDir, 0, false, runner)
}

// TestMigrationRollbackSequence tests rolling back multiple migrations in sequence
func TestMigrationRollbackSequence(t *testing.T) {
	// Using localhost PostgreSQL
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

	// Create a temporary directory for migrations under tests/tmp (returns relative path)
	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "reversibility_")
	defer cleanupTemp()
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create database connection
	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	// Create multiple migrations
	migrations := []struct {
		version string
		up      string
		down    string
		table   string
	}{
		{
			version: "000001",
			up:      `CREATE TABLE table1 (id BIGSERIAL PRIMARY KEY, name VARCHAR(200));`,
			down:    `DROP TABLE IF EXISTS table1;`,
			table:   "table1",
		},
		{
			version: "000002",
			up:      `CREATE TABLE table2 (id BIGSERIAL PRIMARY KEY, name VARCHAR(200));`,
			down:    `DROP TABLE IF EXISTS table2;`,
			table:   "table2",
		},
		{
			version: "000003",
			up:      `CREATE TABLE table3 (id BIGSERIAL PRIMARY KEY, name VARCHAR(200));`,
			down:    `DROP TABLE IF EXISTS table3;`,
			table:   "table3",
		},
	}

	for _, mig := range migrations {
		upPath := filepath.Join(migrationsDir, mig.version+"_create_"+mig.table+".up.sql")
		downPath := filepath.Join(migrationsDir, mig.version+"_create_"+mig.table+".down.sql")
		require.NoError(t, os.WriteFile(upPath, []byte(mig.up), 0644))
		require.NoError(t, os.WriteFile(downPath, []byte(mig.down), 0644))
	}

	// Apply all migrations
	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Verify all tables exist
	for _, mig := range migrations {
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", mig.table)
	}

	// Verify migration state using the existing runner
	helpers.AssertMigrationStateWithRunner(ctx, t, database, migrationsDir, 3, false, runner)

	// Rollback all migrations using the existing runner
	err = helpers.RollbackMigrationSequence(ctx, t, database, migrationsDir, 3, runner)
	require.NoError(t, err)

	// Verify all tables are gone
	for _, mig := range migrations {
		var exists int
		err = postgresDB.QueryRowContext(ctx, `
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = $1 AND table_schema = 'public'
		`, mig.table).Scan(&exists)
		require.Error(t, err, "table %s should not exist after rollback", mig.table)
	}

	// Verify migration state
	helpers.AssertMigrationState(ctx, t, database, migrationsDir, 0, false)
}

// TestMigrationReapplyAfterRollback tests reapplying migrations after rollback
func TestMigrationReapplyAfterRollback(t *testing.T) {
	// Using localhost PostgreSQL
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

	// Create a temporary directory for migrations under tests/tmp (returns relative path)
	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "reversibility_")
	defer cleanupTemp()
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create database connection
	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	// Create migration
	createMigrationUp := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(150) NOT NULL,
			email VARCHAR(254) NOT NULL UNIQUE
		);
	`
	createMigrationDown := `DROP TABLE IF EXISTS users;`

	migrationUpPath := filepath.Join(migrationsDir, "000001_create_users.up.sql")
	migrationDownPath := filepath.Join(migrationsDir, "000001_create_users.down.sql")

	require.NoError(t, os.WriteFile(migrationUpPath, []byte(createMigrationUp), 0644))
	require.NoError(t, os.WriteFile(migrationDownPath, []byte(createMigrationDown), 0644))

	// Apply migration
	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Verify table exists
	helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "users")

	// Rollback
	err = runner.Rollback(ctx)
	require.NoError(t, err)

	// Verify table is gone
	var exists int
	err = postgresDB.QueryRowContext(ctx, `
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'users' AND table_schema = 'public'
	`).Scan(&exists)
	require.Error(t, err, "users table should not exist after rollback")

	// Reapply migration
	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Verify table exists again
	helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "users")

	// Verify migration state
	helpers.AssertMigrationState(ctx, t, database, migrationsDir, 1, false)
}

// TestMigrationDownWithData tests rolling back migrations that have data
func TestMigrationDownWithData(t *testing.T) {
	// Using localhost PostgreSQL
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

	// Create a temporary directory for migrations under tests/tmp (returns relative path)
	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "reversibility_")
	defer cleanupTemp()
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create database connection
	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	// Create migration
	createMigrationUp := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(150) NOT NULL,
			email VARCHAR(254) NOT NULL UNIQUE
		);
	`
	createMigrationDown := `DROP TABLE IF EXISTS users;`

	migrationUpPath := filepath.Join(migrationsDir, "000001_create_users.up.sql")
	migrationDownPath := filepath.Join(migrationsDir, "000001_create_users.down.sql")

	require.NoError(t, os.WriteFile(migrationUpPath, []byte(createMigrationUp), 0644))
	require.NoError(t, os.WriteFile(migrationDownPath, []byte(createMigrationDown), 0644))

	// Apply migration
	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Insert data
	insertSQL := `INSERT INTO users (username, email) VALUES ('testuser', 'test@example.com')`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, insertSQL)

	// Verify data exists
	helpers.AssertRowCount(ctx, t, postgresDB, "users", 1)

	// Rollback (should succeed even with data if using DROP TABLE IF EXISTS)
	err = runner.Rollback(ctx)
	require.NoError(t, err)

	// Verify table is gone
	var exists int
	err = postgresDB.QueryRowContext(ctx, `
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'users' AND table_schema = 'public'
	`).Scan(&exists)
	require.Error(t, err, "users table should not exist after rollback")
}

// TestMigrationPartialRollback tests rolling back to a specific version
func TestMigrationPartialRollback(t *testing.T) {
	// Using localhost PostgreSQL
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

	// Create a temporary directory for migrations under tests/tmp (returns relative path)
	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "reversibility_")
	defer cleanupTemp()
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create database connection
	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	// Create multiple migrations
	migrations := []struct {
		version uint
		up      string
		down    string
		table   string
	}{
		{
			version: 1,
			up:      `CREATE TABLE table1 (id BIGSERIAL PRIMARY KEY);`,
			down:    `DROP TABLE IF EXISTS table1;`,
			table:   "table1",
		},
		{
			version: 2,
			up:      `CREATE TABLE table2 (id BIGSERIAL PRIMARY KEY);`,
			down:    `DROP TABLE IF EXISTS table2;`,
			table:   "table2",
		},
		{
			version: 3,
			up:      `CREATE TABLE table3 (id BIGSERIAL PRIMARY KEY);`,
			down:    `DROP TABLE IF EXISTS table3;`,
			table:   "table3",
		},
	}

	for _, mig := range migrations {
		versionStr := fmt.Sprintf("%06d", mig.version)
		upPath := filepath.Join(migrationsDir, versionStr+"_create_"+mig.table+".up.sql")
		downPath := filepath.Join(migrationsDir, versionStr+"_create_"+mig.table+".down.sql")
		require.NoError(t, os.WriteFile(upPath, []byte(mig.up), 0644))
		require.NoError(t, os.WriteFile(downPath, []byte(mig.down), 0644))
	}

	// Apply all migrations
	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Verify all tables exist
	for _, mig := range migrations {
		helpers.AssertTableExists(ctx, t, postgresDB, "postgres", mig.table)
	}

	// Rollback to version 1
	err = runner.RollbackTo(ctx, 1)
	require.NoError(t, err)

	// Verify only table1 exists
	helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "table1")

	// Verify table2 and table3 are gone
	var exists int
	err = postgresDB.QueryRowContext(ctx, `
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'table2' AND table_schema = 'public'
	`).Scan(&exists)
	require.Error(t, err, "table2 should not exist")

	err = postgresDB.QueryRowContext(ctx, `
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'table3' AND table_schema = 'public'
	`).Scan(&exists)
	require.Error(t, err, "table3 should not exist")

	// Verify migration state
	helpers.AssertMigrationState(ctx, t, database, migrationsDir, 1, false)
}
