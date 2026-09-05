---
sidebar_position: 1
---

# Getting Started Tutorial

This is the guided version of the quick start. Follow in order.

## 1. Install and scaffold

```bash
go install github.com/forgego/forge/cli/cmd@latest
forge new myapp
cd myapp
```

## 2. Configure the database

Edit `config/config.yaml` with your DB settings, then run:

```bash
forge migrate
```

## 3. Create your first model

```go
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
```

## 4. Generate code

```bash
forge generate
forge makemigrations init --auto
forge migrate
```

## 5. Run the server

```bash
forge runserver
```

## Next steps

- [Models guide](/docs/guides/models)
- [REST API guide](/docs/guides/rest-api)
- [Admin guide](/docs/guides/admin)
