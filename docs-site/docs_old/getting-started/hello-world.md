---
sidebar_position: 2
description: Build a minimal Hello World API.
keywords:
  - forge hello world
  - forge tutorial
image: /forge-social-card.svg
---

# Hello World

Build a tiny JSON API.

## 1. Create the project

```bash
forge new hello-world
cd hello-world
```

## 2. Configure the database

Edit `config/config.yaml` and set your DB credentials, then:

```bash
forge migrate
```

## 3. Add a route

```go
server.RegisterRoutes(func(router *httplib.Router) {
    router.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"message":"Hello, World!"}`))
    })
})
```

## 4. Run

```bash
forge runserver
```

## Next steps

- [Quick start](/docs/getting-started/quickstart)
- [First API](/docs/getting-started/first-api)
