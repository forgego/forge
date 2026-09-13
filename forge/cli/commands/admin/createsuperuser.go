package admin

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity"
	"github.com/forgego/forge/identity/service"
	"github.com/spf13/cobra"
)

// CreateSuperUserCommand creates the createsuperuser command
type CreateSuperUserCommand struct{}

// NewCreateSuperUserCommand creates a new instance of CreateSuperUserCommand
func NewCreateSuperUserCommand() *CreateSuperUserCommand {
	return &CreateSuperUserCommand{}
}

// Definition returns the cobra command definition
func (c *CreateSuperUserCommand) Definition() *cobra.Command {
	return &cobra.Command{
		Use:   "createsuperuser",
		Short: "Create admin superuser",
		Long:  "Create a superuser account for the admin interface",
	}
}

// Execute runs the command logic
func (c *CreateSuperUserCommand) Execute(ctx *core.Context, args []string) error {
	// Connect to database
	database, err := db.NewDBFromConfig(ctx.Config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	// Prompt for user details
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Read password (without echo)
	fmt.Print("Password: ")
	password, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	fmt.Print("\nPassword (again): ")
	passwordConfirm, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read password confirmation: %w", err)
	}
	if password != passwordConfirm {
		return fmt.Errorf("passwords do not match")
	}

	// Create superuser using new user system
	cmdCtx := context.Background()
	userSystem, err := identity.SetupIdentitySystem(database, nil)
	if err != nil {
		return fmt.Errorf("failed to setup user system: %w", err)
	}

	// Create superuser request
	createReq := &service.CreateUserRequest{
		Username: username,
		Email:    email,
		Password: password,
		IsStaff:  true,
		IsActive: true,
	}

	// Create user first
	user, err := userSystem.UserService.CreateUser(cmdCtx, createReq)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Update to superuser
	isSuperuser := true
	updateReq := &service.UpdateUserRequest{
		IsSuperuser: &isSuperuser,
	}
	user, err = userSystem.UserService.UpdateUser(cmdCtx, user.ID, updateReq)
	if err != nil {
		return fmt.Errorf("failed to set superuser flag: %w", err)
	}

	fmt.Printf("\n✓ Superuser created successfully (ID: %d, Username: %s)\n", user.ID, user.Username)
	return nil
}

// readPassword reads a password from stdin without echoing
func readPassword() (string, error) {
	// For MVP, use simple input (no password hiding)
	// Full implementation would use term package for password hiding
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(password), nil
}
