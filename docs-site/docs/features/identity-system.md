---
id: identity-system
title: Identity System
sidebar_label: Identity System
---

# Identity System Feature

## Overview

The Identity System provides complete user management and authentication functionality. It includes user models, authentication backends, permission management, session handling, and password management.

## Location

`forge/identity/`

## Status

✅ **Complete** - Production ready

## Core Components

### 1. User Models (`models/`)

**Purpose:** User and related model definitions

**Models:**
- `User` - Main user model with all Django-like fields
- `Session` - User session management
- `Permission` - Permission definitions
- `Group` - User groups
- `Token` - Authentication tokens

**User Model Features:**
- Core identification (ID, Username, Email, Password)
- Personal information (FirstName, LastName, Bio, Website, Location, Avatar)
- Contact information (PhoneNumber, PhoneVerified)
- Preferences (Timezone, Locale, Language)
- Status flags (IsActive, IsStaff, IsSuperuser, IsLocked, EmailVerified)
- Account security (PasswordChangedAt, PasswordExpiresAt, MustChangePassword)
- Account lockout (LockedAt, LockedReason, FailedLoginAttempts)
- Timestamps (CreatedAt, UpdatedAt, LastLogin)

### 2. Repositories (`repository/`)

**Purpose:** Data access layer for identity models

**Repositories:**
- `UserRepository` - User data access
- `SessionRepository` - Session data access
- `TokenRepository` - Token data access
- `PermissionRepository` - Permission data access
- `GroupRepository` - Group data access

**Features:**
- CRUD operations
- Query methods
- Transaction support
- Error handling

### 3. Services (`service/`)

**Purpose:** Business logic layer

**Services:**
- `UserService` - User management
- `AuthService` - Authentication
- `PasswordService` - Password management
- `PermissionService` - Permission management

**UserService:**
- Create, Update, Delete users
- User validation
- User queries
- User activation/deactivation

**AuthService:**
- Login/Logout
- Authentication
- Session management
- Token management

**PasswordService:**
- Password hashing (bcrypt)
- Password validation
- Password reset
- Password change
- Password policy enforcement

**PermissionService:**
- Permission checking
- Permission assignment
- Group management
- Role-based access control

### 4. Authentication Backends (`backends/`)

**Purpose:** Authentication strategy implementations

**Backends:**
- `PasswordBackend` - Username/password authentication
- `TokenBackend` - Token-based authentication
- `BackendRegistry` - Backend registration and management

**Backend Interface:**
```go
type AuthenticationBackend interface {
    Authenticate(ctx context.Context, credentials map[string]string) (*models.User, error)
    GetUser(ctx context.Context, identifier string) (*models.User, error)
    Supports(credentialType string) bool
    Name() string
}
```

**Features:**
- Multiple authentication methods
- Backend chaining
- Custom backend support
- Strategy pattern

### 5. HTTP Handlers (`handlers/`)

**Purpose:** HTTP request handlers

**Handlers:**
- `UserHandler` - User CRUD operations
- `AuthHandler` - Authentication endpoints

**Endpoints:**
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `GET /api/auth/me` - Current user
- `GET /api/users` - List users
- `POST /api/users` - Create user
- `GET /api/users/{id}` - Get user
- `PUT /api/users/{id}` - Update user
- `DELETE /api/users/{id}` - Delete user

### 6. Middleware (`middleware/`)

**Purpose:** Authentication and authorization middleware

**Middleware:**
- `AuthenticationMiddleware` - Require authentication
- `AuthorizationMiddleware` - Require permissions
- `SuperuserMiddleware` - Require superuser

**Features:**
- Request authentication
- User context injection
- Permission checking
- Error handling

### 7. Serializers (`serializers/`)

**Purpose:** API serialization

**Serializers:**
- User serialization
- Session serialization
- Token serialization
- Permission serialization

### 8. Configuration (`config/`)

**Purpose:** Identity system configuration

**Configuration:**
- Password policy
- Session settings
- Token settings
- Authentication settings

**Password Policy:**
- Minimum length
- Complexity requirements
- Expiration settings
- History requirements

### 9. Factory (`factory.go`)

**Purpose:** System setup and initialization

**Features:**
- Complete system setup
- Dependency injection
- Default configuration
- Service initialization

## Features

### ✅ Complete Features

1. **User Management** - Complete user CRUD operations
2. **Authentication** - Multiple authentication backends
3. **Password Management** - Secure password handling
4. **Session Management** - User session handling
5. **Permission System** - Permission and group management
6. **Token Authentication** - Token-based authentication
7. **Account Security** - Password policies, lockout, expiration
8. **HTTP Handlers** - REST API endpoints
9. **Middleware** - Authentication and authorization middleware
10. **Serializers** - API serialization
11. **Configuration** - Flexible configuration system

## Usage Examples

### System Setup

