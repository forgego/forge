// Package audit writes durable, transaction-aware audit events.
package audit

import "time"

type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeDenied  Outcome = "denied"
	OutcomeFailed  Outcome = "failed"
)

// Event describes a security relevant or business significant operation.
type Event struct {
	ID            int64
	OccurredAt    time.Time
	ActorID       string
	ActorKind     string
	Operation     string
	Model         string
	ObjectID      string
	ObjectRepr    string
	Outcome       Outcome
	Changes       string
	CorrelationID string
	Source        string
}
