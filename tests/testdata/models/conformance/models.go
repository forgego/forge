package conformance

import "github.com/forgego/forge/schema"

type Owner struct {
	schema.BaseSchema
	ID int64 `json:"id" db:"id"`
}

func (Owner) Fields() []schema.Field {
	return []schema.Field{schema.Int64Field("id", schema.Primary(), schema.AutoIncrement())}
}

func (Owner) Meta() schema.Meta { return schema.Meta{TableName: "owners"} }

type Tag struct {
	schema.BaseSchema
	ID int64 `json:"id" db:"id"`
}

func (Tag) Fields() []schema.Field {
	return []schema.Field{schema.Int64Field("id", schema.Primary(), schema.AutoIncrement())}
}

func (Tag) Meta() schema.Meta { return schema.Meta{TableName: "tags"} }

type CatalogItem struct {
	schema.BaseSchema
	ID           int64   `json:"id" db:"id"`
	OptionalNote *string `json:"optional_note" db:"optional_note"`
	Name         string  `json:"name" db:"name"`
	State        string  `json:"state" db:"state"`
	ExternalCode string  `json:"external_code" db:"external_code"`
	OwnerID      int64   `json:"owner_id" db:"owner_id"`
}

func (CatalogItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("optional_note", schema.Optional(), schema.MaxLength(40)),
		schema.StringField("name", schema.Required(), schema.MaxLength(60)),
		schema.StringField("state", schema.Required(), schema.MaxLength(20), schema.Default("draft")),
		schema.StringField("external_code", schema.Required(), schema.MaxLength(32), schema.Unique()),
		schema.Int64Field("owner_id", schema.Required()),
	}
}

func (CatalogItem) Meta() schema.Meta { return schema.Meta{TableName: "catalog_items"} }

func (CatalogItem) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("owner_id", "Owner", schema.OnDelete(schema.CascadeCASCADE)),
		schema.ManyToManyField("tags", "Tag", schema.Through("catalog_item_tags")),
	}
}

type CatalogItemTag struct {
	schema.BaseSchema
	ID            int64 `json:"id" db:"id"`
	CatalogItemID int64 `json:"catalog_item_id" db:"catalog_item_id"`
	TagID         int64 `json:"tag_id" db:"tag_id"`
}

func (CatalogItemTag) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("catalog_item_id", schema.Required()),
		schema.Int64Field("tag_id", schema.Required()),
	}
}

func (CatalogItemTag) Meta() schema.Meta { return schema.Meta{TableName: "catalog_item_tags"} }

func (CatalogItemTag) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("catalog_item_id", "CatalogItem", schema.OnDelete(schema.CascadeCASCADE)),
		schema.ForeignKeyField("tag_id", "Tag", schema.OnDelete(schema.CascadeCASCADE)),
	}
}
