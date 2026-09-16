// Command doclinks checks public documentation for relative file links
// that do not exist. Usage: go run ./internal/tools/doclinks -root ..
// when run from ./forge.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	root := flag.String("root", ".", "repository root to check (contains README.md, docs/, skills/)")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "doclinks: unexpected positional arguments")
		os.Exit(2)
	}
	code, err := run(*root, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
}
