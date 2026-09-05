---
sidebar_position: 5
description: UI customization and component overrides.
image: /forge-social-card.svg
---

# Admin UI

Customize the admin UI with component overrides and plugin pages.

## UI overrides

```go
adminSite.UIOverrides = map[string]string{
    "sidebar.brand": "MyCustomLogo",
    "form.footer": "CustomFooter",
}
```

## Dynamic pages

Admin pages are generated from server-provided metadata. Adding a field to a model updates the UI automatically.

## Next steps

- [Admin Config](/docs/admin/config/)
- [Admin Overview](/docs/admin/overview/)
