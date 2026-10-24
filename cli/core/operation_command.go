package core

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/forgego/forge/identity/principal"
	"github.com/forgego/forge/operation"
)

const (
	// OperationExitDenied indicates an authorization failure.
	OperationExitDenied = 2
	// OperationExitInvalidInput indicates invalid operation input.
	OperationExitInvalidInput = 3
	// OperationExitConflict indicates an optimistic-concurrency or state conflict.
	OperationExitConflict = 4
	// OperationExitFailed indicates an unexpected operation failure.
	OperationExitFailed = 5
)

// OperationError carries a stable CLI category and exit code with its cause.
type OperationError struct {
	Category string
	ExitCode int
	Err      error
}

func (e *OperationError) Error() string { return e.Err.Error() }
func (e *OperationError) Unwrap() error { return e.Err }

// OperationCategoryDenied identifies authorization failures.
const OperationCategoryDenied = "denied"

// OperationCategoryInvalidInput identifies malformed or invalid input.
const OperationCategoryInvalidInput = "invalid_input"

// OperationCategoryConflict identifies optimistic-concurrency or state conflicts.
const OperationCategoryConflict = "conflict"

// OperationCategoryFailed identifies unexpected operation failures.
const OperationCategoryFailed = "failed"

// OperationCommand runs a typed operation. Input is read as JSON from the
// supplied reader (typically stdin), and the successful result is JSON-encoded.
type OperationCommand[In, Out any] struct {
	Runner    *operation.Runner
	Operation operation.Operation[In, Out]
	Input     io.Reader
	Output    io.Writer
}

func (c OperationCommand[In, Out]) Execute(ctx context.Context, cli *Context) error {
	var input In
	reader := c.Input
	if reader == nil {
		return operationError("invalid_input", OperationExitInvalidInput, errors.New("operation input reader is not configured"))
	}
	if err := json.NewDecoder(reader).Decode(&input); err != nil {
		return operationError("invalid_input", OperationExitInvalidInput, err)
	}
	if cli == nil || cli.Principal == nil {
		return operationError("denied", OperationExitDenied, operation.ErrDenied)
	}
	result, err := operation.Run(operation.WithActor(ctx, principal.Actor(*cli.Principal)), c.Runner, c.Operation, input)
	if err != nil {
		if errors.Is(err, operation.ErrDenied) {
			return operationError("denied", OperationExitDenied, err)
		}
		if errors.Is(err, operation.ErrInvalidOperation) {
			return operationError("invalid_input", OperationExitInvalidInput, err)
		}
		if errors.Is(err, operation.ErrConflict) {
			return operationError(OperationCategoryConflict, OperationExitConflict, err)
		}
		return operationError("failed", OperationExitFailed, err)
	}
	output := c.Output
	if output == nil {
		output = io.Discard
	}
	if err := json.NewEncoder(output).Encode(result); err != nil {
		return operationError("failed", OperationExitFailed, err)
	}
	return nil
}

func operationError(category string, code int, err error) error {
	return &OperationError{Category: category, ExitCode: code, Err: err}
}
