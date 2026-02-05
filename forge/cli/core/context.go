package core

import (
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/log"
	"github.com/spf13/cobra"
)

// Context provides command execution context with shared dependencies
type Context struct {
	Config   *config.Config
	Logger   *log.Logger
	Database *db.DB
	Cmd      *cobra.Command
}

// NewContext creates a new command context
func NewContext() *Context {
	return &Context{
		Config: config.NewConfig(),
	}
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
