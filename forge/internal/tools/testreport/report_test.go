package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeLongOutput(n int) string {
	var sb strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&sb, "{\"Action\":\"output\",\"Package\":\"p\",\"Test\":\"TestLong\",\"Output\":\"line %d\\n\"}\n", i)
	}
	sb.WriteString("{\"Action\":\"fail\",\"Package\":\"p\",\"Test\":\"TestLong\"}\n")
	return sb.String()
}

func TestRun(t *testing.T) {
	cases := []struct {
		name, input string
		required    patterns
		code        int
		want        string
	}{
		{"pass only", "{\"Action\":\"pass\",\"Package\":\"p\",\"Test\":\"TestOK\"}\n{\"Action\":\"pass\",\"Package\":\"p\"}\n", nil, 0, "TOTAL 1 0 0 0"},
		{"required skip", `{"Action":"skip","Package":"p","Test":"TestDB"}`, patterns{"^p$"}, 1, "REQUIRED TEST SKIPPED: p TestDB"},
		{"other skip", `{"Action":"skip","Package":"q","Test":"TestDB"}`, patterns{"^p$"}, 0, "TOTAL 0 0 1 0"},
		{"failed test", `{"Action":"fail","Package":"p","Test":"TestBad"}`, nil, 1, "TOTAL 0 1 0 0"},
		{"build failure", `{"Action":"fail","Package":"p"}`, nil, 1, "TOTAL 0 0 0 1"},
		{"malformed", "not json", nil, 2, ""},
		{"null", "null", nil, 2, ""},
		{"bad regexp", "", patterns{"["}, 2, ""},
		{"blank lines", "\n \n", nil, 0, "TOTAL 0 0 0 0"},
		{
			"failing test output",
			"{\"Action\":\"output\",\"Package\":\"p\",\"Test\":\"TestBad\",\"Output\":\"bad_test.go:10: assertion failed\\n\"}\n{\"Action\":\"fail\",\"Package\":\"p\",\"Test\":\"TestBad\"}\n",
			nil,
			1,
			"=== FAIL p TestBad bad_test.go:10: assertion failed",
		},
		{
			"package failure output",
			"{\"Action\":\"output\",\"Package\":\"p\",\"Output\":\"syntax error: unexpected token\\n\"}\n{\"Action\":\"fail\",\"Package\":\"p\"}\n",
			nil,
			1,
			"=== PACKAGE FAIL p syntax error: unexpected token",
		},
		{
			"required skip output",
			"{\"Action\":\"output\",\"Package\":\"p\",\"Test\":\"TestDB\",\"Output\":\"skipping: db unavailable\\n\"}\n{\"Action\":\"skip\",\"Package\":\"p\",\"Test\":\"TestDB\"}\n",
			patterns{"^p$"},
			1,
			"=== REQUIRED SKIP p TestDB skipping: db unavailable",
		},
		{
			"60-line cap message",
			makeLongOutput(65),
			nil,
			1,
			"... 5 earlier lines omitted",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			code, err := run(strings.NewReader(tc.input), &output, options{requireNoSkip: tc.required})
			if code != tc.code || (err != nil) != (tc.code == 2) {
				t.Fatalf("run = (%d, %v), want code %d", code, err, tc.code)
			}
			if !strings.Contains(strings.Join(strings.Fields(output.String()), " "), tc.want) {
				t.Fatalf("report %q missing %q", output.String(), tc.want)
			}
		})
	}
}

