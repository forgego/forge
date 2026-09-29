package rest

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/forgego/forge/internal/dberrors"
	validation "github.com/forgego/forge/validate"
)

// respondWriteError reports a failed single-object create or update.
//
// Validation failures and database constraint violations are the client's
// to fix, so they become 4xx responses with per-field details the admin form
// renders next to each input. Anything else is an unexpected storage failure:
// it is logged and answered with a generic 500 so driver text (SQL, table and
// constraint names) does not leak to the browser.
func respondWriteError(w http.ResponseWriter, failureCode string, err error) {
	f := classifyWriteError(failureCode, err)
	respondError(w, f.status, f.code, f.message, f.details)
}

// writeFailure is the client-safe description of a failed create or update.
type writeFailure struct {
	status  int
	code    string
	message string
	details map[string]interface{}
}

// classifyWriteError describes a failed create or update for the single and
// bulk write endpoints alike. Client-fixable failures keep a 4xx status and a
// message safe to show; anything else is logged and reported as failureCode
// with a generic message, never the driver text.
func classifyWriteError(failureCode string, err error) writeFailure {
	var verrs *validation.ValidationErrors
	if errors.As(err, &verrs) {
		return writeFailure{http.StatusBadRequest, "validation_error", err.Error(), validationDetails(err)}
	}
	if v, ok := classifyConstraintViolation(err); ok {
		details := map[string]interface{}{v.field: []string{v.message}}
		return writeFailure{v.status, v.code, v.message, details}
	}
	if !isDriverError(err) && isValidationError(err) {
		return writeFailure{http.StatusBadRequest, "validation_error", err.Error(), nil}
	}
	log.Printf("admin: %s: %s", failureCode, strconv.Quote(err.Error()))
	return writeFailure{http.StatusInternalServerError, failureCode, "The change could not be saved because of a server error", nil}
}

// deleteFailure describes a failed delete without echoing driver text. A
// foreign-key violation means other rows still reference the record, which
// the operator can resolve; anything else is logged as a server error.
func deleteFailure(err error) (status int, code, message string) {
	if v, ok := classifyConstraintViolation(err); ok && v.code == "invalid_reference" {
		return http.StatusConflict, "in_use", "Other records still refer to this record, so it cannot be deleted."
	}
	log.Printf("admin: delete_failed: %s", strconv.Quote(err.Error()))
	return http.StatusInternalServerError, "delete_failed", "The record could not be deleted because of a server error"
}

type constraintViolation struct {
	status  int
	code    string
	field   string
	message string
}

// classifyConstraintViolation recognizes PostgreSQL and SQLite integrity
// errors and describes them without echoing the driver message.
func classifyConstraintViolation(err error) (constraintViolation, bool) {
	v, ok := dberrors.Classify(err)
	if !ok {
		return constraintViolation{}, false
	}
	switch v.Kind {
	case dberrors.Unique:
		return uniqueViolation(v.Field), true
	case dberrors.ForeignKey:
		return foreignKeyViolation(v.Field), true
	case dberrors.NotNull:
		return notNullViolation(v.Field), true
	case dberrors.Check:
		return checkViolation(), true
	case dberrors.DataException:
		// The driver message quotes the rejected input, so it is replaced.
		return dataException(v.Field), true
	}
	return constraintViolation{}, false
}

// isDriverError reports whether err comes from a supported database driver.
// Driver messages quote SQL, identifiers and rejected input, so the keyword
// fallback in isValidationError must never echo them.
func isDriverError(err error) bool {
	return dberrors.IsDriverError(err)
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

func dataException(field string) constraintViolation {
	f := singleField(field)
	msg := "A value is not valid for its field."
	if f != "non_field_errors" {
		msg = fmt.Sprintf("%s has an invalid value.", f)
	}
	return constraintViolation{status: http.StatusBadRequest, code: "validation_error", field: f, message: msg}
}
