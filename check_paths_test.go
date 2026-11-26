package main

import (
	"path/filepath"
	"testing"
)

func TestScanMarkersMissingDirectory(t *testing.T) {
	_, err := scanMarkers(filepath.Join(t.TempDir(), "missing"), false, []string{"draft-only"}, "")
	if err == nil {
		t.Fatal("missing directory reported as clean")
	}
}
