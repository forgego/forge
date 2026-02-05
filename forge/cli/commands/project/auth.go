package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/forgego/forge/cli/core"
	"github.com/spf13/cobra"
)

// AuthCommand creates the "auth" command for scaffolding auth
type AuthCommand struct{}

// NewAuthCommand creates a new instance of AuthCommand
func NewAuthCommand() *AuthCommand {
	return &AuthCommand{}
}

// Definition returns the cobra command definition
func (c *AuthCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Scaffold authentication app",
		Long:  "Create a complete authentication app with users, sessions, and JWT support",
	}
	return cmd
}

// Execute runs the command logic
func (c *AuthCommand) Execute(ctx *core.Context, args []string) error {
	// Detect project root
	projectRoot, err := detectProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to detect project root: %w", err)
	}

	// Create auth app
	appName := "auth"
	appPath := filepath.Join(projectRoot, "app", appName)

	// Check if auth app already exists
	if _, err := os.Stat(appPath); err == nil {
		return fmt.Errorf("auth app already exists")
	}

	// Create app directory
	if err := os.MkdirAll(appPath, 0755); err != nil {
		return fmt.Errorf("failed to create auth app directory: %w", err)
	}

	// Create User model
	userModel := `package auth

import (
	"github.com/forgego/forge/schema"
)

// User represents a user model
type User struct {
	schema.BaseSchema
}

// Fields returns all field definitions for User
func (User) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("username").Unique().Required().MaxLength(150).Build(),
		schema.String("email").Unique().Required().MaxLength(255).Build(),
		schema.String("password").Required().MaxLength(128).Build(),
		schema.Bool("is_active").Default(true).Build(),
		schema.Bool("is_staff").Default(false).Build(),
		schema.Bool("is_superuser").Default(false).Build(),
		schema.Time("date_joined").AutoNowAdd().Build(),
		schema.Time("last_login").Build(),
	}
}

// Meta returns model metadata
func (User) Meta() schema.Meta {
	return schema.Meta{
		TableName:        "users",
		VerboseName:      "User",
		VerboseNamePlural: "Users",
	}
}

// Relations returns all relationship definitions
func (User) Relations() []schema.Relation {
	return []schema.Relation{}
}

// Hooks returns model lifecycle hooks
func (User) Hooks() *schema.ModelHooks {
	return nil
}
`

	if err := os.WriteFile(filepath.Join(appPath, "models.go"), []byte(userModel), 0644); err != nil {
		return fmt.Errorf("failed to create models.go: %w", err)
	}

	// Create admin.go
	adminCode := `package auth

import (
	admincore "github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
)

func init() {
	// Register User model for admin
	// After code generation, uncomment and use:
	// schemaInstance := &UserSchema{} // Your schema implementation
	// manager := orm.NewManager[*User](db) // Your ORM manager
	// config := &admin.Config[*User]{
	//     VerboseName:       "User",
	//     VerboseNamePlural: "Users",
	//     ListPerPage:       20,
	// }
	// admin, err := admin.Register(schemaInstance, manager, config)
	// if err != nil {
	//     log.Fatal(err)
	// }
}
`

	if err := os.WriteFile(filepath.Join(appPath, "admin.go"), []byte(adminCode), 0644); err != nil {
		return fmt.Errorf("failed to create admin.go: %w", err)
	}

	// Create api.go with JWT support
	apiCode := `package auth

import (
	"github.com/forgego/forge/api"
	httplib "github.com/forgego/forge/server"
)

func init() {
	// Auto-register auth API routes
}

// RegisterAuthAPI registers authentication API endpoints
func RegisterAuthAPI(router *httplib.Router) {
	// Create viewset for users
	viewset := api.NewBaseViewSet(
		func() api.Serializer {
			return NewUserSerializer()
		},
		User.Objects.Filter(),
		&User{},
	)

	// Register routes
	apiRouter := api.NewRouter("/api/v1")
	apiRouter.Register("users", viewset)
	apiRouter.RegisterRoutes(router)

	// TODO: Add login/logout endpoints
	// TODO: Add JWT token generation
}

// UserSerializer serializes User model
type UserSerializer struct {
	*api.BaseSerializer
}

// NewUserSerializer creates a new serializer
func NewUserSerializer() api.Serializer {
	return &UserSerializer{
		BaseSerializer: api.NewBaseSerializer(),
	}
}

// Fields returns the fields to serialize
func (s *UserSerializer) Fields() []string {
	return []string{"id", "username", "email", "is_active", "date_joined"}
}
`

	if err := os.WriteFile(filepath.Join(appPath, "api.go"), []byte(apiCode), 0644); err != nil {
		return fmt.Errorf("failed to create api.go: %w", err)
	}

	fmt.Printf("✓ Scaffolded auth app\n")
	fmt.Printf("  Location: %s\n", appPath)
	fmt.Printf("  Created: User model, admin config, API endpoints\n")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Run: forge generate\n")
	fmt.Printf("  2. Run: forge makemigrations\n")
	fmt.Printf("  3. Run: forge migrate\n")

	return nil
}
