package rest

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	validation "github.com/forgego/forge/validate"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
)

// respondWriteError reports a failed single-object create or update.
//
// Validation failures and database constraint violations are the client's
// to fix, so they become 4xx responses with per-field details the admin form
// renders next to each input. Anything else is an unexpected storage failure:
// it is logged and answered with a generic 500 so driver text (SQL, table and
// constraint names) does not leak to the browser.
func respondWriteError(w http.ResponseWriter, failureCode string, err error) {
	var verrs *validation.ValidationErrors
	if errors.As(err, &verrs) {
		respondError(w, http.StatusBadRequest, "validation_error", err.Error(), validationDetails(err))
		return
	}
	if v, ok := classifyConstraintViolation(err); ok {
		details := map[string]interface{}{v.field: []string{v.message}}
		respondError(w, v.status, v.code, v.message, details)
		return
	}
	if isValidationError(err) {
		respondError(w, http.StatusBadRequest, "validation_error", err.Error(), validationDetails(err))
		return
	}
	log.Printf("admin: %s: %v", failureCode, err)
	respondError(w, http.StatusInternalServerError, failureCode, "The change could not be saved because of a server error", nil)
}

type constraintViolation struct {
	status  int
	code    string
	field   string
	message string
}

// pqKeyColumns extracts the column list from a PostgreSQL constraint detail
// such as `Key (slug)=(shoes) already exists.`
var pqKeyColumns = regexp.MustCompile(`^Key \(([^)]+)\)=`)

// classifyConstraintViolation recognises PostgreSQL and SQLite integrity
// errors and describes them without echoing the driver message.
func classifyConstraintViolation(err error) (constraintViolation, bool) {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		field := pgErr.Column
		if m := pqKeyColumns.FindStringSubmatch(pgErr.Detail); m != nil {
			field = m[1]
		}
		switch string(pgErr.Code) {
		case "23505":
			return uniqueViolation(field), true
		case "23503":
			return foreignKeyViolation(field), true
		case "23502":
			return notNullViolation(field), true
		case "23514":
			return checkViolation(), true
		}
		return constraintViolation{}, false
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
			return uniqueViolation(field), true
		case sqlite3.ErrConstraintForeignKey:
			return foreignKeyViolation(field), true
		case sqlite3.ErrConstraintNotNull:
			return notNullViolation(field), true
		default:
			return checkViolation(), true
		}
	}
	return constraintViolation{}, false
}

func singleField(field string) string {
	field = strings.TrimSpace(field)
	if field == "" || strings.ContainsAny(field, ", ") {
		return "non_field_errors"
	}
	return field
}

func uniqueViolation(field string) constraintViolation {
	f := singleField(field)
	msg := "A record with this value already exists."
	if f != "non_field_errors" {
		msg = fmt.Sprintf("A record with this %s already exists.", f)
	}
	return constraintViolation{status: http.StatusConflict, code: "conflict", field: f, message: msg}
}

func foreignKeyViolation(field string) constraintViolation {
	f := singleField(field)
	msg := "A referenced record does not exist, or other records still refer to this one."
	if f != "non_field_errors" {
		msg = fmt.Sprintf("The selected %s does not exist.", f)
	}
	return constraintViolation{status: http.StatusBadRequest, code: "invalid_reference", field: f, message: msg}
}

func notNullViolation(field string) constraintViolation {
	f := singleField(field)
	msg := "A required value is missing."
	if f != "non_field_errors" {
		msg = fmt.Sprintf("%s is required.", f)
	}
	return constraintViolation{status: http.StatusBadRequest, code: "validation_error", field: f, message: msg}
}

func checkViolation() constraintViolation {
	return constraintViolation{
		status:  http.StatusBadRequest,
		code:    "validation_error",
		field:   "non_field_errors",
		message: "A value is not allowed by a database constraint.",
	}
}
