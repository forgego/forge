---
sidebar_position: 5
---

# Migrations Guide

forge includes a built-in migration system based on golang-migrate that automatically manages your database schema changes.

## Overview

Migrations are version-controlled database schema changes. They allow you to:

- Track schema changes over time
- Apply changes to different environments consistently
- Roll back changes if needed
- Collaborate with team members on schema changes

## Creating Migrations

### Automatic Migration Generation

Generate migrations from your model definitions:

```bash
forge makemigrations <name> --auto
```

This will:
1. Scan your models directory
2. Compare current models with database state
3. **Automatically detect dependencies** from foreign keys and table references
4. Generate migration files for any changes
5. Save migrations to `migrations/` directory

### Verbose Mode

Use `--verbose` to see detailed information about parse errors and warnings:

```bash
forge makemigrations <name> --auto --verbose
```

This helps debug issues with migration parsing and state loading.

### Migration Files

Migrations are stored in `migrations/` directory:

```
migrations/
├── 000001_initial.up.sql
├── 000001_initial.down.sql
├── 000002_add_user_table.up.sql
├── 000002_add_user_table.down.sql
└── ...
```

Each migration has:
- **Up migration** (`.up.sql`) - Applies the change
- **Down migration** (`.down.sql`) - Reverts the change

### Dependency Auto-Detection

The migration system automatically detects dependencies between migrations:

```sql
-- Auto-detected dependencies:
-- DEPENDS: 000001

CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    author_id BIGINT NOT NULL,
    FOREIGN KEY (author_id) REFERENCES users(id)
);
```

Dependencies are automatically detected from:
- Foreign key relationships in model definitions
- Table references in SQL (REFERENCES, ALTER TABLE, etc.)

The system validates that all dependencies exist before generating migrations.

## Applying Migrations

### Apply All Pending Migrations

```bash
forge migrate
```

This applies all migrations that haven't been run yet.

### Apply Specific Migration

```bash
forge migrate up 2
```

Applies migrations up to version 2.

### Rollback

```bash
forge migrate down 1
```

Rolls back the last migration.

### Check Migration Status

```bash
forge migrate status
```

Shows which migrations have been applied.

### Show Migration Plan

Preview what migration would be generated:

```bash
forge migrate show
```

Shows the migration plan that would be generated from current models. Use `--sql` to see the full SQL, or `--verbose` to see parse warnings.

### Lint Migrations

Check migration files for common issues:

```bash
forge migrate lint
```

Use `--verbose` to also see parse errors from the state loader:

```bash
forge migrate lint --verbose
```

### Fake Migrations

Mark migrations as applied without running them. Useful when:
- Migrating an existing database
- Marking initial migrations that already exist in the database
- Marking all pending migrations as applied

#### Fake Initial Migrations

If you have an existing database with tables, mark the initial migrations that created those tables:

```bash
forge migrate fake --fake-initial
```

This will:
1. Query the database for existing tables
2. Find migrations that create those tables
3. Mark them as applied without running them

#### Mark All Pending as Applied

Mark all pending migrations as applied:

```bash
forge migrate fake
```

This marks all migrations with versions greater than the current version as applied. A confirmation prompt appears if more than 5 migrations would be faked.

#### Fake Specific Migration

Mark a specific migration version as applied:

```bash
forge migrate fake <version>
```

**Warning:** Use fake commands with caution. Only use when you're certain the migration has already been applied or when you want to skip it.

## Migration Examples

### Creating a Table

**Up Migration:**
```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(150) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(128) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    date_joined TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

**Down Migration:**
```sql
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

### Adding a Column

**Up Migration:**
```sql
ALTER TABLE users ADD COLUMN last_login TIMESTAMP;
```

**Down Migration:**
```sql
ALTER TABLE users DROP COLUMN last_login;
```

### Adding a Foreign Key

**Up Migration:**
```sql
ALTER TABLE posts ADD COLUMN author_id BIGINT;
ALTER TABLE posts ADD CONSTRAINT fk_posts_author 
    FOREIGN KEY (author_id) REFERENCES users(id) 
    ON DELETE CASCADE;
```

**Down Migration:**
```sql
ALTER TABLE posts DROP CONSTRAINT fk_posts_author;
ALTER TABLE posts DROP COLUMN author_id;
```

### Creating an Index

**Up Migration:**
```sql
CREATE INDEX idx_posts_created_at ON posts(created_at);
```

**Down Migration:**
```sql
DROP INDEX IF EXISTS idx_posts_created_at;
```

