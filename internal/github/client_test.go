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

func TestApplyUsesGoGitHubComponentEndpoints(t *testing.T) {
	type request struct {
		method string
		path   string
		body   map[string]any
	}
	var got []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization header = %q", r.Header.Get("Authorization"))
		}
		entry := request{method: r.Method, path: r.URL.Path}
		if r.Method != http.MethodGet && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&entry.body); err != nil {
				t.Errorf("decode %s %s body: %v", r.Method, r.URL.Path, err)
			}
		}
		got = append(got, entry)

		switch r.Method + " " + r.URL.Path {
		case "GET /repos/acme/widget/actions/permissions":
			_, _ = w.Write([]byte(`{"enabled":true,"allowed_actions":"all","sha_pinning_required":true}`))
		case "GET /repos/acme/widget/actions/permissions/selected-actions":
			_, _ = w.Write([]byte(`{"github_owned_allowed":true,"verified_allowed":true,"patterns_allowed":["old/*"]}`))
		case "GET /repos/acme/widget/actions/permissions/workflow":
			_, _ = w.Write([]byte(`{"default_workflow_permissions":"read","can_approve_pull_request_reviews":false}`))
		case "GET /repos/acme/widget/environments/production":
			_, _ = w.Write([]byte(`{"wait_timer":4,"reviewers":[{"type":"User","id":11}],"deployment_branch_policy":{"protected_branches":true,"custom_branch_policies":false},"can_admins_bypass":false,"protection_rules":[{"type":"required_reviewers","prevent_self_review":false}]}`))
		case "GET /repos/acme/widget/rulesets":
			_, _ = w.Write([]byte(`[{"id":99,"name":"protect-main"}]`))
		case "GET /repos/acme/widget/pages":
			_, _ = w.Write([]byte(`{"build_type":"workflow","cname":"docs.example.com","public":true,"https_enforced":true}`))
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test-token", "acme/widget", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	waitTimer := 10
	pagesEnabled := true
	allowedActions := "selected"
	selectedPatterns := []string{"acme/*"}
	topics := []string{"go"}
	settings := settings.Settings{
		Repository: &settings.Repository{Description: new("managed"), Topics: &topics},
		Actions: &settings.Actions{
			Permissions: &settings.ActionPermissions{
				Enabled:        new(false),
				AllowedActions: &allowedActions,
				SelectedActions: &settings.SelectedActions{
					PatternsAllowed: &selectedPatterns,
				},
			},
			Workflow: &settings.WorkflowPermissions{
				DefaultWorkflowPermissions:   new("write"),
				CanApprovePullRequestReviews: new(true),
			},
		},
		Collaborators: []settings.Collaborator{{Username: "dev", Permission: "push"}},
		Environments:  []settings.Environment{{Name: "production", WaitTimer: &waitTimer}},
		Rulesets: []settings.Ruleset{
			{
				Name:        "protect-main",
				Target:      "branch",
				Enforcement: "active",
				Conditions:  settings.RulesetConditions{RefName: &settings.RefNameCondition{Include: []string{"~DEFAULT_BRANCH"}, Exclude: []string{}}},
				Rules:       []settings.Rule{{Type: "pull_request", Parameters: &settings.RuleParameters{RequiredApprovingReviewCount: new(1)}}},
			},
			{
				Name:        "protect-release",
				Target:      "tag",
				Enforcement: "active",
				Conditions:  settings.RulesetConditions{RefName: &settings.RefNameCondition{Include: []string{"v*"}, Exclude: []string{}}},
				Rules:       []settings.Rule{{Type: "creation"}},
			},
		},
		Pages: &settings.Pages{Enabled: &pagesEnabled, BuildType: "workflow", HTTPSEnforced: new(false)},
	}
	if err := client.Apply(context.Background(), settings); err != nil {
		t.Fatal(err)
	}

	want := []request{
		{method: http.MethodPatch, path: "/repos/acme/widget", body: map[string]any{"description": "managed"}},
		{method: http.MethodPut, path: "/repos/acme/widget/topics"},
		{method: http.MethodGet, path: "/repos/acme/widget/actions/permissions"},
		{method: http.MethodPut, path: "/repos/acme/widget/actions/permissions"},
		{method: http.MethodGet, path: "/repos/acme/widget/actions/permissions/selected-actions"},
		{method: http.MethodPut, path: "/repos/acme/widget/actions/permissions/selected-actions"},
		{method: http.MethodGet, path: "/repos/acme/widget/actions/permissions/workflow"},
		{method: http.MethodPut, path: "/repos/acme/widget/actions/permissions/workflow"},
		{method: http.MethodPut, path: "/repos/acme/widget/collaborators/dev"},
		{method: http.MethodGet, path: "/repos/acme/widget/environments/production"},
		{method: http.MethodPut, path: "/repos/acme/widget/environments/production"},
		{method: http.MethodGet, path: "/repos/acme/widget/rulesets"},
		{method: http.MethodPut, path: "/repos/acme/widget/rulesets/99"},
		{method: http.MethodPost, path: "/repos/acme/widget/rulesets"},
		{method: http.MethodGet, path: "/repos/acme/widget/pages"},
		{method: http.MethodPut, path: "/repos/acme/widget/pages"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d requests, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].method != want[i].method || got[i].path != want[i].path {
			t.Errorf("request %d = %s %s, want %s %s", i, got[i].method, got[i].path, want[i].method, want[i].path)
		}
	}
	if got[0].body["description"] != "managed" {
		t.Errorf("repository update = %#v", got[0].body)
	}
	if !reflect.DeepEqual(got[1].body["names"], []any{"go"}) {
		t.Errorf("topics update = %#v", got[1].body)
	}
	if got[3].body["enabled"] != false || got[3].body["allowed_actions"] != "selected" || got[3].body["sha_pinning_required"] != true {
		t.Errorf("Actions permissions update = %#v", got[3].body)
	}
	if got[5].body["patterns_allowed"] == nil || got[7].body["default_workflow_permissions"] != "write" {
		t.Errorf("selected Actions/workflow updates = %#v, %#v", got[5].body, got[7].body)
	}
	if got[8].body["permission"] != "push" {
		t.Errorf("collaborator update = %#v", got[8].body)
	}
	if got[10].body["wait_timer"] != float64(10) || got[10].body["can_admins_bypass"] != false || got[10].body["prevent_self_review"] != false {
		t.Errorf("environment update did not preserve omitted settings and apply requested timer: %#v", got[10].body)
	}
	if got[12].body["name"] != "protect-main" || got[12].body["source"] != "acme/widget" {
		t.Errorf("ruleset update payload = %#v", got[12].body)
	}
	if got[13].body["name"] != "protect-release" || got[13].body["target"] != "tag" {
		t.Errorf("ruleset create payload = %#v", got[13].body)
	}
	if got[15].body["cname"] != "docs.example.com" || got[15].body["https_enforced"] != false {
		t.Errorf("Pages update did not preserve CNAME or apply HTTPS setting: %#v", got[15].body)
	}
}

