package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCapcheck(t *testing.T) {
	tests := []struct {
		name         string
		markdown     string
		goFiles      map[string]string // relPath -> content
		wantCode     int
		wantProblems []string
		wantSummary  string
	}{
		{
			name: "missing evidence",
			markdown: strings.Join([]string{
				"# Capability Status",
				"",
				"| Capability | Tier | Evidence |",
				"| --- | --- | --- |",
				"| Feature A | Supported | |",
			}, "\n") + "\n",
			goFiles: map[string]string{
				"feature_test.go": "package test\n",
			},
			wantCode: 1,
			wantProblems: []string{
				"status.md:5: missing evidence",
			},
		},
		{
			name: "unknown test",
			markdown: strings.Join([]string{
				"# Capability Status",
				"",
				"| Capability | Tier | Evidence |",
				"| --- | --- | --- |",
				"| Feature A | Supported | `TestDoesNotExist` |",
			}, "\n") + "\n",
			goFiles: map[string]string{
				"feature_test.go": "package test\nfunc TestOther(t *testing.T) {}\n",
			},
			wantCode: 1,
			wantProblems: []string{
				"status.md:5: unknown test TestDoesNotExist",
			},
		},
		{
			name: "prefix match",
			markdown: strings.Join([]string{
				"# Capability Status",
				"",
				"| Capability | Tier | Evidence |",
				"| --- | --- | --- |",
				"| Feature A | Supported | `TestPrefix_*` |",
			}, "\n") + "\n",
			goFiles: map[string]string{
				"feature_test.go": "package test\nfunc TestPrefix_Something(t *testing.T) {}\n",
			},
			wantCode:    0,
			wantSummary: "capcheck: 1 supported rows ok\n",
		},
		{
			name: "non-Supported row being ignored",
			markdown: strings.Join([]string{
				"# Capability Status",
				"",
				"| Capability | Tier | Evidence |",
				"| --- | --- | --- |",
				"| Feature A | Partially tested | |",
				"| Feature B | Experimental | `TestDoesNotExist` |",
				"| Feature C | Not implemented | `TestAlsoMissing` |",
				"| Feature D | Not supported | None |",
			}, "\n") + "\n",
			goFiles: map[string]string{
				"feature_test.go": "package test\n",
			},
			wantCode:    0,
			wantSummary: "capcheck: 0 supported rows ok\n",
		},
		{
			name: "unknown prefix match fails",
			markdown: strings.Join([]string{
				"# Capability Status",
				"",
				"| Capability | Tier | Evidence |",
				"| --- | --- | --- |",
				"| Feature A | Supported | `TestPrefix_*` |",
			}, "\n") + "\n",
			goFiles: map[string]string{
				"feature_test.go": "package test\nfunc TestOther_Thing(t *testing.T) {}\n",
			},
			wantCode: 1,
			wantProblems: []string{
				"status.md:5: unknown test TestPrefix_*",
			},
		},
		{
			name: "non-test evidence passes when non-empty",
			markdown: strings.Join([]string{
				"# Capability Status",
				"",
				"| Capability | Tier | Evidence |",
				"| --- | --- | --- |",
				"| Feature A | Supported | `unit-tests` matrix, all other jobs |",
				"| Feature B | Supported | [`go.mod`](https://example.com) |",
				"| Feature C | Supported | database/sql |",
			}, "\n") + "\n",
			goFiles: map[string]string{
				"feature_test.go": "package test\n",
			},
			wantCode:    0,
			wantSummary: "capcheck: 3 supported rows ok\n",
		},
		{
			name: "skip ignored directories",
			markdown: strings.Join([]string{
				"# Capability Status",
				"",
				"| Capability | Tier | Evidence |",
				"| --- | --- | --- |",
				"| Feature A | Supported | `TestInNodeModules` |",
			}, "\n") + "\n",
			goFiles: map[string]string{
				"node_modules/pkg/foo_test.go":           "package pkg\nfunc TestInNodeModules(t *testing.T) {}\n",
				"docs-site/node_modules/pkg/bar_test.go": "package pkg\nfunc TestInDocsNodeModules(t *testing.T) {}\n",
				".git/hooks_test.go":                     "package git\nfunc TestInGit(t *testing.T) {}\n",
			},
			wantCode: 1,
			wantProblems: []string{
				"status.md:5: unknown test TestInNodeModules",
			},
		},
		{
			name: "table without Evidence column is ignored",
			markdown: strings.Join([]string{
				"# Capability Status",
				"",
				"| Tier | Meaning |",
				"| --- | --- |",
				"| Supported | Implemented and tested |",
			}, "\n") + "\n",
			goFiles: map[string]string{
				"feature_test.go": "package test\n",
			},
			wantCode:    0,
			wantSummary: "capcheck: 0 supported rows ok\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			// Write status.md
			statusPath := filepath.Join(dir, "status.md")
			if err := os.WriteFile(statusPath, []byte(tt.markdown), 0o644); err != nil {
				t.Fatalf("failed to write status.md: %v", err)
			}

			// Write Go test files
			for relPath, content := range tt.goFiles {
				fullPath := filepath.Join(dir, filepath.FromSlash(relPath))
				if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
					t.Fatalf("failed to create dir for %s: %v", relPath, err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
					t.Fatalf("failed to write %s: %v", relPath, err)
				}
			}

			var out bytes.Buffer
			code, err := run(statusPath, dir, &out)
			if err != nil {
				t.Fatalf("run returned unexpected error: %v", err)
			}
			if code != tt.wantCode {
				t.Fatalf("got exit code %d, want %d; output:\n%s", code, tt.wantCode, out.String())
			}

			output := out.String()
			if tt.wantCode == 0 {
				if output != tt.wantSummary {
					t.Errorf("got summary %q, want %q", output, tt.wantSummary)
				}
			} else {
				for _, wp := range tt.wantProblems {
					if !strings.Contains(output, wp) {
						t.Errorf("output does not contain expected problem %q\ngot:\n%s", wp, output)
					}
				}
			}
		})
	}
}
