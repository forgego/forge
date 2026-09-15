package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
)

type options struct {
	out           string
	requireNoSkip patterns
	allowSkip     patterns
}

type event struct {
	Action     string
	Package    string
	Test       string
	Elapsed    float64
	Output     string
	ImportPath string
}

type counts struct{ passed, failed, skipped, packageFailed int }

type testID struct {
	pkg  string
	test string
}

const (
	maxBufferedLines = 60
	maxPartialBytes  = 64 * 1024
)

type outputBuffer struct {
	lines     []string
	omitted   int
	partial   string
	truncated int
}

func (b *outputBuffer) addLine(line string) {
	line = strings.TrimSuffix(line, "\r")
	if len(b.lines) < maxBufferedLines {
		b.lines = append(b.lines, line)
	} else {
		copy(b.lines, b.lines[1:])
		b.lines[maxBufferedLines-1] = line
		b.omitted++
	}
}

func (b *outputBuffer) add(s string) {
	if s == "" {
		return
	}
	full := b.partial + s
	lastNL := strings.LastIndex(full, "\n")
	if lastNL == -1 {
		if len(full) > maxPartialBytes {
			b.truncated += len(full) - maxPartialBytes
			full = full[len(full)-maxPartialBytes:]
		}
		b.partial = full
		return
	}
	b.partial = full[lastNL+1:]
	if len(b.partial) > maxPartialBytes {
		b.truncated += len(b.partial) - maxPartialBytes
		b.partial = b.partial[len(b.partial)-maxPartialBytes:]
	}
	for _, line := range strings.Split(full[:lastNL], "\n") {
		b.addLine(line)
	}
}

func (b *outputBuffer) finish() {
	if b.partial != "" {
		b.addLine(b.partial)
		b.partial = ""
	}
}

func (b *outputBuffer) contains(substr string) bool {
	if b == nil {
		return false
	}
	if strings.Contains(b.partial, substr) {
		return true
	}
	for _, l := range b.lines {
		if strings.Contains(l, substr) {
			return true
		}
	}
	return false
}

type section struct {
	header  string
	lines   []string
	omitted int
}

type reportCollector struct {
	packages       map[string]*counts
	forbidden      []string
	testOutputs    map[testID]*outputBuffer
	packageOutputs map[string]*outputBuffer
	failedTests    []testID
	failedSeen     map[testID]bool
	requiredSkips  []testID
	skipSeen       map[testID]bool
	exitCode       int
}

