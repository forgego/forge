package users

import (
	"time"

	"github.com/forgego/forge/schema"
)

type User struct {
	schema.BaseSchema
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	IsActive     bool       `json:"is_active"`
	IsStaff      bool       `json:"is_staff"`
	IsSuperuser  bool       `json:"is_superuser"`
	Avatar       string     `json:"avatar"`
	LastLogin    *time.Time `json:"last_login"`
	DateJoined   time.Time  `json:"date_joined"`
}

func (User) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "users_user",
		VerboseName:       "User",
		VerboseNamePlural: "Users",
	}
}

func (User) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("username", schema.Required(), schema.MaxLength(150),
			schema.HelpText("Unique username")),
		schema.StringField("email", schema.Required(), schema.MaxLength(254)),
		schema.StringField("first_name", schema.MaxLength(150), schema.Optional()),
		schema.StringField("last_name", schema.MaxLength(150), schema.Optional()),
		schema.BoolField("is_active", schema.Default(true)),
		schema.BoolField("is_staff", schema.Default(false)),
		schema.BoolField("is_superuser", schema.Default(false)),
		schema.StringField("avatar", schema.MaxLength(500), schema.Optional()),
		schema.TimeField("last_login", schema.Optional()),
		schema.TimeField("date_joined", schema.AutoNowAdd()),
	}
}

func (User) Relations() []schema.Relation {
	return []schema.Relation{}
}

type Group struct {
	schema.BaseSchema
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Group) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "auth_group",
		VerboseName:       "Group",
		VerboseNamePlural: "Groups",
	}
}

func (Group) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(150),
			schema.HelpText("Group name")),
		schema.StringField("description", schema.MaxLength(500), schema.Optional()),
		schema.JSONField("permissions", schema.Optional()),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Group) Relations() []schema.Relation {
	return []schema.Relation{}
}
