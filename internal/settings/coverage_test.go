package settings

import (
	"reflect"
	"testing"
)

func TestDecodeAcceptsAdditionalRepositoryActionsAndTeamSettings(t *testing.T) {
	got, err := Decode([]byte(`repository:
  default_branch: trunk
  is_template: false
  has_pull_requests: true
  pull_request_creation_policy: collaborators_only
  dependabot_security_updates: true
  vulnerability_alerts: false
  private_vulnerability_reporting: true
  code_security: true
  secret_scanning: true
  secret_scanning_push_protection: false
actions:
  permissions:
    sha_pinning_required: true
  fork_pr:
    run_workflows_from_fork_pull_requests: true
    send_write_tokens_to_workflows: false
    send_secrets_and_variables: false
    require_approval_for_fork_pr_workflows: true
  retention:
    days: 30
teams:
  - slug: platform
    permission: maintain
`))
	if err != nil {
		t.Fatal(err)
	}
	if got.Repository == nil || got.Repository.DefaultBranch == nil || *got.Repository.DefaultBranch != "trunk" ||
		got.Repository.IsTemplate == nil || *got.Repository.IsTemplate ||
		got.Repository.PullRequestCreationPolicy == nil || *got.Repository.PullRequestCreationPolicy != "collaborators_only" ||
		got.Repository.VulnerabilityAlerts == nil || *got.Repository.VulnerabilityAlerts ||
		got.Repository.PrivateVulnerabilityReporting == nil || !*got.Repository.PrivateVulnerabilityReporting ||
		got.Repository.SecretScanningPushProtection == nil || *got.Repository.SecretScanningPushProtection {
		t.Fatalf("repository settings were not decoded with explicit values preserved: %#v", got.Repository)
	}
	if got.Actions == nil || got.Actions.Permissions == nil || got.Actions.Permissions.SHAPinningRequired == nil ||
		!*got.Actions.Permissions.SHAPinningRequired || got.Actions.ForkPR == nil || got.Actions.ForkPR.RunWorkflows == nil ||
		!*got.Actions.ForkPR.RunWorkflows || got.Actions.Retention == nil || got.Actions.Retention.Days == nil ||
		*got.Actions.Retention.Days != 30 {
		t.Fatalf("Actions settings were not decoded: %#v", got.Actions)
	}
	if len(got.Teams) != 1 || got.Teams[0].Slug != "platform" || got.Teams[0].Permission != "maintain" {
		t.Fatalf("team grant was not decoded: %#v", got.Teams)
	}
}

func TestDecodeRejectsInvalidAdditionalSettings(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "pull request creation policy",
			data: "repository:\n  pull_request_creation_policy: everyone\n",
		},
		{
			name: "empty fork PR settings",
			data: "actions:\n  fork_pr: {}\n",
		},
		{
			name: "retention above API limit",
			data: "actions:\n  retention:\n    days: 401\n",
		},
		{
			name: "invalid team permission",
			data: "teams:\n  - slug: platform\n    permission: owner\n",
		},
		{
			name: "duplicate team slug",
			data: "teams:\n  - slug: platform\n    permission: push\n  - slug: PLATFORM\n    permission: pull\n",
		},
		{
			name: "empty Copilot review rule",
			data: "rulesets:\n  - name: copilot\n    target: branch\n    enforcement: active\n    conditions:\n      ref_name: {include: [main], exclude: []}\n    rules:\n      - type: copilot_code_review\n        parameters: {}\n",
		},
		{
			name: "invalid interaction limit",
			data: "interaction_limits:\n  limit: everyone\n",
		},
		{
			name: "autolink removal with update fields",
			data: "autolinks:\n  - key_prefix: JIRA-\n    state: absent\n    url_template: https://issues.example/<num>/\n",
		},
		{
			name: "non-string custom property value",
			data: "custom_properties:\n  - name: team\n    value: 42\n",
		},
		{
			name: "null custom property value",
			data: "custom_properties:\n  - name: team\n    value: null\n",
		},
		{
			name: "literal webhook secret rejected",
			data: "webhooks:\n  - url: https://hooks.example/notify\n    events: [push]\n    secret: forbidden\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Decode([]byte(test.data)); err == nil {
				t.Fatal("Decode accepted invalid settings")
			}
		})
	}
}
func TestDecodeAcceptsRemainingAuditCandidateSettings(t *testing.T) {
	got, err := Decode([]byte(`autolinks:
  - key_prefix: JIRA-
    url_template: https://issues.example/<num>/
    is_alphanumeric: false
custom_properties:
  - name: team
    value: platform
  - name: services
    value: [api, web]
deploy_keys:
  - key: ssh-ed25519 YWJj ci
    title: automation
    read_only: true
interaction_limits:
  limit: collaborators_only
  expiry: one_week
webhooks:
  - url: https://hooks.example/notify
    events: [push]
    secret_env: WEBHOOK_SECRET
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Autolinks) != 1 || got.Autolinks[0].KeyPrefix != "JIRA-" ||
		got.Autolinks[0].IsAlphanumeric == nil || *got.Autolinks[0].IsAlphanumeric {
		t.Fatalf("autolink settings = %#v", got.Autolinks)
	}
	if len(got.CustomProperties) != 2 || got.CustomProperties[0].Value.APIValue() != "platform" ||
		!reflect.DeepEqual(got.CustomProperties[1].Value.APIValue(), []string{"api", "web"}) {
		t.Fatalf("custom property settings = %#v", got.CustomProperties)
	}
	if len(got.DeployKeys) != 1 || got.DeployKeys[0].Title != "automation" ||
		got.DeployKeys[0].ReadOnly == nil || !*got.DeployKeys[0].ReadOnly {
		t.Fatalf("deploy-key settings = %#v", got.DeployKeys)
	}
	if got.InteractionLimits == nil || got.InteractionLimits.Limit != "collaborators_only" ||
		got.InteractionLimits.Expiry != "one_week" {
		t.Fatalf("interaction limits = %#v", got.InteractionLimits)
	}
	if len(got.Webhooks) != 1 || got.Webhooks[0].URL != "https://hooks.example/notify" ||
		got.Webhooks[0].SecretEnv != "WEBHOOK_SECRET" {
		t.Fatalf("webhook settings = %#v", got.Webhooks)
	}
}
