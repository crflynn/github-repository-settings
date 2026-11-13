package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestRunDryRunReportsPlannedWritesWithoutSendingThem(t *testing.T) {
	var reads, writes atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writes.Add(1)
			t.Errorf("dry-run sent %s %s to the API server", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		reads.Add(1)
		if r.URL.Path != "/repos/acme/widget/actions/permissions" {
			t.Errorf("unexpected read request path %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"enabled":true,"allowed_actions":"all"}`))
	}))
	defer server.Close()

	manifestPath := filepath.Join(t.TempDir(), "settings.yml")
	if err := os.WriteFile(manifestPath, []byte("actions:\n  permissions:\n    enabled: false\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	environment := map[string]string{
		"INPUT_DRY-RUN":       "true",
		"INPUT_SETTINGS-FILE": manifestPath,
		"INPUT_TOKEN":         "read-only-token",
		"GITHUB_API_URL":      server.URL,
		"GITHUB_REPOSITORY":   "acme/widget",
	}
	getenv := func(key string) string { return environment[key] }
	var output bytes.Buffer
	if err := run(context.Background(), getenv, &output); err != nil {
		t.Fatal(err)
	}
	if got := reads.Load(); got != 1 {
		t.Errorf("API reads = %d, want 1", got)
	}
	if got := writes.Load(); got != 0 {
		t.Errorf("API writes = %d, want 0", got)
	}
	if want := "Dry-run plan for acme/widget (no write requests were sent):\nWould PUT /repos/acme/widget/actions/permissions\n"; output.String() != want {
		t.Fatalf("dry-run output = %q, want %q", output.String(), want)
	}
	if strings.Contains(output.String(), "read-only-token") {
		t.Fatal("dry-run output exposed the API token")
	}
}
