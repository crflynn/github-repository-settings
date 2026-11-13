package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github-repository-settings/internal/settings"
)

func TestApplyWebhooksDeletesOnlyExplicitlyAbsentURL(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch r.Method + " " + r.URL.Path {
		case "GET /repos/acme/widget/hooks":
			_, _ = w.Write([]byte(`[{"id":4,"config":{"url":"https://hooks.example.test/remove"}},{"id":5,"config":{"url":"https://hooks.example.test/keep"}}]`))
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "test-token", "acme/widget", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if err := client.applyWebhooks(context.Background(), []settings.Webhook{{URL: "https://hooks.example.test/remove", State: "absent"}}); err != nil {
		t.Fatal(err)
	}
	want := []string{"GET /repos/acme/widget/hooks", "DELETE /repos/acme/widget/hooks/4"}
	if !reflect.DeepEqual(requests, want) {
		t.Fatalf("requests = %#v, want %#v", requests, want)
	}
}
func TestApplyWebhooksRejectsEventUpdateWithoutSecretBeforeMutation(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[{"id":4,"events":["push"],"active":true,"config":{"url":"https://hooks.example.test/keep","content_type":"json","insecure_ssl":"0"}}]`))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "test-token", "acme/widget", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	err = client.applyWebhooks(context.Background(), []settings.Webhook{{
		URL: "https://hooks.example.test/keep", Events: []string{"issues"},
	}})
	if err == nil {
		t.Fatal("event update without a reusable secret unexpectedly succeeded")
	}
	if want := []string{"GET /repos/acme/widget/hooks"}; !reflect.DeepEqual(requests, want) {
		t.Fatalf("requests = %#v, want no mutation before secret validation", requests)
	}
}

func TestApplyWebhooksConfigUpdatePreservesWriteOnlySecret(t *testing.T) {
	var update map[string]any
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[{"id":4,"events":["push"],"active":true,"config":{"url":"https://hooks.example.test/keep","content_type":"json","insecure_ssl":"0"}}]`))
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			t.Errorf("decode webhook config update: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "test-token", "acme/widget", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if err := client.applyWebhooks(context.Background(), []settings.Webhook{{
		URL: "https://hooks.example.test/keep", Events: []string{"push"}, ContentType: "form",
	}}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"GET /repos/acme/widget/hooks",
		"PATCH /repos/acme/widget/hooks/4/config",
	}
	if !reflect.DeepEqual(requests, want) {
		t.Fatalf("requests = %#v, want %#v", requests, want)
	}
	if _, included := update["secret"]; included {
		t.Fatalf("config update included a write-only secret: %#v", update)
	}
	if update["content_type"] != "form" {
		t.Errorf("config update = %#v", update)
	}
}
