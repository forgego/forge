package db

import (
	"context"
	"database/sql"
)

// Tx wraps *sql.Tx with additional functionality
type Tx struct {
	*sql.Tx
	db *DB
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
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

// CreateSavepoint creates a savepoint
func (tx *Tx) CreateSavepoint(name string) (*Savepoint, error) {
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
