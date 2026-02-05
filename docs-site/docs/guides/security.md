---
sidebar_position: 6
---

# Security Guide

forge includes built-in security features to protect your application from common vulnerabilities.

## Overview

forge provides protection against:

- **SQL Injection** - Parameterized queries and proper escaping
- **XSS (Cross-Site Scripting)** - Output sanitization
- **CSRF (Cross-Site Request Forgery)** - CSRF tokens
- **Session Hijacking** - Secure session management
- **Password Attacks** - Secure password hashing

## SQL Injection Protection

forge uses parameterized queries by default, preventing SQL injection:

```go
// Safe: Uses parameterized queries
users, err := User.Objects.
    Filter(User.Fields.Username.Equals(username)).
    All(ctx)

// The SQL builder automatically escapes and parameterizes
// SQL: SELECT * FROM users WHERE username = $1
// Args: [username]
```

### Never Use String Concatenation

```go
// ❌ BAD: Vulnerable to SQL injection
query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", username)

// ✅ GOOD: Use QuerySet API
users, err := User.Objects.
    Filter(User.Fields.Username.Equals(username)).
    All(ctx)
```

## XSS Protection

### Automatic Escaping

When rendering templates, forge automatically escapes output:

```go
// In templates, output is automatically escaped
{{ .User.Username }}  // Safe: HTML entities escaped
```

### Manual Escaping

For custom output, use the HTML escape function:

```go
import "html"

escaped := html.EscapeString(userInput)
```

### Safe HTML

If you need to output HTML, mark it as safe:

```go
// Only if you trust the content
safeHTML := template.HTML(trustedContent)
```

## CSRF Protection

### Enable CSRF Protection

CSRF protection is enabled by default. Configure it in `config.yaml`:

```yaml
security:
  csrf:
    enabled: true
    secret_key: "your-secret-key-change-in-production"
    cookie_name: "csrf_token"
```

### Using CSRF Tokens

In forms, include the CSRF token:

```html
<form method="POST" action="/posts/create">
    <input type="hidden" name="csrf_token" value="{{ .CSRFToken }}">
    <!-- form fields -->
</form>
```

### API Requests

For API requests, include CSRF token in header:

```javascript
fetch('/api/posts/', {
    method: 'POST',
    headers: {
        'X-CSRF-Token': getCSRFToken(),
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
});
```

### Exempt Routes

Exempt certain routes from CSRF protection:

```go
router.Use(csrf.Protect(
    csrf.Secure(true),
    csrf.Exempt("/api/public/"),
))
```

## Session Security

### Secure Session Configuration

Configure secure sessions:

```yaml
security:
  session:
    secret: "your-secret-key-change-in-production"
    secure: true      # HTTPS only
    http_only: true   # Not accessible via JavaScript
    same_site: "strict"
    max_age: 86400    # 24 hours
```

### Session Management

```go
// Create session
session, err := sessionManager.Create(r.Context(), userID)

// Get session
userID, err := sessionManager.Get(r.Context())

// Destroy session
err := sessionManager.Destroy(r.Context())
```

## Password Security

### Password Hashing

Always hash passwords using bcrypt:

```go
import "github.com/forgego/forge/pkg/users/backends"

// Hash password
hashed, err := backends.HashPassword(password)

// Verify password
err := backends.VerifyPassword(hashed, password)
```

### Password Validation

Enforce strong passwords:

```go
func (User) Hooks() *schema.ModelHooks {
    return &schema.ModelHooks{
        BeforeCreate: func(ctx context.Context, instance interface{}) error {
            user := instance.(*User)
            
            // Validate password strength
            if len(user.Password) < 8 {
                return errors.New("password must be at least 8 characters")
            }
            
            // Hash password
            hashed, err := backends.HashPassword(user.Password)
            if err != nil {
                return err
            }
            user.Password = hashed
            
            return nil
        },
    }
}
```

## Authentication

### Secure Authentication

Use secure authentication practices:

```go
// Rate limit login attempts
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    // Check rate limit
    if rateLimitExceeded(r) {
        http.Error(w, "Too many attempts", http.StatusTooManyRequests)
        return
    }
    
    // Authenticate
    user, err := authenticate(r)
    if err != nil {
        // Don't reveal if user exists
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    
    // Create secure session
    session, err := createSession(user)
    // ...
}
```

## Input Validation

### Validate All Input

Always validate user input:

```go
// Use schema validators
schema.String("email").
    Required().
    Validators(schema.EmailValidator()).
    Build()

// Custom validation in hooks
func (User) Hooks() *schema.ModelHooks {
    return &schema.ModelHooks{
        Clean: func(ctx context.Context, instance interface{}) error {
            user := instance.(*User)
            
            // Validate email format
            if !isValidEmail(user.Email) {
                return errors.New("invalid email format")
            }
            
            return nil
        },
    }
}
```

## HTTPS

### Force HTTPS in Production

```yaml
server:
  tls:
    enabled: true
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
```

### Redirect HTTP to HTTPS

```go
router.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
            http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
            return
        }
        next.ServeHTTP(w, r)
    })
})
```

## Security Headers

### Add Security Headers

```go
router.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000")
        next.ServeHTTP(w, r)
    })
})
```

## File Upload Security

### Validate File Uploads

```go
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
    // Limit file size
    r.ParseMultipartForm(10 << 20) // 10 MB
    
    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()
    
    // Validate file type
    if !isAllowedFileType(header.Filename) {
        http.Error(w, "Invalid file type", http.StatusBadRequest)
        return
    }
    
    // Validate file size
    if header.Size > maxFileSize {
        http.Error(w, "File too large", http.StatusBadRequest)
        return
    }
    
    // Save file securely
    // ...
}
```

## Best Practices

1. **Always Use Parameterized Queries** - Never concatenate SQL
2. **Validate All Input** - Don't trust user input
3. **Hash Passwords** - Never store plain text passwords
4. **Use HTTPS** - Encrypt data in transit
5. **Keep Dependencies Updated** - Update packages regularly
6. **Limit Rate** - Prevent brute force attacks
7. **Log Security Events** - Monitor for suspicious activity
8. **Regular Security Audits** - Review code for vulnerabilities

## Security Checklist

- [ ] SQL injection protection (parameterized queries)
- [ ] XSS protection (output escaping)
- [ ] CSRF protection (tokens in forms)
- [ ] Secure sessions (httpOnly, secure, sameSite)
- [ ] Password hashing (bcrypt)
- [ ] Input validation (all user input)
- [ ] HTTPS enabled (production)
- [ ] Security headers (X-Frame-Options, etc.)
- [ ] Rate limiting (login attempts)
- [ ] File upload validation (type, size)
- [ ] Error messages (don't reveal sensitive info)
- [ ] Dependency updates (regular updates)

## Next Steps

- [Deployment Guide](/docs/guides/deployment) - Secure deployment practices
- [Admin Guide](/docs/guides/admin) - Secure admin interface
- [REST API Guide](/docs/guides/rest-api) - Secure API endpoints
