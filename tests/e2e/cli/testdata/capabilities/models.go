package probes

import "github.com/forgego/forge/schema"

type Asset struct{ AssetGenerated }

func (Asset) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.DecimalField("amount", schema.MaxDigits(12), schema.DecimalPlaces(2)),
		schema.JSONField("payload", schema.Optional()),
		schema.UUIDField("external_id", schema.Optional()),
		schema.StringField("code", schema.DBIndex()),
		schema.StringField("tenant"),
	}
}
func (Asset) Meta() schema.Meta {
	return schema.Meta{TableName: "probe_assets", UniqueTogether: [][]string{{"tenant", "code"}}, Indexes: []schema.Index{{Name: "probe_assets_meta_index", Fields: []string{"external_id"}}}, Constraints: []schema.Constraint{{Name: "probe_assets_meta_unique", Type: "UNIQUE", Fields: []string{"tenant", "external_id"}}}}
}
func (Asset) Relations() []schema.Relation {
	return []schema.Relation{schema.ManyToManyField("labels", "Label")}
}

type Label struct{ LabelGenerated }

func (Label) Fields() []schema.Field {
	return []schema.Field{schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()), schema.StringField("name")}
}
func (Label) Meta() schema.Meta { return schema.Meta{TableName: "probe_labels"} }
