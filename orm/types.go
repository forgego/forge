package orm

import (
	"context"
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

// DBTX is satisfied by *sql.DB and *sql.Tx.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// GetDBTX extracts a DBTX from a database connection or transaction.
func GetDBTX(conn any) (DBTX, error) {
	switch v := conn.(type) {
	case *db.DB:
		if v == nil || v.DB == nil {
			return nil, fmt.Errorf("database connection is nil")
		}
		return v.DB, nil
	case *sql.DB:
		if v == nil {
			return nil, fmt.Errorf("database connection is nil")
		}
		return v, nil
	case *db.Tx:
		if v == nil || v.Tx == nil {
			return nil, fmt.Errorf("transaction is nil")
		}
		return v.Tx, nil
	case *sql.Tx:
		if v == nil {
			return nil, fmt.Errorf("transaction is nil")
		}
		return v, nil
	case interface{ DBTX() (DBTX, error) }:
		return v.DBTX()
	default:
		return nil, fmt.Errorf("unsupported database connection type: %T", conn)
	}
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
// A bare *sql.Tx exposes no driver metadata, so it uses PostgreSQL, matching
// the SQL builder's default for a raw *sql.DB without dialect information.
// Use *db.Tx to retain the parent database's dialect.
func GetDialect(conn interface{}) (dialect.Dialect, error) {
	switch v := conn.(type) {
	case *db.DB:
		if v == nil {
			return nil, fmt.Errorf("database connection is nil")
		}
		return v.Dialect(), nil
	case *db.Tx:
		if v == nil || v.DB() == nil {
			return nil, fmt.Errorf("transaction has no parent database")
		}
		return v.DB().Dialect(), nil
	case *sql.Tx:
		if v == nil {
			return nil, fmt.Errorf("transaction is nil")
		}
		return dialect.NewPostgreSQLDialect(), nil
	case interface {
		Dialect() (dialect.Dialect, error)
	}:
		return v.Dialect()
	default:
		return nil, fmt.Errorf("unsupported database connection type for dialect: %T", conn)
	}
}