func TestReportSortedAndCopied(t *testing.T) {
	input := `{"Action":"pass","Package":"z","Test":"TestZ"}
{"Action":"skip","Package":"a","Test":"TestA"}
{"Action":"fail","Package":"z"}
`
	path := filepath.Join(t.TempDir(), "report.txt")
	var output bytes.Buffer
	code, err := run(strings.NewReader(input), &output, options{out: path})
	if code != 1 || err != nil {
		t.Fatalf("run = (%d, %v)", code, err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != output.String() {
		t.Fatalf("file differs from stdout: %v", err)
	}
	lines := strings.Split(output.String(), "\n")
	want := []string{"PACKAGE PASSED FAILED SKIPPED PACKAGE_FAILURES", "a 0 0 1 0", "z 1 0 0 1", "TOTAL 1 0 1 1"}
	for i, expected := range want {
		if got := strings.Join(strings.Fields(lines[i]), " "); got != expected {
			t.Errorf("line %d = %q, want %q", i, got, expected)
		}
	}
}

func TestDetailedFailureReporting(t *testing.T) {
	t.Run("failing test output appears after fail header", func(t *testing.T) {
		input := `{"Action":"output","Package":"pkg/a","Test":"TestCrash","Output":"line from test crash\n"}
{"Action":"fail","Package":"pkg/a","Test":"TestCrash"}
`
		var out bytes.Buffer
		code, err := run(strings.NewReader(input), &out, options{})
		if code != 1 || err != nil {
			t.Fatalf("run = (%d, %v), want code 1", code, err)
		}
		s := out.String()
		headerIdx := strings.Index(s, "=== FAIL pkg/a TestCrash")
		outputIdx := strings.Index(s, "line from test crash")
		if headerIdx == -1 {
			t.Fatalf("missing === FAIL header in %q", s)
		}
		if outputIdx == -1 {
			t.Fatalf("missing output line in %q", s)
		}
		if outputIdx <= headerIdx {
			t.Fatalf("output line appears before or at header: headerIdx=%d, outputIdx=%d", headerIdx, outputIdx)
		}
	})

	t.Run("package-level failure prints output under package fail", func(t *testing.T) {
		input := `{"Action":"output","Package":"pkg/b","Output":"build error: cannot find package\n"}
{"Action":"fail","Package":"pkg/b"}
`
		var out bytes.Buffer
		code, err := run(strings.NewReader(input), &out, options{})
		if code != 1 || err != nil {
			t.Fatalf("run = (%d, %v), want code 1", code, err)
		}
		s := out.String()
		headerIdx := strings.Index(s, "=== PACKAGE FAIL pkg/b")
		outputIdx := strings.Index(s, "build error: cannot find package")
		if headerIdx == -1 {
			t.Fatalf("missing === PACKAGE FAIL header in %q", s)
		}
		if outputIdx == -1 {
			t.Fatalf("missing output line in %q", s)
		}
		if outputIdx <= headerIdx {
			t.Fatalf("output line appears before or at header: headerIdx=%d, outputIdx=%d", headerIdx, outputIdx)
		}
	})

	t.Run("required skip prints skip message", func(t *testing.T) {
		input := `{"Action":"output","Package":"pkg/c","Test":"TestNetwork","Output":"skipping: no network\n"}
{"Action":"skip","Package":"pkg/c","Test":"TestNetwork"}
`
		var out bytes.Buffer
		code, err := run(strings.NewReader(input), &out, options{requireNoSkip: patterns{"^pkg/c$"}})
		if code != 1 || err != nil {
			t.Fatalf("run = (%d, %v), want code 1", code, err)
		}
		s := out.String()
		headerIdx := strings.Index(s, "=== REQUIRED SKIP pkg/c TestNetwork")
		outputIdx := strings.Index(s, "skipping: no network")
		if headerIdx == -1 {
			t.Fatalf("missing === REQUIRED SKIP header in %q", s)
		}
		if outputIdx == -1 {
			t.Fatalf("missing output line in %q", s)
		}
		if outputIdx <= headerIdx {
			t.Fatalf("output line appears before or at header: headerIdx=%d, outputIdx=%d", headerIdx, outputIdx)
		}
	})

	t.Run("60-line cap omits earlier lines and keeps last 60", func(t *testing.T) {
		input := makeLongOutput(70)
		var out bytes.Buffer
		code, err := run(strings.NewReader(input), &out, options{})
		if code != 1 || err != nil {
			t.Fatalf("run = (%d, %v), want code 1", code, err)
		}
		s := out.String()
		wantCapMsg := "... 10 earlier lines omitted"
		if !strings.Contains(s, wantCapMsg) {
			t.Fatalf("expected %q in %q", wantCapMsg, s)
		}
		if strings.Contains(s, "\nline 1\n") || strings.Contains(s, "\nline 10\n") {
			t.Fatalf("earlier omitted lines should not appear in %q", s)
		}
		if !strings.Contains(s, "\nline 11\n") || !strings.Contains(s, "\nline 70\n") {
			t.Fatalf("retained lines (11..70) should appear in %q", s)
		}
	})

	t.Run("cap at 50 failure sections", func(t *testing.T) {
		var sb strings.Builder
		for i := 1; i <= 55; i++ {
			fmt.Fprintf(&sb, "{\"Action\":\"fail\",\"Package\":\"p\",\"Test\":\"TestFail%02d\"}\n", i)
		}
		var out bytes.Buffer
		code, err := run(strings.NewReader(sb.String()), &out, options{})
		if code != 1 || err != nil {
			t.Fatalf("run = (%d, %v), want code 1", code, err)
		}
		s := out.String()
		wantMoreMsg := "... 5 more failures omitted"
		if !strings.Contains(s, wantMoreMsg) {
			t.Fatalf("expected %q in %q", wantMoreMsg, s)
		}
		// Count occurrences of "=== FAIL"
		count := strings.Count(s, "=== FAIL")
		if count != 50 {
			t.Fatalf("expected 50 === FAIL sections, got %d", count)
		}
	})
}

func TestOutputBufferingBounded(t *testing.T) {
	t.Run("10,000 output lines for a passing test leave no retained buffer", func(t *testing.T) {
		var sb strings.Builder
		for i := 1; i <= 10000; i++ {
			fmt.Fprintf(&sb, "{\"Action\":\"output\",\"Package\":\"p\",\"Test\":\"TestPass10k\",\"Output\":\"line %d\\n\"}\n", i)
		}
		sb.WriteString("{\"Action\":\"pass\",\"Package\":\"p\",\"Test\":\"TestPass10k\"}\n")
		sb.WriteString("{\"Action\":\"pass\",\"Package\":\"p\"}\n")

		col, err := collectEvents(strings.NewReader(sb.String()), options{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		id := testID{pkg: "p", test: "TestPass10k"}
		if buf, exists := col.testOutputs[id]; exists && buf != nil {
			t.Fatalf("expected passing test to leave no retained buffer, got %v", buf)
		}
	})

	t.Run("10,000 output lines for a failing test keep only 60 lines plus omitted count", func(t *testing.T) {
		var sb strings.Builder
		for i := 1; i <= 10000; i++ {
			fmt.Fprintf(&sb, "{\"Action\":\"output\",\"Package\":\"p\",\"Test\":\"TestFail10k\",\"Output\":\"line %d\\n\"}\n", i)
		}
		sb.WriteString("{\"Action\":\"fail\",\"Package\":\"p\",\"Test\":\"TestFail10k\"}\n")
		sb.WriteString("{\"Action\":\"fail\",\"Package\":\"p\"}\n")

		col, err := collectEvents(strings.NewReader(sb.String()), options{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		id := testID{pkg: "p", test: "TestFail10k"}
		buf, exists := col.testOutputs[id]
		if !exists || buf == nil {
			t.Fatalf("expected failing test to have retained buffer")
		}
		buf.finish()
		if len(buf.lines) != 60 {
			t.Fatalf("expected 60 lines, got %d", len(buf.lines))
		}
		if buf.omitted != 9940 {
			t.Fatalf("expected 9940 omitted lines, got %d", buf.omitted)
		}
		if buf.lines[0] != "line 9941" || buf.lines[59] != "line 10000" {
			t.Fatalf("unexpected line contents: first=%q, last=%q", buf.lines[0], buf.lines[59])
		}

		// Verify end-to-end report generation
		var out bytes.Buffer
		code, err := run(strings.NewReader(sb.String()), &out, options{})
		if code != 1 || err != nil {
			t.Fatalf("run = (%d, %v), want code 1", code, err)
		}
		s := out.String()
		wantMsg := "... 9940 earlier lines omitted"
		if !strings.Contains(s, wantMsg) {
			t.Fatalf("expected %q in %q", wantMsg, s)
		}
		if strings.Contains(s, "\nline 1\n") || strings.Contains(s, "\nline 9940\n") {
			t.Fatalf("omitted lines should not appear in report: %q", s)
		}
		if !strings.Contains(s, "\nline 9941\n") || !strings.Contains(s, "\nline 10000\n") {
			t.Fatalf("retained lines should appear in report: %q", s)
		}
	})

	t.Run("10 MiB output event without newlines keeps at most 64 KiB", func(t *testing.T) {
		buf := &outputBuffer{}
		tenMB := strings.Repeat("a", 10*1024*1024)
		buf.add(tenMB)
		if len(buf.partial) > 64*1024 {
			t.Fatalf("expected buf.partial <= 64 KiB, got %d bytes", len(buf.partial))
		}
		if buf.truncated != 10*1024*1024-64*1024 {
			t.Fatalf("expected truncated == %d, got %d", 10*1024*1024-64*1024, buf.truncated)
		}
	})
}

func TestBuildEvents(t *testing.T) {
	t.Run("build-output and build-fail events record output and fail package", func(t *testing.T) {
		input := `{"Action":"build-output","ImportPath":"pkg/buildfail","Output":"syntax error: unexpected token\n"}
{"Action":"build-fail","ImportPath":"pkg/buildfail"}
`
		var out bytes.Buffer
		code, err := run(strings.NewReader(input), &out, options{})
		if code != 1 || err != nil {
			t.Fatalf("run = (%d, %v), want code 1, err nil", code, err)
		}
		s := out.String()
		norm := strings.Join(strings.Fields(s), " ")
		if !strings.Contains(s, "=== PACKAGE FAIL pkg/buildfail") {
			t.Errorf("expected '=== PACKAGE FAIL pkg/buildfail' in report, got: %s", s)
		}
		if !strings.Contains(s, "syntax error: unexpected token") {
			t.Errorf("expected build output in report, got: %s", s)
		}
		if !strings.Contains(norm, "pkg/buildfail 0 0 0 1") {
			t.Errorf("expected package failure count 1 in table, got: %s", s)
		}
	})

	t.Run("build failure with subsequent package fail event deduplicates package failure count", func(t *testing.T) {
		input := `{"Action":"build-output","ImportPath":"pkg/buildfail","Output":"syntax error: unexpected token\n"}
{"Action":"build-fail","ImportPath":"pkg/buildfail"}
{"Action":"output","Package":"pkg/buildfail","Output":"FAIL\tpkg/buildfail [setup failed]\n"}
{"Action":"fail","Package":"pkg/buildfail"}
`
		var out bytes.Buffer
		code, err := run(strings.NewReader(input), &out, options{})
		if code != 1 || err != nil {
			t.Fatalf("run = (%d, %v), want code 1, err nil", code, err)
		}
		s := out.String()
		norm := strings.Join(strings.Fields(s), " ")
		if !strings.Contains(s, "=== PACKAGE FAIL pkg/buildfail") {
			t.Errorf("expected '=== PACKAGE FAIL pkg/buildfail' in report, got: %s", s)
		}
		if !strings.Contains(norm, "pkg/buildfail 0 0 0 1") {
			t.Errorf("expected package failure count 1 in table, got: %s", s)
		}
	})
}

func TestPackageLevelSkip(t *testing.T) {
	t.Run("unrequired package-level skip counts in table", func(t *testing.T) {
		input := `{"Action":"output","Package":"pkg/notests","Output":"?   \tpkg/notests [no test files]\n"}
{"Action":"skip","Package":"pkg/notests"}
`
		var out bytes.Buffer
		code, err := run(strings.NewReader(input), &out, options{})
		if code != 0 || err != nil {
			t.Fatalf("run = (%d, %v), want code 0, err nil", code, err)
		}
		s := out.String()
		norm := strings.Join(strings.Fields(s), " ")
		if !strings.Contains(norm, "pkg/notests 0 0 1 0") {
			t.Errorf("expected skipped package count 1 in table, got: %s", s)
		}
		if strings.Contains(s, "REQUIRED TEST SKIPPED") {
			t.Errorf("did not expect required test skipped in report, got: %s", s)
		}
	})

	t.Run("required package-level skip with no test files passes run", func(t *testing.T) {
		input := `{"Action":"output","Package":"pkg/required","Output":"?   \tpkg/required [no test files]\n"}
{"Action":"skip","Package":"pkg/required"}
`
		var out bytes.Buffer
		code, err := run(strings.NewReader(input), &out, options{requireNoSkip: patterns{"^pkg/required$"}})
		if code != 0 || err != nil {
			t.Fatalf("run = (%d, %v), want code 0, err nil", code, err)
		}
		s := out.String()
		norm := strings.Join(strings.Fields(s), " ")
		if !strings.Contains(norm, "pkg/required 0 0 1 0") {
			t.Errorf("expected skipped package count 1 in table, got: %s", s)
		}
		if strings.Contains(s, "REQUIRED TEST SKIPPED") {
			t.Errorf("did not expect required test skipped in report, got: %s", s)
		}
	})

	t.Run("real package-level skip in a required package fails run", func(t *testing.T) {
		input := `{"Action":"output","Package":"pkg/required","Output":"skipping package tests\n"}
{"Action":"skip","Package":"pkg/required"}
`
		var out bytes.Buffer
		code, err := run(strings.NewReader(input), &out, options{requireNoSkip: patterns{"^pkg/required$"}})
		if code != 1 || err != nil {
			t.Fatalf("run = (%d, %v), want code 1, err nil", code, err)
		}
		s := out.String()
		norm := strings.Join(strings.Fields(s), " ")
		if !strings.Contains(norm, "pkg/required 0 0 1 0") {
			t.Errorf("expected skipped package count 1 in table, got: %s", s)
		}
		if !strings.Contains(s, "REQUIRED TEST SKIPPED: pkg/required") {
			t.Errorf("expected 'REQUIRED TEST SKIPPED: pkg/required' in report, got: %s", s)
		}
		if !strings.Contains(s, "=== REQUIRED SKIP pkg/required") {
			t.Errorf("expected '=== REQUIRED SKIP pkg/required' section in report, got: %s", s)
		}
	})
}

func TestAllowSkip(t *testing.T) {
	input := `{"Action":"output","Package":"github.com/forgego/forge/tests/pkg_migrations","Test":"TestMigrationApplySQLite","Output":"SQLite migration apply is unverified\n"}
{"Action":"skip","Package":"github.com/forgego/forge/tests/pkg_migrations","Test":"TestMigrationApplySQLite"}
`
	t.Run("fails without allow-skip", func(t *testing.T) {
		var out bytes.Buffer
		code, err := run(strings.NewReader(input), &out, options{
			requireNoSkip: patterns{"^github.com/forgego/forge/tests/(integration|pkg_migrations|e2e)"},
		})
		if code != 1 || err != nil {
			t.Fatalf("run = (%d, %v), want code 1", code, err)
		}
		if !strings.Contains(out.String(), "REQUIRED TEST SKIPPED:") {
			t.Errorf("expected required test skipped, got: %s", out.String())
		}
	})

	t.Run("passes with matching allow-skip", func(t *testing.T) {
		var out bytes.Buffer
		code, err := run(strings.NewReader(input), &out, options{
			requireNoSkip: patterns{"^github.com/forgego/forge/tests/(integration|pkg_migrations|e2e)"},
			allowSkip:     patterns{"tests/pkg_migrations$ ^TestMigrationApplySQLite$"},
		})
		if code != 0 || err != nil {
			t.Fatalf("run = (%d, %v), want code 0, err nil", code, err)
		}
		s := out.String()
		norm := strings.Join(strings.Fields(s), " ")
		if strings.Contains(s, "REQUIRED TEST SKIPPED") {
			t.Errorf("did not expect required test skipped, got: %s", s)
		}
		if !strings.Contains(norm, "github.com/forgego/forge/tests/pkg_migrations 0 0 1 0") {
			t.Errorf("expected 1 skip recorded in table, got: %s", s)
		}
	})

	t.Run("fails if allow-skip pattern does not match test", func(t *testing.T) {
		var out bytes.Buffer
		code, err := run(strings.NewReader(input), &out, options{
			requireNoSkip: patterns{"^github.com/forgego/forge/tests/(integration|pkg_migrations|e2e)"},
			allowSkip:     patterns{"tests/pkg_migrations$ ^OtherTest$"},
		})
		if code != 1 || err != nil {
			t.Fatalf("run = (%d, %v), want code 1", code, err)
		}
	})

	t.Run("invalid allow-skip flag format gives exit code 2", func(t *testing.T) {
		var out bytes.Buffer
		code, err := run(strings.NewReader(input), &out, options{
			allowSkip: patterns{"invalid-pattern-without-space"},
		})
		if code != 2 || err == nil {
			t.Fatalf("run = (%d, %v), want code 2 and error", code, err)
		}
	})
}
