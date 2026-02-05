package models

import "github.com/forgego/forge/schema"

// ProductWithJSONB demonstrates JSONB columns
type ProductWithJSONB struct {
	schema.BaseSchema
}

func (ProductWithJSONB) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().MaxLength(255).Build(),
		schema.JSON("attributes").Build(), // Maps to JSONB in Postgres
		schema.JSON("metadata").Build(),
	}
}

func (ProductWithJSONB) Meta() schema.Meta {
	return schema.Meta{
		TableName: "products_jsonb",
	}
}

func (ProductWithJSONB) Relations() []schema.Relation {
	return []schema.Relation{}
}

// UserWithUUID demonstrates UUID fields
type UserWithUUID struct {
	schema.BaseSchema
}

func (UserWithUUID) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.UUID("uuid").Required().DBDefault("gen_random_uuid()").Build(),
		schema.String("email").Required().MaxLength(255).Unique().Build(),
	}
}

func (UserWithUUID) Meta() schema.Meta {
	return schema.Meta{
		TableName: "users_uuid",
	}
}

func (UserWithUUID) Relations() []schema.Relation {
	return []schema.Relation{}
}

// DocumentWithTimestamps demonstrates timestamp with time zone
type DocumentWithTimestamps struct {
	schema.BaseSchema
}

func (DocumentWithTimestamps) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("title").Required().MaxLength(255).Build(),
		schema.DateTime("created_at").Build(),
		schema.DateTime("updated_at").Build(),
	}
}

func (DocumentWithTimestamps) Meta() schema.Meta {
	return schema.Meta{
		TableName: "documents",
	}
}

func (DocumentWithTimestamps) Relations() []schema.Relation {
	return []schema.Relation{}
}