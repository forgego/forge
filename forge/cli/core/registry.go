package core

import (
	"fmt"
	"sync"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/log"
	"github.com/spf13/cobra"
)

// Registry manages command registration and discovery
type Registry struct {
	commands map[string]Command
	groups   map[string]CommandGroup
	plugins  []CLIPlugin
	mu       sync.RWMutex
}

var globalRegistry = &Registry{
	commands: make(map[string]Command),
	groups:   make(map[string]CommandGroup),
	plugins:  []CLIPlugin{},
}

// GetRegistry returns the global command registry
func GetRegistry() *Registry {
	return globalRegistry
}

// RegisterCommand registers a command
func (r *Registry) RegisterCommand(name string, cmd Command) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[name] = cmd
}

// RegisterGroup registers a command group
func (r *Registry) RegisterGroup(name string, group CommandGroup) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.groups[name] = group
}

// RegisterPlugin registers a CLI plugin
func (r *Registry) RegisterPlugin(plugin CLIPlugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Register commands from plugin
	for _, cmd := range plugin.Commands() {
		cmdDef := cmd.Definition()
		if cmdDef == nil {
			continue
		}
		r.commands[cmdDef.Use] = cmd
	}

	// Register command groups from plugin
	for _, group := range plugin.CommandGroups() {
		groupDef := group.Definition()
		if groupDef == nil {
			continue
		}
		r.groups[groupDef.Use] = group
	}

	r.plugins = append(r.plugins, plugin)
	return nil
}

// GetCommand retrieves a command by name
func (r *Registry) GetCommand(name string) (Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, exists := r.commands[name]
	if !exists {
		return nil, fmt.Errorf("command %s not found", name)
	}
	return cmd, nil
}

// GetGroup retrieves a command group by name
func (r *Registry) GetGroup(name string) (CommandGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	group, exists := r.groups[name]
	if !exists {
		return nil, fmt.Errorf("command group %s not found", name)
	}
	return group, nil
}

// BuildRootCommand builds the root cobra command with all registered commands
func (r *Registry) BuildRootCommand() *cobra.Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rootCmd := &cobra.Command{
		Use:   "forge",
		Short: "Forge framework CLI",
		Long:  "Command-line interface for Forge framework",
	}

	// Add all standalone commands
	for _, cmd := range r.commands {
		cobraCmd := cmd.Definition()
		if cobraCmd != nil {
			// Wrap RunE to use our Execute method
			cobraCmd.RunE = func(c *cobra.Command, args []string) error {
				ctx := NewContext()
				ctx.Cmd = c
				// Try to create logger if needed
				if ctx.Config != nil {
					logger, err := createLogger(ctx.Config)
					if err == nil {
						ctx.WithLogger(logger)
					}
				}
				// Execute using our interface
				return cmd.Execute(ctx, args)
			}
			rootCmd.AddCommand(cobraCmd)
		}
	}

	// Add all command groups
	for _, group := range r.groups {
		groupCmd := group.Definition()
		if groupCmd != nil {
			// Wrap RunE for the group command itself (if it has one)
			if groupCmd.RunE != nil {
				groupCmd.RunE = func(c *cobra.Command, args []string) error {
					ctx := NewContext()
					ctx.Cmd = c
					if ctx.Config != nil {
						logger, err := createLogger(ctx.Config)
						if err == nil {
							ctx.WithLogger(logger)
						}
					}
					return group.Execute(ctx, args)
				}
			}

			// Add subcommands to the group
			for _, subCmd := range group.Commands() {
				subCobraCmd := subCmd.Definition()
				if subCobraCmd != nil {
					subCobraCmd.RunE = func(c *cobra.Command, args []string) error {
						ctx := NewContext()
						ctx.Cmd = c
						if ctx.Config != nil {
							logger, err := createLogger(ctx.Config)
							if err == nil {
								ctx.WithLogger(logger)
							}
						}
						return subCmd.Execute(ctx, args)
					}
					groupCmd.AddCommand(subCobraCmd)
				}
			}

			rootCmd.AddCommand(groupCmd)
		}
	}

	return rootCmd
}

// Helper function to create logger
func createLogger(cfg *config.Config) (*log.Logger, error) {
	if cfg == nil {
		return nil, nil
	}
	settings := config.LoadSettings(cfg)
	return log.NewLogger(settings.App.Debug)
}