func collectEvents(r io.Reader, opts options) (*reportCollector, error) {
	col := &reportCollector{
		packages:       make(map[string]*counts),
		testOutputs:    make(map[testID]*outputBuffer),
		packageOutputs: make(map[string]*outputBuffer),
		failedSeen:     make(map[testID]bool),
		skipSeen:       make(map[testID]bool),
	}

	var required []*regexp.Regexp
	for _, pattern := range opts.requireNoSkip {
		re, err := regexp.Compile(pattern)
		if err != nil {
			col.exitCode = 2
			return col, fmt.Errorf("invalid --require-no-skip pattern: %w", err)
		}
		required = append(required, re)
	}

	type allowRule struct {
		pkgRe  *regexp.Regexp
		testRe *regexp.Regexp
	}
	var allowed []allowRule
	for _, pattern := range opts.allowSkip {
		parts := strings.Fields(pattern)
		if len(parts) != 2 {
			col.exitCode = 2
			return col, fmt.Errorf("invalid --allow-skip pattern %q: must be '<package-regexp> <test-regexp>'", pattern)
		}
		pkgRe, err := regexp.Compile(parts[0])
		if err != nil {
			col.exitCode = 2
			return col, fmt.Errorf("invalid --allow-skip package pattern: %w", err)
		}
		testRe, err := regexp.Compile(parts[1])
		if err != nil {
			col.exitCode = 2
			return col, fmt.Errorf("invalid --allow-skip test pattern: %w", err)
		}
		allowed = append(allowed, allowRule{pkgRe: pkgRe, testRe: testRe})
	}

	isAllowed := func(pkg, test string) bool {
		for _, rule := range allowed {
			if rule.pkgRe.MatchString(pkg) && rule.testRe.MatchString(test) {
				return true
			}
		}
		return false
	}

	reader := bufio.NewReader(r)
	for lineNo := 1; ; lineNo++ {
		line, readErr := reader.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			col.exitCode = 2
			return col, fmt.Errorf("read input: %w", readErr)
		}
		if strings.TrimSpace(line) != "" {
			var e event
			if err := json.Unmarshal([]byte(line), &e); err != nil {
				col.exitCode = 2
				return col, fmt.Errorf("line %d: %w", lineNo, err)
			}
			if e.Package == "" && e.ImportPath != "" {
				e.Package = e.ImportPath
			}
			if e.Action == "" || e.Package == "" {
				col.exitCode = 2
				return col, fmt.Errorf("line %d: missing Action or Package", lineNo)
			}
			c := col.packages[e.Package]
			if c == nil {
				c = &counts{}
				col.packages[e.Package] = c
			}

			if e.Action == "output" || e.Action == "build-output" {
				if e.Test != "" {
					id := testID{pkg: e.Package, test: e.Test}
					b := col.testOutputs[id]
					if b == nil {
						b = &outputBuffer{}
						col.testOutputs[id] = b
					}
					b.add(e.Output)
				} else {
					b := col.packageOutputs[e.Package]
					if b == nil {
						b = &outputBuffer{}
						col.packageOutputs[e.Package] = b
					}
					b.add(e.Output)
				}
			}

			if e.Test == "" {
				if e.Action == "fail" || e.Action == "build-fail" {
					if c.packageFailed == 0 {
						c.packageFailed = 1
					}
					col.exitCode = 1
				} else if e.Action == "pass" {
					delete(col.packageOutputs, e.Package)
				} else if e.Action == "skip" {
					c.skipped++
					noTestFiles := col.packageOutputs[e.Package].contains("[no test files]")
					isRequired := false
					if !noTestFiles {
						for _, re := range required {
							if re.MatchString(e.Package) {
								if !isAllowed(e.Package, "") {
									isRequired = true
									col.forbidden = append(col.forbidden, e.Package)
									col.exitCode = 1
									id := testID{pkg: e.Package, test: ""}
									if !col.skipSeen[id] {
										col.skipSeen[id] = true
										col.requiredSkips = append(col.requiredSkips, id)
									}
								}
								break
							}
						}
					}
					if !isRequired {
						delete(col.packageOutputs, e.Package)
					}
				}
			} else {
				switch e.Action {
				case "pass":
					c.passed++
					delete(col.testOutputs, testID{pkg: e.Package, test: e.Test})
				case "fail":
					c.failed++
					col.exitCode = 1
					id := testID{pkg: e.Package, test: e.Test}
					if !col.failedSeen[id] {
						col.failedSeen[id] = true
						col.failedTests = append(col.failedTests, id)
					}
				case "skip":
					c.skipped++
					isRequired := false
					for _, re := range required {
						if re.MatchString(e.Package) {
							if !isAllowed(e.Package, e.Test) {
								isRequired = true
								col.forbidden = append(col.forbidden, e.Package+" "+e.Test)
								col.exitCode = 1
								id := testID{pkg: e.Package, test: e.Test}
								if !col.skipSeen[id] {
									col.skipSeen[id] = true
									col.requiredSkips = append(col.requiredSkips, id)
								}
							}
							break
						}
					}
					if !isRequired {
						delete(col.testOutputs, testID{pkg: e.Package, test: e.Test})
					}
				}
			}
		}
		if readErr == io.EOF {
			break
		}
	}
	return col, nil
}

