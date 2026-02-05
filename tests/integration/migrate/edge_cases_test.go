package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/tests/helpers"
	"github.com/forgego/forge/tests/testhelpers"
)

// TestColumnRename tests renaming a column
func TestColumnRename(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_column_rename_%d", time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"users"})

	// Create initial table
	createTableSQL := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(150) NOT NULL,
			email VARCHAR(254) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Insert test data
	insertSQL := `INSERT INTO users (username, email) VALUES ('testuser', 'test@example.com')`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, insertSQL)

	// Rename column
	renameSQL := `ALTER TABLE users RENAME COLUMN username TO user_name`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, renameSQL)

	// Verify old column doesn't exist
	var exists int
	err = postgresDB.QueryRowContext(ctx, `
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = 'users' AND column_name = 'username' AND table_schema = 'public'
	`).Scan(&exists)
	require.Error(t, err, "old column should not exist")
	require.Equal(t, sql.ErrNoRows, err)

	// Verify new column exists
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "users", "user_name")

	// Verify data is preserved
	var username string
	err = postgresDB.QueryRowContext(ctx, `SELECT user_name FROM users WHERE id = 1`).Scan(&username)
	require.NoError(t, err)
	require.Equal(t, "testuser", username)
}

// TestColumnTypeChange tests changing a column type
func TestColumnTypeChange(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_column_type_change_%d", time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"products"})

	// Create initial table
	createTableSQL := `
		CREATE TABLE products (
			id BIGSERIAL PRIMARY KEY,
			price INTEGER NOT NULL,
			name VARCHAR(200) NOT NULL
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Insert test data
	insertSQL := `INSERT INTO products (name, price) VALUES ('Test Product', 100)`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, insertSQL)

	// Change column type from INTEGER to NUMERIC
	alterSQL := `ALTER TABLE products ALTER COLUMN price TYPE NUMERIC(12, 2) USING price::NUMERIC(12, 2)`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, alterSQL)

	// Verify column type changed
	var dataType string
	err = postgresDB.QueryRowContext(ctx, `
		SELECT data_type FROM information_schema.columns 
		WHERE table_name = 'products' AND column_name = 'price' AND table_schema = 'public'
	`).Scan(&dataType)
	require.NoError(t, err)
	require.Equal(t, "numeric", dataType)

	// Verify data is preserved
	var price float64
	err = postgresDB.QueryRowContext(ctx, `SELECT price FROM products WHERE id = 1`).Scan(&price)
	require.NoError(t, err)
	require.Equal(t, 100.0, price)
}

// TestTableRename tests renaming a table
func TestTableRename(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_table_rename_%d", time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"old_table", "new_table"})

	// Create initial table
	createTableSQL := `
		CREATE TABLE old_table (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Insert test data
	insertSQL := `INSERT INTO old_table (name) VALUES ('Test Data')`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, insertSQL)

	// Rename table
	renameSQL := `ALTER TABLE old_table RENAME TO new_table`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, renameSQL)

	// Verify old table doesn't exist
	var exists int
	err = postgresDB.QueryRowContext(ctx, `
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'old_table' AND table_schema = 'public'
	`).Scan(&exists)
	require.Error(t, err, "old table should not exist")
	require.Equal(t, sql.ErrNoRows, err)

	// Verify new table exists
	helpers.AssertTableExists(ctx, t, postgresDB, "postgres", "new_table")

	// Verify data is preserved
	var name string
	err = postgresDB.QueryRowContext(ctx, `SELECT name FROM new_table WHERE id = 1`).Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "Test Data", name)
}

