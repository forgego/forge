package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Backtick and fence literals are spelled with hex escapes to keep the
// test fixtures explicit about which characters they contain.
const (
	bt    = "\x60"
	fence = "\x60\x60\x60"
)

func writeFixture(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func checkFixture(t *testing.T, root string) []problem {
	t.Helper()
	_, problems, err := check(root)
	if err != nil {
		t.Fatal(err)
	}
	return problems
}

func TestValidRelativeLinkPasses(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\nSee [guide](docs/guide.md).\n")
	writeFixture(t, root, "docs/guide.md", "# Guide\n")
	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected no problems, got %v", problems)
	}
}

func TestMissingRelativeLinkReportedWithLineNumber(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\nIntro line.\nSee [gone](docs/nope.md).\n")
	problems := checkFixture(t, root)
	if len(problems) != 1 {
		t.Fatalf("expected 1 problem, got %v", problems)
	}
	got := problems[0]
	if got.file != "README.md" || got.line != 4 || got.target != "docs/nope.md" {
		t.Fatalf("unexpected problem: %v", got)
	}
}

func TestExternalAndAnchorLinksSkipped(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\n"+
		"[web](https://example.com/docs/guide.md)\n"+
		"[plain](http://example.com/x.md)\n"+
		"[mail](mailto:someone@example.com)\n"+
		"[anchor](#section)\n")
	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected no problems, got %v", problems)
	}
}

func TestDocsSiteRouteSkipped(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\n[guide](/docs/guides/models)\n")
	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected no problems, got %v", problems)
	}
}

func TestBacktickedPaths(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\n"+
		"See "+bt+"docs/MISSING.md"+bt+" for details.\n"+
		"Run "+bt+"go test ./..."+bt+" to test.\n")
	problems := checkFixture(t, root)
	if len(problems) != 1 {
		t.Fatalf("expected 1 problem, got %v", problems)
	}
	got := problems[0]
	if got.file != "README.md" || got.line != 3 || got.target != "docs/MISSING.md" {
		t.Fatalf("unexpected problem: %v", got)
	}
}

func TestBacktickedPathsMustStartAtRepositoryRoot(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\n"+
		"Skip "+bt+"config/config.yaml"+bt+" and "+bt+"models/post.go"+bt+".\n"+
		"Report "+bt+"docs/MISSING.md"+bt+".\n")
	problems := checkFixture(t, root)
	if len(problems) != 1 {
		t.Fatalf("expected 1 problem, got %v", problems)
	}
	if got := problems[0]; got.file != "README.md" || got.line != 4 || got.target != "docs/MISSING.md" {
		t.Fatalf("unexpected problem: %v", got)
	}
}

func TestIgnoredLineSkipped(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\n"+
		"[gone](docs/nope.md) <!-- doclinks:ignore -->\n"+
		"See "+bt+"docs/MISSING.md"+bt+" <!-- doclinks:ignore -->\n")
	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected no problems, got %v", problems)
	}
}

func TestFencedCodeBlocksIgnored(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\n"+
		fence+"\n"+
		"[gone](docs/nope.md)\n"+
		"see "+bt+"docs/nope.md"+bt+"\n"+
		fence+"\n")
	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected no problems, got %v", problems)
	}
}

func TestFragmentStripped(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "docs/REVIEWING.md", "# Reviewing\n")
	writeFixture(t, root, "docs/page.md", "# Page\n\nSee [review](REVIEWING.md#acceptance-review).\n")
	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected no problems, got %v", problems)
	}
}

func TestQueryStripped(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "docs/guide.md", "# Guide\n")
	writeFixture(t, root, "docs/page.md", "# Page\n\nSee [guide](guide.md?x=1).\n")
	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected no problems, got %v", problems)
	}
}

func TestDirectoryLinkPasses(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\nSee [sample](examples/ecommerce/).\n")
	if err := os.MkdirAll(filepath.Join(root, "examples", "ecommerce"), 0o755); err != nil {
		t.Fatal(err)
	}
	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected no problems, got %v", problems)
	}
}

