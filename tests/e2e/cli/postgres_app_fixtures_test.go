package cli

// Application code the release-gate journey writes into a project scaffolded
// by `forge new`, as a developer would. Everything else in the project comes
// from `forge new`, `forge add app` and `forge generate`, unedited.

// trackerModels declares the models; the journey later replaces
// trackerExtraTaskField with a new field to exercise a schema change.
const trackerModels = `package work

import "github.com/forgego/forge/schema"

// Member is a user of the API; members authenticate with their API token.
type Member struct {
	MemberGenerated
}

func (Member) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100)),
		schema.StringField("api_token", schema.Required(), schema.MaxLength(64), schema.Unique(), schema.WriteOnly()),
	}
}

func (Member) Meta() schema.Meta {
	return schema.Meta{TableName: "members", VerboseName: "Member"}
}

func (Member) Relations() []schema.Relation { return nil }

func (Member) Hooks() *schema.ModelHooks { return nil }

type Project struct {
	ProjectGenerated
}

func (Project) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100)),
		schema.TextField("description", schema.Optional()),
		schema.DateTimeField("created_at", schema.AutoNowAdd()),
		schema.DateTimeField("updated_at", schema.AutoNow()),
	}
}

func (Project) Meta() schema.Meta {
	return schema.Meta{TableName: "projects", VerboseName: "Project"}
}

func (Project) Relations() []schema.Relation { return nil }

func (Project) Hooks() *schema.ModelHooks { return nil }

// Task belongs to a project and is owned by a member.
type Task struct {
	TaskGenerated
}

func (Task) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("project_id", schema.Required()),
		schema.Int64Field("owner_id", schema.Required()),
		schema.StringField("title", schema.Required(), schema.MaxLength(200)),
		schema.BoolField("done", schema.Default(false)),
		schema.DateTimeField("created_at", schema.AutoNowAdd()),
		// EXTRA TASK FIELD
	}
}

func (Task) Meta() schema.Meta {
	return schema.Meta{TableName: "tasks", VerboseName: "Task"}
}

func (Task) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("project_id", "Project", schema.OnDelete(schema.CascadeCASCADE), schema.RelatedName("tasks")),
		schema.ForeignKeyField("owner_id", "Member", schema.OnDelete(schema.CascadePROTECT), schema.RelatedName("tasks")),
	}
}

func (Task) Hooks() *schema.ModelHooks { return nil }
`

const trackerExtraTaskField = "// EXTRA TASK FIELD"

// trackerRoutes wires the generated ViewSets: every request needs a member's
// API token, and only a task's owner may change or delete it. Members have
// no API; the journey creates them in the database.
const trackerRoutes = `package work

import (
	"context"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/server"
)

// Register binds the generated managers to the database and serves the
// project and task APIs under /api/v1.
func Register(router *server.Router, database *db.DB) {
	MemberObjects.SetDB(database)
	ProjectObjects.SetDB(database)
	TaskObjects.SetDB(database)

	api.SetDefaultAuthentication(authentication.NewTokenAuthentication(memberForToken))
	api.SetDefaultPermissions(permissions.NewIsAuthenticated())

	tasks := NewTaskViewSet()
	tasks.Permissions = []permissions.Permission{
		permissions.NewIsAuthenticated(),
		permissions.NewIsOwnerOrReadOnly("owner_id"),
	}

	apiRouter := api.NewRouter("/api/v1")
	apiRouter.Register("projects", NewProjectViewSet())
	apiRouter.Register("tasks", tasks)
	apiRouter.RegisterRoutes(router)
}

func memberForToken(token string) (interface{}, error) {
	qs, err := MemberObjects.Filter(MemberFieldsInstance.ApiToken.Eq(token))
	if err != nil {
		return nil, err
	}
	members, err := qs.Limit(1).All(context.Background())
	if err != nil || len(members) == 0 {
		return nil, err
	}
	return members[0], nil
}
`

// trackerTxCommand creates a project and its first task in one transaction
// through the generated managers. Given an owner that does not exist, the
// second write violates the foreign key and the whole transaction must roll
// back. It exits 3 when the transaction fails.
const trackerTxCommand = `package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"

	"tracker/app/work"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: txcheck <project name> <owner id>")
		os.Exit(2)
	}
	ownerID, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	database, err := db.NewDBFromConfig(config.NewConfig())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer database.Close()
	work.ProjectObjects.SetDB(database)
	work.TaskObjects.SetDB(database)

	ctx := context.Background()
	err = database.WithTx(ctx, func(tx *db.Tx) error {
		project := &work.Project{}
		project.Name = os.Args[1]
		if err := work.ProjectObjects.WithTx(tx).Create(ctx, project); err != nil {
			return fmt.Errorf("create project: %w", err)
		}
		task := &work.Task{}
		task.ProjectId = project.Id
		task.OwnerId = ownerID
		task.Title = "first task"
		if err := work.TaskObjects.WithTx(tx).Create(ctx, task); err != nil {
			return fmt.Errorf("create task: %w", err)
		}
		return nil
	})
	if err != nil {
		fmt.Println("rolled back:", err)
		os.Exit(3)
	}
	fmt.Println("committed")
}
`
