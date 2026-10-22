package development

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
)

// selected checks execute in the requested directory.
func TestCheckRunsSelectedQualityToolsInChosenDirectory(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(dir, "tools.log")
	for _, name := range []string{"go", "staticcheck", "golangci-lint"} {
		script := "#!/bin/sh\nprintf '%s %s %s\\n' '" + name + "' \"$*\" \"$PWD\" >> \"$TOOL_LOG\"\n"
		if err := os.WriteFile(filepath.Join(bin, name), []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TOOL_LOG", log)
	cases := []struct {
		name  string
		flags []string
		want  []string
	}{
		{"default vet", nil, []string{"go vet ./... "}},
		{"explicit vet", []string{"--vet"}, []string{"go vet ./... "}},
		{"staticcheck", []string{"--vet=false", "--staticcheck"}, []string{"staticcheck ./... "}},
		{"golangci lint", []string{"--vet=false", "--golangci-lint"}, []string{"golangci-lint run "}},
		{"all checks", []string{"--all"}, []string{"go vet ./... ", "staticcheck ./... ", "golangci-lint run "}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(log, nil, 0600); err != nil {
				t.Fatal(err)
			}
			command := NewCheckCommand()
			def := command.Definition()
			if err := def.ParseFlags(append(tc.flags, "--path", dir)); err != nil {
				t.Fatal(err)
			}
			if err := command.Execute(&core.Context{Cmd: def}, nil); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			if len(lines) != len(tc.want) {
				t.Fatalf("calls: %q", data)
			}
			for i, want := range tc.want {
				if lines[i] != want+dir {
					t.Fatalf("call %d = %q, want %q", i, lines[i], want+dir)
				}
			}
		})
	}
}

// the shell accepts commands and exits on request.
func TestShellStartsInteractiveSessionAndExits(t *testing.T) {
	input := filepath.Join(t.TempDir(), "input")
	if err := os.WriteFile(input, []byte("help\nexit\n"), 0600); err != nil {
		t.Fatal(err)
	}
	in, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	oldIn := os.Stdin
	os.Stdin = in
	defer func() { os.Stdin = oldIn }()
	output := captureDevelopmentStdout(t, func() error { return NewShellCommand().Execute(&core.Context{}, nil) })
	for _, part := range []string{"forge Interactive Shell", "forge> ", "Available commands:", "Goodbye!"} {
		if !strings.Contains(output, part) {
			t.Fatalf("missing %q in %q", part, output)
		}
	}
}

func captureDevelopmentStdout(t *testing.T, run func() error) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	runErr := run()
	w.Close()
	os.Stdout = old
	var out bytes.Buffer
	_, err = out.ReadFrom(r)
	r.Close()
	if err != nil {
		t.Fatal(err)
	}
	if runErr != nil {
		t.Fatal(runErr)
	}
	return out.String()
}

// test invokes the Go test runner for the supplied project directory.
func TestTestRunsProjectGoTests(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(dir, "go.log")
	script := "#!/bin/sh\nprintf '%s %s\\n' \"$*\" \"$PWD\" > \"$TOOL_LOG\"\n"
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TOOL_LOG", log)
	command := NewTestCommand()
	def := command.Definition()
	def.SetOut(&bytes.Buffer{})
	def.SetContext(context.Background())
	if err := command.Execute(&core.Context{Cmd: def}, []string{dir}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "test ./... "+dir+"\n" {
		t.Fatalf("go invocation: %q", data)
	}
}
