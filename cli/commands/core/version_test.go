package core

import (
	"bytes"
	"os"
	"testing"

	clicore "github.com/forgego/forge/cli/core"
)

// version prints the module version.
func TestVersionPrintsForgeVersion(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	runErr := NewVersionCommand().Execute(nil, nil)
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
	want := "Forge " + clicore.Version() + "\n"
	if out.String() != want {
		t.Fatalf("version output = %q, want %q", out.String(), want)
	}
}
