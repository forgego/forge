---
sidebar_position: 11
description: The OpenAPI document skeleton in api/docs and what it does not generate yet.
image: /social-card.png
---

# OpenAPI

Forge does not generate an OpenAPI description of your API yet. The
`github.com/forgego/forge/api/docs` package serves an OpenAPI 3.0 document that
carries only the `info` block: `paths` and `components.schemas` are empty,
because viewsets, serializers and routes are not introspected. There is no
Swagger UI and no CLI command that exports a specification.

## Serving the document

```go
import "github.com/forgego/forge/api/docs"

generator := docs.NewOpenAPIGenerator("Shop API", "1.0.0")
generator.Description = "Catalog and orders"
router.Get("/api/openapi.json", generator.Handler())
```

`GET /api/openapi.json` then returns:

```json
{
  "openapi": "3.0.0",
  "info": {"title": "Shop API", "version": "1.0.0", "description": "Catalog and orders"},
  "paths": {},
  "components": {}
}
```

`generator.Generate()` returns the same document as an `*docs.OpenAPISpec`.
To publish a complete description today, fill `Paths` and
`Components.Schemas` on that value yourself, or maintain an OpenAPI file by
hand and serve it as a static file.

## Next Steps

- **[Platform Security](/docs/server/security/)**: Protecting endpoints with CSRF, CORS, and session cookies.
- **[Full Framework Features](/docs/features/)**: Complete capability matrix.
