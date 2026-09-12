package models

import (
	"github.com/forgego/forge/schema"
)

// CustomerProfile represents extended customer profile information
type CustomerProfile struct {
	schema.BaseSchema
}

// Fields returns all field definitions for CustomerProfile
func (CustomerProfile) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.Int64("customer_id").Required().Unique().VerboseName("Customer ID").Build(),
		schema.Date("date_of_birth").Optional().VerboseName("Date of Birth").Build(),
		schema.String("gender").MaxLength(20).Choices(
			schema.Choice{Value: "male", Label: "Male"},
			schema.Choice{Value: "female", Label: "Female"},
			schema.Choice{Value: "other", Label: "Other"},
			schema.Choice{Value: "prefer_not_to_say", Label: "Prefer Not to Say"},
		).VerboseName("Gender").Build(),
		schema.URL("avatar_url").Optional().MaxLength(500).VerboseName("Avatar URL").Build(),
		schema.Text("bio").Optional().VerboseName("Biography").Build(),
		schema.String("preferred_language").MaxLength(10).Default("en").VerboseName("Preferred Language").Build(),
		schema.String("timezone").MaxLength(50).Default("UTC").VerboseName("Timezone").Build(),
		schema.JSON("preferences").Optional().VerboseName("User Preferences").Build(),
		schema.JSON("social_links").Optional().VerboseName("Social Media Links").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (CustomerProfile) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "customer_profiles",
		VerboseName:       "Customer Profile",
		VerboseNamePlural: "Customer Profiles",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			{Name: "idx_customer_profile_customer_id", Fields: []string{"customer_id"}, Unique: true},
		},
	}
}

// Relations returns all relationship definitions
func (CustomerProfile) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("customer_id", "Customer").Required().OnDelete(schema.CascadeCASCADE).RelatedName("profile").Build(),
	}
}

// Hooks returns model lifecycle hooks
func (CustomerProfile) Hooks() *schema.ModelHooks {
	return nil
}
