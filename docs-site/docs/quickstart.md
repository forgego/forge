---
sidebar_position: 2
description: Create a Forge project, define models, generate an API, and open the admin.
image: /social-card.png
---

# Quick Start

This tutorial follows the same project and database workflow exercised by the
consumer journey. You will create a small PostgreSQL-backed shop with customers
and orders, generate its model and API code, apply the schema, and open the
admin screen.

## Before you begin

- Go 1.26 or later ([install Go](https://go.dev/dl/)).
- PostgreSQL running locally, with a database and credentials you can use.
- A C compiler if you choose SQLite instead. This walkthrough uses PostgreSQL.

Install Forge and check that the command is available:

```bash
go install github.com/forgego/forge/cmd/forge@latest
forge version
```

If `forge` is not found, add `$(go env GOPATH)/bin` to your `PATH` and open a
new terminal.

## 1. Create a project

```bash
forge new shop --template simple --database postgres --docker=false
cd shop
```

The flags select the simple project layout, PostgreSQL, and no Docker files.
The project generator writes local secrets to `.env`. Keep that file private.

Set `config/config.yaml`'s database connection values for your local PostgreSQL
instance. Keep the generated server and admin settings unless you need to
change them. The default development address is `localhost:8000`.

## 2. Define the models

Create `app/work/models.go`. The `work` package groups these related models;
the embedded generated types will be added by the next command.

```go title="app/work/models.go"
package work

import "github.com/forgego/forge/schema"

type Customer struct{ CustomerGenerated }

func (Customer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("email", schema.Required(), schema.MaxLength(255), schema.Unique()),
		schema.StringField("name", schema.Required(), schema.MaxLength(255)),
		schema.BoolField("is_active", schema.Default(true)),
	}
}

func (Customer) Meta() schema.Meta {
	return schema.Meta{TableName: "customers", VerboseName: "Customer"}
}

func (Customer) Relations() []schema.Relation { return nil }
func (Customer) Hooks() *schema.ModelHooks   { return nil }

type Order struct{ OrderGenerated }

func (Order) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("order_number", schema.Required(), schema.MaxLength(50), schema.Unique()),
		schema.Int64Field("customer_id", schema.Required()),
		schema.DecimalField("total_amount", schema.Required()),
		schema.StringField("status", schema.Required(), schema.MaxLength(50), schema.Default("pending")),
	}
}

func (Order) Meta() schema.Meta {
	return schema.Meta{TableName: "orders", VerboseName: "Order"}
}

func (Order) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer", schema.OnDelete(schema.CascadePROTECT)),
	}
}

func (Order) Hooks() *schema.ModelHooks { return nil }
```

## 3. Generate code and apply the initial migration

Generate model managers, typed fields, serializers, and viewsets. Then create
and apply the initial migration:

```bash
forge generate --api --models ./app/work --output ./app/work
forge makemigrations initial --auto --models ./app/work
forge migrate up
```

`makemigrations --auto` reads the model definitions and writes the SQL migration
files under `migrations/`. `migrate up` applies pending migrations to the
configured database.

## 4. Register the admin and API routes

API generation creates viewsets and serializers, but the app must connect its
model managers to the database and register its routes. Create
`app/work/routes.go`:

```go
package work

import (
	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/api"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/server"
)

func Register(router *server.Router, database *db.DB) {
	OrderObjects.SetDB(database)
	CustomerObjects.SetDB(database)
	if _, err := admin.Register[Order](&admin.Config[Order]{}); err != nil { panic(err) }
	if _, err := admin.Register[Customer](&admin.Config[Customer]{}); err != nil { panic(err) }
	apiRouter := api.NewRouter("/api/v1")
	apiRouter.Register("orders", NewOrderViewSet())
	apiRouter.Register("customers", NewCustomerViewSet())
	apiRouter.RegisterRoutes(router)
}
```

In `cmd/server/main.go`, add `"shop/app/work"` to the imports (use the module
name shown in your `go.mod` if you chose another project name). Inside the
existing `srv.RegisterRoutes` callback, call `work.Register(router, database)`.
Place it before the scaffold's admin route mounting block. The simple template
already mounts the admin site at `/admin` when admin is enabled.

## 5. Start the server and open admin

From the project root, run:

```bash
forge runserver
```

Open [http://localhost:8000/admin/](http://localhost:8000/admin/) to sign in
and manage customers and orders. The command also accepts `--port` to override
the configured port; for example, `forge runserver --port 8080`.

## Next steps

- [Models & fields](/docs/models)
- [ORM & QuerySets](/docs/orm)
- [Migrations](/docs/migrations)
- [Admin](/docs/admin/overview)
- [REST API](/docs/api/overview)
