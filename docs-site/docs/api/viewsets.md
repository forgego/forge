---
sidebar_position: 3
description: CRUD endpoints with viewsets.
image: /forge-social-card.svg
---

# ViewSets

ViewSets provide CRUD endpoints with built-in auth, permissions, and throttling hooks.

## Example

```go
router.Register("posts", viewsets.NewModelViewSet(PostSerializer{}))
```

## Next steps

- [Authentication](/docs/api/authentication/)
- [Permissions](/docs/api/permissions/)
