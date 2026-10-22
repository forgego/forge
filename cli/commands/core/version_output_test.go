package core

import (
	"io"
	"os"
	"testing"

	clientcore "github.com/forgego/forge/cli/core"
)

// The CLI prints the module version selected by build information.
func TestVersionCommandPrintsInstalledModuleVersion(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = old; _ = r.Close(); _ = w.Close() })
	if err := NewVersionCommand().Execute(nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "Forge "+clientcore.Version()+"\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
