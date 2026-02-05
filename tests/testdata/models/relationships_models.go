package models

import "github.com/forgego/forge/schema"

// Author represents an author
type Author struct {
	schema.BaseSchema
}

func (Author) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().MaxLength(255).Build(),
		schema.String("email").Required().MaxLength(255).Unique().Build(),
	}
}

func (Author) Meta() schema.Meta {
	return schema.Meta{
		TableName: "authors",
	}
}

func (Author) Relations() []schema.Relation {
	return []schema.Relation{}
}

// Article with foreign key to Author
type Article struct {
	schema.BaseSchema
}

func (Article) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("title").Required().MaxLength(200).Build(),
		schema.Text("body").Build(),
		schema.Int64("author_id").Required().Build(),
	}
}

func (Article) Meta() schema.Meta {
	return schema.Meta{
		TableName: "articles",
	}
}

func (Article) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("author_id", "Author").OnDelete("CASCADE").Build(),
	}
}

// Profile with OneToOne relationship to Author
type Profile struct {
	schema.BaseSchema
}

func (Profile) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.Text("bio").Build(),
		schema.String("website").MaxLength(255).Build(),
		schema.Int64("author_id").Required().Unique().Build(),
	}
}

func (Profile) Meta() schema.Meta {
	return schema.Meta{
		TableName: "profiles",
	}
}

func (Profile) Relations() []schema.Relation {
	return []schema.Relation{
		schema.OneToOne("author_id", "Author").OnDelete("CASCADE").Build(),
	}
}

// Tag for many-to-many relationship
type Tag struct {
	schema.BaseSchema
}

func (Tag) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().MaxLength(100).Unique().Build(),
		schema.String("slug").Required().MaxLength(100).Unique().Build(),
	}
}

func (Tag) Meta() schema.Meta {
	return schema.Meta{
		TableName: "tags",
	}
}

func (Tag) Relations() []schema.Relation {
	return []schema.Relation{}
}

// ArticleTag for many-to-many relationship between Article and Tag
type ArticleTag struct {
	schema.BaseSchema
}

func (ArticleTag) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.Int64("article_id").Required().Build(),
		schema.Int64("tag_id").Required().Build(),
	}
}

func (ArticleTag) Meta() schema.Meta {
	return schema.Meta{
		TableName: "article_tags",
		Indexes: []schema.Index{
			schema.Index("article_id", "tag_id").Unique().Build(),
		},
	}
}

func (ArticleTag) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("article_id", "Article").OnDelete("CASCADE").Build(),
		schema.ForeignKey("tag_id", "Tag").OnDelete("CASCADE").Build(),
	}
}