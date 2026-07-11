package main

import (
	"fmt"
	"regexp"
)

// Version is the CLI version. Override at build time with -ldflags.
var Version = "v0.0.0-dev"

// nameRe is the allowed project-name shape: a lowercase letter followed by
// lowercase letters, digits and single dashes.
var nameRe = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// reserved are project names that must not be used.
var reserved = map[string]bool{
	"go": true, "test": true, "goloop": true, "new": true,
	"build": true, "internal": true, "vendor": true,
}

// validateName reports whether name is an acceptable project name.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("project name is required")
	}
	if !nameRe.MatchString(name) {
		return fmt.Errorf("invalid name %q: use lowercase letters, digits and single dashes, starting with a letter", name)
	}
	if reserved[name] {
		return fmt.Errorf("%q is a reserved name", name)
	}
	return nil
}
