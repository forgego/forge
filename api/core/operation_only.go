package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ErrOperationOnlyField indicates an attempt to modify an operation-only field
// through a generic write endpoint.
type ErrOperationOnlyField struct {
	Field string
}

func (e *ErrOperationOnlyField) Error() string {
	return fmt.Sprintf("field %q is operation-only and cannot be changed directly", e.Field)
}

func (e *ErrOperationOnlyField) Code() string {
	return "operation_only_field"
}

func (e *ErrOperationOnlyField) StatusCode() int {
	return http.StatusForbidden
}

// CheckOperationOnlyFields checks whether data contains any operation-only field.
// It returns a non-nil *ErrOperationOnlyField with the offending field name if found.
func CheckOperationOnlyFields(data map[string]interface{}, operationOnlyFields []string) *ErrOperationOnlyField {
	if len(data) == 0 || len(operationOnlyFields) == 0 {
		return nil
	}
	for key := range data {
		for _, opField := range operationOnlyFields {
			if strings.EqualFold(key, opField) {
				return &ErrOperationOnlyField{Field: key}
			}
		}
	}
	return nil
}

// EnforceOperationOnlyFields checks data against operationOnlyFields. If any
// operation-only field is present, it writes an HTTP 403 Forbidden response with
// error code "operation_only_field" and returns false. Otherwise it returns true.
func EnforceOperationOnlyFields(w http.ResponseWriter, r *http.Request, data map[string]interface{}, operationOnlyFields []string) bool {
	err := CheckOperationOnlyFields(data, operationOnlyFields)
	if err == nil {
		return true
	}
	RespondOperationOnly(w, r, err)
	return false
}

// RespondOperationOnly writes the standard 403 response for an operation-only field rejection.
func RespondOperationOnly(w http.ResponseWriter, r *http.Request, err *ErrOperationOnlyField) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    err.Code(),
			"message": err.Error(),
			"details": map[string]interface{}{
				"field": err.Field,
			},
		},
		"code":    err.Code(),
		"message": err.Error(),
		"detail":  err.Error(),
		"status":  http.StatusForbidden,
	})
}
