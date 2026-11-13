package settings

// Actions controls repository-level GitHub Actions permissions.
type Actions struct {
	Permissions *ActionPermissions   `json:"-" yaml:"permissions,omitempty"`
	Workflow    *WorkflowPermissions `json:"-" yaml:"workflow,omitempty"`
	ForkPR      *ForkPRSettings      `json:"-" yaml:"fork_pr,omitempty"`
	Retention   *ArtifactRetention   `json:"-" yaml:"retention,omitempty"`
}

type ForkPRSettings struct {
	RunWorkflows            *bool `json:"run_workflows_from_fork_pull_requests,omitempty" yaml:"run_workflows_from_fork_pull_requests,omitempty"`
	SendWriteTokens         *bool `json:"send_write_tokens_to_workflows,omitempty" yaml:"send_write_tokens_to_workflows,omitempty"`
	SendSecretsAndVariables *bool `json:"send_secrets_and_variables,omitempty" yaml:"send_secrets_and_variables,omitempty"`
	RequireApproval         *bool `json:"require_approval_for_fork_pr_workflows,omitempty" yaml:"require_approval_for_fork_pr_workflows,omitempty"`
}

type ArtifactRetention struct {
	Days *int `json:"-" yaml:"days,omitempty"`
}

type ActionPermissions struct {
	Enabled            *bool            `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	AllowedActions     *string          `json:"allowed_actions,omitempty" yaml:"allowed_actions,omitempty"`
	SHAPinningRequired *bool            `json:"sha_pinning_required,omitempty" yaml:"sha_pinning_required,omitempty"`
	SelectedActions    *SelectedActions `json:"-" yaml:"selected_actions,omitempty"`
}

type SelectedActions struct {
	GitHubOwnedAllowed *bool     `json:"github_owned_allowed,omitempty" yaml:"github_owned_allowed,omitempty"`
	VerifiedAllowed    *bool     `json:"verified_allowed,omitempty" yaml:"verified_allowed,omitempty"`
	PatternsAllowed    *[]string `json:"patterns_allowed,omitempty" yaml:"patterns_allowed,omitempty"`
}

type WorkflowPermissions struct {
	DefaultWorkflowPermissions   *string `json:"default_workflow_permissions,omitempty" yaml:"default_workflow_permissions,omitempty"`
	CanApprovePullRequestReviews *bool   `json:"can_approve_pull_request_reviews,omitempty" yaml:"can_approve_pull_request_reviews,omitempty"`
}
