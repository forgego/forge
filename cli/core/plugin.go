package core

import (
	"github.com/forgego/forge/registry"
)

// CLIPlugin allows plugins to add CLI commands
type CLIPlugin interface {
	registry.Plugin
	// Commands returns commands provided by this plugin
	Commands() []Command
	// CommandGroups returns command groups provided by this plugin
	CommandGroups() []CommandGroup
}
