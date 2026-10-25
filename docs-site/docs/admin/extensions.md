---
sidebar_position: 6
description: Custom admin UI extensions, client registration API, security trust model, and build setup.
image: /social-card.png
---

# Admin UI Extensions

Forge Admin allows applications to extend the admin console with custom UI components without forking or rebuilding the core SPA bundle.

Extensions are declared in Go, embedded into your application binary via `embed.FS`, and loaded by the admin Single Page Application as same-origin ES modules at runtime.

---

## Go Registration

An extension is defined by the `admin.Extension` struct and registered using `site.RegisterExtension` or the package-level `admin.RegisterExtension`.

```go
package main

import (
    "embed"
    "fmt"
    "io/fs"
    "log"

    "github.com/forgego/forge/admin"
)

//go:embed web/dist/*
var extensionAssets embed.FS

func RegisterExtension(site *admin.Site) error {
    distFS, err := fs.Sub(extensionAssets, "web/dist")
    if err != nil {
        return fmt.Errorf("failed to extract asset sub-filesystem: %w", err)
    }

    ext := admin.Extension{
        Name:    "analytics",
        Assets:  distFS,
        Version: admin.ExtensionAPIVersion, // 1
    }

    return site.RegisterExtension(ext)
}
```

The Go registry requires three fields:
- `Name`: The unique extension identifier matching `^[a-z0-9-]+$`.
- `Assets`: An `fs.FS` filesystem containing the extension's compiled assets (must contain `index.js` at its root).
- `Version`: The targeted client API version, which must equal `admin.ExtensionAPIVersion` (currently `1`).

---

## Name Rules & Startup Validation

Extension registration validates all invariants at application startup. Any violation fails immediately:

- **Name format:** The extension name must match `^[a-z0-9-]+$` (lowercase alphanumeric characters and hyphens only). If invalid, `Validate()` returns:
  `admin extension: invalid name "<name>": must match ^[a-z0-9-]+$`
- **Name uniqueness:** Each extension name must be unique per `admin.Site`. Duplicate names return:
  `admin extension "<name>" already registered`
- **API version compatibility:** The extension's `Version` must match `admin.ExtensionAPIVersion` (`1`). If mismatched, validation returns:
  `admin extension "<name>": unsupported version <v> (supported: 1)`
- **Non-nil assets filesystem:** The `Assets` filesystem cannot be `nil`:
  `admin extension "<name>": assets filesystem cannot be nil`
- **Entry point file:** The assets filesystem must contain a regular file named `index.js`. If missing or if `index.js` is a directory, validation returns:
  `admin extension "<name>": missing index.js: <err>`
  `admin extension "<name>": index.js must be a file, not a directory`

Startup validation is enforced during `site.RegisterExtension(ext)` and re-checked by `site.ValidateExtensions()`.

---

## Asset URL & Delivery

Extension assets are mounted automatically on the admin HTTP router:

- **Mount path:** Assets are served under `<admin-prefix>/ext/<name>/` (for example, `/admin/ext/analytics/index.js`).
- **Filesystem-backed:** Files are streamed directly from the registered `fs.FS` via `http.ServeFileFS`.
- **Files only:** Directory listings are disabled. Requests for directory paths return `404 Not Found`.
- **Path traversal prevention:** Paths are checked with `fs.ValidPath`. Requests attempting directory traversal return `404 Not Found`.
- **MIME & security headers:** All asset responses include the `X-Content-Type-Options: nosniff` header.
- **Allowed HTTP methods:** Only `GET` and `HEAD` requests are allowed; other methods return `405 Method Not Allowed`.

---

## Client Registration API

When the admin SPA boots, it discovers registered extensions via the staff-only metadata endpoint, dynamically imports each extension's `index.js`, and invokes the default exported function passing the `ForgeAdminAPI` object (`api`):

```tsx
export default function (api) {
  // Use api to register widgets, pages, field editors, and action forms
}
```

The `ForgeAdminAPI` interface (`window.ForgeAdmin`) provides:

- `version`: The integer client API version supported by the admin runtime (`1`).
- `registerWidget(type, Component)`: Registers a custom dashboard widget component for the specified widget `type`.
- `registerPage(options, Component)`: Mounts a custom page component at `/ext/<name>/<path>` inside the admin shell.
- `registerFieldEditor(target, Component)`: Registers a custom field or relationship editor component for the target `{ model, field }`.
- `registerActionForm(actionName, Component)`: Registers a custom parameter form component for action `actionName`.
- `react`: The host admin's shared React 19 instance.
- `ui`: A frozen dictionary of design-system UI primitives (`Button`, `Card`, `Table`, `Dialog`, `Select`, `Input`, `Badge`, and more).

