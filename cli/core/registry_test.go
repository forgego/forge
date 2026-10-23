package core

import (
	"testing"

	"github.com/forgego/forge/log"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type registryTestCommand struct {
	use  string
	seen *Context
}

func (c *registryTestCommand) Definition() *cobra.Command { return &cobra.Command{Use: c.use} }
func (c *registryTestCommand) Execute(ctx *Context, _ []string) error {
	c.seen = ctx
	return nil
}

type registryTestGroup struct {
	*registryTestCommand
	commands []Command
}

func (g *registryTestGroup) Commands() []Command { return g.commands }
func (g *registryTestGroup) Definition() *cobra.Command {
	return &cobra.Command{Use: g.use, RunE: func(*cobra.Command, []string) error { return nil }}
}

func TestBuildRootCommandProvidesContextAndLogger(t *testing.T) {
	standalone := &registryTestCommand{use: "standalone"}
	groupCommand := &registryTestCommand{use: "group"}
	child := &registryTestCommand{use: "child"}
	group := &registryTestGroup{registryTestCommand: groupCommand, commands: []Command{child}}
	registry := &Registry{commands: map[string]Command{"standalone": standalone}, groups: map[string]CommandGroup{"group": group}}

	root := registry.BuildRootCommand()
	groupCmd, _, err := root.Find([]string{"group"})
	require.NoError(t, err)
	standaloneCmd, _, err := root.Find([]string{"standalone"})
	require.NoError(t, err)
	childCmd, _, err := root.Find([]string{"group", "child"})
	require.NoError(t, err)
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
		impl *registryTestCommand
	}{
		{"standalone", standaloneCmd, standalone},
		{"group", groupCmd, groupCommand},
		{"child", childCmd, child},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, tc.cmd.RunE(tc.cmd, nil))
			require.NotNil(t, tc.impl.seen)
			require.Same(t, tc.cmd, tc.impl.seen.Cmd)
			require.NotNil(t, tc.impl.seen.Config)
			require.NotNil(t, tc.impl.seen.Logger)
			require.IsType(t, (*log.Logger)(nil), tc.impl.seen.Logger)
		})
	}
}
