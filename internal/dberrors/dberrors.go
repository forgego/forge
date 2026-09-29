// Package dberrors recognizes integrity and data errors from the supported
// database drivers (PostgreSQL via lib/pq, SQLite via go-sqlite3). It only
// classifies: each layer (the REST API, the admin API) decides the status
// code and message it answers with, and never echoes the driver text, which
// quotes SQL, identifiers and rejected input.
package dberrors

import (
	"errors"
	"regexp"
	"strings"

	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
)

// Kind is the class of a recognized database error.
type Kind int

const (
	// Unique is a unique or primary key violation.
	Unique Kind = iota + 1
	// ForeignKey is a foreign key violation.
	ForeignKey
	// NotNull is a NOT NULL violation.
	NotNull
	// Check is a CHECK violation, or another SQLite constraint failure.
	Check
	// DataException is a value the column type cannot hold (PostgreSQL
	// class 22: too long, bad timestamp, bad integer, ...).
	DataException
)

// Violation describes a recognized database error.
type Violation struct {
	Kind Kind
	// Field is the column the driver names, "" when it names none. It can
	// list several columns ("a, b") for a composite key.
	Field string
}

// pqKeyColumns extracts the column list from a PostgreSQL constraint detail
// such as `Key (slug)=(shoes) already exists.`
var pqKeyColumns = regexp.MustCompile(`^Key \(([^)]+)\)=`)

// Classify recognizes PostgreSQL integrity (class 23) and data (class 22)
// errors and SQLite constraint errors anywhere in err's chain.
func Classify(err error) (Violation, bool) {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		field := pgErr.Column
		if m := pqKeyColumns.FindStringSubmatch(pgErr.Detail); m != nil {
			field = m[1]
		}
		switch string(pgErr.Code) {
		case "23505":
			return Violation{Kind: Unique, Field: field}, true
		case "23503":
			return Violation{Kind: ForeignKey, Field: field}, true
		case "23502":
			return Violation{Kind: NotNull, Field: field}, true
		case "23514":
			return Violation{Kind: Check, Field: field}, true
		}
		if pgErr.Code.Class() == "22" {
			return Violation{Kind: DataException, Field: field}, true
		}
		return Violation{}, false
	}

	var liteErr sqlite3.Error
	if errors.As(err, &liteErr) && liteErr.Code == sqlite3.ErrConstraint {
		// SQLite reports e.g. "UNIQUE constraint failed: categories.slug".
		field := ""
		if _, cols, ok := strings.Cut(liteErr.Error(), "constraint failed: "); ok && !strings.Contains(cols, ",") {
			if _, col, ok := strings.Cut(cols, "."); ok {
				field = col
			}
		}
		switch liteErr.ExtendedCode {
		case sqlite3.ErrConstraintUnique, sqlite3.ErrConstraintPrimaryKey:
			return Violation{Kind: Unique, Field: field}, true
		case sqlite3.ErrConstraintForeignKey:
			return Violation{Kind: ForeignKey, Field: field}, true
		case sqlite3.ErrConstraintNotNull:
			return Violation{Kind: NotNull, Field: field}, true
		default:
			return Violation{Kind: Check, Field: field}, true
		}
	}
	return Violation{}, false
}

// IsDriverError reports whether err comes from a supported database driver.
func IsDriverError(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return true
	}
	var liteErr sqlite3.Error
	return errors.As(err, &liteErr)
}
