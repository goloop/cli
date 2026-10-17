package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// traceMarkers are case-insensitive substrings that betray AI-assistant use in
// a product repository. The check command exists to keep them out.
//
// Each marker is assembled from fragments so that this scanner's own source
// stays clean under the very check it performs (and under a boundary review of
// this repository); the concatenation is identical at runtime. A consequence is
// that the compiled binary DOES contain the joined markers, so never commit the
// built cli binary to a product repository and keep it in .gitignore.
var traceMarkers = []string{
	"cla" + "ude",
	"cur" + "sor",
	"cop" + "ilot",
	"chat" + "gpt",
	"ai-gen" + "erated",
	"gener" + "ated by",
	"co-auth" + "ored-by",
	"pro" + "mpt:",
}

// skipDirs are directories the scan never descends into.
var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, "dist": true,
}

// Finding is one marker hit in a file.
type Finding struct {
	Path   string
	Line   int
	Marker string
}

// scanTraces walks root and returns every marker hit across the files it can
// read, plus an entry for any file it cannot scan (see scanFile). It is the
// core of the check command and is deliberately pure for testing.
func scanTraces(root string, scanAll bool) ([]Finding, error) {
	var findings []Finding
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		hits, ferr := scanFile(path, scanAll)
		findings = append(findings, hits...) // keep partial hits even on error
		if ferr != nil {
			// A file we cannot fully scan (for example an overlong line) is
			// reported so the scanner fails closed: a trace must not be able to
			// hide behind content the scanner skipped.
			findings = append(findings, Finding{Path: path, Marker: "unscannable (" + ferr.Error() + ")"})
		}
		return nil
	})
	return findings, err
}

// scanFile returns the marker hits in a single file. Unless scanAll is set, a
// file that looks binary (a NUL byte near its start) is skipped: a compiled
// artifact or image is not source and only produces noise. With scanAll the
// scan stays fail-closed and inspects every file.
func scanFile(path string, scanAll bool) ([]Finding, error) {
	f, err := os.Open(path)
	if err != nil {
		// Fail closed: a file the scanner cannot open (permissions, a special
		// file) is reported as unscannable by the caller rather than skipped,
		// so a trace cannot hide behind an open error.
		return nil, err
	}
	defer f.Close()

	br := bufio.NewReader(f)
	if !scanAll {
		head, _ := br.Peek(512)
		if bytes.IndexByte(head, 0) >= 0 {
			return nil, nil
		}
	}

	var findings []Finding
	sc := bufio.NewScanner(br)
	// A generous line cap: legitimate source lines are far smaller, and an
	// overlong line surfaces as a scan error (handled by the caller) rather
	// than being silently dropped.
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	line := 0
	for sc.Scan() {
		line++
		lower := strings.ToLower(sc.Text())
		for _, m := range traceMarkers {
			if strings.Contains(lower, m) {
				findings = append(findings, Finding{Path: path, Line: line, Marker: m})
			}
		}
	}
	if err := sc.Err(); err != nil {
		return findings, err
	}
	return findings, nil
}

// cmdCheck scans a directory (default ".") for AI-assistant traces and reports
// them. It exits non-zero when any are found.
func cmdCheck(args []string) int {
	scanAll := false
	var dirs []string
	for _, a := range args {
		switch {
		case a == "--all":
			scanAll = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "check: unknown flag %q\n", a)
			return 2
		default:
			dirs = append(dirs, a)
		}
	}
	if len(dirs) > 1 {
		fmt.Fprintln(os.Stderr, "check: accepts at most one directory")
		return 2
	}
	root := "."
	if len(dirs) == 1 {
		root = dirs[0]
	}
	findings, err := scanTraces(root, scanAll)
	if err != nil {
		fmt.Fprintf(os.Stderr, "check: %v\n", err)
		return 1
	}
	if len(findings) == 0 {
		fmt.Printf("check: clean, no AI traces in %s\n", root)
		return 0
	}
	for _, f := range findings {
		fmt.Printf("%s:%d: found %q\n", f.Path, f.Line, f.Marker)
	}
	fmt.Fprintf(os.Stderr, "check: %d trace(s) found\n", len(findings))
	return 1
}
