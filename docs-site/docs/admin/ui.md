---
sidebar_position: 5
description: UI customization and component overrides.
image: /forge-social-card.svg
---

# Admin UI

Customize the admin UI with component overrides and custom pages.

## UI overrides

```go
adminSite.UIOverrides = map[string]string{
    "sidebar.brand": "MyCustomLogo",
}
```

## Custom pages

Admin supports plugin pages and dynamic rendering driven by server metadata.

## Next steps

- [Admin Config](/docs/admin/config/)
- [Admin Overview](/docs/admin/overview/)
