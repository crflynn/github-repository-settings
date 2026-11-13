package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github-repository-settings/internal/settings"
)

func TestApplyAutolinksCreatesMissingEntry(t *testing.T) {
	var gotBody map[string]any
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`[]`))
		case http.MethodPost:
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Errorf("decode create request: %v", err)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":12,"key_prefix":"JIRA-","url_template":"https://issues.example/<num>/","is_alphanumeric":false}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()
	client := autolinkTestClient(t, server)
	falseValue := false
	if err := client.applyAutolinks(context.Background(), []settings.Autolink{{KeyPrefix: "JIRA-", URLTemplate: "https://issues.example/<num>/", IsAlphanumeric: &falseValue}}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(methods, []string{"GET /repos/acme/widget/autolinks", "POST /repos/acme/widget/autolinks"}) {
		t.Fatalf("requests = %#v", methods)
	}
	if !reflect.DeepEqual(gotBody, map[string]any{"key_prefix": "JIRA-", "url_template": "https://issues.example/<num>/", "is_alphanumeric": false}) {
		t.Errorf("create body = %#v", gotBody)
	}
}

func TestApplyAutolinksKeepsUnlistedAndMatchingEntries(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet {
			t.Errorf("unexpected mutating request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_, _ = w.Write([]byte(`[{"id":3,"key_prefix":"OLD-","url_template":"https://old.example/<num>","is_alphanumeric":true},{"id":4,"key_prefix":"JIRA-","url_template":"https://issues.example/<num>","is_alphanumeric":true}]`))
	}))
	defer server.Close()
	client := autolinkTestClient(t, server)
	if err := client.applyAutolinks(context.Background(), []settings.Autolink{{KeyPrefix: "JIRA-", URLTemplate: "https://issues.example/<num>"}}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(methods, []string{"GET /repos/acme/widget/autolinks"}) {
		t.Errorf("requests = %#v, want only list", methods)
	}
}

func TestApplyAutolinksRejectsUnupdatableDifference(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		_, _ = w.Write([]byte(`[{"id":4,"key_prefix":"JIRA-","url_template":"https://old.example/<num>","is_alphanumeric":true}]`))
	}))
	defer server.Close()
	client := autolinkTestClient(t, server)
	err := client.applyAutolinks(context.Background(), []settings.Autolink{{KeyPrefix: "JIRA-", URLTemplate: "https://new.example/<num>"}})
	if err == nil || !strings.Contains(err.Error(), "GitHub does not support updating autolinks") {
		t.Fatalf("applyAutolinks() error = %v, want unsupported-update error", err)
	}
	if !reflect.DeepEqual(methods, []string{"GET /repos/acme/widget/autolinks"}) {
		t.Errorf("requests = %#v, want only list", methods)
	}
}

func TestApplyAutolinksDeletesOnlyExplicitlyAbsentEntry(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`[{"id":17,"key_prefix":"JIRA-","url_template":"https://issues.example/<num>"},{"id":18,"key_prefix":"KEEP-","url_template":"https://keep.example/<num>"}]`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()
	client := autolinkTestClient(t, server)
	if err := client.applyAutolinks(context.Background(), []settings.Autolink{{KeyPrefix: "JIRA-", State: "absent"}}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(methods, []string{"GET /repos/acme/widget/autolinks", "DELETE /repos/acme/widget/autolinks/17"}) {
		t.Errorf("requests = %#v", methods)
	}
}

func autolinkTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	client, err := NewClient(server.URL, "test-token", "acme/widget", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	return client
}
