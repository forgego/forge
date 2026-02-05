# Comprehensive Implementation Plan - Full-Featured & Extensible Admin Framework

This plan details the complete redesign and implementation of the Forge Admin Framework, focusing on extensibility, embedding, and a high-end React/TypeScript frontend.

## Phase 1: Core Architecture & Extensibility (Backend)

### 1.1 Extended Admin Registry
Enhance `forge/admin/core/registry.go` and `site.go` to support:
- Custom Pages: Pages not tied to a specific model.
- Dashboard Configuration: Global and per-user dashboard widget layouts.
- Menu System: Extensible navigation structure with nesting and icons.

### 1.2 Plugin System (Go)
Implement a formal Plugin interface in `forge/admin/core/plugins.go`:
```go
type Plugin interface {
    Name() string
    Initialize(site *Site) error
    RegisterRoutes(router chi.Router)
    GetMetadata() PluginMetadata
    // UI extension points
    RegisterWidgets() []Widget
    RegisterCustomPages() []CustomPage
}
```

### 1.3 UI Configuration & Serving
Update `forge/admin/site.go` to handle:
- `Source`: Embedded, Static, or External.
- `Template Overrides`: Ability for users to provide their own React build easily.
- SPA Handler: Robust routing that serves `index.html` for all `/admin/*` paths.

## Phase 2: React UI Redesign (Frontend)

### 2.1 Component Registry & Overrides
Implement a registry system in `admin-ui/src/lib/registry.ts`:
- **Model Components**: Override List, Detail, or Form per model.
- **Fields & Widgets**: Register custom input types and display components.
- **Slots**: Define layout slots (TopBar, SidebarFooter, etc.) where users can inject components.

### 2.2 Dashboard & Widget System
- **Grid Layout**: Use `react-grid-layout` for a customisable dashboard.
- **Widget Base**: Standard interface for dashboard widgets (Stats, Charts, Tables).
- **Communication**: Widgets fetch data from defined backend endpoints.

### 2.3 Tab Management
- Support multi-tab interface or sub-tabs within model views for large datasets or related data.

### 2.4 State Management & API
- Move to `TanStack Query` for all data fetching.
- Use `Zustand` for global state (navigation, open tabs, user preferences).

## Phase 3: Embedding & Project Integration

### 3.1 `forge/admin/ui` Package
- Create a bridge package that uses `embed.FS` to bundle the `admin-ui/dist`.
- Provide a simple `admin.Serve(site)` helper that sets up the full stack.

### 3.2 Example/Sample Integration
- Update `examples/ecommerce` to use the new full-featured admin.
- Demonstrate adding a custom page (e.g., "Sales Analytics") and a custom widget.

## Phase 4: CLI Tooling

### 4.1 UI Commands
- `forge admin build`: Compiles the UI.
- `forge admin eject`: Copies UI source to PROJECT_ROOT/admin-ui for total control.
- `forge admin add-page <Name>`: Scaffolds a custom admin page.

## Phase 5: Verification & Testing
- Integrated tests for API-UI communication.
- End-to-end testing with Playwright/Cypress for key admin workflows.
- Performance audit for large table views.
