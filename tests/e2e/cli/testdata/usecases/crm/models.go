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

type Organization struct{ OrganizationGenerated }

func (Organization) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(80)),
		schema.StringField("domain", schema.Required(), schema.MaxLength(80), schema.Unique()),
		schema.StringField("internal_token", schema.Default("system-only"), schema.WriteOnly(), schema.Editable(false)),
		schema.DateTimeField("created_at", schema.AutoNowAdd()),
		schema.DateTimeField("updated_at", schema.AutoNow()),
		// ACCEPTANCE EXTRA FIELD
	}
}
func (Organization) Meta() schema.Meta {
	return schema.Meta{TableName: "organizations", VerboseName: "Organization", Indexes: []schema.Index{{Name: "idx_organizations_name", Fields: []string{"name"}}}, OrderBy: []string{"name"}}
}
func (Organization) Relations() []schema.Relation {
	return []schema.Relation{}
}
func (Organization) Hooks() *schema.ModelHooks { return nil }

type Contact struct{ ContactGenerated }

func (Contact) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("organization_id", schema.Required()),
		schema.StringField("name", schema.Required(), schema.MaxLength(80)),
		schema.EmailField("email", schema.Required(), schema.Unique()),
	}
}
func (Contact) Meta() schema.Meta { return schema.Meta{TableName: "contacts", VerboseName: "Contact"} }
func (Contact) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("organization_id", "Organization", schema.OnDelete(schema.CascadePROTECT)),
	}
}
func (Contact) Hooks() *schema.ModelHooks { return nil }

type Ticket struct{ TicketGenerated }

func (Ticket) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("contact_id", schema.Required()),
		schema.Int64Field("owner_id", schema.Required()),
		schema.StringField("title", schema.Required(), schema.MaxLength(80)),
		schema.StringField("status", schema.Required(), schema.MaxLength(80), schema.Default("open")),
		schema.Int64Field("priority", schema.Required(), schema.Default(0), schema.MinValue(0), schema.MaxValue(5)),
	}
}
func (Ticket) Meta() schema.Meta { return schema.Meta{TableName: "tickets", VerboseName: "Ticket"} }
func (Ticket) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("contact_id", "Contact", schema.OnDelete(schema.CascadePROTECT)),
		schema.ForeignKeyField("owner_id", "Member", schema.OnDelete(schema.CascadePROTECT)),
	}
}
func (Ticket) Hooks() *schema.ModelHooks { return nil }

type Comment struct{ CommentGenerated }

func (Comment) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("ticket_id", schema.Required()),
		schema.TextField("body", schema.Required()),
	}
}
func (Comment) Meta() schema.Meta { return schema.Meta{TableName: "comments", VerboseName: "Comment"} }
func (Comment) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("ticket_id", "Ticket", schema.OnDelete(schema.CascadeCASCADE)),
	}
}
func (Comment) Hooks() *schema.ModelHooks { return nil }
