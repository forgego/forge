---
sidebar_position: 5
---

# Relations

Relations define how models are connected to each other. forge supports ForeignKey, OneToOne, and ManyToMany relationships (one-to-many is represented by a ForeignKey with a `RelatedName`).

Complete reference for model relationships.

## Relation Types

### ForeignKey (Many-to-One)

A many-to-one relationship:

```go
import "github.com/forgego/forge/schema"

func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ForeignKeyField("author", "User",
            schema.OnDelete(schema.CascadeCASCADE),
            schema.RelatedName("posts"),
        ),
    }
}
```

**Access:**
```go
post.Author  // *User
user.Posts   // []*Post (reverse)
```

### OneToOne

A one-to-one relationship:

```go
func (User) Relations() []schema.Relation {
    return []schema.Relation{
        schema.OneToOneField("profile", "UserProfile",
            schema.OnDelete(schema.CascadeCASCADE),
        ),
    }
}
```

**Access:**
```go
user.Profile    // *UserProfile
profile.User    // *User (reverse)
```

### ManyToMany

A many-to-many relationship:

```go
func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ManyToManyField("tags", "Tag",
            schema.Through("post_tags"),
            schema.RelatedName("posts"),
        ),
    }
}
```

**Access:**
```go
post.Tags  // []*Tag
tag.Posts  // []*Post (reverse)
```

## Relation Options

### OnDelete

Cascade behavior on delete:

```go
schema.ForeignKeyField("author", "User",
    schema.OnDelete(schema.CascadeCASCADE),
    // schema.OnDelete(schema.CascadeSET_NULL),
    // schema.OnDelete(schema.CascadeSET_DEFAULT),
    // schema.OnDelete(schema.CascadePROTECT),   // Prevent deletion
    // schema.OnDelete(schema.CascadeDO_NOTHING), // No action
)
```

### OnUpdate

Cascade behavior on update:

```go
schema.ForeignKeyField("author", "User",
    schema.OnUpdate(schema.CascadeCASCADE),
)
```

### RelatedName

Reverse relation name:

```go
schema.ForeignKeyField("author", "User",
    schema.RelatedName("posts"),
)
```

### Through

Through table for ManyToMany:

```go
schema.ManyToManyField("tags", "Tag",
    schema.Through("post_tags"),
)
```

### DBConstraint

Control FK constraint creation:

```go
schema.ForeignKeyField("author", "User",
    schema.DBConstraint(true),
)
```

## Cascade Options

| Option | SQL | Description |
|--------|-----|-------------|
| `CascadeCASCADE` | `ON DELETE CASCADE` | Delete related objects |
| `CascadeSET_NULL` | `ON DELETE SET NULL` | Set foreign key to NULL |
| `CascadeSET_DEFAULT` | `ON DELETE SET DEFAULT` | Set foreign key to default |
| `CascadePROTECT` | `ON DELETE RESTRICT` | Prevent deletion if related objects exist |
| `CascadeDO_NOTHING` | `ON DELETE NO ACTION` | No action (database default) |

## Examples

### Blog Post with Author

```go
type Post struct {
    schema.BaseSchema
}

func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ForeignKeyField("author", "User",
            schema.OnDelete(schema.CascadeCASCADE),
            schema.RelatedName("posts"),
        ),
    }
}
```

### User with Profile

```go
type User struct {
    schema.BaseSchema
}

func (User) Relations() []schema.Relation {
    return []schema.Relation{
        schema.OneToOneField("profile", "UserProfile",
            schema.OnDelete(schema.CascadeCASCADE),
        ),
    }
}
```

### Post with Tags

```go
type Post struct {
    schema.BaseSchema
}

func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ManyToManyField("tags", "Tag",
            schema.Through("post_tags"),
            schema.RelatedName("posts"),
        ),
    }
}
```

## See Also

- [Models Guide](/docs/guides/models) - Model usage guide
- [Queries Guide](/docs/guides/queries) - Querying relations

