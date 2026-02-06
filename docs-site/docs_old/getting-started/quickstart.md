---
sidebar_position: 3
description: Fast path to a working forge app.
keywords:
  - forge quickstart
  - forge tutorial
  - get started with forge
image: /forge-social-card.svg
---

# Quick Start

This is the shortest path to a working forge app.

## 1) Install the CLI

```bash
go install github.com/forgego/forge/cli/cmd@latest
forge --help
```

## 2) Create a project

```bash
forge new myapp
cd myapp
```

## 3) Define a model

```go
type Post struct {
    schema.BaseSchema
}

func (Post) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("title", schema.Required(), schema.MaxLength(200)),
        schema.TextField("content", schema.Required()),
        schema.TimeField("created_at", schema.AutoNowAdd()),
    }
}
```

## 4) Generate and migrate

```bash
forge generate
forge makemigrations init --auto
forge migrate
```

## 5) Run the server

```bash
forge runserver
```

Expected output:

```text
Starting forge server on http://localhost:8000
```

## Next steps

- [Hello World](/docs/getting-started/hello-world/)
- [First API](/docs/getting-started/first-api/)
- [Models guide](/docs/guides/models/)
