# 🔥 Forge Framework

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=for-the-badge)](LICENSE)
[![Tests](https://img.shields.io/github/actions/workflow/status/hamidrabedi/foreit/test.yml?branch=main&label=Tests&style=for-the-badge)](https://github.com/hamidrabedi/foreit/actions)
[![Security](https://img.shields.io/github/actions/workflow/status/hamidrabedi/foreit/security.yml?branch=main&label=Security&style=for-the-badge)](https://github.com/hamidrabedi/foreit/security)
[![Documentation](https://img.shields.io/badge/docs-online-success?style=for-the-badge)](https://hamidrabedi.github.io/foreit/)

**A Django-like Go framework with full type safety, code generation, and extensibility.**

[Documentation](https://hamidrabedi.github.io/foreit/) • [Examples](examples/) • [API Reference](https://hamidrabedi.github.io/foreit/docs/api-reference/schema) • [Contributing](CONTRIBUTING.md)

</div>

---

## Features

- **Type-Safe ORM**: Full Django ORM features with compile-time type checking
- **Code Generation**: AST-based code generation for models, managers, and querysets
- **Auto-Generated Admin**: Django-style admin interface auto-generated from model registry
- **Dual API**: Type-safe primary API + dynamic secondary API for runtime flexibility
- **Migration System**: Built-in migration system with golang-migrate
- **Security**: Built-in CSRF, XSS, and SQL injection protection
- **Extensible**: Everything is extendable/overridable via plugins
- **SQL Builder**: Type-safe SQL generation with proper escaping and parameter binding

## Quick Start

### 1. Create a New Project

```bash
forge new myapp
cd myapp
```

### 2. Configure Database

Edit `config/config.yaml` with your PostgreSQL credentials.

### 3. Define Models

Edit `models/example.go` or create new model files:

```go
package models

import "github.com/forgego/forge/internal/schema"

type Post struct {
	schema.BaseSchema
}

func (Post) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("title").Required().MaxLength(200).Build(),
		schema.String("content").Required().Build(),
	}
}

func (Post) Meta() schema.Meta {
	return schema.Meta{
		TableName: "posts",
		VerboseName: "Post",
	}
}

func (Post) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Post) Hooks() *schema.ModelHooks {
	return nil
}
```

### 4. Generate Code

```bash
forge generate
```

### 5. Run Migrations

```bash
forge migrate
```

### 6. Start Server

```bash
forge runserver
```

Visit `http://localhost:8000/admin/` for the auto-generated admin interface!

For detailed instructions, see the [Getting Started Guide](docs/GETTING_STARTED.md).

## Documentation

- **[Getting Started](docs/GETTING_STARTED.md)** - 10-minute quick start guide ⭐
- **[REST API Guide](docs/REST_API.md)** - Build APIs for React/Vue frontends
- **[HTMX Patterns](docs/HTMX_PATTERNS.md)** - Admin interface patterns and best practices
- **[UI Strategy](docs/UI_STRATEGY.md)** - UI technology choices and approach
- **[Architecture](docs/ARCHITECTURE.md)** - Framework architecture
- **[Schema Reference](docs/SCHEMA_REFERENCE.md)** - Schema definition guide
- **[API Reference](docs/API_REFERENCE.md)** - API documentation
- **[Usage Guide](docs/USAGE_GUIDE.md)** - Step-by-step tutorials
- **[Features](docs/FEATURES.md)** - Current and planned features
- **[Roadmap](docs/ROADMAP.md)** - Development roadmap
- **[Development](docs/DEVELOPMENT.md)** - Contributing guide

## Example

See `examples/blog/` for a complete example application.

## Technology Stack

- **Go 1.24+**
- **chi/v5** - HTTP router
- **database/sql** - Standard SQL interface
- **golang-migrate** - Migrations
- **zap** - Logging
- **viper** - Configuration
- **testify** - Testing

## 🔒 Security

Forge takes security seriously with:

- ✅ **Zero Known Vulnerabilities** - All dependencies scanned and updated
- 🛡️ **10+ Security Tools** - Automated scanning with govulncheck, CodeQL, Trivy, Snyk, gosec, and more
- 🔐 **Secret Detection** - TruffleHog prevents credential leaks
- 📊 **Daily Security Scans** - Continuous monitoring
- 🚨 **PR Security Gates** - Automatic security review on every pull request
- 📋 **Security Policy** - Responsible disclosure process ([SECURITY.md](SECURITY.md))

See our [Security Audit Report](SECURITY_AUDIT.md) for details.

## 📊 Status

**Current:** Production Ready - Secure, tested, and documented! 🎉

**Implementation Status:**
- ✅ Schema system - Complete
- ✅ Code generation - Complete
- ✅ Type-safe ORM - Complete (All, Get, First, Last, Count, Exists)
- ✅ Q Objects - Complete (And, Or, Not)
- ✅ Manager CRUD - Complete (Create, Update, Delete with hooks)
- ✅ Admin Interface - Complete (List view, Create/Edit forms)
- ✅ REST API - Complete (Full CRUD with pagination, filtering, ordering)
- ✅ Plugin System - Complete
- ✅ CLI - Complete (new, generate, migrate, runserver)

See [Features](docs/FEATURES.md) and [Roadmap](docs/ROADMAP.md) for details.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs-site/docs/contributing/development.md) for:

- Code of Conduct
- Development setup
- Pull request process
- Testing guidelines
- Code style guidelines

### Quick Start for Contributors

```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/foreit.git
cd foreit

# Install Go 1.25+
go version  # Should be 1.25+

# Build and test
cd forge
go mod download
go build ./...
go test ./...

# See tests/ for integration tests
cd ../tests
go test ./integration/...
```

## 📈 Project Stats

- **78 Test Functions** - Comprehensive test coverage
- **60+ Documentation Pages** - Complete guides and references
- **679 Tracked Files** - Clean, organized codebase
- **10+ Security Scans** - Enterprise-grade security
- **Zero Vulnerabilities** - All dependencies secure

## 🌟 Why Forge?

- **🚀 Productivity**: Build full-stack apps 10x faster than traditional Go
- **✨ Type Safety**: Full type safety with Go generics, catch errors at compile time
- **🎨 Modern UI**: Beautiful React admin interface with TanStack Router
- **🔐 Secure**: Built-in security features and continuous scanning
- **📚 Documented**: Comprehensive documentation with examples
- **🧪 Tested**: 78 test functions with real PostgreSQL integration
- **🔧 Extensible**: Plugin system for custom functionality

## 📞 Support

- **Documentation**: [https://hamidrabedi.github.io/foreit/](https://hamidrabedi.github.io/foreit/)
- **Issues**: [GitHub Issues](https://github.com/hamidrabedi/foreit/issues)
- **Security**: See [SECURITY.md](SECURITY.md)
- **Discussions**: [GitHub Discussions](https://github.com/hamidrabedi/foreit/discussions)

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Built with ❤️ using:
- [Go](https://go.dev/) - Programming language
- [chi](https://github.com/go-chi/chi) - HTTP router
- [React](https://react.dev/) - Frontend framework
- [TanStack](https://tanstack.com/) - Query and Router
- [Radix UI](https://www.radix-ui.com/) - UI components
- [Docusaurus](https://docusaurus.io/) - Documentation

---

<div align="center">

**[⭐ Star us on GitHub](https://github.com/hamidrabedi/foreit) • [📖 Read the Docs](https://hamidrabedi.github.io/foreit/) • [🐛 Report a Bug](https://github.com/hamidrabedi/foreit/issues)**

Made with ❤️ by the Forge community

</div>
