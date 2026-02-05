# Forge Admin Extensibility Guide

The Forge Admin Framework is designed for deep customization at both the backend and frontend levels. This guide outlines the core hooks and override mechanisms.

## Backend Extensibility

### 1. Custom Model Configuration
Every registered model can be configured using `admin.Config[T]`.

```go
admin.Register(&admin.Config[Product]{
    // Custom logic hooks
    GetQueryset: func(ctx context.Context, a *Admin[Product], qs orm.QuerySet[Product]) (orm.QuerySet[Product], error) {
        return qs.Filter(orm.Q{"deleted_at__isnull": true}), nil
    },
    
    // UI Metadata
    Icon: "Package",
    ListDisplay: []string{"name", "price"},
})
```

### 2. Custom Actions
Define bulk actions that show up in the list view.

```go
Actions: []admin.Action[Product]{
    {
        Name: "mark_as_featured",
        Label: "Feature Product",
        Handler: func(ctx context.Context, instances []*Product) error {
            // Your logic here
            return nil
        },
    },
},
```

### 3. Smart Filtering
Implement sidebar filters with custom logic handlers.

```go
Filters: []admin.Filter[Product]{
    {
        Name: "price_range",
        Label: "Price Range",
        Handler: func(ctx context.Context, qs orm.QuerySet[Product], value interface{}) orm.QuerySet[Product] {
            // Custom filtering logic
            return qs.Filter(orm.Q{"price__gt": value})
        },
    },
},
```

## Frontend Extensibility

### 1. UI Component Overrides
Use the `UIOverrides` map in `admin.Config` to swap out entire components.

```go
// Backend
adminSite.UIOverrides = map[string]string{
    "sidebar.brand": "MyCustomLogo",
    "form.footer": "PremiumFormFooter",
}
```

### 2. Field-Level Overrides
You can override specific fields in the form builder by registering them in the frontend `ComponentRegistry`.

```typescript
// Frontend (lib/boot.ts)
registerUIComponent("CustomPriceEditor", PriceEditor);
```

### 3. Dynamic Page Engine
The `ModelListPage` and `ModelUpsertPage` are reactive to the metadata provided by the server. Adding a new field to your Go struct automatically populates it in the UI with an appropriate widget.

---
For deep integration details, refer to the `forge/admin/core/interface.go` for all available method hooks.
