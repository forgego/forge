package orm

import (
	"database/sql"
	"fmt"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/dialect"
	"github.com/forgego/forge/schema"
)

// ModelWithID is a type constraint for models that have an ID field
// This allows type-safe operations without reflection
type ModelWithID interface {
	GetID() int64
	SetID(int64)
}

// SchemaModel is a type constraint for models that implement schema.Schema
type SchemaModel interface {
	schema.Schema
}

// GetSQLDB extracts *sql.DB from a database connection
// Uses type assertion instead of reflection
func GetSQLDB(conn interface{}) (*sql.DB, error) {
	switch v := conn.(type) {
	case *db.DB:
		return v.DB, nil
	case *sql.DB:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported database connection type: %T", conn)
	}
}

// GetDialect extracts the SQL dialect from a database connection.
// Returns the dialect for generating database-agnostic SQL queries.
func GetDialect(conn interface{}) (dialect.Dialect, error) {
	switch v := conn.(type) {
	case *db.DB:
		return v.Dialect(), nil
	default:
		return nil, fmt.Errorf("unsupported database connection type for dialect: %T", conn)
	}
}
