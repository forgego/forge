---
sidebar_position: 34
title: Extension Contracts
description: Architectural and runtime contracts for Forge extension points.
---

# Extension Contracts

This document defines the architectural and runtime contracts for Forge extension points. Each contract specifies the interface or type signature, resource ownership, concurrency requirements, error semantics, semantic versioning compatibility, cleanup lifecycle, and corresponding test coverage.

Every contract is derived directly from runtime call sites in the codebase. Where behavior is intentionally omitted or left undefined in code, this document explicitly notes it as unspecified.

---

## 1. Authentication Backends

Authentication backends implement pluggable authentication strategies (e.g., password, token, OAuth, SAML) managed by a backend registry.

### Interface and File Path

Source: `identity/backends/interface.go`

```go
// AuthenticationBackend defines the interface for authentication backends
// This follows the Strategy pattern - different backends implement different auth methods
type AuthenticationBackend interface {
	// Authenticate attempts to authenticate using this backend
	// credentials is a map of credential types to values (e.g., {"username": "user", "password": "pass"})
	// Returns the authenticated user if successful, nil if not applicable, error if failed
	Authenticate(ctx context.Context, credentials map[string]string) (*models.User, error)

	// GetUser retrieves a user by identifier (for token-based auth)
	// identifier could be a token, session key, etc.
	GetUser(ctx context.Context, identifier string) (*models.User, error)

	// Supports returns true if this backend can handle the given credential type
	// credentialType examples: "password", "token", "oauth", "saml"
	Supports(credentialType string) bool

	// Name returns the backend name (e.g., "password", "token", "oauth")
	Name() string
}
```

The registry interface is also defined in `identity/backends/interface.go`:

```go
// BackendRegistry manages authentication backends
type BackendRegistry interface {
	Register(backend AuthenticationBackend)
	Get(name string) (AuthenticationBackend, error)
	GetByCredentialType(credentialType string) []AuthenticationBackend
	Authenticate(ctx context.Context, credentials map[string]string) (*models.User, error)
	GetUser(ctx context.Context, identifier string) (*models.User, error)
}
```

### Ownership Rules

Authentication backends are instantiated by the application or by factory constructors (such as `backends.NewPasswordBackend` or `backends.NewTokenBackend`) during system startup and registered into a `BackendRegistry` via `Register(backend)`.

The backend implementation does not implement `io.Closer`, and `BackendRegistry` does not own the backend lifecycle or close it. Ownership of any underlying dependencies (such as database connections, user repositories, or token caches) remains with the creator/caller.

### Concurrency

Must be safe for concurrent use.

**Derivation:** `BackendRegistry.Authenticate` and `BackendRegistry.GetUser` are called directly by HTTP request handlers (`identity/handlers/auth.go`), authentication middleware (`identity/middleware/auth.go`), and application services (`identity/service/auth.go`). Because HTTP servers dispatch incoming requests simultaneously across multiple goroutines, the registered `AuthenticationBackend` instances receive concurrent calls to `Authenticate`, `GetUser`, and `Supports`.

### Error Semantics

- **In `BackendRegistry.Authenticate(ctx, credentials)`:** The registry checks backends supporting the requested credential type (or all backends if none explicitly match). If a backend's `Authenticate` returns a non-nil error, the registry ignores the error and continues trying subsequent backends. If any backend returns a non-nil user without error (`*models.User, nil`), that user is returned immediately. If all candidate backends fail or return `nil`, the registry returns `nil, fmt.Errorf("authentication failed")`.
- **In `BackendRegistry.GetUser(ctx, identifier)`:** If a backend returns an error or a `nil` user, the registry continues to the next registered backend. If all backends fail to locate the user, the registry returns `nil, repository.ErrUserNotFound`.
- **In `StaffLoginAuthenticator.AuthenticateAdmin(ctx, username, password)` (`identity/backends/staff_login.go`):** If `backend.Authenticate` returns an error matching `ErrInvalidCredentials`, `ErrUserInactive`, or `ErrUserLocked`, it maps the failure to `&AdminLoginDeniedError{cause: err}`. Any other non-nil error is wrapped and returned as `fmt.Errorf("staff login authenticate: %w", err)`. If authentication succeeds but the returned user has an account state error or lacks `IsStaff` and `IsSuperuser` flags, `&AdminLoginDeniedError{}` is returned.

