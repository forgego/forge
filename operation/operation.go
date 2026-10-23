// Package operation runs authorized business operations atomically with audit
// records and post-commit effects.
package operation

import (
	"context"

	"github.com/forgego/forge/audit"
)

// Actor is the minimal identity contract needed by the runner. Applications
// can adapt their identity/principal type by implementing these methods.
// Permission checks must fail closed when permission state cannot be resolved.
type Actor interface {
	ID() string
	Kind() string
	HasPermission(context.Context, string) (bool, error)
}

// ActorResolver resolves the actor for an invocation. The default resolver
// reads an Actor placed in context with WithActor.
type ActorResolver interface {
	ResolveActor(context.Context) (Actor, error)
}

// RunFunc performs the business write using the transaction supplied by the
// runner. It returns effects that are called only after a successful commit.
type RunFunc[In, Out any] func(context.Context, Transaction, Actor, In) (Out, []Effect, error)

// Effect is work that must happen only after the database transaction commits.
// Errors are logged and do not change the already committed operation result.
type Effect func(context.Context) error

// Operation declares an operation name, its required permission, and its
// transactional implementation. An empty permission is always denied.
type Operation[In, Out any] struct {
	Name       string
	Permission string
	Run        RunFunc[In, Out]
	Audit      AuditDetails
}

// AuditDetails supplies optional information attached to the operation event.
type AuditDetails struct {
	Model         string
	ObjectID      string
	ObjectRepr    string
	CorrelationID string
	Source        string
}

func successEvent(actor Actor, name string, details AuditDetails) audit.Event {
	return event(actor, name, details, audit.OutcomeSuccess)
}

func event(actor Actor, name string, details AuditDetails, outcome audit.Outcome) audit.Event {
	result := audit.Event{
		Operation:     name,
		Model:         details.Model,
		ObjectID:      details.ObjectID,
		ObjectRepr:    details.ObjectRepr,
		Outcome:       outcome,
		CorrelationID: details.CorrelationID,
		Source:        details.Source,
	}
	if actor != nil {
		result.ActorID = actor.ID()
		result.ActorKind = actor.Kind()
	}
	return result
}
