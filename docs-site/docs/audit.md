# Durable audit events

The `audit` package provides a SQL sink for durable operation events. Run
`(*audit.SQLSink).EnsureSchema` during application setup, or apply the matching
PostgreSQL or SQLite migration in `audit/migrations`.

`SQLSink.Record(ctx, tx, event)` accepts the transaction that owns the business
write. When `tx` is supplied, the event commits or rolls back with that
transaction. A nil transaction writes directly through the sink's database.
The caller owns transaction boundaries; the sink never commits a transaction.

`audit.FailClosed` is the zero value and returns a write error to the caller.
`audit.FailOpen` logs the correlation ID and increments the configured
`audit.Counter`, then permits the operation to continue. Fail-open should be
selected explicitly. The audit table has no automatic retention policy; plan
retention and database access controls for your deployment.

Admin create, update, and delete run the model write and history write in one
database transaction when the ORM manager has a database. The built-in database
history store joins that transaction, so fail-closed history errors roll back
the model write and fail-open errors allow it to commit. Custom
`core.HistoryManager` implementations remain source-compatible. To participate
in the same transaction, they can also implement
`core.TransactionalHistoryManager`; other custom stores receive the legacy
`LogAction` call while the model transaction is open. A non-transactional store
cannot guarantee atomic history persistence with the model: its successful
write may remain even if the model transaction later fails.
