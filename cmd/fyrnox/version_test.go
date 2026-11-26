package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSingleSourceVersionConfiguration(t *testing.T) {
	root := filepath.Join("..", "..")
	raw, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if err != nil {
		t.Fatalf("failed to read root VERSION file: %v", err)
	}
	want := strings.TrimSpace(string(raw))
	if want == "" {
		t.Fatalf("root VERSION file is empty")
	}
	if AppVersion != want {
		t.Fatalf("AppVersion=%q does not match root VERSION=%q", AppVersion, want)
	}

	// Ensure no active release configuration hardcodes the release literal independently
	checkNoHardcodedVersion := func(relPath string) {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(root, relPath))
		if err != nil {
			t.Fatalf("failed to read %s: %v", relPath, err)
		}
		if strings.Contains(string(b), want) {
			t.Errorf("%s contains hardcoded version %q instead of deriving from root VERSION", relPath, want)
		}
	}

	checkNoHardcodedVersion(filepath.Join("cmd", "fyrnox", "version.go"))
	checkNoHardcodedVersion(filepath.Join("cmd", "installer", "main.go"))
	checkNoHardcodedVersion("installer.iss")
	checkNoHardcodedVersion(filepath.Join("scripts", "build_dist.ps1"))
}
