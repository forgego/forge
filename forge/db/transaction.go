package db

import (
	"context"
	"database/sql"
	"errors"
	"unicode"
)

// Tx wraps *sql.Tx with additional functionality
type Tx struct {
	*sql.Tx
	db *DB
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	if db == nil || db.DB == nil {
		return nil, errors.New("database connection is nil")
	}

	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx: tx,
		db: db,
	}, nil
}

// WithTx executes a function within a transaction
func (db *DB) WithTx(ctx context.Context, fn func(*Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			// nolint:errcheck // Rollback error in defer can't be handled meaningfully
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			// nolint:errcheck // Rollback error in defer can't be handled meaningfully
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// Savepoint represents a savepoint in a transaction
type Savepoint struct {
	tx   *Tx
	name string
}

// Error definitions for savepoint validation
var (
	// ErrEmptySavepointName is returned when savepoint name is empty
	ErrEmptySavepointName = errors.New("savepoint name cannot be empty")
	// ErrSavepointNameTooLong is returned when savepoint name exceeds maximum length
	ErrSavepointNameTooLong = errors.New("savepoint name too long: maximum 128 characters")
	// ErrInvalidSavepointName is returned when savepoint name contains invalid characters
	ErrInvalidSavepointName = errors.New("invalid savepoint name: must contain only alphanumeric characters and underscores")
)

// validateSavepointName validates that a savepoint name is safe to use in SQL
func validateSavepointName(name string) error {
	if name == "" {
		return ErrEmptySavepointName
	}
	if len(name) > 128 {
		return ErrSavepointNameTooLong
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return ErrInvalidSavepointName
		}
	}
	return nil
}

// CreateSavepoint creates a savepoint
func (tx *Tx) CreateSavepoint(name string) (*Savepoint, error) {
	if err := validateSavepointName(name); err != nil {
		return nil, err
	}
	_, err := tx.Exec("SAVEPOINT " + name)
	if err != nil {
		return nil, err
	}

	return &Savepoint{
		name: name,
		tx:   tx,
	}, nil
}

// RollbackToSavepoint rolls back to a savepoint
func (sp *Savepoint) RollbackToSavepoint() error {
	_, err := sp.tx.Exec("ROLLBACK TO SAVEPOINT " + sp.name)
	return err
}

// ReleaseSavepoint releases a savepoint
func (sp *Savepoint) ReleaseSavepoint() error {
	_, err := sp.tx.Exec("RELEASE SAVEPOINT " + sp.name)
	return err
}

