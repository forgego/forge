---
sidebar_position: 1
description: Register models and mount the Forge admin site.
image: /social-card.png
---

# Admin Console Overview

Forge serves an admin UI and API from an `admin.Site`. The UI includes model pages, a dashboard, and server-driven plugin pages.

The frontend uses React 19 and TanStack Router. Its theme controls switch between light and dark modes. Press Ctrl+K to open the global search palette; press `?` outside a text field for keyboard shortcut help. The default dashboard shows model and record summaries. A configured server-driven dashboard can render Recharts charts; see [dashboard configuration](/docs/admin/ui/).

## Set up a site

```go
package main

import (
    "github.com/go-chi/chi/v5"
    "github.com/forgego/forge/admin"
    "myapp/models"
)

func SetupAdmin(r chi.Router) error {
    site := admin.NewSite("commerce")
    site.Title = "Forge Commerce Admin"
    ui := admin.DefaultUIConfig()
    ui.Prefix = "/admin"
    site.WithUIConfig(ui)

    _, err := admin.RegisterWithSite(site, &admin.Config[models.Product]{
        ListDisplay: []admin.Field{admin.Computed("name")},
        SearchFields: []admin.Field{admin.Computed("name")},
    })
    if err != nil { return err }

    r.Mount("/admin", site.Handler())
    return nil
}
```

`admin.Register(&admin.Config[T]{...})` registers on `admin.DefaultSite`. Use `RegisterWithSite` for a separate site. Set the site's database with `site.SetDB(database)` before registering models that need it.

## Permissions

Autocomplete and global search apply `HasViewPermission` to each suggested object. Objects denied by the callback do not appear in suggestions. CSV exports also omit denied objects and escape formula-like text cells before download.

Permission callbacks receive a context, the registered admin, a user value, and (except for add and module permissions) the object. The user is an `interface{}` value; check its type before using application-specific roles.

```go
HasViewPermission: func(ctx context.Context, modelAdmin *admin.Admin[models.Product], user interface{}, obj *models.Product) bool {
    return user != nil && obj != nil
},
HasAddPermission: func(ctx context.Context, modelAdmin *admin.Admin[models.Product], user interface{}) bool {
    return user != nil
},
```

See [model configuration](/docs/admin/config/), [actions](/docs/admin/actions/), [filters](/docs/admin/filters/), and [plugin pages](/docs/admin/ui/).
