package orm

import (
	"context"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComparisonExpression_DateParts_Postgres(t *testing.T) {
	field := NewField[time.Time]("created_at", "")

	tests := []struct {
		name        string
		op          Operator
		val         interface{}
		expectedSQL string
	}{
		{
			name:        "OpYear",
			op:          OpYear,
			val:         2024,
			expectedSQL: `EXTRACT(YEAR FROM "created_at") = $1`,
		},
		{
			name:        "OpMonth",
			op:          OpMonth,
			val:         5,
			expectedSQL: `EXTRACT(MONTH FROM "created_at") = $1`,
		},
		{
			name:        "OpDay",
			op:          OpDay,
			val:         15,
			expectedSQL: `EXTRACT(DAY FROM "created_at") = $1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := ComparisonExpression[time.Time]{
				Field: field,
				Op:    tt.op,
				Value: tt.val,
			}
			builder := NewSQLBuilder()
			sql, args, err := expr.ToSQL(builder)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedSQL, sql)
			assert.Equal(t, []interface{}{tt.val}, args)
		})
	}

	t.Run("with table prefix", func(t *testing.T) {
		tableField := NewField[time.Time]("created_at", "users")
		expr := ComparisonExpression[time.Time]{
			Field: tableField,
			Op:    OpYear,
			Value: 2024,
		}
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, `EXTRACT(YEAR FROM "users"."created_at") = $1`, sql)
		assert.Equal(t, []interface{}{2024}, args)
	})
}

func TestComparisonExpression_DateParts_SQLite(t *testing.T) {
	field := NewField[time.Time]("created_at", "")
	sqliteDialect := dialect.NewSQLiteDialect()

	tests := []struct {
		name        string
		op          Operator
		val         interface{}
		expectedSQL string
	}{
		{
			name:        "OpYear",
			op:          OpYear,
			val:         2024,
			expectedSQL: `CAST(strftime('%Y', "created_at") AS INTEGER) = ?`,
		},
		{
			name:        "OpMonth",
			op:          OpMonth,
			val:         5,
			expectedSQL: `CAST(strftime('%m', "created_at") AS INTEGER) = ?`,
		},
		{
			name:        "OpDay",
			op:          OpDay,
			val:         15,
			expectedSQL: `CAST(strftime('%d', "created_at") AS INTEGER) = ?`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := ComparisonExpression[time.Time]{
				Field: field,
				Op:    tt.op,
				Value: tt.val,
			}
			builder := NewSQLBuilderWithDialect(sqliteDialect)
			sql, args, err := expr.ToSQL(builder)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedSQL, sql)
			assert.Equal(t, []interface{}{tt.val}, args)
		})
	}

	t.Run("with table prefix", func(t *testing.T) {
		tableField := NewField[time.Time]("created_at", "users")
		expr := ComparisonExpression[time.Time]{
			Field: tableField,
			Op:    OpYear,
			Value: 2024,
		}
		builder := NewSQLBuilderWithDialect(sqliteDialect)
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, `CAST(strftime('%Y', "users"."created_at") AS INTEGER) = ?`, sql)
		assert.Equal(t, []interface{}{2024}, args)
	})
}

type customDialect struct {
	dialect.Dialect
	name string
}

func (c *customDialect) Name() string {
	return c.name
}

func TestSQLBuilder_isSQLite(t *testing.T) {
	// Nil builder or nil dialect
	var nilBuilder *SQLBuilder
	assert.False(t, nilBuilder.isSQLite())

	builderDefault := NewSQLBuilder()
	assert.False(t, builderDefault.isSQLite())

	// Postgres dialect
	builderPG := NewSQLBuilderWithDialect(dialect.NewPostgreSQLDialect())
	assert.False(t, builderPG.isSQLite())

	// SQLite dialect
	builderSQLite := NewSQLBuilderWithDialect(dialect.NewSQLiteDialect())
	assert.True(t, builderSQLite.isSQLite())

	// Case-insensitivity and sqlite3 alias
	builderSQLite3 := NewSQLBuilderWithDialect(&customDialect{name: "sqlite3"})
	assert.True(t, builderSQLite3.isSQLite())

	builderUpperSQLite := NewSQLBuilderWithDialect(&customDialect{name: "SQLITE"})
	assert.True(t, builderUpperSQLite.isSQLite())

	builderUpperSQLite3 := NewSQLBuilderWithDialect(&customDialect{name: "SQLite3"})
	assert.True(t, builderUpperSQLite3.isSQLite())

	// Other dialect
	builderMySQL := NewSQLBuilderWithDialect(&customDialect{name: "mysql"})
	assert.False(t, builderMySQL.isSQLite())
}

func TestBaseQuerySet_newSQLBuilder(t *testing.T) {
	// When db is nil, newSQLBuilder should return default builder (not SQLite)
	qs, err := NewQuerySet[User]("users")
	require.NoError(t, err)
	baseQS := qs.(*BaseQuerySet[User])

	builder := baseQS.newSQLBuilder()
	require.NotNil(t, builder)
	assert.False(t, builder.isSQLite())

	// When db has PostgreSQL dialect
	pgDB := &db.DB{Driver: "postgres"}
	pgDB.SetDialect(dialect.NewPostgreSQLDialect())
	qsPG := baseQS.SetDB(pgDB).(*BaseQuerySet[User])
	builderPG := qsPG.newSQLBuilder()
	require.NotNil(t, builderPG)
	assert.False(t, builderPG.isSQLite())

	// When db has SQLite dialect
	sqliteDB := &db.DB{Driver: "sqlite"}
	sqliteDB.SetDialect(dialect.NewSQLiteDialect())
	qsSQLite := baseQS.SetDB(sqliteDB).(*BaseQuerySet[User])
	builderSQLite := qsSQLite.newSQLBuilder()
	require.NotNil(t, builderSQLite)
	assert.True(t, builderSQLite.isSQLite())
}

func TestQuerySet_DateParts_SQL(t *testing.T) {
	// PostgreSQL dialect (default when db is nil)
	qs, err := NewQuerySet[User]("users")
	require.NoError(t, err)
	qs = qs.Filter(Where("created_at", OpYear, 2024))
	sql, args, err := qs.(*BaseQuerySet[User]).buildSQL()
	require.NoError(t, err)
	assert.Contains(t, sql, `EXTRACT(YEAR FROM "created_at") = $1`)
	assert.Equal(t, []interface{}{2024}, args)

	// SQLite dialect
	sqliteDB := &db.DB{Driver: "sqlite"}
	sqliteDB.SetDialect(dialect.NewSQLiteDialect())
	qs2, err := NewQuerySet[User]("users")
	require.NoError(t, err)
	qs2 = qs2.SetDB(sqliteDB).Filter(Where("created_at", OpYear, 2024))
	sql2, args2, err := qs2.(*BaseQuerySet[User]).buildSQL()
	require.NoError(t, err)
	assert.Contains(t, sql2, `CAST(strftime('%Y', "created_at") AS INTEGER) = ?`)
	assert.Equal(t, []interface{}{2024}, args2)
}

func TestDateParts_SQLiteIntegration(t *testing.T) {
	database, err := db.NewDBWithDriver("sqlite3", ":memory:", db.WithMaxOpenConns(1))
	require.NoError(t, err)
	defer database.Close()

	_, err = database.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			created_at TIMESTAMP
		);
	`)
	require.NoError(t, err)

	_, err = GetModelSchema[User]()
	require.NoError(t, err)

	t2023 := time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC)
	t2024May := time.Date(2024, 5, 20, 12, 0, 0, 0, time.UTC)
	t2024Nov := time.Date(2024, 11, 5, 14, 0, 0, 0, time.UTC)

	_, err = database.Exec(`INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)`, "Alice", "alice@example.com", t2023)
	require.NoError(t, err)
	_, err = database.Exec(`INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)`, "Bob", "bob@example.com", t2024May)
	require.NoError(t, err)
	_, err = database.Exec(`INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)`, "Charlie", "charlie@example.com", t2024Nov)
	require.NoError(t, err)

	ctx := context.Background()
	qs, err := NewQuerySet[User]("users")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	// Filter by year 2024
	users2024, err := qs.Filter(Where("created_at", OpYear, 2024)).All(ctx)
	require.NoError(t, err)
	assert.Len(t, users2024, 2)

	// Filter by month 5
	usersMay, err := qs.Filter(Where("created_at", OpMonth, 5)).All(ctx)
	require.NoError(t, err)
	require.Len(t, usersMay, 1)
	assert.Equal(t, "Bob", usersMay[0].Name)

	// Filter by day 15
	usersDay15, err := qs.Filter(Where("created_at", OpDay, 15)).All(ctx)
	require.NoError(t, err)
	require.Len(t, usersDay15, 1)
	assert.Equal(t, "Alice", usersDay15[0].Name)
}