### Compatibility

Public API. The package `identity/backends` is an exported Go package in `github.com/forgego/forge` outside `internal/`, covered by semantic versioning guarantees under [Versioning and Support](versioning-and-support.md).

### Cleanup

None. Neither `AuthenticationBackend` nor `BackendRegistry` defines a teardown or close method. Clean shutdown of backing stores or connections is the responsibility of the application.

### Tested by

`identity/backends/registry_test.go` (also exercised by `identity/backends/password_test.go` and `identity/backends/staff_login_test.go`).

---

## 2. `Principal` Adapters

`Principal` represents authenticated actor identity and permissions attached to contexts and operations. Adapters convert concrete identity models and backend users into this unified representation.

### Interface or Type and File Path

Source: `identity/principal/principal.go`

```go
// Kind identifies the source and trust category of a principal.
type Kind string

const (
	KindUser      Kind = "user"
	KindService   Kind = "service"
	KindAnonymous Kind = "anonymous"
	KindSystem    Kind = "system"
)

// Principal represents the actor associated with an operation. Roles and
// permissions are snapshots supplied by the principal's trusted adapter.
type Principal struct {
	ID              string
	Kind            Kind
	DisplayName     string
	RoleNames       []string
	PermissionNames []string
	accountStateErr error
}
```

Adapter functions in `identity/principal/principal.go`:

```go
// FromUser adapts an active models.User into a user principal.
func FromUser(user *models.User) (Principal, error)

// FromBackend adapts common custom authentication backend values. Values that
// do not expose a non-empty GetID and optional GetDisplayName are ignored.
func FromBackend(user interface{}) (Principal, bool)

// Actor adapts a Principal to the operation runner's authorization contract.
func Actor(p Principal) operation.Actor

// Anonymous creates an anonymous principal.
func Anonymous() Principal

// System creates a system principal with the supplied identifier.
func System(id string) Principal

// WithPrincipal returns a context containing p. Invalid principals are stored
// as anonymous to avoid propagating an unusable identity.
func WithPrincipal(ctx context.Context, p Principal) context.Context

// PrincipalFromContext retrieves a principal from ctx.
func PrincipalFromContext(ctx context.Context) (Principal, bool)
```

Adapter function in `identity/principal_adapter.go`:

```go
// PrincipalFromUser adapts an active user model to a principal.
func PrincipalFromUser(user *models.User) (principal.Principal, error)
```

### Ownership Rules

`Principal` is an immutable value struct created on demand by adapter functions (`FromUser`, `FromBackend`, `Anonymous`, `System`). It holds no external operating system or network handles and is managed entirely by Go runtime memory allocation and garbage collection. Neither caller nor framework needs to close it.

### Concurrency

Must be safe for concurrent use.

**Derivation:** `Principal` instances are value types passed by value and stored in request contexts using `WithPrincipal`. The slice accessor methods `Roles()` and `Permissions()` return new defensive slice allocations (`append([]string(nil), ...)`). Because `Principal` values are immutable and have no mutable internal state, they are inherently safe for concurrent access across multiple goroutines.

### Error Semantics

- **`FromUser(user *models.User)`:** Evaluates `accountStateError(user)`. If the user has an invalid state (e.g., `models.ErrUserInactive` or `models.ErrUserLocked`), it returns a zero `Principal{}` and the state error. If `user.ID == 0` and `user.Username != ""`, the username is used as the principal ID. Call sites reject authentication or authorization when an error is returned.
- **`FromBackend(user interface{})`:** Uses reflection to inspect the user object. If `accountStateError(user) != nil`, the extracted ID is empty, or the user is disabled (`IsActive == false` or `IsLocked == true`), it returns `(Principal{}, false)`. Call sites treat `false` as unadaptable/unusable credentials.
- **`WithPrincipal(ctx, p)`:** If `p.ID == ""` or `p.Kind == ""`, the function normalizes the stored principal to `Anonymous()` rather than propagating an invalid principal.
- **`Actor(p).HasPermission(ctx, permission)`:** Checks `p.Permissions()`. Returns `(true, nil)` if the permission is present, and `(false, nil)` otherwise. It never returns an error.

