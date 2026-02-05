---
sidebar_position: 5
---

# Relations

Relations define how models are connected to each other. forge supports ForeignKey, OneToOne, OneToMany, and ManyToMany relationships.

Complete reference for model relationships.

## Relation Types

### ForeignKey (Many-to-One)

A many-to-one relationship:

```go
import "github.com/forgego/forge/pkg/schema/relations"

func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        relations.ForeignKey("author", "User").
            Required().
            OnDelete(schema.Cascade).
            RelatedName("posts"),
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
        relations.OneToOne("profile", "UserProfile").
            OnDelete(schema.Cascade),
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
        relations.ManyToMany("tags", "Tag").
            Through("post_tags").
            RelatedName("posts"),
    }
}
```

**Access:**
```go
post.Tags  // []*Tag
tag.Posts  // []*Post (reverse)
```

## Relation Options

### Required

Make relation required:

```go
relations.ForeignKey("author", "User").Required()
```

### OnDelete

Cascade behavior on delete:

```go
relations.ForeignKey("author", "User").
    OnDelete(schema.Cascade)
    // OnDelete(schema.SetNull)
    // OnDelete(schema.SetDefault)
    // OnDelete(schema.Restrict)  // Prevent deletion
    // OnDelete(schema.DoNothing) // No action
```

### OnUpdate

Cascade behavior on update:

```go
relations.ForeignKey("author", "User").
    OnUpdate(schema.Cascade)
```

### RelatedName

Reverse relation name:

```go
relations.ForeignKey("author", "User").
    RelatedName("posts")
```

### Through

Through table for ManyToMany:

```go
relations.ManyToMany("tags", "Tag").
    Through("post_tags")
```

### FromField / ToField

Custom field names:

```go
relations.ManyToMany("tags", "Tag").
    Through("post_tags").
    FromField("post_id").
    ToField("tag_id")
```

## Cascade Options

| Option | SQL | Description |
|--------|-----|-------------|
| `Cascade` | `ON DELETE CASCADE` | Delete related objects |
| `SetNull` | `ON DELETE SET NULL` | Set foreign key to NULL |
| `SetDefault` | `ON DELETE SET DEFAULT` | Set foreign key to default |
| `Restrict` | `ON DELETE RESTRICT` | Prevent deletion if related objects exist |
| `DoNothing` | `ON DELETE NO ACTION` | No action (database default) |

## Examples

### Blog Post with Author

```go
type Post struct {
    schema.BaseSchema
}

func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        relations.ForeignKey("author", "User").
            Required().
            OnDelete(schema.Cascade).
            RelatedName("posts"),
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
        relations.OneToOne("profile", "UserProfile").
            OnDelete(schema.Cascade),
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
        relations.ManyToMany("tags", "Tag").
            Through("post_tags").
            RelatedName("posts"),
    }
}
```

## See Also

- [Models Guide](/docs/guides/models) - Model usage guide
- [Queries Guide](/docs/guides/queries) - Querying relations

