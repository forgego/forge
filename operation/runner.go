package operation

import (
	"context"
	"errors"
	"fmt"

	"github.com/forgego/forge/audit"
	"github.com/forgego/forge/db"
	forgelog "github.com/forgego/forge/log"
	"go.uber.org/zap"
)

var (
	// ErrDenied indicates that an invocation has no authorized actor.
	ErrDenied = errors.New("operation denied")
	// ErrConflict marks optimistic-concurrency or state conflicts. Operation
	// callbacks should wrap this sentinel when reporting those conflicts.
	ErrConflict = errors.New("operation conflict")
	// ErrInvalidOperation indicates that an operation is incomplete.
	ErrInvalidOperation = errors.New("invalid operation")
)

// Transaction is the subset of a database transaction available to operation
// code. Transaction ownership stays with the runner.
type Transaction interface {
	audit.DBTX
}

// ManagedTransaction adds the boundary methods used by a Beginner. Operation
// callbacks receive only Transaction, so they cannot commit or roll back early.
type ManagedTransaction interface {
	Transaction
	Commit() error
	Rollback() error
}

// Beginner starts transactions for business work and separate failure audits.
type Beginner interface {
	Begin(context.Context) (ManagedTransaction, error)
}

type databaseBeginner struct{ database *db.DB }

func (b databaseBeginner) Begin(ctx context.Context) (ManagedTransaction, error) {
	return b.database.BeginTx(ctx, nil)
}

// Runner enforces authorization, transactional execution, audit, commit, and
// post-commit effect ordering. Its zero value is not configured; use
// NewRunner.
type Runner struct {
	beginner Beginner
	sink     audit.Sink
	resolver ActorResolver
	policy   audit.Policy
	logger   *forgelog.Logger
}

// Option configures a Runner.
type Option func(*Runner)

// WithActorResolver selects the source of invocation actors.
func WithActorResolver(resolver ActorResolver) Option {
	return func(r *Runner) { r.resolver = resolver }
}

// WithAuditPolicy selects the audit failure policy. FailClosed is the default.
func WithAuditPolicy(policy audit.Policy) Option {
	return func(r *Runner) { r.policy = policy }
}

// WithLogger configures logging for post-commit effect failures.
func WithLogger(logger *forgelog.Logger) Option {
	return func(r *Runner) { r.logger = logger }
}

// WithBeginner replaces the database transaction provider. It is useful for
// adapters and tests that implement the Transaction contract.
func WithBeginner(beginner Beginner) Option {
	return func(r *Runner) { r.beginner = beginner }
}

// NewRunner creates a runner backed by database. Actors are resolved from
// context unless another resolver is supplied. Audit failures fail closed.
func NewRunner(database *db.DB, sink audit.Sink, options ...Option) *Runner {
	runner := &Runner{
		sink:     sink,
		resolver: contextActorResolver{},
		policy:   audit.FailClosed,
		logger:   forgelog.QuickLogger(),
	}
	if database != nil {
		runner.beginner = databaseBeginner{database: database}
	}
	for _, option := range options {
		if option != nil {
			option(runner)
		}
	}
	return runner
}

// WithActor returns a derived context that resolves to actor during Run.
func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorContextKey{}, actor)
}

type actorContextKey struct{}

type contextActorResolver struct{}

func (contextActorResolver) ResolveActor(ctx context.Context) (Actor, error) {
	actor, _ := ctx.Value(actorContextKey{}).(Actor)
	if actor == nil {
		return nil, nil
	}
	return actor, nil
}

