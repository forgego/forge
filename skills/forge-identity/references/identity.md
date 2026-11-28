# Identity Architecture (`identity/`)

## Components Created by `identity.SetupIdentitySystem(database, cfg)`

- **Repositories (`identity/repository/`)**: `UserRepo`, `SessionRepo`, `TokenRepo`, `PermissionRepo`, `GroupRepo`
- **Services (`identity/service/`)**: `UserService`, `AuthService`, `PasswordService`, `PermissionService`
- **Backends (`identity/backends/`)**: `PasswordBackend`, `TokenBackend` registered on `BackendRegistry`
- **HTTP Handlers & Middleware (`identity/handlers/`, `identity/middleware/`)**: Login/logout, token auth, password reset (hashed-token flow + notifier), and `RequirePermission`

## Integration with Admin & API

- **Admin**: Override default env-var login (`FORGE_ADMIN_USERNAME`/`FORGE_ADMIN_PASSWORD`) via `adminSite.SetLoginAuthenticator(...)`.
- **REST API**: Configure `ViewSetConfig.Authentication` (`api/authentication`) and `ViewSetConfig.Permissions` (`api/permissions`).
