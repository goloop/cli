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

// readMarkers loads one case-insensitive literal per nonblank line.
func readMarkers(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var markers []string
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		if m := strings.TrimSpace(sc.Text()); m != "" {
			markers = append(markers, strings.ToLower(m))
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(markers) == 0 {
		return nil, fmt.Errorf("empty rule file %q", path)
	}
	return markers, nil
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

// scanMarkers walks root and returns every marker hit across the files it can
// read, plus an entry for any file it cannot scan (see scanFile). It is the
// core of the check command and is deliberately pure for testing.
func scanMarkers(root string, scanAll bool, markers []string, ruleFile string) ([]Finding, error) {
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
		if ruleFile != "" {
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			if abs == ruleFile {
				return nil
			}
		}
		hits, ferr := scanFile(path, scanAll, markers)
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
func scanFile(path string, scanAll bool, markers []string) ([]Finding, error) {
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
		for _, m := range markers {
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

// cmdCheck scans a directory (default ".") for project-defined markers and reports
// them. It exits non-zero when any are found.
func cmdCheck(args []string) int {
	scanAll := false
	var dirs []string
	ruleFile := os.Getenv("GOLOOP_DENY_FILE")
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--deny-file":
			i++
			if i >= len(args) || strings.HasPrefix(args[i], "-") {
				fmt.Fprintln(os.Stderr, "check: --deny-file requires a path")
				return 2
			}
			ruleFile = args[i]
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
	if ruleFile == "" {
		fmt.Fprintln(os.Stderr, "check: specify --deny-file or GOLOOP_DENY_FILE")
		return 2
	}
	ruleFile, err := filepath.Abs(ruleFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	markers, err := readMarkers(ruleFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	findings, err := scanMarkers(root, scanAll, markers, ruleFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "check: %v\n", err)
		return 1
	}
	if len(findings) == 0 {
		fmt.Printf("check: clean, no matches in %s\n", root)
		return 0
	}
	for _, f := range findings {
		fmt.Printf("%s:%d: found %q\n", f.Path, f.Line, f.Marker)
	}
	fmt.Fprintf(os.Stderr, "check: %d match(es) found\n", len(findings))
	return 1
}
