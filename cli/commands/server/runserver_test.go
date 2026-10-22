package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
)

func runWithFakeGo(t *testing.T, flags []string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd", "server"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd", "server", "main.go"), []byte("package main\nfunc main() {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(root, "invocation.txt")
	script := "#!/bin/sh\nprintf '%s|%s' \"$*\" \"$FORGE_SERVER_PORT\" > \"$FORGE_TEST_INVOCATION\"\n"
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FORGE_TEST_INVOCATION", logPath)
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	command := NewRunServerCommand()
	def := command.Definition()
	if err := def.ParseFlags(flags); err != nil {
		t.Fatal(err)
	}
	if err := command.Execute(&core.Context{Cmd: def}, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRunServerStartsDevelopmentServer(t *testing.T) {
	if got := runWithFakeGo(t, nil); !strings.HasPrefix(got, "run cmd/server/main.go|") {
		t.Fatalf("invocation: %q", got)
	}
}

func TestRunServerPortOverridesConfiguredPort(t *testing.T) {
	t.Setenv("FORGE_SERVER_PORT", "8000")
	if got := runWithFakeGo(t, []string{"--port", "9123"}); got != "run cmd/server/main.go|9123" {
		t.Fatalf("invocation: %q", got)
	}
}
