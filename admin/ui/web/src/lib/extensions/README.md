# Admin UI Extension System

Admin UI extensions allow mounting custom widgets, full pages, field editors, and action parameter forms at runtime without rebuilding the core admin interface.

## Registration API (`window.ForgeAdmin`)

Extensions are loaded dynamically via ES module import and passed the frozen `window.ForgeAdmin` API object as the argument to their `default(api)` export:

```ts
export default function init(api: typeof window.ForgeAdmin) {
  // 1. Register a custom dashboard widget
  api.registerWidget("my-stats", MyStatsWidget);

  // 2. Register a page mounted at /ext/<extension-name>/<path>
  api.registerPage(
    {
      path: "settings",
      title: "Settings",
      nav: { label: "Settings", order: 10 }
    },
    SettingsPage
  );

  // 3. Register a custom field editor for model and field
  api.registerFieldEditor(
    { model: "orders", field: "status" },
    CustomStatusEditor
  );

  // 4. Register a parameter form for actions
  api.registerActionForm("export_custom", CustomExportForm);
}
```

## Shared React & Vite Build Helper

Extensions must NOT bundle their own duplicate copy of React. Instead, they share the admin's React runtime (`api.react`) and design primitives (`api.ui`).

Recommended `vite.config.ts` for extension authors:

```ts
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig({
  plugins: [react()],
  build: {
    lib: {
      entry: path.resolve(__dirname, "src/index.ts"),
      formats: ["es"],
      fileName: () => "index.js",
    },
    rollupOptions: {
      external: ["react", "react-dom"],
      output: {
        // Map external react imports directly to window.ForgeAdmin.react
        globals: {
          react: "ForgeAdmin.react",
        },
        paths: {
          react: "window.ForgeAdmin.react",
        },
      },
    },
  },
});
```

## Isolation & Error Handling

- Every component registered through `ForgeAdmin` is automatically wrapped in an `ExtensionBoundary`.
- Failures in dynamic import, version checks, or module initialization are caught, logged, and surfaced in a non-blocking `ExtensionErrorBanner` while allowing the rest of the admin interface to load safely.