### Option Types

```typescript
export interface RegisterPageOptions {
  path: string;
  title: string;
  nav?: PageNavConfig | boolean;
  extension?: string;
}

export interface PageNavConfig {
  label?: string;
  icon?: string | React.ComponentType<any>;
  order?: number;
  group?: string;
}

export interface FieldEditorTarget {
  model: string;
  field: string;
}
```

---

## No Duplicate React Rule & Build Helper

> [!IMPORTANT]
> Extensions **must not** bundle their own copy of React. React 19 requires a single runtime instance per window. Bundling React inside an extension creates duplicate dispatchers, causing runtime `Invalid hook call` errors and breaking context providers.

Always access React from `api.react` or reference `window.ForgeAdmin.react`.

### Vite Build Helper

Configure Vite in library mode targeting ES modules (`es`) with an entry point named `index.js`, and mark React as external:

```typescript
// vite.config.ts
import { defineConfig } from "vite";
import path from "path";

export default defineConfig({
  build: {
    outDir: "dist",
    emptyOutDir: true,
    lib: {
      entry: path.resolve(__dirname, "src/index.tsx"),
      name: "AnalyticsExtension",
      formats: ["es"],
      fileName: () => "index.js",
    },
    rollupOptions: {
      external: ["react", "react-dom"],
      output: {
        globals: {
          react: "ForgeAdmin.react",
        },
      },
    },
  },
});
```

See the complete example in [build-helper example](https://github.com/forgego/forge/tree/master/examples/admin-extension/).

---

## Error Isolation

Extensions are isolated to prevent third-party UI errors from crashing the admin workspace:

- **Component Error Boundaries:** Every component registered via `registerWidget`, `registerPage`, `registerFieldEditor`, or `registerActionForm` is automatically wrapped in an `ExtensionBoundary`. If an extension component throws an unhandled render error, only that component displays an error fallback, leaving the rest of the page fully operational.
- **Boot Failure Banner:** If an extension module fails to download, fails to parse, or throws during initialization, the error is logged to the browser console and displayed in a non-blocking alert banner at the top of the admin console. The core admin shell and all other healthy extensions load normally.

---

## Security Trust Model

The extension architecture operates under the following security model:

- **Same-origin execution:** Extensions execute within the admin user's active browser session under the same origin. Consequently, extension code is trusted and equivalent to code shipped by the application owner.
- **Embedded Go assets:** Extension assets are served exclusively from Go `fs.FS` filesystems embedded into or configured on the server. Remote URLs and external script CDNs are not supported.
- **Strict Content Security Policy:** Because all extension scripts originate from `<admin-prefix>/ext/<name>/`, the admin console maintains a strict `script-src 'self'` Content Security Policy.
- **REST permission checks unchanged:** Extensions only introduce client-side UI. All data fetching, mutations, and actions continue to flow through standard Forge REST API endpoints. Every backend authentication, authorization, role, and permission invariant is enforced on the server and cannot be bypassed by client extension code.

---

## Limits & Constraints

- **Parameterized action forms:** Custom action forms registered via `registerActionForm` wait on parameterized actions (T021 adapter).
- **UI only:** Extensions cannot define new backend database tables, alter server routes, or introduce unauthenticated endpoints. Server-side behavior and workflow transitions must use existing backend APIs and hooks.
- **No hot reload or marketplace:** Extensions do not support runtime hot-reloading, untrusted iframe sandboxing, dynamic third-party marketplaces, or server-side rendering.

---

## Next Steps

- **[Admin Console Overview](./overview.md)**: Architecture and security overview.
- **[Admin UI & Custom Plugins](./ui.md)**: Dashboard layout, components, and server-driven plugin pages.
- **[Admin Actions](./actions.md)**: Batch actions, row actions, and confirmation dialogs.
- **[Model Configuration](./config.md)**: Admin list, filter, and field configurations.
- **[Admin Extension Example](https://github.com/forgego/forge/tree/master/examples/admin-extension/)**: Full runnable example with Go embedding and Vite build helper.
