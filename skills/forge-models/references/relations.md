# Relations

Relations are defined in `Relations() []schema.Relation` using functional builders from `schema/relations_functional.go`.

## Relation Constructors

### ForeignKey (Many-to-One)

```go
func (Post) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("author_id", "User",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("posts"),
			schema.RelatedQueryName("post"),
		),
	}
}
```

### OneToOne

```go
func (UserProfile) Relations() []schema.Relation {
	return []schema.Relation{
		schema.OneToOneField("user_id", "User",
			schema.OnDelete(schema.CascadeCASCADE),
			schema.RelatedName("profile"),
		),
	}
}
```

### ManyToMany

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

## Cascade Options (`schema/relation.go`)

- `schema.CascadeCASCADE`
- `schema.CascadeSET_NULL`
- `schema.CascadeRESTRICT`
- `schema.CascadeNO_ACTION`
- `schema.CascadeSET_DEFAULT`

## Struct Fields & Loading

`forge generate` does **not** automatically inject relation pointer/slice fields (such as `Author *User` or `Tags []*Tag`) onto `<Model>Generated`. Declare those fields on your model struct with `db:"-"` if you hydrate them via `SelectRelated` or `PrefetchRelated` in `orm`.
