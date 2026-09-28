package api

import (
	"errors"
	"net/http"

	apierrors "github.com/forgego/forge/api/errors"
	"github.com/forgego/forge/api/exceptions"
	forgeerrors "github.com/forgego/forge/errors"
	"github.com/forgego/forge/validate"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
)

// persistenceException maps manager persistence failures to API exceptions
// without leaking internal details. Errors already in API form are returned
// unchanged. Validation-style errors become 400 ValidationErrors, missing
// rows become 404 NotFound, and anything else is returned unchanged so the
// default error writer renders a generic 500 with no internal message.
func persistenceException(err error) error {
	if err == nil {
		return nil
	}

	var problemErr apierrors.ProblemError
	if errors.As(err, &problemErr) {
		return problemErr
	}
	var problem *apierrors.Problem
	if errors.As(err, &problem) {
		return problem
	}

	var apiExc *exceptions.APIException
	if errors.As(err, &apiExc) {
		return apiExc
	}
	var validationExc *exceptions.ValidationError
	if errors.As(err, &validationExc) {
		return validationExc
	}
	var authFailed *exceptions.AuthenticationFailed
	if errors.As(err, &authFailed) {
		return authFailed
	}
	var notAuth *exceptions.NotAuthenticated
	if errors.As(err, &notAuth) {
		return notAuth
	}
	var permDenied *exceptions.PermissionDenied
	if errors.As(err, &permDenied) {
		return permDenied
	}
	var notFoundExc *exceptions.NotFound
	if errors.As(err, &notFoundExc) {
		return notFoundExc
	}
	var throttled *exceptions.Throttled
	if errors.As(err, &throttled) {
		return throttled
	}
	var methodNotAllowed *exceptions.MethodNotAllowed
	if errors.As(err, &methodNotAllowed) {
		return methodNotAllowed
	}
	var parseErr *exceptions.ParseError
	if errors.As(err, &parseErr) {
		return parseErr
	}
	var notAcceptable *exceptions.NotAcceptable
	if errors.As(err, &notAcceptable) {
		return notAcceptable
	}
	var unsupportedMedia *exceptions.UnsupportedMediaType
	if errors.As(err, &unsupportedMedia) {
		return unsupportedMedia
	}

	var invalidInput *forgeerrors.InvalidInputError
	if errors.As(err, &invalidInput) {
		field := invalidInput.Field
		if field == "" {
			field = "non_field_errors"
		}
		return exceptions.NewValidationError(map[string][]string{field: {invalidInput.Message}})
	}
	var notFoundErr *forgeerrors.NotFoundError
	if errors.As(err, &notFoundErr) {
		return exceptions.NewNotFound("Not found")
	}

	var validationErrs *validation.ValidationErrors
	if errors.As(err, &validationErrs) {
		grouped := make(map[string][]string, len(validationErrs.Errors))
		for _, e := range validationErrs.Errors {
			field := e.Field
			if field == "" {
				field = "non_field_errors"
			}
			grouped[field] = append(grouped[field], e.Message)
		}
		if len(grouped) == 0 {
			grouped["non_field_errors"] = []string{"Invalid input"}
		}
		return exceptions.NewValidationError(grouped)
	}
	var validationErr *validation.ValidationError
	if errors.As(err, &validationErr) {
		field := validationErr.Field
		if field == "" {
			field = "non_field_errors"
		}
		return exceptions.NewValidationError(map[string][]string{field: {validationErr.Message}})
	}

	if status, ok := constraintViolationStatus(err); ok {
		if status == http.StatusConflict {
			return exceptions.NewAPIException(status, "conflict", "Resource conflicts with an existing value", nil)
		}
		return exceptions.NewAPIException(status, "invalid_request", "Request violates a database constraint", nil)
	}

	return err
}

func constraintViolationStatus(err error) (int, bool) {
	var postgresErr *pq.Error
	if errors.As(err, &postgresErr) {
		switch string(postgresErr.Code) {
		case "23505":
			return http.StatusConflict, true
		case "23502", "23503", "23514":
			return http.StatusBadRequest, true
		}
	}

	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint {
		switch sqliteErr.ExtendedCode {
		case sqlite3.ErrConstraintPrimaryKey, sqlite3.ErrConstraintUnique:
			return http.StatusConflict, true
		case sqlite3.ErrConstraintCheck, sqlite3.ErrConstraintForeignKey, sqlite3.ErrConstraintNotNull:
			return http.StatusBadRequest, true
		default:
			return http.StatusBadRequest, true
		}
	}
	return 0, false
}
