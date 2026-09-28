---
sidebar_position: 2
description: Install Forge, create a project, and apply your first migration.
image: /social-card.png
---

# Quick Start

This guide installs the `forge` command, creates a project, adds an app with a
model, and applies its first migration. It uses SQLite, so you do not need a
database server.

## Prerequisites

- Go 1.26 or later ([download](https://go.dev/dl/))
- A C compiler, which the SQLite driver needs (`gcc` on Linux, Xcode command
  line tools on macOS)

## Step 1: Install the CLI

```bash
go install github.com/forgego/forge/cmd/forge@latest
forge version
```

If your shell cannot find `forge`, add Go's bin directory to your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Step 2: Create a project

```bash
forge new myapp --database sqlite --template simple --docker=false
cd myapp
```

Leave out the flags to choose interactively. The project looks like this:

```
myapp/
├── app/                # One directory per app
├── cmd/server/main.go  # Server entry point
├── config/config.yaml  # Settings (secrets go in .env)
├── migrations/         # SQL migrations
├── static/
├── templates/
├── .env                # Generated secrets, not committed
└── go.mod
```

## Step 3: Add an app

An app groups related models with their admin and API code:

```bash
forge add app blog --example
```

This creates `app/blog/models.go` with an `Example` model, plus `admin.go` and
`api.go`. Replace the example with your own model, for instance:

```go title="app/blog/models.go"
package blog

import "github.com/forgego/forge/schema"

type Article struct {
	schema.BaseSchema
}

func (Article) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required(), schema.MaxLength(200)),
		schema.TextField("body", schema.Required()),
		schema.BoolField("published", schema.Default(false)),
	}
}

func (Article) Meta() schema.Meta {
	return schema.Meta{TableName: "articles", VerboseName: "Article"}
}

func (Article) Relations() []schema.Relation { return nil }

func (Article) Hooks() *schema.ModelHooks { return nil }
```

## Step 4: Generate code

```bash
forge generate --models ./app/blog --output ./app/blog
```

This writes `app/blog/gen.go` with a manager (`ArticleObjects`) and typed field
accessors (`ArticleFieldsInstance`) that you use to query the model.

## Step 5: Create and apply migrations

```bash
forge makemigrations initial --auto --models ./app/blog
forge migrate up
forge migrate status
```

`makemigrations --auto` compares the models with the existing migrations and
writes `migrations/000001_initial.up.sql` and its `.down.sql` counterpart.

## Step 6: Query your model

With the generated code, queries are typed:

```go
f := blog.ArticleFieldsInstance

qs, err := blog.ArticleObjects.Filter(f.Published.Eq(true))
if err != nil {
	return err
}
articles, err := qs.OrderBy(f.Id.Desc()).Limit(10).All(ctx)
```

See [ORM & QuerySets](/docs/orm) for filters, relations and aggregates.

## Running the server

```bash
go run ./cmd/server
```

The server listens on the host and port in `config/config.yaml`
(`localhost:8000` by default) and mounts the admin at `/admin`. To serve an
app's REST API, see [Generating a REST API](/docs/models#generating-a-rest-api).

The [ecommerce example](https://github.com/forgego/forge/tree/master/examples/ecommerce)
is a complete Forge app with the admin and REST API wired up. To run it with
Docker and PostgreSQL:

```bash
git clone https://github.com/forgego/forge.git
cd forge/examples/ecommerce
docker compose up --build
```

Then open [localhost:8020/admin](http://localhost:8020/admin/) and sign in as `admin` / `admin123`.

## Common commands

```bash
forge new <name>               # Create a project
forge add app <name>           # Add an app to the project
forge generate                 # Generate managers and typed fields
forge makemigrations <name>    # Create a migration (--auto writes the SQL)
forge migrate up               # Apply migrations
forge migrate status           # Show applied and pending migrations
forge migrate rollback         # Roll back the last migration
forge createsuperuser          # Create an admin user
forge --help                   # List every command
```

## Next steps

- [Models & fields](/docs/models)
- [ORM & QuerySets](/docs/orm)
- [Migrations](/docs/migrations)
- [Admin](/docs/admin/overview)
- [REST API](/docs/api/overview)
