---
sidebar_position: 2
description: Define API serializers and field mappings.
image: /forge-social-card.svg
---

# Serializers

Serializers map model fields to API responses and validate input.

## Example

```go
type PostSerializer struct {
    serializers.ModelSerializer
}

func (PostSerializer) Meta() serializers.Meta {
    return serializers.Meta{
        Model: Post{},
        Fields: []string{"id", "title", "content"},
    }
}
```

## Next steps

- [ViewSets](/docs/api/viewsets/)
- [Errors](/docs/api/errors/)