func (col *reportCollector) render(w io.Writer, opts options) (int, error) {
	var names []string
	for name := range col.packages {
		names = append(names, name)
	}
	sort.Strings(names)
	var report bytes.Buffer
	table := tabwriter.NewWriter(&report, 0, 4, 2, ' ', 0)
	fmt.Fprintln(table, "PACKAGE\tPASSED\tFAILED\tSKIPPED\tPACKAGE_FAILURES")
	var total counts
	row := func(name string, c counts) {
		fmt.Fprintf(table, "%s\t%d\t%d\t%d\t%d\n", name, c.passed, c.failed, c.skipped, c.packageFailed)
	}
	for _, name := range names {
		c := *col.packages[name]
		row(name, c)
		total.passed += c.passed
		total.failed += c.failed
		total.skipped += c.skipped
		total.packageFailed += c.packageFailed
	}
	row("TOTAL", total)
	if err := table.Flush(); err != nil {
		return 2, err
	}
	sort.Strings(col.forbidden)
	for _, test := range col.forbidden {
		fmt.Fprintf(&report, "REQUIRED TEST SKIPPED: %s\n", test)
	}

	var sections []section

	sort.Slice(col.failedTests, func(i, j int) bool {
		if col.failedTests[i].pkg != col.failedTests[j].pkg {
			return col.failedTests[i].pkg < col.failedTests[j].pkg
		}
		return col.failedTests[i].test < col.failedTests[j].test
	})
	for _, id := range col.failedTests {
		var lines []string
		var omitted int
		if b := col.testOutputs[id]; b != nil {
			b.finish()
			lines = b.lines
			omitted = b.omitted
		}
		sections = append(sections, section{
			header:  fmt.Sprintf("=== FAIL %s %s", id.pkg, id.test),
			lines:   lines,
			omitted: omitted,
		})
	}

	for _, name := range names {
		c := col.packages[name]
		if c.packageFailed > 0 && c.failed == 0 {
			var lines []string
			var omitted int
			if b := col.packageOutputs[name]; b != nil {
				b.finish()
				lines = b.lines
				omitted = b.omitted
			}
			sections = append(sections, section{
				header:  fmt.Sprintf("=== PACKAGE FAIL %s", name),
				lines:   lines,
				omitted: omitted,
			})
		}
	}

	sort.Slice(col.requiredSkips, func(i, j int) bool {
		if col.requiredSkips[i].pkg != col.requiredSkips[j].pkg {
			return col.requiredSkips[i].pkg < col.requiredSkips[j].pkg
		}
		return col.requiredSkips[i].test < col.requiredSkips[j].test
	})
	for _, id := range col.requiredSkips {
		var lines []string
		var omitted int
		if id.test != "" {
			if b := col.testOutputs[id]; b != nil {
				b.finish()
				lines = b.lines
				omitted = b.omitted
			}
		} else {
			if b := col.packageOutputs[id.pkg]; b != nil {
				b.finish()
				lines = b.lines
				omitted = b.omitted
			}
		}
		header := fmt.Sprintf("=== REQUIRED SKIP %s %s", id.pkg, id.test)
		if id.test == "" {
			header = fmt.Sprintf("=== REQUIRED SKIP %s", id.pkg)
		}
		sections = append(sections, section{
			header:  header,
			lines:   lines,
			omitted: omitted,
		})
	}

	const maxSections = 50
	const maxLines = 60

	toPrint := sections
	var omittedSections int
	if len(toPrint) > maxSections {
		omittedSections = len(toPrint) - maxSections
		toPrint = toPrint[:maxSections]
	}

	for _, sec := range toPrint {
		fmt.Fprintln(&report, sec.header)
		lines := sec.lines
		omitted := sec.omitted
		if len(lines) > maxLines {
			omitted += len(lines) - maxLines
			lines = lines[len(lines)-maxLines:]
		}
		if omitted > 0 {
			fmt.Fprintf(&report, "... %d earlier lines omitted\n", omitted)
		}
		for _, line := range lines {
			fmt.Fprintln(&report, line)
		}
	}

	if omittedSections > 0 {
		fmt.Fprintf(&report, "... %d more failures omitted\n", omittedSections)
	}

	if opts.out != "" {
		if err := os.WriteFile(opts.out, report.Bytes(), 0o644); err != nil {
			return 2, fmt.Errorf("write report: %w", err)
		}
	}
	if _, err := w.Write(report.Bytes()); err != nil {
		return 2, fmt.Errorf("write report: %w", err)
	}
	return col.exitCode, nil
}

func run(r io.Reader, w io.Writer, opts options) (exitCode int, err error) {
	col, err := collectEvents(r, opts)
	if err != nil {
		return col.exitCode, err
	}
	return col.render(w, opts)
}
