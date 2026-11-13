package settings

import "testing"

func TestDecodeAcceptsTypedRepositoryAndRulesetManifest(t *testing.T) {
	data := []byte(`repository:
  visibility: private
  has_issues: false
  topics: []
rulesets:
  - name: protect-main
    target: branch
    enforcement: active
    conditions:
      ref_name:
        include: ["~DEFAULT_BRANCH"]
        exclude: []
    rules:
      - type: pull_request
        parameters:
          required_approving_review_count: 2
          require_code_owner_review: true
`)
	settings, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if settings.Repository == nil || settings.Repository.HasIssues == nil || *settings.Repository.HasIssues {
		t.Fatalf("false repository setting was not preserved: %#v", settings.Repository)
	}
	if settings.Repository.Topics == nil || len(*settings.Repository.Topics) != 0 {
		t.Fatalf("explicit empty topics were not preserved: %#v", settings.Repository.Topics)
	}
	if len(settings.Rulesets) != 1 || settings.Rulesets[0].Rules[0].Parameters.RequiredApprovingReviewCount == nil || *settings.Rulesets[0].Rules[0].Parameters.RequiredApprovingReviewCount != 2 {
		t.Fatalf("ruleset was not decoded into typed fields: %#v", settings.Rulesets)
	}
}

func TestDecodeAcceptsCoreComponents(t *testing.T) {
	_, err := Decode([]byte(`actions:
  permissions:
    enabled: true
    allowed_actions: selected
    selected_actions:
      github_owned_allowed: true
      verified_allowed: false
      patterns_allowed: []
  workflow:
    default_workflow_permissions: read
    can_approve_pull_request_reviews: false
collaborators:
  - username: octocat
    permission: push
environments:
  - name: production
    reviewers:
      - type: Team
        id: 42
    deployment_branch_policy:
      protected_branches: true
      custom_branch_policies: false
rulesets:
  - name: protect-main
    target: branch
    enforcement: active
    conditions:
      ref_name:
        include: ["~DEFAULT_BRANCH"]
        exclude: []
    rules:
      - type: deletion
pages:
  enabled: true
  build_type: workflow
`))
	if err != nil {
		t.Fatalf("Decode rejected a valid core manifest: %v", err)
	}
}

func TestDecodeRejectsUnknownFields(t *testing.T) {
	_, err := Decode([]byte("repository:\n  has_issue: true\n"))
	if err == nil {
		t.Fatal("Decode accepted an unknown repository key")
	}
}

func TestDecodeRejectsInvalidComponentValues(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{name: "invalid collaborator permission", data: "collaborators:\n  - username: dev\n    permission: owner\n"},
		{name: "unknown ruleset rule", data: "rulesets:\n  - name: main\n    target: branch\n    enforcement: active\n    conditions:\n      ref_name:\n        include: [main]\n        exclude: []\n    rules:\n      - type: unknown_rule\n"},
		{name: "pages requires explicit enabled", data: "pages:\n  build_type: workflow\n"},
		{name: "unknown component", data: "unknown_component: []\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Decode([]byte(test.data)); err == nil {
				t.Fatal("Decode accepted invalid manifest")
			}
		})
	}
}

func TestDecodeAcceptsExpandedRulesetRules(t *testing.T) {
	_, err := Decode([]byte(`rulesets:
  - name: protect-main
    target: branch
    enforcement: active
    conditions:
      ref_name:
        include: [main]
        exclude: []
    bypass_actors:
      - actor_type: OrganizationAdmin
        bypass_mode: always
    rules:
      - type: pull_request
        parameters:
          required_approving_review_count: 1
          dismissal_restriction:
            enabled: true
            allowed_actors:
              - id: 11
                type: Team
          required_reviewers:
            - minimum_approvals: 1
              file_patterns: ["*.go"]
              reviewer:
                id: 11
                type: Team
      - type: workflows
        parameters:
          workflows:
            - path: .github/workflows/ci.yml
              repository_id: 42
              sha: abc123
      - type: code_scanning
        parameters:
          code_scanning_tools:
            - tool: CodeQL
              alerts_threshold: errors
              security_alerts_threshold: high_or_higher
      - type: code_coverage
        parameters:
          minimum_coverage: 80
      - type: copilot_code_review
        parameters:
          review_on_push: true
          review_draft_pull_requests: false
  - name: restrict-push
    target: push
    enforcement: active
    rules:
      - type: file_path_restriction
        parameters:
          restricted_file_paths: ["*.pem"]
      - type: file_extension_restriction
        parameters:
          restricted_file_extensions: [".exe"]
      - type: max_file_path_length
        parameters:
          max_file_path_length: 200
      - type: max_file_size
        parameters:
          max_file_size: 1000000
`))
	if err != nil {
		t.Fatalf("Decode rejected expanded ruleset settings: %v", err)
	}
}

func TestDecodeRejectsRulesetTargetMismatches(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "push rule on branch target",
			data: `rulesets:
  - name: branch
    target: branch
    enforcement: active
    conditions:
      ref_name: {include: [main], exclude: []}
    rules:
      - type: max_file_size
        parameters: {max_file_size: 100}
`,
		},
		{
			name: "branch rule on push target",
			data: `rulesets:
  - name: push
    target: push
    enforcement: active
    rules:
      - type: creation
`,
		},
		{
			name: "invalid push file size",
			data: `rulesets:
  - name: push
    target: push
    enforcement: active
    rules:
      - type: max_file_size
        parameters: {max_file_size: 0}
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Decode([]byte(test.data)); err == nil {
				t.Fatal("Decode accepted invalid ruleset")
			}
		})
	}
}

func TestDecodeRejectsMultipleYAMLDocuments(t *testing.T) {
	_, err := Decode([]byte("repository: {}\n---\npages:\n  enabled: false\n"))
	if err == nil {
		t.Fatal("Decode accepted multiple YAML documents")
	}
}
func TestDecodeRejectsInvalidDeployKeyDeclarations(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{name: "private key", data: "deploy_keys:\n  - key: '-----BEGIN OPENSSH PRIVATE KEY-----'\n"},
		{name: "missing title", data: "deploy_keys:\n  - key: 'ssh-ed25519 YWJj ci'\n    read_only: true\n"},
		{name: "missing read_only", data: "deploy_keys:\n  - key: 'ssh-ed25519 YWJj ci'\n    title: ci\n"},
		{name: "invalid state", data: "deploy_keys:\n  - key: 'ssh-ed25519 YWJj ci'\n    state: deleted\n"},
		{name: "absent with settings", data: "deploy_keys:\n  - key: 'ssh-ed25519 YWJj ci'\n    state: absent\n    title: ci\n"},
		{name: "duplicate key ignoring comment", data: "deploy_keys:\n  - key: 'ssh-ed25519 YWJj first'\n    title: first\n    read_only: true\n  - key: 'ssh-ed25519 YWJj second'\n    title: second\n    read_only: true\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Decode([]byte(test.data)); err == nil {
				t.Fatal("Decode accepted invalid deploy-key declaration")
			}
		})
	}
}