```go
import (
    "github.com/forgego/forge/db"
    "github.com/forgego/forge/identity"
    "github.com/forgego/forge/identity/config"
)

// Create database connection
db, err := db.NewDB(dsn)
if err != nil {
    log.Fatal(err)
}

// Create identity config
identityConfig := config.DefaultIdentityConfig()
identityConfig.PasswordPolicy.MinLength = 8
identityConfig.PasswordPolicy.RequireUppercase = true
identityConfig.PasswordPolicy.RequireLowercase = true
identityConfig.PasswordPolicy.RequireNumbers = true

// Setup identity system
identitySystem, err := identity.SetupIdentitySystem(db, identityConfig)
if err != nil {
    log.Fatal(err)
}

// Register routes
identitySystem.RegisterRoutes(router)
```

### User Registration

```go
user, err := identitySystem.UserService.Create(ctx, &identity.CreateUserRequest{
    Username: "john",
    Email:    "john@example.com",
    Password: "securepassword123",
    FirstName: "John",
    LastName:  "Doe",
})
if err != nil {
    log.Fatal(err)
}
```

### Authentication

```go
// Login
user, session, err := identitySystem.AuthService.Login(ctx, &identity.LoginRequest{
    Username: "john",
    Password: "securepassword123",
})
if err != nil {
    log.Fatal(err)
}

// Set session cookie
http.SetCookie(w, &http.Cookie{
    Name:     "session",
    Value:    session.Key,
    HttpOnly: true,
    Secure:   true,
})
```

### Password Management

```go
// Change password
err := identitySystem.PasswordService.ChangePassword(ctx, userID, &identity.ChangePasswordRequest{
    OldPassword: "oldpassword",
    NewPassword: "newpassword123",
})
if err != nil {
    log.Fatal(err)
}

// Reset password
token, err := identitySystem.PasswordService.RequestPasswordReset(ctx, "john@example.com")
if err != nil {
    log.Fatal(err)
}

// Use token to reset
err = identitySystem.PasswordService.ResetPassword(ctx, token, "newpassword123")
```

### Permission Checking

```go
// Check permission
hasPerm, err := identitySystem.PermissionService.HasPermission(ctx, userID, "can_edit_post")
if err != nil {
    log.Fatal(err)
}

// Assign permission
err = identitySystem.PermissionService.AssignPermission(ctx, userID, "can_edit_post")
if err != nil {
    log.Fatal(err)
}

// Check group membership
isMember, err := identitySystem.PermissionService.IsInGroup(ctx, userID, "editors")
if err != nil {
    log.Fatal(err)
}
```

### Middleware Usage

```go
import "github.com/forgego/forge/identity/middleware"

// Require authentication
authMiddleware := middleware.NewAuthenticationMiddleware(
    identitySystem.BackendRegistry,
    identitySystem.SessionRepo,
    identitySystem.UserRepo,
)

router.Use(authMiddleware.RequireAuth)

// Require specific permission
router.Get("/posts/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
    user := middleware.GetUser(r)
    hasPerm, _ := identitySystem.PermissionService.HasPermission(
        r.Context(),
        user.ID,
        "can_edit_post",
    )
    if !hasPerm {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    // ... edit post
})
```

### Custom Backend

```go
type OAuthBackend struct {
    // ... fields
}

func (b *OAuthBackend) Authenticate(ctx context.Context, credentials map[string]string) (*models.User, error) {
    token := credentials["oauth_token"]
    // ... OAuth authentication
    return user, nil
}

func (b *OAuthBackend) Supports(credentialType string) bool {
    return credentialType == "oauth"
}

func (b *OAuthBackend) Name() string {
    return "oauth"
}

// Register backend
identitySystem.BackendRegistry.Register(&OAuthBackend{})
```

## Integration Points

### Admin System
- User management in admin
- Permission integration
- User editing

### API Framework
- Authentication middleware
- Permission classes
- User serializers

### HTTP Server
- Authentication middleware
- Session management
- Request context

### Database Layer
- User data storage
- Session storage
- Transaction support

## Security Features

### Password Security
- bcrypt hashing
- Password policy enforcement
- Password history
- Password expiration

### Account Security
- Account lockout
- Failed login tracking
- Email verification
- Phone verification

### Session Security
- Secure session storage
- Session expiration
- Session invalidation
- CSRF protection

### Token Security
- Secure token generation
- Token expiration
- Token revocation
- Token validation

## Best Practices

1. **Password Policy** - Enforce strong passwords
2. **Session Security** - Use secure session storage
3. **Token Security** - Use secure token generation
4. **Permission Checks** - Always check permissions
5. **Error Handling** - Don't leak sensitive information
6. **Logging** - Log security events
7. **Rate Limiting** - Limit authentication attempts
8. **HTTPS** - Always use HTTPS in production
9. **Input Validation** - Validate all user input
10. **Output Sanitization** - Sanitize all output

## Future Enhancements

- [ ] OAuth2 support
- [ ] SAML support
- [ ] Two-factor authentication
- [ ] Social authentication
- [ ] Account recovery
- [ ] Email verification flow
- [ ] Phone verification flow
- [ ] Advanced permission system
- [ ] Audit logging
- [ ] User activity tracking