// TestAddColumnWithDefault tests adding a column with a default value
func TestAddColumnWithDefault(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_add_column_default_%d", time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"users"})

	// Create initial table
	createTableSQL := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(150) NOT NULL
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Insert test data
	insertSQL := `INSERT INTO users (username) VALUES ('testuser')`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, insertSQL)

	// Add column with default
	addColumnSQL := `ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT true NOT NULL`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, addColumnSQL)

	// Verify column exists
	helpers.AssertColumnExists(ctx, t, postgresDB, "postgres", "users", "is_active")

	// Verify existing row has default value
	var isActive bool
	err = postgresDB.QueryRowContext(ctx, `SELECT is_active FROM users WHERE id = 1`).Scan(&isActive)
	require.NoError(t, err)
	require.True(t, isActive, "existing row should have default value")
}

// TestAddNotNullColumnWithoutDefault tests adding a NOT NULL column without default (should fail)
func TestAddNotNullColumnWithoutDefault(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_add_notnull_no_default_%d", time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"users"})

	// Create initial table with data
	createTableSQL := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(150) NOT NULL
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	insertSQL := `INSERT INTO users (username) VALUES ('testuser')`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, insertSQL)

	// Try to add NOT NULL column without default - should fail
	addColumnSQL := `ALTER TABLE users ADD COLUMN email VARCHAR(254) NOT NULL`
	err = helpers.RunSQLExpectError(ctx, postgresDB, addColumnSQL)
	require.Error(t, err, "adding NOT NULL column without default should fail when table has data")
}

// TestDropColumn tests dropping a column
func TestDropColumn(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_drop_column_%d", time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Create initial table
	createTableSQL := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(150) NOT NULL,
			old_field VARCHAR(100)
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Drop column
	dropColumnSQL := `ALTER TABLE users DROP COLUMN old_field`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, dropColumnSQL)

	// Verify column doesn't exist
	var exists int
	err = postgresDB.QueryRowContext(ctx, `
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = 'users' AND column_name = 'old_field' AND table_schema = 'public'
	`).Scan(&exists)
	require.Error(t, err, "dropped column should not exist")
	require.Equal(t, sql.ErrNoRows, err)
}

// TestAddForeignKey tests adding a foreign key constraint
func TestAddForeignKey(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_add_fk_%d", time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"users", "posts"})

	// Create parent table
	createUsersSQL := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(150) NOT NULL
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createUsersSQL)

	// Create child table without FK
	createPostsSQL := `
		CREATE TABLE posts (
			id BIGSERIAL PRIMARY KEY,
			title VARCHAR(200) NOT NULL,
			user_id BIGINT
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createPostsSQL)

	// Add foreign key
	addFKSQL := `ALTER TABLE posts ADD CONSTRAINT fk_posts_user_id 
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, addFKSQL)

	// Verify FK exists
	helpers.AssertForeignKeyExists(ctx, t, postgresDB, "postgres", "posts", "user_id")
	helpers.AssertForeignKeyCascade(ctx, t, postgresDB, "posts", "user_id", helpers.CascadeCASCADE)
}

