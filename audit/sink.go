package audit

import (
	"context"
	"database/sql"
	"log/slog"
	"sync/atomic"

	apierrors "github.com/forgego/forge/api/errors"
)

// DBTX is implemented by database connections and transactions.
type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// Sink records an event, using tx when supplied to join the caller's transaction.
type Sink interface {
	Record(ctx context.Context, tx DBTX, event Event) error
}

type Policy uint8

const (
	// FailClosed is the zero value and returns sink failures to the caller.
	FailClosed Policy = iota
	// FailOpen logs and counts sink failures while allowing the operation to continue.
	FailOpen
)

type Options struct {
	Logger  *slog.Logger
	Counter *Counter
}
type Counter struct{ count atomic.Uint64 }

func (c *Counter) Count() uint64 {
	if c == nil {
		return 0
	}
	return c.count.Load()
}

// Record applies the audit failure policy. FailClosed is the default policy.
func Record(ctx context.Context, sink Sink, tx DBTX, event Event, policy Policy, options Options) error {
	if event.CorrelationID == "" && ctx != nil {
		event.CorrelationID = apierrors.GetRequestIDFromContext(ctx)
	}
	if err := sink.Record(ctx, tx, event); err != nil {
		return HandleError(ctx, err, event.CorrelationID, policy, options)
	}
	return nil
}

// HandleError applies the package failure policy to an audit write error.
func HandleError(ctx context.Context, err error, correlationID string, policy Policy, options Options) error {
	if err == nil {
		return nil
	}
	if correlationID == "" && ctx != nil {
		correlationID = apierrors.GetRequestIDFromContext(ctx)
	}
	if policy != FailOpen {
		return err
	}
	if options.Counter != nil {
		options.Counter.count.Add(1)
	}
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}
	logger.ErrorContext(ctx, "audit write failed", "correlation_id", correlationID, "error", err)
	return nil
}
