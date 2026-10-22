package work

import "github.com/forgego/forge/schema"

type Member struct{ MemberGenerated }

func (Member) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(80)),
		schema.StringField("email", schema.Required(), schema.MaxLength(80), schema.Unique()),
		schema.StringField("api_token", schema.Required(), schema.WriteOnly(), schema.Editable(false)),
		schema.BoolField("is_staff", schema.Default(false), schema.WriteOnly(), schema.Editable(false)),
	}
}
func (Member) Meta() schema.Meta { return schema.Meta{TableName: "members", VerboseName: "Member"} }
func (Member) Relations() []schema.Relation {
	return []schema.Relation{}
}
func (Member) Hooks() *schema.ModelHooks { return nil }

type Author struct{ AuthorGenerated }

func (Author) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(80)),
		schema.EmailField("email", schema.Required(), schema.Unique()),
		schema.StringField("internal_token", schema.Default("system-only"), schema.WriteOnly(), schema.Editable(false)),
		schema.DateTimeField("created_at", schema.AutoNowAdd()),
		schema.DateTimeField("updated_at", schema.AutoNow()),
		// ACCEPTANCE EXTRA FIELD
	}
}
func (Author) Meta() schema.Meta {
	return schema.Meta{TableName: "authors", VerboseName: "Author", Indexes: []schema.Index{{Name: "idx_authors_name", Fields: []string{"name"}}}, OrderBy: []string{"name"}}
}
func (Author) Relations() []schema.Relation {
	return []schema.Relation{}
}
func (Author) Hooks() *schema.ModelHooks { return nil }

type Post struct{ PostGenerated }

func (Post) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("author_id", schema.Required()),
		schema.Int64Field("owner_id", schema.Required()),
		schema.StringField("title", schema.Required(), schema.MaxLength(80)),
		schema.StringField("slug", schema.Required(), schema.MaxLength(80), schema.Unique()),
		schema.TextField("body", schema.Optional()),
		schema.BoolField("published", schema.Default(false)),
		schema.DateTimeField("published_at", schema.Optional()),
	}
}
func (Post) Meta() schema.Meta { return schema.Meta{TableName: "posts", VerboseName: "Post"} }
func (Post) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("author_id", "Author", schema.OnDelete(schema.CascadePROTECT)),
		schema.ForeignKeyField("owner_id", "Member", schema.OnDelete(schema.CascadePROTECT)),
	}
}
func (Post) Hooks() *schema.ModelHooks { return nil }

type Tag struct{ TagGenerated }

func (Tag) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(80), schema.Unique()),
	}
}
func (Tag) Meta() schema.Meta { return schema.Meta{TableName: "tags", VerboseName: "Tag"} }
func (Tag) Relations() []schema.Relation {
	return []schema.Relation{}
}
func (Tag) Hooks() *schema.ModelHooks { return nil }

type PostTag struct{ PostTagGenerated }

func (PostTag) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("post_id", schema.Required()),
		schema.Int64Field("tag_id", schema.Required()),
	}
}
func (PostTag) Meta() schema.Meta {
	return schema.Meta{TableName: "post_tags", VerboseName: "PostTag", Constraints: []schema.Constraint{{Name: "unique_post_tag", Type: schema.ConstraintTypeUnique, Fields: []string{"post_id", "tag_id"}}}}
}
func (PostTag) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("post_id", "Post", schema.OnDelete(schema.CascadeCASCADE)),
		schema.ForeignKeyField("tag_id", "Tag", schema.OnDelete(schema.CascadeCASCADE)),
	}
}
func (PostTag) Hooks() *schema.ModelHooks { return nil }

type Revision struct{ RevisionGenerated }

func (Revision) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("post_id", schema.Required()),
		schema.Int64Field("number", schema.Required(), schema.MinValue(1)),
		schema.TextField("body", schema.Required()),
	}
}
func (Revision) Meta() schema.Meta {
	return schema.Meta{TableName: "revisions", VerboseName: "Revision"}
}
func (Revision) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("post_id", "Post", schema.OnDelete(schema.CascadeCASCADE)),
	}
}
func (Revision) Hooks() *schema.ModelHooks { return nil }