func TestRunSuccessOutput(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\nSee [guide](docs/guide.md).\n")
	writeFixture(t, root, "docs/guide.md", "# Guide\n")
	var out bytes.Buffer
	code, err := run(root, &out)
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if got := out.String(); got != "doclinks: 2 files, 0 missing\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRunReportsProblems(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\nSee [gone](docs/nope.md).\n")
	var out bytes.Buffer
	code, err := run(root, &out)
	if err != nil {
		t.Fatal(err)
	}
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if got := out.String(); got != "README.md:3: missing docs/nope.md\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestCollectFilesScope(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n")
	writeFixture(t, root, "docs/top.md", "# Top\n")
	writeFixture(t, root, "docs/nested/deep.md", "# Deep\n")
	writeFixture(t, root, "skills/demo/SKILL.md", "# Skill\n")
	writeFixture(t, root, "skills/demo/references/ref.md", "# Ref\n")
	writeFixture(t, root, "skills/demo/references/notes.txt", "notes\n")
	writeFixture(t, root, "tests/README.md", "# Tests\n")
	files, err := collectFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	var rels []string
	for _, f := range files {
		rel, err := filepath.Rel(root, f)
		if err != nil {
			t.Fatal(err)
		}
		rels = append(rels, filepath.ToSlash(rel))
	}
	got := strings.Join(rels, ",")
	want := "README.md,docs/nested/deep.md,docs/top.md,skills/demo/SKILL.md,skills/demo/references/ref.md,tests/README.md"
	if got != want {
		t.Fatalf("got files %q, want %q", got, want)
	}
}

func TestNestedDocumentationLinkReported(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "docs/guides/nested.md", "[gone](missing.md)\n")
	problems := checkFixture(t, root)
	if len(problems) != 1 || problems[0].file != "docs/guides/nested.md" || problems[0].line != 1 {
		t.Fatalf("unexpected problems: %v", problems)
	}
}

func TestMarkdownLinksResolveOnlyFromContainingDirectory(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Root\n")
	writeFixture(t, root, "models.md", "# Root model\n")
	writeFixture(t, root, "docs/guides/a.md", "[model](models.md)\n")
	if problems := checkFixture(t, root); len(problems) != 1 {
		t.Fatalf("expected nested link to be missing, got %v", problems)
	}
	writeFixture(t, root, "docs/guides/models.md", "# Nested model\n")
	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected nested link to pass, got %v", problems)
	}
}

func TestInvalidRootFails(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	if _, _, err := check(missing); err == nil || err.Error() != "doclinks: root "+missing+" is not a directory" {
		t.Fatalf("unexpected error: %v", err)
	}
	file := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := check(file); err == nil || err.Error() != "doclinks: root "+file+" is not a directory" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNoDocumentationFilesFails(t *testing.T) {
	root := t.TempDir()
	if _, _, err := check(root); err == nil || err.Error() != "doclinks: no documentation files found under "+root {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReferenceDefinitionsChecked(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "[guide]: docs/missing.md \"Guide\"\n")
	problems := checkFixture(t, root)
	if len(problems) != 1 || problems[0].file != "README.md" || problems[0].line != 1 || problems[0].target != "docs/missing.md" {
		t.Fatalf("unexpected problems: %v", problems)
	}
	writeFixture(t, root, "docs/missing.md", "# Guide\n")
	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected valid reference definition to pass, got %v", problems)
	}
}

func TestBacktickedDirectoriesChecked(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "Missing "+bt+"docs/missing-dir/"+bt+".\nValid "+bt+"forge/orm/"+bt+".\nIgnore "+bt+"go test ./..."+bt+".\n")
	if err := os.MkdirAll(filepath.Join(root, "forge", "orm"), 0o755); err != nil {
		t.Fatal(err)
	}
	problems := checkFixture(t, root)
	if len(problems) != 1 || problems[0].line != 1 || problems[0].target != "docs/missing-dir/" {
		t.Fatalf("unexpected problems: %v", problems)
	}
}

