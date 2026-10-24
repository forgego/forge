package core

import (
	"context"
	"errors"
	"fmt"

	api "github.com/forgego/forge/api/core"
	"github.com/forgego/forge/identity/principal"
	"github.com/forgego/forge/operation"
)

// OperationAction adapts a typed business operation for an admin request.
// Parameters and results are scoped to each Execute call.
type OperationAction[In, Out any] struct {
	Runner               *operation.Runner
	Operation            operation.Operation[In, Out]
	RequiresConfirmation bool
}

// ActionHTTPError reports an HTTP status suitable for an admin adapter.
type ActionHTTPError struct {
	StatusCode int
	Err        error
}

func (e *ActionHTTPError) Error() string { return e.Err.Error() }
func (e *ActionHTTPError) Unwrap() error { return e.Err }

// Execute runs the operation with request identity and invocation parameters.
func (a OperationAction[In, Out]) Execute(ctx context.Context, params In, confirmed bool) (Out, error) {
	var zero Out
	if a.RequiresConfirmation && !confirmed {
		return zero, &ActionHTTPError{StatusCode: 400, Err: errors.New("operation confirmation required")}
	}
	if actor, ok := api.PrincipalFromContext(ctx); ok {
		ctx = operation.WithActor(ctx, principal.Actor(actor))
	}
	result, err := operation.Run(ctx, a.Runner, a.Operation, params)
	if err != nil {
		if errors.Is(err, operation.ErrDenied) {
			return zero, &ActionHTTPError{StatusCode: 403, Err: err}
		}
		if errors.Is(err, operation.ErrConflict) {
			return zero, &ActionHTTPError{StatusCode: 409, Err: err}
		}
		return zero, fmt.Errorf("run admin operation %q: %w", a.Operation.Name, err)
	}
	return result, nil
}
