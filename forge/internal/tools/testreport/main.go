// Command testreport summarizes a go test -json stream from stdin.
// --out writes a copy of the report; --require-no-skip is repeatable and
// accepts a regular expression matching package paths whose tests must not skip.
// --allow-skip is repeatable and accepts '<package-regexp> <test-regexp>'
// to exempt documented skips.
package main

import (
	"flag"
	"fmt"
	"os"
)

type patterns []string

func (p *patterns) String() string         { return fmt.Sprint([]string(*p)) }
func (p *patterns) Set(value string) error { *p = append(*p, value); return nil }

func main() {
	var opts options
	flags := flag.NewFlagSet("testreport", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&opts.out, "out", "", "write the report to a file")
	flags.Var(&opts.requireNoSkip, "require-no-skip", "package regexp that forbids skipped tests (repeatable)")
	flags.Var(&opts.allowSkip, "allow-skip", "package and test regexp pair '<package-regexp> <test-regexp>' that permits skipped tests (repeatable)")
	if err := flags.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "testreport: unexpected positional arguments")
		os.Exit(2)
	}
	code, err := run(os.Stdin, os.Stdout, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "testreport:", err)
	}
	os.Exit(code)
}