// TestDropForeignKey tests dropping a foreign key constraint
func TestDropForeignKey(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_drop_fk_%d", time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"users", "posts"})

	// Create tables with FK
	createUsersSQL := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(150) NOT NULL
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createUsersSQL)

	createPostsSQL := `
		CREATE TABLE posts (
			id BIGSERIAL PRIMARY KEY,
			title VARCHAR(200) NOT NULL,
			user_id BIGINT REFERENCES users(id) ON DELETE CASCADE
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createPostsSQL)

	// Get constraint name
	var constraintName string
	err = postgresDB.QueryRowContext(ctx, `
		SELECT constraint_name FROM information_schema.table_constraints
		WHERE table_name = 'posts' AND constraint_type = 'FOREIGN KEY' AND table_schema = 'public'
		LIMIT 1
	`).Scan(&constraintName)
	require.NoError(t, err)

	// Drop foreign key
	dropFKSQL := `ALTER TABLE posts DROP CONSTRAINT ` + constraintName
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, dropFKSQL)

	// Verify FK doesn't exist
	var exists int
	err = postgresDB.QueryRowContext(ctx, `
		SELECT 1 FROM information_schema.table_constraints
		WHERE table_name = 'posts' AND constraint_name = $1 AND table_schema = 'public'
	`, constraintName).Scan(&exists)
	require.Error(t, err, "dropped FK constraint should not exist")
	require.Equal(t, sql.ErrNoRows, err)
}

// TestAddUniqueConstraint tests adding a unique constraint
func TestAddUniqueConstraint(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_add_unique_%d", time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Create table
	createTableSQL := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			email VARCHAR(254) NOT NULL
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Add unique constraint
	addUniqueSQL := `ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE (email)`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, addUniqueSQL)

	// Verify constraint exists
	helpers.AssertConstraintExistsEnhanced(ctx, t, postgresDB, "users", "unique_email", "UNIQUE")

	// Verify uniqueness is enforced
	insert1 := `INSERT INTO users (email) VALUES ('test@example.com')`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, insert1)

	insert2 := `INSERT INTO users (email) VALUES ('test@example.com')`
	err = helpers.RunSQLExpectError(ctx, postgresDB, insert2)
	require.Error(t, err, "duplicate email should fail")
}

// TestCompositeUniqueConstraint tests composite unique constraints
func TestCompositeUniqueConstraint(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_composite_unique_%d", time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Create table
	createTableSQL := `
		CREATE TABLE user_roles (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			role_id BIGINT NOT NULL
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Add composite unique constraint
	addUniqueSQL := `ALTER TABLE user_roles ADD CONSTRAINT unique_user_role UNIQUE (user_id, role_id)`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, addUniqueSQL)

	// Verify constraint exists
	helpers.AssertConstraintExistsEnhanced(ctx, t, postgresDB, "user_roles", "unique_user_role", "UNIQUE")

	// Verify uniqueness is enforced
	insert1 := `INSERT INTO user_roles (user_id, role_id) VALUES (1, 1)`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, insert1)

	insert2 := `INSERT INTO user_roles (user_id, role_id) VALUES (1, 1)`
	err = helpers.RunSQLExpectError(ctx, postgresDB, insert2)
	require.Error(t, err, "duplicate user_id+role_id should fail")

	// But different combinations should work
	insert3 := `INSERT INTO user_roles (user_id, role_id) VALUES (1, 2)`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, insert3)

	insert4 := `INSERT INTO user_roles (user_id, role_id) VALUES (2, 1)`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, insert4)
}

// TestAddIndex tests adding an index
func TestAddIndex(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_add_index_%d", time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Cleanup tables before creating
	helpers.CleanupTables(ctx, t, postgresDB, "postgres", []string{"users"})

	// Create table
	createTableSQL := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			email VARCHAR(254) NOT NULL,
			username VARCHAR(150) NOT NULL
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	// Add index
	addIndexSQL := `CREATE INDEX idx_users_email ON users(email)`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, addIndexSQL)

	// Verify index exists
	helpers.AssertIndexExists(ctx, t, postgresDB, "postgres", "users", "idx_users_email")
}

// TestDropIndex tests dropping an index
func TestDropIndex(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_drop_index_%d", time.Now().UnixNano()),
	}
	postgresDB, _, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	// Create table with index
	createTableSQL := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			email VARCHAR(254) NOT NULL
		)
	`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, createTableSQL)

	addIndexSQL := `CREATE INDEX idx_users_email ON users(email)`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, addIndexSQL)

	// Drop index
	dropIndexSQL := `DROP INDEX idx_users_email`
	helpers.RunSQLExpectSuccess(ctx, t, postgresDB, dropIndexSQL)

	// Verify index doesn't exist
	var exists int
	err = postgresDB.QueryRowContext(ctx, `
		SELECT 1 FROM pg_indexes 
		WHERE tablename = 'users' AND indexname = 'idx_users_email'
	`).Scan(&exists)
	require.Error(t, err, "dropped index should not exist")
	require.Equal(t, sql.ErrNoRows, err)
}
