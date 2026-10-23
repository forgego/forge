package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	apierrors "github.com/forgego/forge/api/errors"
	"github.com/forgego/forge/operation"
)

// OperationActorResolver adapts the authenticated principal stored in a request
// context to the operation runner's actor contract.
type OperationActorResolver func(*http.Request) (operation.Actor, error)

// OperationAction returns a JSON POST handler for an operation. The resolver
// must read the principal from request context and fail closed when absent.
func OperationAction[In, Out any](runner *operation.Runner, op operation.Operation[In, Out], resolve OperationActorResolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input In
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := decoder.Decode(&input); err != nil {
			apierrors.WriteError(w, r, apierrors.NewProblem(apierrors.ErrorTypeValidation, apierrors.CodeInvalidJSON, http.StatusBadRequest, "Invalid request", "Request body must contain valid JSON"))
			return
		}
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			apierrors.WriteError(w, r, apierrors.NewProblem(apierrors.ErrorTypeValidation, apierrors.CodeInvalidJSON, http.StatusBadRequest, "Invalid request", "Request body must contain one JSON value"))
			return
		}
		if resolve == nil {
			apierrors.WriteError(w, r, errors.New("operation actor resolver is not configured"))
			return
		}
		actor, err := resolve(r)
		if err != nil {
			apierrors.WriteError(w, r, err)
			return
		}
		result, err := operation.Run(operation.WithActor(r.Context(), actor), runner, op, input)
		if err != nil {
			if errors.Is(err, operation.ErrDenied) {
				apierrors.WriteError(w, r, apierrors.NewProblem(apierrors.ErrorTypeAuthorization, apierrors.CodePermissionDenied, http.StatusForbidden, "Permission denied", "You are not allowed to perform this operation"))
				return
			}
			if errors.Is(err, operation.ErrInvalidOperation) {
				apierrors.WriteError(w, r, apierrors.NewProblem(apierrors.ErrorTypeValidation, apierrors.CodeValidationError, http.StatusBadRequest, "Invalid operation", "The operation request is invalid"))
				return
			}
			apierrors.WriteError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
	}
}
