package main

import (
	"path/filepath"
	"testing"
)

func TestDefaultDotEnvFileTargetsWorkingDirectory(t *testing.T) {
	workspace := t.TempDir()
	t.Chdir(workspace)

	if got, want := defaultDotEnvFile(), filepath.Join(workspace, ".env"); got != want {
		t.Fatalf("dotenv path = %q, want %q", got, want)
	}
}
