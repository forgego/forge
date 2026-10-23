package core

import (
	"context"
	"fmt"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/principal"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/log"
	"github.com/spf13/cobra"
)

// Context provides command execution context with shared dependencies
type Context struct {
	Config    *config.Config
	Logger    *log.Logger
	Database  *db.DB
	Cmd       *cobra.Command
	Principal *principal.Principal
}

// PrincipalFromContext makes a principal already attached to a context
// available to CLI command setup without changing authorization behavior.
func (c *Context) PrincipalFromContext(ctx context.Context) *Context {
	if p, ok := principal.PrincipalFromContext(ctx); ok {
		c.Principal = &p
	}
	return c
}

// NewContext creates a new command context with default configuration.
func NewContext() *Context {
	return NewContextWithConfig(config.NewConfig())
}

// NewContextWithConfig creates a new command context from the CLI configuration.
func NewContextWithConfig(cfg *config.Config) *Context {
	return &Context{
		Config:    cfg,
		Principal: principalPointer(principal.System("cli")),
	}
}

func principalPointer(p principal.Principal) *principal.Principal { return &p }

// AssumeUser resolves an explicitly selected account and refuses identities
// that cannot authenticate.
func (c *Context) AssumeUser(ctx context.Context, users repository.UserRepository, username string) error {
	user, err := users.GetByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("resolve CLI principal %q: %w", username, err)
	}
	p, err := principal.FromUser(user)
	if err != nil {
		return fmt.Errorf("resolve CLI principal %q: %w", username, err)
	}
	c.Principal = &p
	return nil
}

// WithLogger sets the logger in the context
func (c *Context) WithLogger(logger *log.Logger) *Context {
	c.Logger = logger
	return c
}

// WithDatabase sets the database in the context
func (c *Context) WithDatabase(database *db.DB) *Context {
	c.Database = database
	return c
}
