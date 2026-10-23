---
sidebar_position: 24
title: Business operations
---

# Business operations

The `operation` package runs multi-step business actions through one authorized,
audited database transaction. It is intended for actions whose invariants do not
fit a single CRUD request. Admin, REST, and CLI adapters can call the same
operation runner.

Each operation declares a stable name, one required permission, and a callback.
The runner resolves an actor from the invocation context by default. Put an
implementation of `operation.Actor` in context with `operation.WithActor`; an
application can instead supply an `operation.ActorResolver` that adapts its
identity principal. An actor must implement `HasPermission`, and an empty
permission, missing actor, permission lookup error, or negative result is denied.

```go
op := operation.Operation[Input, Result]{
    Name:       "orders.approve",
    Permission: "orders.approve_order",
    Run: func(ctx context.Context, tx operation.Transaction, actor operation.Actor, in Input) (Result, []operation.Effect, error) {
        // Read and write through tx; the runner controls commit and rollback.
        result, err := approveOrder(ctx, tx, in.OrderID)
        if err != nil {
            return Result{}, nil, err
        }
        effect := func(ctx context.Context) error { return sendApprovalNotice(ctx, result) }
        return result, []operation.Effect{effect}, nil
    },
}

result, err := operation.Run(ctx, runner, op, input)
```

The order is authorization, transaction begin, callback, success audit, commit,
then effects. Success audit records share the business transaction. On denial,
callback error, or panic, business writes roll back; a `denied` or `failed`
event is attempted in its own short transaction. Audit writes fail closed by
default. Select `audit.FailOpen` explicitly only when losing an audit event is
acceptable. Post-commit effect errors are logged and never undo the committed
business operation.

The runner guarantees transaction boundaries, but the callback owns its
domain-specific validation, state checks, and any required target-row locking.
Use the database's appropriate locking strategy where concurrent updates could
violate an invariant.

## REST adapter

`api.OperationAction` provides a JSON `POST` handler for an operation. Supply
an `OperationActorResolver` that reads the authenticated principal from the
request context and adapts it to `operation.Actor`; a missing principal must
resolve to no actor so the runner denies the call. Register the returned
handler with `Router.Action`. Invalid JSON, denied calls, and execution errors
use the API's problem error response format. Successful calls return the
operation result as JSON.
