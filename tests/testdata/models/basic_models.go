package models

import "github.com/forgego/forge/schema"

// User is a basic user model
type User struct {
	schema.BaseSchema
}

func (User) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("email").Required().MaxLength(255).Unique().Build(),
		schema.String("username").Required().MaxLength(150).Build(),
		schema.Bool("is_active").Default(true).Build(),
		schema.DateTime("created_at").Build(),
		schema.DateTime("updated_at").Build(),
	}
}

func (User) Meta() schema.Meta {
	return schema.Meta{
		TableName: "users",
	}
}

func (User) Relations() []schema.Relation {
	return []schema.Relation{}
}

// Post is a basic post model
type Post struct {
	schema.BaseSchema
}

func (Post) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("title").Required().MaxLength(200).Build(),
		schema.Text("content").Build(),
		schema.String("status").Default("draft").MaxLength(50).Build(),
		schema.DateTime("published_at").Build(),
		schema.DateTime("created_at").Build(),
		schema.DateTime("updated_at").Build(),
	}
}

func (Post) Meta() schema.Meta {
	return schema.Meta{
		TableName: "posts",
	}
}

func (Post) Relations() []schema.Relation {
	return []schema.Relation{}
}

// Comment is a basic comment model
type Comment struct {
	schema.BaseSchema
}

func (Comment) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.Text("content").Required().Build(),
		schema.Bool("is_approved").Default(false).Build(),
		schema.DateTime("created_at").Build(),
		schema.DateTime("updated_at").Build(),
	}
}

func (Comment) Meta() schema.Meta {
	return schema.Meta{
		TableName: "comments",
	}
}

func (Comment) Relations() []schema.Relation {
	return []schema.Relation{}
}