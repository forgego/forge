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

type Resource struct{ ResourceGenerated }

func (Resource) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(80)),
		schema.StringField("code", schema.Required(), schema.MaxLength(80), schema.Unique()),
		schema.BoolField("active", schema.Default(true)),
		schema.StringField("internal_token", schema.Default("system-only"), schema.WriteOnly(), schema.Editable(false)),
		schema.DateTimeField("created_at", schema.AutoNowAdd()),
		schema.DateTimeField("updated_at", schema.AutoNow()),
		// ACCEPTANCE EXTRA FIELD
	}
}
func (Resource) Meta() schema.Meta {
	return schema.Meta{TableName: "resources", VerboseName: "Resource", Indexes: []schema.Index{{Name: "idx_resources_name", Fields: []string{"name"}}}, OrderBy: []string{"name"}}
}
func (Resource) Relations() []schema.Relation {
	return []schema.Relation{}
}
func (Resource) Hooks() *schema.ModelHooks { return nil }

type Slot struct{ SlotGenerated }

func (Slot) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("resource_id", schema.Required()),
		schema.DateTimeField("starts_at", schema.Required()),
		schema.DateTimeField("ends_at", schema.Required()),
	}
}
func (Slot) Meta() schema.Meta {
	return schema.Meta{TableName: "slots", VerboseName: "Slot", Constraints: []schema.Constraint{{Name: "unique_resource_start", Type: schema.ConstraintTypeUnique, Fields: []string{"resource_id", "starts_at"}}, {Name: "positive_slot", Type: schema.ConstraintTypeCheck, Condition: "ends_at > starts_at"}}}
}
func (Slot) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("resource_id", "Resource", schema.OnDelete(schema.CascadePROTECT)),
	}
}
func (Slot) Hooks() *schema.ModelHooks { return nil }

type Booking struct{ BookingGenerated }

func (Booking) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("slot_id", schema.Required(), schema.Unique()),
		schema.Int64Field("owner_id", schema.Required()),
		schema.StringField("reference", schema.Required(), schema.MaxLength(80), schema.Unique()),
		schema.StringField("status", schema.Required(), schema.MaxLength(80), schema.Default("confirmed")),
	}
}
func (Booking) Meta() schema.Meta { return schema.Meta{TableName: "bookings", VerboseName: "Booking"} }
func (Booking) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("slot_id", "Slot", schema.OnDelete(schema.CascadePROTECT)),
		schema.ForeignKeyField("owner_id", "Member", schema.OnDelete(schema.CascadePROTECT)),
	}
}
func (Booking) Hooks() *schema.ModelHooks { return nil }
