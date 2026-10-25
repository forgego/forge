// Command capcheck verifies that capabilities marked Supported in status.md
// have valid evidence backing them. Usage: go run ./internal/tools/capcheck
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type problem struct {
	file   string
	line   int
	reason string
}

func (p problem) String() string {
	return fmt.Sprintf("%s:%d: %s", p.file, p.line, p.reason)
}

var (
	funcTestRe  = regexp.MustCompile(`^\s*func\s+(Test[A-Za-z0-9_]*)\s*\(`)
	backtickRe  = regexp.MustCompile("`([^`]+)`")
	testIdentRe = regexp.MustCompile(`^Test[a-zA-Z0-9_]*\*?$`)
)

type testIndex struct {
	exact map[string]bool
	list  []string
}

func (idx *testIndex) match(ident string) bool {
	if strings.HasSuffix(ident, "*") {
		prefix := strings.TrimSuffix(ident, "*")
		for _, t := range idx.list {
			if strings.HasPrefix(t, prefix) {
				return true
			}
		}
		return false
	}
	return idx.exact[ident]
}

// collectTests scans the repository tree under root for func TestXxx(
// declarations in all _test.go files, skipping node_modules, .git and docs-site/node_modules.
func collectTests(root string) (*testIndex, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absRoot)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("capcheck: root %s is not a directory", root)
	}

	testMap := make(map[string]bool)
	var testList []string

	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			rel, err := filepath.Rel(absRoot, path)
			if err == nil {
				rel = filepath.ToSlash(rel)
				if rel == "node_modules" || rel == ".git" || rel == "docs-site/node_modules" ||
					d.Name() == "node_modules" || d.Name() == ".git" {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(bytes.NewReader(data))
		inBlockComment := false
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if inBlockComment {
				if idx := strings.Index(line, "*/"); idx >= 0 {
					inBlockComment = false
					line = strings.TrimSpace(line[idx+2:])
				} else {
					continue
				}
			}
			if strings.HasPrefix(line, "/*") {
				if idx := strings.Index(line, "*/"); idx >= 0 {
					line = strings.TrimSpace(line[idx+2:])
				} else {
					inBlockComment = true
					continue
				}
			}
			if strings.HasPrefix(line, "//") {
				continue
			}
			if m := funcTestRe.FindStringSubmatch(line); m != nil {
				name := m[1]
				if !testMap[name] {
					testMap[name] = true
					testList = append(testList, name)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &testIndex{exact: testMap, list: testList}, nil
}

func splitPipes(s string) []string {
	var cells []string
	var cur strings.Builder
	inBacktick := false
	escaped := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escaped {
			cur.WriteByte(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			cur.WriteByte(ch)
			continue
		}
		if ch == '`' {
			inBacktick = !inBacktick
			cur.WriteByte(ch)
			continue
		}
		if ch == '|' && !inBacktick {
			cells = append(cells, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteByte(ch)
	}
	cells = append(cells, cur.String())
	return cells
}

func parseTableRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	if !strings.Contains(trimmed, "|") {
		return nil
	}
	if strings.HasPrefix(trimmed, "|") {
		trimmed = trimmed[1:]
	}
	if strings.HasSuffix(trimmed, "|") {
		trimmed = trimmed[:len(trimmed)-1]
	}
	rawCells := splitPipes(trimmed)
	cells := make([]string, len(rawCells))
	for i, c := range rawCells {
		cells[i] = strings.TrimSpace(c)
	}
	return cells
}

func isDelimiterRow(cells []string) bool {
	if len(cells) == 0 {
		return false
	}
	for _, c := range cells {
		c = strings.TrimSpace(c)
		if len(c) == 0 {
			return false
		}
		for _, ch := range c {
			if ch != '-' && ch != ':' && ch != ' ' {
				return false
			}
		}
		if !strings.Contains(c, "-") {
			return false
		}
	}
	return true
}

// check parses the markdown tables in statusPath and validates capability evidence
// against tests under rootPath.
func check(statusPath, rootPath string) (int, []problem, error) {
	tests, err := collectTests(rootPath)
	if err != nil {
		return 0, nil, err
	}

	data, err := os.ReadFile(statusPath)
	if err != nil {
		return 0, nil, err
	}

	lines := bytes.Split(data, []byte("\n"))
	inFence := false
	var prevRow []string

	inTable := false
	tierColIdx := -1
	evidenceColIdx := -1

	supportedCount := 0
	var problems []problem

	for lineIdx, rawLine := range lines {
		lineNo := lineIdx + 1
		line := strings.TrimRight(string(rawLine), "\r")
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			inTable = false
			prevRow = nil
			continue
		}
		if inFence {
			continue
		}

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			inTable = false
			prevRow = nil
			continue
		}

		cells := parseTableRow(line)
		if cells == nil {
			inTable = false
			prevRow = nil
			continue
		}

		if isDelimiterRow(cells) {
			if len(prevRow) > 0 {
				tIdx, eIdx := -1, -1
				for i, h := range prevRow {
					clean := strings.Trim(strings.TrimSpace(h), "*_ ")
					if strings.EqualFold(clean, "Tier") {
						tIdx = i
					} else if strings.EqualFold(clean, "Evidence") {
						eIdx = i
					}
				}
				if tIdx >= 0 && eIdx >= 0 {
					inTable = true
					tierColIdx = tIdx
					evidenceColIdx = eIdx
				} else {
					inTable = false
				}
			} else {
				inTable = false
			}
			prevRow = nil
			continue
		}

		if !inTable {
			prevRow = cells
			continue
		}

		var tier, evidence string
		if tierColIdx < len(cells) {
			tier = cells[tierColIdx]
		}
		if evidenceColIdx < len(cells) {
			evidence = cells[evidenceColIdx]
		}

		cleanTier := strings.Trim(strings.TrimSpace(tier), "*_ ")
		if cleanTier == "Supported" {
			supportedCount++
			trimmedEvidence := strings.TrimSpace(evidence)
			if trimmedEvidence == "" {
				problems = append(problems, problem{
					file:   filepath.Base(statusPath),
					line:   lineNo,
					reason: "missing evidence",
				})
			} else {
				matches := backtickRe.FindAllStringSubmatch(trimmedEvidence, -1)
				seenInRow := make(map[string]bool)
				for _, m := range matches {
					raw := strings.TrimSpace(m[1])
					if testIdentRe.MatchString(raw) {
						if seenInRow[raw] {
							continue
						}
						seenInRow[raw] = true
						if !tests.match(raw) {
							problems = append(problems, problem{
								file:   filepath.Base(statusPath),
								line:   lineNo,
								reason: fmt.Sprintf("unknown test %s", raw),
							})
						}
					}
				}
			}
		}

		prevRow = cells
	}

	sort.Slice(problems, func(i, j int) bool {
		if problems[i].file != problems[j].file {
			return problems[i].file < problems[j].file
		}
		if problems[i].line != problems[j].line {
			return problems[i].line < problems[j].line
		}
		return problems[i].reason < problems[j].reason
	})

	return supportedCount, problems, nil
}

func run(statusPath, rootPath string, w io.Writer) (int, error) {
	n, problems, err := check(statusPath, rootPath)
	if err != nil {
		return 2, err
	}
	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintln(w, p.String())
		}
		return 1, nil
	}
	fmt.Fprintf(w, "capcheck: %d supported rows ok\n", n)
	return 0, nil
}

func main() {
	status := flag.String("status", "", "path to status markdown file (default docs-site/docs/status.md under root)")
	root := flag.String("root", ".", "repository root to scan for test files")
	flag.Parse()

	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "capcheck: unexpected positional arguments")
		os.Exit(2)
	}

	statusPath := *status
	if statusPath == "" {
		statusPath = filepath.Join(*root, "docs-site", "docs", "status.md")
	} else if _, err := os.Stat(statusPath); err != nil {
		relToRoot := filepath.Join(*root, statusPath)
		if _, err2 := os.Stat(relToRoot); err2 == nil {
			statusPath = relToRoot
		}
	}

	code, err := run(statusPath, *root, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
}