func TestImageWrappedLinkOuterDestinationChecked(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\n"+
		"[![License](img/license.png)](LICENSE)\n")
	writeFixture(t, root, "img/license.png", "image data\n")

	// Outer destination LICENSE does not exist yet; must be reported as missing
	problems := checkFixture(t, root)
	if len(problems) != 1 {
		t.Fatalf("expected 1 problem for missing outer link, got %v", problems)
	}
	if got := problems[0]; got.file != "README.md" || got.line != 3 || got.target != "LICENSE" {
		t.Fatalf("unexpected problem: %v", got)
	}

	// After LICENSE is created, both destinations exist and pass
	writeFixture(t, root, "LICENSE", "MIT License\n")
	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected 0 problems after creating LICENSE, got %v", problems)
	}
}

func TestFootnoteDefinitionsExcludedFromLinkReferences(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\n"+
		"Statement referencing footnote.[^1]\n\n"+
		"[^1]: explanatory text\n"+
		"[^note]: This is another note.\n")
	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected no problems for footnote definitions, got %v", problems)
	}
}

func TestFencesLongerThanThreeBackticksOrTildes(t *testing.T) {
	root := t.TempDir()
	sixTicks := strings.Repeat(bt, 6)
	writeFixture(t, root, "README.md", "# Title\n\n"+
		sixTicks+"markdown\n"+
		"```go\n"+
		"[gone](docs/nope.md)\n"+
		"```\n"+
		sixTicks+"\n"+
		"~~~~\n"+
		"[gone2](docs/nope2.md)\n"+
		"~~~~\n"+
		"See [after](docs/after.md).\n")

	problems := checkFixture(t, root)
	if len(problems) != 1 || problems[0].file != "README.md" || problems[0].line != 11 || problems[0].target != "docs/after.md" {
		t.Fatalf("expected 1 problem for link after closed fence, got %v", problems)
	}
}

func TestURISchemeDetectionCaseInsensitiveAndNonFileSkipped(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\n"+
		"[web](HTTPS://example.com/guide.md)\n"+
		"[call](tel:+1234567890)\n"+
		"[files](ftp://example.com/archive.zip)\n"+
		"[inline](data:text/plain;base64,SGVsbG8=)\n"+
		"[chat](irc://irc.libera.chat/channel)\n")
	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected no problems for non-file URI schemes, got %v", problems)
	}
}

func TestPercentEncodedLocalLinksDecoded(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\n"+
		"See [guide](docs/a%20b.md).\n"+
		"See [literal](docs/100%25real.md).\n"+
		"See [fallback](docs/invalid%zz.md).\n")
	writeFixture(t, root, "docs/a b.md", "# Guide A B\n")
	writeFixture(t, root, "docs/100%real.md", "# Real 100%\n")
	writeFixture(t, root, "docs/invalid%zz.md", "# Invalid fallback\n")

	if problems := checkFixture(t, root); len(problems) != 0 {
		t.Fatalf("expected no problems for existing decoded local links, got %v", problems)
	}

	writeFixture(t, root, "README.md", "# Title\n\n"+
		"See [guide](docs/missing%20file.md).\n")
	problems := checkFixture(t, root)
	if len(problems) != 1 || problems[0].target != "docs/missing file.md" {
		t.Fatalf("expected missing decoded target, got %v", problems)
	}
}

func TestInlineCodeSpansExcludedFromMarkdownLinks(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "# Title\n\n"+
		"Example syntax: "+bt+"[gone](docs/missing.md)"+bt+".\n"+
		"Double backtick example: "+bt+bt+"[also gone](docs/nope.md)"+bt+bt+".\n"+
		"Real backtick path: "+bt+"docs/real.md"+bt+".\n"+
		"Missing backtick path: "+bt+"docs/MISSING.md"+bt+".\n")
	writeFixture(t, root, "docs/real.md", "# Real\n")

	problems := checkFixture(t, root)
	if len(problems) != 1 {
		t.Fatalf("expected exactly 1 problem for missing backticked path, got %v", problems)
	}
	got := problems[0]
	if got.file != "README.md" || got.line != 6 || got.target != "docs/MISSING.md" {
		t.Fatalf("unexpected problem: %v", got)
	}
}