// Run executes op atomically. Denied and failed audit events are written in
// their own short transaction after the business transaction is absent or has
// rolled back. Only successful commits trigger effects.
func Run[In, Out any](ctx context.Context, runner *Runner, op Operation[In, Out], input In) (out Out, err error) {
	if runner == nil || runner.beginner == nil || runner.sink == nil || runner.resolver == nil {
		return out, fmt.Errorf("%w: runner is not configured", ErrInvalidOperation)
	}
	if op.Name == "" || op.Run == nil {
		return out, fmt.Errorf("%w: name and run function are required", ErrInvalidOperation)
	}

	actor, resolveErr := runner.resolver.ResolveActor(ctx)
	allowed := false
	if resolveErr == nil && actor != nil && op.Permission != "" {
		allowed, resolveErr = actor.HasPermission(ctx, op.Permission)
	}
	if resolveErr != nil || !allowed {
		auditErr := runner.recordOutsideTransaction(ctx, event(actor, op.Name, op.Audit, audit.OutcomeDenied))
		if auditErr != nil {
			return out, errors.Join(ErrDenied, fmt.Errorf("record denied operation: %w", auditErr))
		}
		return out, ErrDenied
	}

	tx, beginErr := runner.beginner.Begin(ctx)
	if beginErr != nil {
		return out, runner.failedWithEvent(ctx, actor, op.Name, op.Audit, fmt.Errorf("begin operation transaction: %w", beginErr))
	}
	active := true
	defer func() {
		if recovered := recover(); recovered != nil {
			if active {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					runner.logger.Error("rollback operation transaction after panic failed", zap.String("operation", op.Name), zap.Error(rollbackErr))
				}
				active = false
				_ = runner.recordOutsideTransaction(ctx, event(actor, op.Name, op.Audit, audit.OutcomeFailed))
			}
			panic(recovered)
		}
		if active {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Join(err, fmt.Errorf("rollback operation transaction: %w", rollbackErr))
			}
		}
	}()
	fail := func(cause error) error {
		rollbackErr := tx.Rollback()
		active = false
		if rollbackErr != nil {
			cause = errors.Join(cause, fmt.Errorf("rollback operation transaction: %w", rollbackErr))
		}
		return runner.failedWithEvent(ctx, actor, op.Name, op.Audit, cause)
	}

	effects, runErr := func() ([]Effect, error) {
		result, pending, runErr := op.Run(ctx, tx, actor, input)
		out = result
		return pending, runErr
	}()
	if runErr != nil {
		return out, fail(fmt.Errorf("execute operation: %w", runErr))
	}
	if auditErr := audit.Record(ctx, runner.sink, tx, successEvent(actor, op.Name, op.Audit), runner.policy, audit.Options{}); auditErr != nil {
		return out, fail(fmt.Errorf("record successful operation: %w", auditErr))
	}
	if commitErr := tx.Commit(); commitErr != nil {
		active = false
		return out, runner.failedWithEvent(ctx, actor, op.Name, op.Audit, fmt.Errorf("commit operation: %w", commitErr))
	}
	active = false
	for _, effect := range effects {
		if effect == nil {
			continue
		}
		if effectErr := effect(ctx); effectErr != nil {
			logger := runner.logger
			if logger == nil {
				logger = forgelog.QuickLogger()
			}
			logger.Error("post-commit operation effect failed", zap.String("operation", op.Name), zap.String("correlation_id", op.Audit.CorrelationID), zap.Error(effectErr))
		}
	}
	return out, nil
}

func (r *Runner) recordOutsideTransaction(ctx context.Context, event audit.Event) error {
	tx, err := r.beginner.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin audit transaction: %w", err)
	}
	if err := audit.Record(ctx, r.sink, tx, event, r.policy, audit.Options{}); err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return errors.Join(err, fmt.Errorf("rollback audit transaction: %w", rollbackErr))
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		commitErr := fmt.Errorf("commit audit transaction: %w", err)
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(commitErr, fmt.Errorf("rollback audit transaction: %w", rollbackErr))
		}
		return commitErr
	}
	return nil
}

func (r *Runner) failedWithEvent(ctx context.Context, actor Actor, name string, details AuditDetails, cause error) error {
	if auditErr := r.recordOutsideTransaction(ctx, event(actor, name, details, audit.OutcomeFailed)); auditErr != nil {
		return errors.Join(cause, fmt.Errorf("record failed operation: %w", auditErr))
	}
	return cause
}

var _ Beginner = databaseBeginner{}
var _ Transaction = (*db.Tx)(nil)
