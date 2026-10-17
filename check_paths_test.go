package main

import (
	"path/filepath"
	"testing"
)

func TestScanTracesMissingDirectory(t *testing.T) {
	_, err := scanTraces(filepath.Join(t.TempDir(), "missing"), false)
	if err == nil {
		t.Fatal("missing directory reported as clean")
	}
}
