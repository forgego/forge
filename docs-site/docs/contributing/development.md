---
sidebar_position: 1
---

# Development Guide

Guide for contributing to forge framework development.

## Development Setup

### Prerequisites

- Go 1.25 or later
- PostgreSQL 12 or later
- Git

### Setup

```bash
# Clone repository
git clone https://github.com/forgego/forge.git
cd forge

# Install dependencies
go mod download

# Run tests
go test ./...

# Build CLI
go build -o forge ./cli/cmd
```

## Project Structure

```
forge/
├── admin/                  # Admin interface
├── api/                    # REST API framework
├── cli/                    # CLI tool
├── codegen/                # Code generation
├── config/                 # Configuration
├── db/                     # Database layer + migrations
├── identity/               # Auth + user management
├── log/                    # Logging
├── orm/                    # ORM + QuerySet
├── registry/               # Plugin/model registry
├── schema/                 # Schema definitions
├── server/                 # HTTP server + router
├── validate/               # Validation
└── ...

examples/
├── ecommerce/              # E-commerce example

docs-site/                  # Documentation site
tests/                      # Integration/e2e tests
```

## Development Workflow

### 1. Schema Changes

1. Modify schema definition
2. Run `forge generate`
3. Review generated code
4. Run tests
5. Update documentation if needed

### 2. Adding Features

1. Create feature branch
2. Implement feature
3. Write tests
4. Update documentation
5. Submit PR

### 3. Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./pkg/query
```

## Code Style

### Naming Conventions

- **Packages:** lowercase, single word
- **Types:** PascalCase
- **Functions:** PascalCase (exported), camelCase (internal)
- **Variables:** camelCase
- **Constants:** PascalCase or UPPER_SNAKE_CASE

### File Organization

- One type per file (for large types)
- Related types in same package
- Keep files under 500 lines when possible

### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("context: %w", err)
}

// Avoid
if err != nil {
    return err  // Loses context
}
```

### Documentation

- Export all public APIs
- Use godoc comments
- Include examples in comments
- Document complex logic

## Testing Guidelines

### Unit Tests

```go
func TestFieldExpr(t *testing.T) {
    field := query.NewFieldExpr[string]("username")

    q := field.Equals("john")
    assert.NotNil(t, q)

    sql, args := q.ToSQL()
    assert.Equal(t, "username = $1", sql)
    assert.Equal(t, []interface{}{"john"}, args)
}
```

### Integration Tests

```go
func TestUserCRUD(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    ctx := context.Background()

    // Create
    user := &User{Username: "test"}
    err := User.Objects.Create(ctx, user)
    require.NoError(t, err)

    // Read
    found, err := User.Objects.Get(ctx, user.ID)
    require.NoError(t, err)
    assert.Equal(t, "test", found.Username)

    // Update
    found.Username = "updated"
    err = User.Objects.Update(ctx, found)
    require.NoError(t, err)

    // Delete
    err = User.Objects.Delete(ctx, found)
    require.NoError(t, err)
}
```

## Contributing

### Pull Request Process

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Make changes
4. Write/update tests
5. Update documentation
6. Commit changes (`git commit -m 'feat: Add amazing feature'`)
7. Push to branch (`git push origin feature/amazing-feature`)
8. Submit PR with description

### Commit Messages

Follow conventional commits:

```
feat: Add UUID field type support
fix: Fix QuerySet.Count() implementation
docs: Update API reference
refactor: Simplify AST parser
test: Add tests for FieldExpr
chore: Update dependencies
```

### Code Review

- All PRs require review
- Tests must pass
- Documentation must be updated
- No breaking changes without discussion

## Architecture Decisions

### Why SQL Builder?

- Type-safe SQL generation from Go code
- Proper identifier escaping to prevent SQL injection
- Parameter binding for all values
- Works with standard database/sql
- SQL-first approach

### Why chi?

- Lightweight, composable
- Standard library compatible
- Framework-friendly

### Why Code Generation?

- Type safety
- Performance
- IDE support

## Performance Guidelines

### Database

- Use indexes for frequently queried fields
- Use SelectRelated for JOINs
- Use PrefetchRelated for separate queries
- Batch operations when possible

### Code Generation

- Keep templates simple
- Avoid complex logic in templates
- Use helper functions

### Memory

- Reuse QuerySet instances when possible
- Close rows properly
- Use connection pooling

## Security Guidelines

### Input Validation

- Always validate user input
- Use parameterized queries
- Sanitize output

### Authentication

- Hash passwords (bcrypt)
- Use secure sessions
- Implement CSRF protection

### Error Handling

- Don't expose internal errors
- Log errors securely
- Return user-friendly messages

## Debugging

### Enable Debug Logging

```go
logger, _ := logging.NewLogger(true)  // development mode
```

### Query Logging

```go
// Log all queries
db.EnableQueryLogging(true)
```

### Profiling

```go
import _ "net/http/pprof"

// In main.go
go http.ListenAndServe(":6060", nil)
```

## Release Process

### Versioning

- Follow semantic versioning
- Major: Breaking changes
- Minor: New features
- Patch: Bug fixes

### Release Checklist

- [ ] All tests passing
- [ ] Documentation updated
- [ ] Changelog updated
- [ ] Version bumped
- [ ] Tagged in git
- [ ] Release notes written

## Getting Help

- **Issues:** [GitHub Issues](https://github.com/forgego/forge/issues)
- **Discussions:** [GitHub Discussions](https://github.com/forgego/forge/discussions)
- **Documentation:** This site

## See Also

- [Architecture](/docs/contributing/architecture) - Framework architecture
- [Implementation status](/docs/status/implementation) - Current build status
- Planning updates live in GitHub issues and discussions; legacy docs like
  `PRODUCTION_READY.md` and the admin UI plan have been removed.
