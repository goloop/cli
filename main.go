// Command goloop scaffolds and inspects goloop-first Go projects.
//
// Usage:
//
//	goloop new <name> [--module path] [--dir path] [--force]
//	goloop check [dir]     scan for AI-assistant traces (exit non-zero on any)
//	goloop doctor          check that expected tools are installed
//	goloop status          summarize the module in the current directory
//	goloop version         print the CLI version
//
// It is a small dispatcher over os.Args (no third-party dependencies) and never
// commits, pushes or deploys.
package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// run dispatches a subcommand and returns the process exit code.
func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "new":
		return cmdNew(args[1:])
	case "check":
		return cmdCheck(args[1:])
	case "doctor":
		return cmdDoctor(args[1:])
	case "status":
		return cmdStatus(args[1:])
	case "version":
		fmt.Printf("goloop %s %s/%s\n", Version, runtime.GOOS, runtime.GOARCH)
		return 0
	case "help", "-h", "--help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "goloop: unknown command %q\n\n", args[0])
		usage()
		return 2
	}
}

// usage prints the command overview.
func usage() {
	fmt.Fprint(os.Stderr, `goloop - scaffold and inspect goloop-first Go projects

usage:
  goloop new <name> [--module path] [--dir path] [--force]
  goloop check [dir]     scan for AI-assistant traces
  goloop doctor          check that expected tools are installed
  goloop status          summarize the module in the current directory
  goloop version         print the CLI version
`)
}
