package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRejectsUnknownManifestKeyBeforeClientCreation(t *testing.T) {
	manifestPath := filepath.Join(t.TempDir(), "settings.yml")
	if err := os.WriteFile(manifestPath, []byte("repository:\n  has_issue: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	getenv := func(key string) string {
		if key == "INPUT_SETTINGS-FILE" {
			return manifestPath
		}
		return ""
	}
	var output bytes.Buffer
	err := run(context.Background(), getenv, &output)
	if err == nil || !strings.Contains(err.Error(), "has_issue") {
		t.Fatalf("run error = %v, want unknown-field error", err)
	}
	if output.Len() != 0 {
		t.Fatalf("run wrote success output before validation: %q", output.String())
	}
}