## Data Migrations

You can also include data migrations in your SQL files:

```sql
-- Update existing data
UPDATE users SET is_active = true WHERE is_active IS NULL;

-- Insert default data
INSERT INTO categories (name, slug) VALUES
    ('Technology', 'technology'),
    ('Science', 'science'),
    ('Arts', 'arts');
```

## Best Practices

### 1. Keep Migrations Small

Break large changes into multiple migrations:

```sql
-- Good: Separate migrations
-- 000001_add_users_table.up.sql
-- 000002_add_posts_table.up.sql
-- 000003_add_comments_table.up.sql

-- Bad: One large migration
-- 000001_add_all_tables.up.sql
```

### 2. Always Write Down Migrations

Every up migration should have a corresponding down migration:

```sql
-- Up
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- Down
ALTER TABLE users DROP COLUMN phone;
```

### 3. Test Migrations

Test both up and down migrations:

```bash
# Apply
forge migrate

# Rollback
forge migrate down 1

# Re-apply
forge migrate up 1
```

### 4. Don't Modify Existing Migrations

Once a migration is applied to production, don't modify it. Create a new migration instead.

### 5. Use Transactions

Wrap migrations in transactions when possible:

```sql
BEGIN;

ALTER TABLE users ADD COLUMN phone VARCHAR(20);
CREATE INDEX idx_users_phone ON users(phone);

COMMIT;
```

## Migration Workflow

### Development

1. Modify your models
2. Generate migrations: `forge makemigrations`
3. Review generated SQL
4. Apply migrations: `forge migrate`
5. Test your application

### Production

1. Review migrations before deploying
2. Backup database
3. Apply migrations: `forge migrate`
4. Verify application works
5. Keep backup until confident

## Troubleshooting

### Migration Conflicts

If you have conflicting migrations:

```bash
# Check status
forge migrate status

# Manually resolve conflicts
# Edit migration files as needed
```

### Failed Migrations

If a migration fails:

1. Check the error message
2. Fix the SQL in the migration file
3. Rollback if needed: `forge migrate down 1`
4. Fix and re-apply: `forge migrate up 1`

### Database State Mismatch

If your database state doesn't match migrations:

```bash
# Check current state
forge migrate status

# Force to specific version (use with caution)
forge migrate force 5
```

### Parse Errors

If you see parse errors or warnings:

```bash
# Use verbose mode to see detailed parse errors
forge migrate show --verbose
forge migrate lint --verbose
```

The migration system uses a robust three-pass retry mechanism to handle:
- Out-of-order CREATE TABLE statements
- Foreign key constraints referencing tables not yet created
- Indexes created before their tables

Parse errors are collected and can be viewed with verbose mode.

## Advanced Topics

### Custom Migration SQL

You can write custom SQL migrations:

```sql
-- Custom migration logic
DO $$
BEGIN
    -- Complex migration logic
    IF EXISTS (SELECT 1 FROM information_schema.tables 
               WHERE table_name = 'old_table') THEN
        -- Migration code
    END IF;
END $$;
```

### Migration Dependencies

Dependencies are automatically detected, but you can also specify them manually:

```sql
-- DEPENDS: 000001
-- DEPENDS: 000002

CREATE TABLE posts (
    -- ...
);
```

For cross-application dependencies:

```sql
-- DEPENDS: app_name:000001

CREATE TABLE shared_table (
    -- ...
);
```

### State Loading Improvements

The migration system includes improved state loading:

- **Three-pass retry mechanism**: Handles out-of-order migrations gracefully
- **Table context tracking**: Automatically infers table names for indexes
- **Parse error collection**: Collects and reports parse errors without failing
- **Verbose logging**: Optional detailed logging for debugging

### Parser Improvements

The SQL parser includes:

- **Two-pass parsing**: Processes CREATE TABLE statements first, then constraints
- **Better error handling**: Fails softly with UnknownChange for unparseable statements
- **Table context inference**: Tracks table context for DROP INDEX and CREATE INDEX
- **Verbose mode**: Logs UnknownChange statements and parse errors

### Migration Hooks

Add hooks to run code during migrations:

```go
// In your migration file or code
func RunMigration(ctx context.Context, db *sql.DB) error {
    // Custom migration logic
    return nil
}
```

## Next Steps

- [Schema Reference](/docs/api-reference/schema) - Learn about model definitions
- [Development Guide](/docs/contributing/development) - Contributing to forge
- [Deployment Guide](/docs/guides/deployment) - Deploying your application
