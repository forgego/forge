package execute

import (
	"context"
	"fmt"
)

// RollbackOptions contains options for rollback operations
type RollbackOptions struct {
	Version uint
	Steps   int // Number of steps to rollback (if Version is 0)
}

// RollbackManager handles rollback operations
type RollbackManager struct {
	executor *Executor
}

// NewRollbackManager creates a new rollback manager
func NewRollbackManager(executor *Executor) *RollbackManager {
	return &RollbackManager{
		executor: executor,
	}
}

// Rollback performs a rollback operation based on options
func (r *RollbackManager) Rollback(ctx context.Context, opts RollbackOptions) error {
	if opts.Version > 0 {
		return r.executor.RollbackTo(ctx, opts.Version)
	}
	if opts.Steps > 0 {
		// Rollback multiple steps
		for i := 0; i < opts.Steps; i++ {
			if err := r.executor.Rollback(ctx); err != nil {
				return fmt.Errorf("failed to rollback step %d: %w", i+1, err)
			}
		}
		return nil
	}
	// Default: rollback one step
	return r.executor.Rollback(ctx)
}
