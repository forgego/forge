# Forge Admin Extension Example

This example demonstrates how to build and register a custom admin UI extension for Forge Admin.

Extensions allow applications to register custom dashboard widgets, full-screen pages, field editors, and action forms without forking or rebuilding the core Forge Admin SPA bundle.

## Project Structure

```text
examples/admin-extension/
├── main.go               # Go entry point embedding assets and registering extension
├── README.md             # This document
└── web/
    ├── package.json      # Frontend package configuration (Vite, React devDependencies)
    ├── vite.config.ts    # Library mode build configuration (ES format, index.js output)
    ├── dist/
    │   └── index.js      # Compiled ES module bundle embedded into Go binary
    └── src/
        └── index.tsx     # Extension registration code (registerWidget, registerPage)
```

## How It Works

1. **Go Registration (`main.go`):**
   The Go backend embeds compiled assets using `embed.FS` and registers an `admin.Extension` on the admin site:
   ```go
   ext := admin.Extension{
       Name:    "analytics",
       Assets:  distFS,
       Version: admin.ExtensionAPIVersion, // 1
   }
   site.RegisterExtension(ext)
   ```
   Assets are served automatically at `<admin-prefix>/ext/<name>/` (e.g. `/admin/ext/analytics/index.js`).

2. **Frontend Build Helper (`web/vite.config.ts`):**
   The extension is built in Vite library mode as an ES module (`es` format) with entry point `index.js`.
   Crucially, `react` and `react-dom` are externalized and mapped to the host admin's `ForgeAdmin.react` instance to prevent duplicate React runtime errors.

3. **Extension Entry Point (`web/src/index.tsx`):**
   The extension exports a default function that receives the `ForgeAdminAPI` object (`api`):
   ```tsx
   export default function (api: any) {
     const React = api.react;
     const { Card, CardHeader, CardTitle, CardContent, Button } = api.ui;

     api.registerWidget("analytics-summary", AnalyticsWidget);
     api.registerPage({ path: "analytics", title: "Analytics", nav: { label: "Analytics" } }, AnalyticsPage);
   }
   ```

## Building & Running

### 1. Build the Frontend Extension

```bash
cd web
npm run build
```

This compiles `web/src/index.tsx` into `web/dist/index.js`. A pre-compiled bundle is included so the example can run immediately.

### 2. Run the Go Application

```bash
go run .
```

Open `http://localhost:8080/admin/` in your browser. The admin console will boot, discover the registered extension, dynamically import `/admin/ext/analytics/index.js`, and mount the custom widget and page.

## Security Considerations

- **Same-Origin Trust:** Extensions execute in the same origin as the admin user's active session and are equivalent to code shipped by the application owner.
- **Embedded Assets:** Assets must be served from Go `fs.FS` filesystems only. External script URLs and CDNs are not permitted, keeping Content Security Policy (`script-src 'self'`) intact.
- **REST Permissions:** Extensions only provide frontend UI. Backend data fetching and actions continue through standard Forge REST APIs subject to all existing permission and authentication checks.
