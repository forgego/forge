---
sidebar_position: 2
description: Build your first forge application in 5 minutes. Learn how to create a simple Hello World API endpoint with forge.
keywords:
  - forge hello world
  - forge tutorial
  - forge first app
  - forge quickstart
  - forge example
image: /forge-social-card.svg
---

# Hello World

Let's build your first forge app. This'll take about 5 minutes.

## What we're building

A simple API that says "Hello, World!" in JSON. Nothing fancy, just enough to see forge in action.

## Step 1: Create the project

```bash
forge new hello-world
cd hello-world
```

This creates a new project with all the boilerplate you need.

## Step 2: Set up the database

Open `config/config.yaml` and add your database settings:

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: your_password
  dbname: hello_world_db
  sslmode: disable

server:
  host: localhost
  port: "8000"
```

Then create the database:

```bash
psql -U postgres -c "CREATE DATABASE hello_world_db;"
```

(If you don't have PostgreSQL running, start it first)

## Step 3: Write a handler

Open `cmd/server/main.go` and add this:

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/forgego/forge/pkg/config"
    "github.com/forgego/forge/pkg/db"
    "github.com/forgego/forge/pkg/logging"
    httplib "github.com/forgego/forge/pkg/http"
)

func main() {
    cfg := config.NewConfig()
    settings := config.LoadSettings(cfg)

    logger, err := logging.NewLogger(cfg.IsDevelopment())
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Sync()

    database, err := db.NewDBFromConfig(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer database.Close()

    server, err := httplib.NewServer(cfg, settings, logger)
    if err != nil {
        log.Fatal(err)
    }

    server.RegisterRoutes(func(router *httplib.Router) {
        router.Get("/", helloHandler)
        router.Get("/api/hello", helloAPIHandler)
    })

    logger.Info("Starting server", "host", settings.Server.Host, "port", settings.Server.Port)
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(`
        <html>
            <body>
                <h1>Hello, World!</h1>
                <p>Welcome to forge!</p>
            </body>
        </html>
    `))
}

func helloAPIHandler(w http.ResponseWriter, r *http.Request) {
    response := map[string]interface{}{
        "message": "Hello, World!",
        "framework": "forge",
        "language": "Go",
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

## Step 4: Run it

Either run it directly:

```bash
go run cmd/server/main.go
```

Or use the forge CLI:

```bash
forge runserver
```

## Step 5: Test it

Open your browser and hit:

- `http://localhost:8000/` - Returns HTML
- `http://localhost:8000/api/hello` - Returns JSON

You should see the JSON response:

```json
{
  "message": "Hello, World!",
  "framework": "forge",
  "language": "Go"
}
```

## What just happened?

1. The server started on port 8000
2. Your routes were registered (`/` and `/api/hello`)
3. When you visited the URL, forge matched the route and called your handler
4. Your handler sent back the response

That's it. You've got a working forge app.

## Next steps

Now that you have a working application, try:

1. **[Add a model](/docs/getting-started/first-api)** - Create your first database model
2. **[Build an API](/docs/guides/rest-api)** - Create a REST API
3. **[Add authentication](/docs/guides/authentication/)** - Secure your API
4. **[Explore examples](/docs/examples/blog)** - See complete examples

## Troubleshooting

**Port 8000 already in use?** Change it in `config/config.yaml`:

```yaml
server:
  port: "8001"
```

**Database connection failing?** Make sure PostgreSQL is running and your credentials are right:

```bash
psql -U postgres -d hello_world_db -c "SELECT 1;"
```

**Module not found?** Run:

```bash
go mod download
```