### Compatibility

Public API. Exported types and functions in `identity/principal` and `identity` are covered by [Versioning and Support](versioning-and-support.md). (The unexported field `accountStateErr` is private; all exported fields and methods are public).

### Cleanup

None. `Principal` is an in-memory value struct. No teardown or cleanup is required.

### Tested by

`identity/principal/principal_test.go`

---

## 3. Admin Extensions and Registry Extensions

Forge provides two distinct extension mechanisms: admin UI extensions (`admin.RegisterExtension`) for serving custom frontend bundles, and registry extensions (`registry/extensions.go`) for augmenting core model, manager, query, and relation behaviors.

### 3A. Admin Extensions

#### Interface or Type and File Path

Source: `admin/extension.go`

```go
// ExtensionAPIVersion is the currently supported client API version.
const ExtensionAPIVersion = 1

// Extension represents an admin UI extension module.
type Extension struct {
	Name    string
	Assets  fs.FS
	Version int
}
```

Registration and validation functions in `admin/extension.go`:

```go
func (e Extension) Validate() error
func (s *Site) RegisterExtension(ext Extension) error
func RegisterExtension(ext Extension) error
func (s *Site) ValidateExtensions() error
func (s *Site) GetExtensions() []Extension
```

#### Ownership Rules

Created by the application or extension author and registered via `admin.RegisterExtension(ext)` or `site.RegisterExtension(ext)`. The admin site stores the `Extension` struct in its internal collection.

The admin site does not close `Assets` (the standard `io/fs.FS` interface defines no `Close()` method). Individual files opened during validation (`defer f.Close()`) or HTTP serving (`http.ServeFileFS`) are closed immediately after reading. If the underlying `Assets` filesystem wraps resources requiring explicit cleanup (such as temporary directories), the creator owns that lifecycle.

#### Concurrency

Must be safe for concurrent use.

**Derivation:** `Site.RegisterExtension`, `Site.ValidateExtensions`, and `Site.GetExtensions` synchronize access via `s.extensionsMu` (`sync.RWMutex`). Asset requests handled by `extensionAssetHandler` invoke `fs.Stat(ext.Assets, cleanPath)` and `http.ServeFileFS(w, r, ext.Assets, cleanPath)` concurrently across incoming HTTP request goroutines. Therefore, the `Assets` `fs.FS` implementation must be safe for concurrent file access across multiple goroutines.

#### Error Semantics

- **`Extension.Validate()`:** Returns an error if `Name` fails regex `^[a-z0-9-]+$`, if `Version != ExtensionAPIVersion` (`1`), if `Assets == nil`, if `index.js` cannot be opened, or if `index.js` is a directory instead of a regular file.
- **`Site.RegisterExtension(ext)`:** Calls `ext.Validate()`. If validation fails, the error is returned immediately. Under write lock, it checks whether an extension with `existing.Name == ext.Name` is already registered; if so, it returns `fmt.Errorf("admin extension %q already registered", ext.Name)`.
- **Asset HTTP Handler:** Requests using methods other than `GET` and `HEAD` receive HTTP 405 (Method Not Allowed). Paths with directory traversal, trailing slashes, invalid syntax (`!fs.ValidPath`), missing files, or directory targets receive HTTP 404 (Not Found). Every response sets `X-Content-Type-Options: nosniff`.

#### Compatibility

Public API. Explicitly governed by [Versioning and Support](versioning-and-support.md): "The admin extension API contract and `admin.ExtensionAPIVersion` (currently `1`)."

#### Cleanup

Unregistering extensions is not supported by `admin.Site` (there is no `UnregisterExtension` method). File handles opened during validation or request handling are closed immediately. The framework performs no teardown of `Assets`.

