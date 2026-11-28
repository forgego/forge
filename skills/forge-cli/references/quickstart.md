# Forge Quickstart Flow

## 1. Create Project and App

```bash
forge new myapp --database sqlite --template simple --docker=false
cd myapp
forge add app blog --example
```

## 2. Define a Model (`app/blog/models.go`)

Embed `<Model>Generated` (emitted by `forge generate`) so your struct inherits all schema fields and tags:

```go
package blog

import "github.com/forgego/forge/schema"

type Article struct {
	ArticleGenerated
}

func (Article) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required(), schema.MaxLength(200)),
		schema.TextField("body", schema.Required()),
		schema.BoolField("published", schema.Default(false)),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (Article) Meta() schema.Meta {
	return schema.Meta{TableName: "articles", VerboseName: "Article"}
}

func (Article) Relations() []schema.Relation { return nil }

func (Article) Hooks() *schema.ModelHooks { return nil }
```

## 3. Generate Code & Migrations

```bash
forge generate --models ./app/blog --output ./app/blog
forge makemigrations initial --auto --models ./app/blog
forge migrate up
forge migrate status
```

## 4. Query with Generated Types

`Filter` on `<Model>Objects` returns `(orm.QuerySet[T], error)`:

```go
f := blog.ArticleFieldsInstance

qs, err := blog.ArticleObjects.Filter(orm.And(
	f.Published.Eq(true),
	f.Id.Gt(0),
))
if err != nil {
	return err
}
articles, err := qs.OrderBy(f.Id.Desc()).Limit(10).All(ctx)
```

## 5. Wire Server & Database Stores (`cmd/server/main.go`)

Pass `server.WithDatabase(database)` to `server.NewServer` and call `adminSite.UseDatabaseStores(ctx)` so sessions, throttling counters, tokens, saved views, and change history live in the database:

```go
blog.ArticleObjects.SetDB(database)

adminSite := admin.DefaultSite
adminSite.SetDB(database)
if err := adminSite.UseDatabaseStores(ctx); err != nil {
	log.Fatal(err)
}

srv, err := server.NewServer(cfg, settings, logger, server.WithDatabase(database))
if err != nil {
	log.Fatal(err)
}

srv.RegisterRoutes(func(router *server.Router) {
	if settings.Admin.Enabled {
		router.Mount(settings.Admin.Path, adminSite.Handler())
	}
})
```

Start the app:

```bash
forge runserver
# or: go run ./cmd/server
```