func TestApplyDeletesOnlyExplicitlyAbsentResources(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodGet && r.URL.Path == "/repos/acme/widget/rulesets" {
			_, _ = w.Write([]byte(`[{"id":7,"name":"old"}]`))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "test-token", "acme/widget", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Apply(context.Background(), settings.Settings{
		Collaborators: []settings.Collaborator{{Username: "former", State: "absent"}},
		Environments:  []settings.Environment{{Name: "retired", State: "absent"}},
		Rulesets:      []settings.Ruleset{{Name: "old", State: "absent"}},
	}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"DELETE /repos/acme/widget/collaborators/former",
		"DELETE /repos/acme/widget/environments/retired",
		"GET /repos/acme/widget/rulesets",
		"DELETE /repos/acme/widget/rulesets/7",
	}
	if !reflect.DeepEqual(methods, want) {
		t.Fatalf("methods = %#v, want %#v", methods, want)
	}
}

func TestApplyPagesCreateAndDisableStates(t *testing.T) {
	tests := []struct {
		name      string
		enabled   bool
		getStatus int
		getBody   string
		want      []string
	}{
		{
			name:      "create",
			enabled:   true,
			getStatus: http.StatusNotFound,
			want: []string{
				"GET /repos/acme/widget/pages",
				"POST /repos/acme/widget/pages",
			},
		},
		{
			name:      "disable",
			enabled:   false,
			getStatus: http.StatusOK,
			getBody:   `{"build_type":"workflow"}`,
			want: []string{
				"GET /repos/acme/widget/pages",
				"DELETE /repos/acme/widget/pages",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var methods []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				methods = append(methods, r.Method+" "+r.URL.Path)
				if r.Method == http.MethodGet {
					w.WriteHeader(test.getStatus)
					_, _ = w.Write([]byte(test.getBody))
					return
				}
				if r.Method == http.MethodPost {
					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write([]byte(`{"build_type":"workflow"}`))
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()
			client, err := NewClient(server.URL, "test-token", "acme/widget", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			pages := &settings.Pages{Enabled: new(test.enabled)}
			if test.enabled {
				pages.BuildType = "workflow"
			}
			if err := client.Apply(context.Background(), settings.Settings{Pages: pages}); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(methods, test.want) {
				t.Fatalf("methods = %#v, want %#v", methods, test.want)
			}
		})
	}
}

func TestNewClientRejectsInvalidRepositoryPathSegments(t *testing.T) {
	for _, repository := range []string{
		"acme",
		"acme/widget/other",
		"../widget",
		"acme/widget?per_page=1",
		"acme//widget",
	} {
		t.Run(repository, func(t *testing.T) {
			if _, err := NewClient("https://api.github.com", "test-token", repository, nil); err == nil {
				t.Fatalf("NewClient accepted repository path %q", repository)
			}
		})
	}
}