#### Tested by

`admin/extension_test.go`

---

### 3B. Registry Extensions

#### Interface or Type and File Path

Source: `registry/extensions.go`

```go
// ExtensionRegistry maintains a registry of all extensions
type ExtensionRegistry struct {
	extensions map[string]Extension
	mu         sync.RWMutex
}

// Extension is the interface that all extensions must implement
type Extension interface {
	Name() string
	Version() string
}

// FieldExtension extends field functionality
type FieldExtension interface {
	Extension
	ExtendField(field interface{}) error
}

// RelationExtension extends relation functionality
type RelationExtension interface {
	Extension
	ExtendRelation(relation interface{}) error
}

// QuerySetExtension extends QuerySet functionality
type QuerySetExtension interface {
	Extension
	ExtendQuerySet(queryset interface{}) error
}

// ManagerExtension extends Manager functionality
type ManagerExtension interface {
	Extension
	ExtendManager(manager interface{}) error
}

// AdminExtension extends Admin functionality
type AdminExtension interface {
	Extension
	ExtendAdmin(admin interface{}) error
}
```

Registration functions in `registry/extensions.go`:

```go
func RegisterExtension(ext Extension) error
func GetExtension(name string) (Extension, error)
func GetAllExtensions() map[string]Extension
```

#### Ownership Rules

Instantiated by the caller or plugin author and registered into the global extension registry via `registry.RegisterExtension(ext)`. The registry retains a reference in its internal map. The framework does not close or release extensions.

#### Concurrency

`ExtensionRegistry` guards its map with a `sync.RWMutex` (`RegisterExtension` acquires `mu.Lock()`; `GetExtension` and `GetAllExtensions` acquire `mu.RLock()`). If extension methods (`ExtendField`, `ExtendRelation`, etc.) are called concurrently by application code, the extension implementation must ensure its own concurrency safety.

#### Error Semantics

- **`RegisterExtension(ext)`:** If an extension with `ext.Name()` already exists in the registry, returns `fmt.Errorf("extension %s is already registered", ext.Name())`.
- **`GetExtension(name)`:** If the extension is not found, returns `nil, fmt.Errorf("extension %s is not registered", name)`.
- **Extension Methods (`ExtendField`, `ExtendRelation`, etc.):** Return an error if extending the target component fails. Call-site handling of returned errors is unspecified; do not rely on it.

#### Compatibility

Public API. Exported Go package `registry` is covered by [Versioning and Support](versioning-and-support.md).

#### Cleanup

Unregistering is not supported. Cleanup lifecycle is unspecified; do not rely on it.

#### Tested by

`registry/extensions_test.go`

---

## 4. Health Checkers

Health checkers report operational readiness, liveness, and component health for HTTP health endpoints.

### Interface or Type and File Path

Source: `server/health.go`

```go
// HealthChecker is the interface for health check implementations
type HealthChecker interface {
	CheckHealth(ctx context.Context) error
}

// HealthCheckFunc is a function type that implements HealthChecker
type HealthCheckFunc func(ctx context.Context) error

// CheckHealth implements HealthChecker
func (f HealthCheckFunc) CheckHealth(ctx context.Context) error {
	return f(ctx)
}
```

Registry in `server/health.go`:

```go
// HealthRegistry stores registered health checkers for an application.
type HealthRegistry struct {
	mu        sync.RWMutex
	checkers  map[string]HealthChecker
	startTime time.Time
}

func NewHealthRegistry() *HealthRegistry
func (r *HealthRegistry) Register(name string, checker HealthChecker)
func (r *HealthRegistry) RegisterFunc(name string, fn func(ctx context.Context) error)
func (r *HealthRegistry) Unregister(name string)
func (r *HealthRegistry) Uptime() time.Duration
```

### Ownership Rules

Created by the application (or adapted from functions using `HealthCheckFunc`) and registered with a `HealthRegistry` (typically via `(*Server).Health().Register(name, checker)` or `registry.Register`). The registry stores the checker in its internal map.

