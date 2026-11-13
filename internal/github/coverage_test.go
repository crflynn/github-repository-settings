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

func TestApplyAdditionalRepositoryActionsAndTeamSettings(t *testing.T) {
	type request struct {
		method string
		body   map[string]any
	}
	var got = make(map[string]request)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry := request{method: r.Method}
		if r.Method != http.MethodGet && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&entry.body); err != nil {
				t.Errorf("decode %s %s body: %v", r.Method, r.URL.Path, err)
			}
		}
		got[r.URL.Path] = entry
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test-token", "acme/widget", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	desired := settings.Settings{
		Repository: &settings.Repository{
			DefaultBranch:                 new("trunk"),
			IsTemplate:                    new(false),
			PullRequestCreationPolicy:     new("collaborators_only"),
			DependabotSecurityUpdates:     new(true),
			VulnerabilityAlerts:           new(false),
			PrivateVulnerabilityReporting: new(true),
			CodeSecurity:                  new(true),
			SecretScanning:                new(true),
			SecretScanningPushProtection:  new(true),
		},
		Actions: &settings.Actions{
			Permissions: &settings.ActionPermissions{SHAPinningRequired: new(true)},
			ForkPR: &settings.ForkPRSettings{
				RunWorkflows:            new(true),
				SendWriteTokens:         new(false),
				SendSecretsAndVariables: new(false),
				RequireApproval:         new(true),
			},
			Retention: &settings.ArtifactRetention{Days: new(30)},
		},
		Teams: []settings.TeamRepository{{Slug: "platform", Permission: "maintain"}},
	}
	if err := client.Apply(context.Background(), desired); err != nil {
		t.Fatal(err)
	}

	for path, method := range map[string]string{
		"/repos/acme/widget":                                                     http.MethodPatch,
		"/repos/acme/widget/vulnerability-alerts":                                http.MethodDelete,
		"/repos/acme/widget/private-vulnerability-reporting":                     http.MethodPut,
		"/repos/acme/widget/actions/permissions":                                 http.MethodPut,
		"/repos/acme/widget/actions/permissions/fork-pr-workflows-private-repos": http.MethodPut,
		"/repos/acme/widget/actions/permissions/artifact-and-log-retention":      http.MethodPut,
		"/orgs/acme/teams/platform/repos/acme/widget":                            http.MethodPut,
	} {
		if got[path].method != method {
			t.Errorf("request %s method = %q, want %q; all requests: %#v", path, got[path].method, method, got)
		}
	}

	repositoryBody := got["/repos/acme/widget"].body
	if repositoryBody["default_branch"] != "trunk" || repositoryBody["is_template"] != false ||
		repositoryBody["pull_request_creation_policy"] != "collaborators_only" {
		t.Errorf("repository update body = %#v", repositoryBody)
	}
	security, ok := repositoryBody["security_and_analysis"].(map[string]any)
	if !ok || security["dependabot_security_updates"].(map[string]any)["status"] != "enabled" ||
		security["secret_scanning_push_protection"].(map[string]any)["status"] != "enabled" {
		t.Errorf("repository security payload = %#v", repositoryBody["security_and_analysis"])
	}
	permissionsBody := got["/repos/acme/widget/actions/permissions"].body
	if permissionsBody["sha_pinning_required"] != true {
		t.Errorf("Actions permissions body = %#v", permissionsBody)
	}
	forkBody := got["/repos/acme/widget/actions/permissions/fork-pr-workflows-private-repos"].body
	if forkBody["run_workflows_from_fork_pull_requests"] != true || forkBody["send_write_tokens_to_workflows"] != false ||
		forkBody["send_secrets_and_variables"] != false || forkBody["require_approval_for_fork_pr_workflows"] != true {
		t.Errorf("fork workflow settings body = %#v", forkBody)
	}
	if retention := got["/repos/acme/widget/actions/permissions/artifact-and-log-retention"].body; retention["days"] != float64(30) {
		t.Errorf("retention body = %#v", retention)
	}
	teamBody := got["/orgs/acme/teams/platform/repos/acme/widget"].body
	if !reflect.DeepEqual(teamBody, map[string]any{"permission": "maintain"}) {
		t.Errorf("team grant body = %#v", teamBody)
	}
}
func TestApplyRemainingAuditCandidatesThroughManifest(t *testing.T) {
	const secret = "smoke-secret"
	t.Setenv("SETTINGS_WEBHOOK_SECRET", secret)
	var calls []string
	var webhookBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := r.Method + " " + r.URL.Path
		calls = append(calls, request)
		var body map[string]any
		if r.Method != http.MethodGet && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode %s body: %v", request, err)
			}
		}
		switch request {
		case "GET /repos/acme/widget/autolinks",
			"GET /repos/acme/widget/keys",
			"GET /repos/acme/widget/properties/values",
			"GET /repos/acme/widget/hooks":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[]`))
		case "POST /repos/acme/widget/hooks":
			webhookBody = body
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":7,"config":{"url":"https://hooks.example/notify"}}`))
		case "POST /repos/acme/widget/autolinks", "POST /repos/acme/widget/keys":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":8}`))
		case "PUT /repos/acme/widget/interaction-limits",
			"PATCH /repos/acme/widget/properties/values":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s", request)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test-token", "acme/widget", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	desired, err := settings.Decode([]byte(`autolinks:
  - key_prefix: JIRA-
    url_template: https://issues.example/<num>/
    is_alphanumeric: false
deploy_keys:
  - key: ssh-ed25519 YWJj ci
    title: automation
    read_only: true
interaction_limits:
  limit: collaborators_only
  expiry: one_week
custom_properties:
  - name: services
    value: [api, web]
webhooks:
  - url: https://hooks.example/notify
    events: [push]
    secret_env: SETTINGS_WEBHOOK_SECRET
`))
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Apply(context.Background(), desired); err != nil {
		t.Fatal(err)
	}

	wantCalls := []string{
		"GET /repos/acme/widget/autolinks",
		"POST /repos/acme/widget/autolinks",
		"GET /repos/acme/widget/keys",
		"POST /repos/acme/widget/keys",
		"PUT /repos/acme/widget/interaction-limits",
		"GET /repos/acme/widget/properties/values",
		"PATCH /repos/acme/widget/properties/values",
		"GET /repos/acme/widget/hooks",
		"POST /repos/acme/widget/hooks",
	}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("API calls = %#v, want %#v", calls, wantCalls)
	}
	config, ok := webhookBody["config"].(map[string]any)
	if !ok || config["secret"] != secret {
		t.Errorf("webhook secret was not supplied from the referenced environment variable: %#v", webhookBody)
	}
}
