package migrations

import (
	"testing"

	"github.com/forgego/forge/cli/core"
	forgeerrors "github.com/forgego/forge/errors"
)

func TestSquashCommand_Execute_NotImplemented(t *testing.T) {
	cmd := NewSquashCommand()
	def := cmd.Definition()
	tmpDir := t.TempDir()
	if err := def.Flags().Set("path", tmpDir); err != nil {
		t.Fatalf("failed to set path flag: %v", err)
	}

	ctx := &core.Context{
		Cmd: def,
	}

	err := cmd.Execute(ctx, []string{"000001", "000002", "squashed"})
	if err == nil {
		t.Fatal("expected error from SquashCommand.Execute, got nil")
	}
	if !forgeerrors.IsNotImplemented(err) {
		t.Fatalf("expected NotImplementedError from SquashCommand.Execute, got: %v", err)
	}
}