The framework does not own or close the checker. Any underlying resources (such as database connection pools or socket connections checked during health checks) are owned and closed by the application.

### Concurrency

Must be safe for concurrent use.

**Derivation:** `HealthHandler()` and `ReadinessHandler()` take a snapshot copy of registered checkers under `r.mu.RLock()`. While checkers within a single request snapshot are evaluated sequentially, multiple incoming HTTP requests to `/health` or `/ready` execute concurrently across separate goroutines. Each request independently invokes `CheckHealth(ctx)` on the registered checker instances. Therefore, implementations of `HealthChecker` must be safe for concurrent execution across multiple goroutines.

### Error Semantics

`CheckHealth(ctx context.Context)` is called with a context bounded by a 5-second timeout (`context.WithTimeout(req.Context(), 5*time.Second)`).

- **In `HealthHandler`:** If `CheckHealth` returns a non-nil error, the server logs the error details using `logCheckFailure("health", name, err)` with quoted names and error strings. The individual check entry `status.Checks[name]` is marked `"unhealthy"`, overall `status.Status` is set to `"unhealthy"`, and the endpoint responds with HTTP 503 (`http.StatusServiceUnavailable`). The raw error message is retained in the server log only and is not exposed in the public JSON response to prevent information leakage.
- **In `ReadinessHandler`:** If `CheckHealth` returns a non-nil error, the failure is logged, `status.Checks[name]` is set to `"not ready"`, overall `status.Status` is set to `"not ready"`, and the endpoint responds with HTTP 503 (`http.StatusServiceUnavailable`).
- **On success:** If all checkers return `nil`, individual checks are marked `"healthy"` (or `"ready"`), overall status is `"healthy"` (or `"ready"`), and the endpoint responds with HTTP 200 (`http.StatusOK`).

### Compatibility

Public API. Exported types and functions in package `server` are covered by [Versioning and Support](versioning-and-support.md). (Global functions `RegisterHealthCheck`, `RegisterHealthCheckFunc`, and `UnregisterHealthCheck` are deprecated in favor of `(*Server).Health()`).

### Cleanup

Checkers can be unregistered from a registry using `(*HealthRegistry).Unregister(name)` (or deprecated `UnregisterHealthCheck(name)`), which removes the checker from the internal map under write lock. The framework does not invoke any cleanup method on the `HealthChecker` itself.

### Tested by

`server/health_test.go` (and `server/health_registry_test.go`).

---

## 5. Idempotency Stores

Idempotency stores persist HTTP responses associated with idempotency keys to enable safe retries of state-changing operations.

### Interface or Type and File Path

Source: `api/errors/idempotency.go`

```go
// CachedResponse represents a cached response for idempotency
type CachedResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	Timestamp  time.Time
}

// IdempotencyStore is the interface for idempotency key storage
type IdempotencyStore interface {
	Get(key string) (*CachedResponse, error)
	Set(key string, response *CachedResponse, ttl time.Duration) error
	Delete(key string) error
}
```

Built-in implementations in `api/errors/idempotency_stores.go`:
- `InMemoryStore`: In-memory store for testing.
- `DatabaseStore`: Relational database store for production (supports PostgreSQL and SQLite).
- `RedisStore`: Redis client store skeleton.

### Ownership Rules

Created by the application or test harness (e.g., via `NewDatabaseStore(database, table)` or `NewInMemoryStore()`) and passed to the idempotency layer.

The store does not implement `io.Closer`, and the framework does not close it. The creator retains ownership of any backing database connections (`*sql.DB` or `*db.DB`) or cache clients and must close them during application shutdown.

### Concurrency

Must be safe for concurrent use.

**Derivation:** Idempotency stores are accessed during incoming HTTP requests that provide idempotency keys. Because HTTP requests are handled concurrently across multiple goroutines, the store's `Get`, `Set`, and `Delete` methods are called simultaneously. Built-in implementations enforce concurrency safety: `InMemoryStore` synchronizes map operations with a `sync.RWMutex`, and `DatabaseStore` relies on concurrent-safe `*sql.DB` connection pools and atomic SQL statements (`ON CONFLICT DO UPDATE`).

