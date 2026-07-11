package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// cmdStatus prints a short overview of the Go module in the current directory.
func cmdStatus(args []string) int {
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "status: takes no arguments")
		return 2
	}

	f, err := os.Open("go.mod")
	if err != nil {
		fmt.Println("status: no go.mod in the current directory")
		return 0
	}
	defer f.Close()

	module, goVersion, goloopDeps, err := parseGoMod(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "status: reading go.mod: %v\n", err)
		return 1
	}

	fmt.Printf("module:  %s\n", module)
	fmt.Printf("go:      %s\n", goVersion)
	if len(goloopDeps) == 0 {
		fmt.Println("goloop:  (none)")
	} else {
		fmt.Printf("goloop:  %s\n", strings.Join(goloopDeps, ", "))
	}
	return 0
}

// parseGoMod extracts the module path, Go version and required goloop
// dependencies from a go.mod file. It understands both single-line and block
// require directives and deliberately ignores replace directives, which point a
// dependency elsewhere rather than declaring one.
func parseGoMod(r *os.File) (module, goVersion string, goloopDeps []string, err error) {
	const goloopPrefix = "github.com/goloop/"

	// block tracks which parenthesised directive, if any, we are inside.
	// Only "require" blocks contribute dependencies.
	var block string

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}

		if block != "" {
			if line == ")" {
				block = ""
				continue
			}
			if block == "require" && strings.HasPrefix(line, goloopPrefix) {
				goloopDeps = append(goloopDeps, strings.Fields(line)[0])
			}
			continue
		}

		switch {
		case strings.HasPrefix(line, "module "):
			module = strings.TrimSpace(strings.TrimPrefix(line, "module"))
		case strings.HasPrefix(line, "go "):
			goVersion = strings.TrimSpace(strings.TrimPrefix(line, "go"))
		case line == "require (":
			block = "require"
		case line == "replace (", line == "exclude (", line == "retract (":
			// Enter a block whose entries are not required dependencies.
			block = "ignore"
		case strings.HasPrefix(line, "require "):
			// Single-line require: fields are [require, path, version].
			if fields := strings.Fields(line); len(fields) >= 2 &&
				strings.HasPrefix(fields[1], goloopPrefix) {
				goloopDeps = append(goloopDeps, fields[1])
			}
		}
	}
	if serr := sc.Err(); serr != nil {
		return "", "", nil, serr
	}
	return module, goVersion, goloopDeps, nil
}
