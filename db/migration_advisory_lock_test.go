package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type migrationLockSpy struct {
	mu         sync.Mutex
	statements []string
}
type migrationLockConnector struct{ spy *migrationLockSpy }
type migrationLockDriver struct{ spy *migrationLockSpy }
type migrationLockConn struct{ spy *migrationLockSpy }

func (c migrationLockConnector) Connect(context.Context) (driver.Conn, error) {
	return migrationLockConn{c.spy}, nil
}
func (c migrationLockConnector) Driver() driver.Driver         { return migrationLockDriver{c.spy} }
func (d migrationLockDriver) Open(string) (driver.Conn, error) { return migrationLockConn{d.spy}, nil }
func (c migrationLockConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (c migrationLockConn) Close() error { return nil }
func (c migrationLockConn) Begin() (driver.Tx, error) {
	return nil, errors.New("unexpected transaction")
}
func (c migrationLockConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.spy.mu.Lock()
	defer c.spy.mu.Unlock()
	switch {
	case strings.Contains(query, "pg_advisory_lock"):
		c.spy.statements = append(c.spy.statements, "lock")
	case strings.Contains(query, "pg_advisory_unlock"):
		c.spy.statements = append(c.spy.statements, "unlock")
	default:
		c.spy.statements = append(c.spy.statements, "ddl")
		return nil, errors.New("stop after lock")
	}
	return driver.RowsAffected(1), nil
}

// PostgreSQL migration work acquires and releases its advisory lock, even on error.
func TestPostgresMigrationUsesAdvisoryLock(t *testing.T) {
	spy := &migrationLockSpy{}
	sqlDB := sql.OpenDB(migrationLockConnector{spy})
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(2)
	runner := &MigrationRunner{db: &DB{DB: sqlDB, Driver: "postgres"}}
	called := false
	err := runner.withMigrationChecksums(context.Background(), func(checksumExecutor) error { called = true; return nil })
	if err == nil || called {
		t.Fatalf("error=%v callback called=%t", err, called)
	}
	spy.mu.Lock()
	got := append([]string(nil), spy.statements...)
	spy.mu.Unlock()
	if want := []string{"lock", "ddl", "unlock"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("statements=%v, want %v", got, want)
	}
}
