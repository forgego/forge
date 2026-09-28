// Package blog is the supported schema DSL fixture for code generation.
//
// Every declaration here uses a form that `forge generate --strict` accepts:
// functional field constructors with option arguments, relations declared with
// the functional relation constructors, and Meta values written as literals.
// codegen/dsl_contract_test.go generates this package twice, compiles it in a
// temporary module, and compares its schema facts with the SQL produced by
// makemigrations. Keep it representative rather than exhaustive.
package blog

import (
	"context"
	"errors"

	"github.com/forgego/forge/schema"
)

// Author has a scalar field of every supported type.
type Author struct {
	AuthorGenerated
}

func (Author) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("username", schema.Required(), schema.MaxLength(150), schema.MinLength(3), schema.Unique()),
		schema.EmailField("email", schema.Required(), schema.Unique()),
		schema.URLField("website", schema.Optional()),
		schema.UUIDField("external_ref", schema.Optional()),
		schema.TextField("bio", schema.Optional(), schema.HelpText("Short biography")),
		schema.BoolField("is_active", schema.Default(true)),
		schema.Int32Field("login_count", schema.Default(0), schema.MinValue(0)),
		schema.Float64Field("reputation", schema.Default(1.5)),
		schema.Float32Field("ratio", schema.Optional()),
		schema.DecimalField("balance", schema.MaxDigits(12), schema.DecimalPlaces(2), schema.Default(0)),
		schema.JSONField("settings", schema.Optional()),
		schema.BytesField("avatar", schema.Optional()),
		schema.DateField("birth_date", schema.Optional()),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.DateTimeField("updated_at", schema.AutoNow()),
	}
}

func (Author) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "blog_authors",
		VerboseName:       "Author",
		VerboseNamePlural: "Authors",
		OrderBy:           []string{"username"},
	}
}

func (Author) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Author) Hooks() *schema.ModelHooks {
	return schema.NewModelHooks().WithBeforeSave(func(ctx context.Context, instance interface{}) error {
		if author, ok := instance.(*Author); ok && author.Username == "" {
			return errors.New("username is required")
		}
		return nil
	})
}

// Post references Author through a foreign key and carries table constraints.
type Post struct {
	PostGenerated
}

func (Post) Fields() []schema.Field {
	fields := []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required(), schema.MaxLength(200)),
		schema.StringField("slug", schema.Required(), schema.MaxLength(200)),
		schema.StringField("status", schema.Required(), schema.MaxLength(20), schema.Default("draft")),
		schema.Int64Field("author_id", schema.Required()),
	}
	fields = append(fields, schema.Int32Field("word_count", schema.Default(0)))
	return fields
}

func (Post) Meta() schema.Meta {
	return schema.Meta{
		TableName: "blog_posts",
		OrderBy:   []string{"-id"},
		Indexes: []schema.Index{
			{Name: "blog_posts_status_idx", Fields: []string{"status"}},
		},
		Constraints: []schema.Constraint{
			{Name: "blog_posts_word_count_nonnegative", Type: "CHECK", Condition: "word_count >= 0"},
			{Name: "blog_posts_author_slug_key", Type: "UNIQUE", Fields: []string{"author_id", "slug"}},
		},
	}
}

func (Post) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("author_id", "Author", schema.OnDelete(schema.CascadeCASCADE), schema.RelatedName("posts")),
	}
}

// Profile is a one-to-one extension of Author.
type Profile struct {
	ProfileGenerated
}

func (Profile) Fields() []schema.Field {
	var fields = []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("author_id", schema.Required(), schema.Unique()),
		schema.StringField("display_name", schema.Optional(), schema.MaxLength(100)),
	}
	return fields
}

func (Profile) Meta() schema.Meta {
	return schema.Meta{TableName: "blog_profiles"}
}

func (Profile) Relations() []schema.Relation {
	return []schema.Relation{
		schema.OneToOneField("author_id", "Author", schema.OnDelete(schema.CascadeCASCADE)),
	}
}
