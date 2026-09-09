---
sidebar_position: 11
description: Automatic OpenAPI 3.0 specification generation, Swagger UI, and schema export.
image: /forge-social-card.svg
---

# OpenAPI 3.0 & Swagger UI

Forge automatically generates OpenAPI 3.0 specifications by introspecting your registered `ViewSets`, `ModelSerializers`, and router endpoints.

---

## Enabling OpenAPI

Mount the OpenAPI generator onto your Chi router:

```go
package main

import (
    "github.com/go-chi/chi/v5"
    "github.com/forgego/forge/api"
)

func RegisterAPIDocumentation(r chi.Router) {
    generator := api.NewOpenAPIGenerator(api.DocConfig{
        Title:       "Forge Ecommerce API",
        Version:     "1.0.0",
        Description: "Type-safe high-performance Go REST API powered by Forge.",
        ContactEmail: "dev@forgego.dev",
        License:     "MIT",
    })

    // Serves OpenAPI 3.0 JSON specification
    r.Get("/api/openapi.json", generator.ServeSpec)

    // Mount interactive Swagger UI explorer
    r.Get("/api/docs/*", generator.ServeSwaggerUI("/api/openapi.json"))
}
```

Navigate to `http://localhost:8000/api/docs/` to explore your interactive API documentation.

---

## Schema Generation Features

1. **Auto-Generated Schemas**: Serializer fields and validation rules (e.g. `MaxLength`, `Required`, `Choices`) are translated directly into JSON Schema types (`string`, `integer`, `minimum`, `maximum`, `enum`).
2. **Standard HTTP Responses**: Status codes `200 OK`, `201 Created`, `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, and `422 Unprocessable Entity` are documented with concrete response models.
3. **Query Parameter Descriptors**: ViewSet pagination parameters (`limit`, `offset`, `page`, `cursor`) and FilterSet query arguments are automatically indexed in the `parameters` section.
4. **Export Specification via CLI**:
   ```bash
   forge routes --openapi > openapi.json
   ```

---

## Next Steps

- **[Platform Security](/docs/server/security/)**: Protecting endpoints with CSRF, CORS, and session cookies.
- **[Full Framework Features](/docs/features/)**: Complete capability matrix.

