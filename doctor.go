package main

import (
	"fmt"
	"os"
	"os/exec"
)

// tool is an external program the doctor checks for.
type tool struct {
	name     string
	required bool
}

// tools are the external programs a goloop project typically needs.
var tools = []tool{
	{"go", true},
	{"git", true},
	{"docker", false},
	{"node", false},
}

// cmdDoctor reports which expected tools are present. It exits non-zero when a
// required tool is missing.
func cmdDoctor(args []string) int {
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "doctor: takes no arguments")
		return 2
	}
	ok := true
	for _, t := range tools {
		path, err := exec.LookPath(t.name)
		switch {
		case err == nil:
			fmt.Printf("  ok       %-8s %s\n", t.name, path)
		case t.required:
			fmt.Printf("  MISSING  %-8s (required)\n", t.name)
			ok = false
		default:
			fmt.Printf("  -        %-8s (optional, not found)\n", t.name)
		}
	}
	if !ok {
		fmt.Println("doctor: some required tools are missing")
		return 1
	}
	fmt.Println("doctor: all required tools present")
	return 0
}
