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

- Type-safe ORM and migrations
- Code generation for models and admin
- Extensible plugin system

## Quick Start

### 1. Install the CLI

Install the `forge` command without cloning the repo:

```bash
go install github.com/forgego/forge/cli/cmd@latest
```

Make sure `$GOBIN` (or `$GOPATH/bin`) is on your `PATH`, then verify:

```bash
forge --help
```

### 2. Create a New Project

```bash
forge new myapp
cd myapp
```

### 3. Configure Database

Edit `config/config.yaml` with your PostgreSQL credentials.

### 4. Define Models

Edit `models/example.go` or create new model files:

```go
package models

import "github.com/forgego/forge/schema"

type Post struct {
	schema.BaseSchema
}

func (Post) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required(), schema.MaxLength(200)),
		schema.TextField("content", schema.Required()),
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

### 5. Generate Code

```bash
forge generate
```

### 6. Run Migrations

```bash
forge makemigrations
forge migrate up
```

### 7. Start Server

```bash
forge runserver
```

Visit `http://localhost:8000/admin/` for the auto-generated admin interface!

For detailed instructions, see the [Getting Started Guide](docs-site/docs/getting-started/quickstart.md).

## Install as a Library

If you want to use Forge packages directly in an existing Go project:

```bash
go get github.com/forgego/forge@latest
```

Then import the packages you need, for example:

```go
import "github.com/forgego/forge/schema"
```

## CLI Usage (Forge Commands)

Once installed, the `forge` binary is your entry point:

```bash
forge new myapp
forge generate
forge makemigrations
forge migrate up
forge runserver
```

Run `forge --help` to see all commands and flags.

## Documentation & Support

- Docs: https://hamidrabedi.github.io/foreit/
- Issues: https://github.com/hamidrabedi/foreit/issues
- Security policy: [SECURITY.md](SECURITY.md)

## License

MIT License - see [LICENSE](LICENSE) file for details.
