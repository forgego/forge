---
sidebar_position: 2
description: Define API serializers and field mappings.
image: /forge-social-card.svg
---

# Serializers

Serializers map model fields to API responses and validate input.

## What you can do

- Model serializers with field lists
- Typed serializers and enhanced serializers
- Field-level validation and mapping

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