### Error Semantics

- **`Get(key string)`:**
  - Returns `(*CachedResponse, nil)` on a valid cache hit.
  - Returns an error (such as `fmt.Errorf("key not found")` or `sql.ErrNoRows`) if the key does not exist.
  - If the record has expired: `InMemoryStore` returns `fmt.Errorf("key expired")`. `DatabaseStore` deletes the expired record and returns `fmt.Errorf("key expired")` (or returns a wrapped error if deletion fails).
  - If a key was claimed but the response is not yet written (`!payload.Valid` in database): `DatabaseStore` returns `fmt.Errorf("response not completed")`.
- **`Set(key string, response *CachedResponse, ttl time.Duration)`:** Returns `nil` on success. Returns an error if payload serialization fails or database execution fails.
- **`Delete(key string)`:** Returns `nil` on successful removal. Returns an error if database execution fails.

### Compatibility

Public API. Exported types and interfaces in package `api/errors` are covered by [Versioning and Support](versioning-and-support.md).

### Cleanup

`IdempotencyStore` defines no cleanup method. For `DatabaseStore`, an explicit `Cleanup() error` method is provided to prune expired records (`DELETE FROM <table> WHERE expires_at <= NOW()`), but invocation of cleanup is caller-managed; the framework does not run automatic background eviction.

### Tested by

`api/errors/idempotency_test.go` (and `api/errors/idempotency_database_test.go`).

---

## 6. Audit Sinks

Audit sinks write durable, transaction-aware audit event records for security-relevant and business-significant operations.

### Interface or Type and File Path

Source: `audit/sink.go`

```go
// DBTX is implemented by database connections and transactions.
type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// Sink records an event, using tx when supplied to join the caller's transaction.
type Sink interface {
	Record(ctx context.Context, tx DBTX, event Event) error
}
```

Failure policies and supporting types in `audit/sink.go` and `audit/event.go`:

```go
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
```

### Ownership Rules

Created by the application (e.g., `audit.NewSQLSink(db)`) and supplied to `audit.Record` or passed as a dependency to `operation.NewRunner(db, sink)`.

The sink does not implement `io.Closer`, and the framework does not close the sink. The creator maintains ownership of any underlying database connection pool (`*db.DB` or `*sql.DB`) and is responsible for its closure.

### Concurrency

Must be safe for concurrent use.

**Derivation:** `audit.Record` is invoked from HTTP handlers, background workers, and operation runners concurrently across multiple goroutines. When `tx` is non-nil, `tx` is scoped to the specific invocation transaction, but when `tx` is nil or across different concurrent transactions, `Sink.Record` is invoked concurrently on the shared sink instance.

### Error Semantics

When `sink.Record(ctx, tx, event)` returns an error during `audit.Record(...)`:
- **`FailClosed` (the default zero-value policy):** The sink error is returned directly to the caller via `HandleError`. In `operation.Runner`, an audit failure under `FailClosed` rolls back the active transaction and aborts the business operation.
- **`FailOpen`:** The sink error is handled without failing the caller. The failure is logged at `Error` level using the configured logger (`slog.Default()` if nil) with `"correlation_id"` and `"error"` attributes, any configured atomic `Counter` is incremented by 1, and `HandleError` returns `nil`. The business operation proceeds normally.

### Compatibility

Public API. Exported types and functions in package `audit` are covered by [Versioning and Support](versioning-and-support.md).

### Cleanup

None. `Sink` defines no teardown or close method. Resource cleanup is the responsibility of the caller.

### Tested by

`audit/audit_test.go`

---

## 7. Operations

Operations encapsulate authorized, transactional business writes executed atomically with audit records and post-commit effects.

### Interface or Type and File Path

Source: `operation/operation.go`

```go
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
	Fields     []string
}

// AuditDetails supplies optional information attached to the operation event.
type AuditDetails struct {
	Model         string
	ObjectID      string
	ObjectRepr    string
	CorrelationID string
	Source        string
}
```

Runner execution and sentinels in `operation/runner.go`:

```go
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

func Run[In, Out any](ctx context.Context, runner *Runner, op Operation[In, Out], input In) (out Out, err error)
```

### Ownership Rules

`Operation[In, Out]` is declared as a value struct by the application and passed by value to `operation.Run[In, Out](ctx, runner, op, input)`.

The database transaction passed to `RunFunc` is owned exclusively by `Runner`. `RunFunc` receives only `Transaction` (which embeds `audit.DBTX`), purposefully omitting `Commit()` and `Rollback()` methods so that callbacks cannot commit or roll back early. The runner manages the entire transaction boundary (begin, rollback, and commit).

### Concurrency

`Operation` struct values are immutable and passed by value.

However, `RunFunc` callbacks may be invoked concurrently by multiple goroutines calling `operation.Run` simultaneously. While the `Transaction` passed to `RunFunc` is unique to that invocation, `RunFunc` implementations must not mutate shared mutable receiver or package state without synchronization. Additionally, `Actor.HasPermission` and `Effect` callbacks may be executed concurrently across operations.

### Error Semantics

`operation.Run[In, Out]` enforces a strict ordering of error checks and recovery steps:

1. **Configuration validation:** If `runner == nil`, `runner.beginner == nil`, `runner.sink == nil`, `runner.resolver == nil`, or `op.Name == ""` or `op.Run == nil`, `Run` returns `ErrInvalidOperation` wrapped in an error.
2. **Authorization check:** The actor is resolved via `runner.resolver.ResolveActor(ctx)`. If resolving fails, the actor is nil, `op.Permission == ""` (empty permission is always denied), or `actor.HasPermission(ctx, op.Permission)` returns `false` or an error:
   - The runner records an audit event with `audit.OutcomeDenied` in a separate transaction via `runner.recordOutsideTransaction`.
   - Returns `ErrDenied` (or `errors.Join(ErrDenied, auditErr)` if audit recording also fails).
3. **Transaction begin:** Starts a transaction via `runner.beginner.Begin(ctx)`. If begin fails, records an audit event with `audit.OutcomeFailed` outside the transaction and returns the error.
4. **Execution and panic handling:**
   - If `RunFunc` panics: the active transaction is rolled back, a failed audit event (`audit.OutcomeFailed`) is recorded outside the transaction, and the panic is re-panicked.
   - If `RunFunc` returns an error: the active transaction is rolled back via `tx.Rollback()`, a failed audit event (`audit.OutcomeFailed`) is recorded outside the transaction, and the error is returned wrapped as `fmt.Errorf("execute operation: %w", runErr)` (joined with any rollback error).
5. **In-transaction audit:** If `RunFunc` succeeds, the runner writes an audit event with `audit.OutcomeSuccess` inside the transaction using `audit.Record(ctx, runner.sink, tx, ...)`. If audit recording fails under `FailClosed`, the transaction rolls back, a failed audit event is recorded outside the transaction, and the error is returned.
6. **Commit:** The runner calls `tx.Commit()`. If commit fails, a failed audit event is recorded outside the transaction and the commit error is returned.
7. **Post-commit effects:** Only after commit succeeds, `Effect` functions returned by `RunFunc` are executed sequentially. If an `Effect` returns an error, the error is logged via `runner.logger.Error` with operation name, correlation ID, and error details; it does NOT alter the operation's outcome or cause `Run` to return an error.
8. **Sentinels:** Callers can check for `ErrDenied` (authorization failure), `ErrConflict` (optimistic concurrency / state conflict), and `ErrInvalidOperation` (unconfigured runner or missing fields).

### Compatibility

Public API. Exported types, functions, and sentinels in package `operation` are covered by [Versioning and Support](versioning-and-support.md).

### Cleanup

Transaction cleanup is guaranteed by the runner: committed on success, or rolled back on error or panic via deferred handlers. Any non-database resources allocated inside `RunFunc` must be cleaned up by the callback author.

### Tested by

`operation/runner_test.go`
