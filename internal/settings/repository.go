package settings

// Repository contains mutable fields accepted by GitHub's repository API.
type Repository struct {
	Description                   *string   `json:"description,omitempty" yaml:"description,omitempty"`
	Homepage                      *string   `json:"homepage,omitempty" yaml:"homepage,omitempty"`
	Visibility                    *string   `json:"visibility,omitempty" yaml:"visibility,omitempty"`
	DefaultBranch                 *string   `json:"default_branch,omitempty" yaml:"default_branch,omitempty"`
	IsTemplate                    *bool     `json:"is_template,omitempty" yaml:"is_template,omitempty"`
	HasPullRequests               *bool     `json:"has_pull_requests,omitempty" yaml:"has_pull_requests,omitempty"`
	PullRequestCreationPolicy     *string   `json:"pull_request_creation_policy,omitempty" yaml:"pull_request_creation_policy,omitempty"`
	HasIssues                     *bool     `json:"has_issues,omitempty" yaml:"has_issues,omitempty"`
	HasProjects                   *bool     `json:"has_projects,omitempty" yaml:"has_projects,omitempty"`
	HasWiki                       *bool     `json:"has_wiki,omitempty" yaml:"has_wiki,omitempty"`
	HasDiscussions                *bool     `json:"has_discussions,omitempty" yaml:"has_discussions,omitempty"`
	HasDownloads                  *bool     `json:"has_downloads,omitempty" yaml:"has_downloads,omitempty"`
	AllowSquashMerge              *bool     `json:"allow_squash_merge,omitempty" yaml:"allow_squash_merge,omitempty"`
	AllowMergeCommit              *bool     `json:"allow_merge_commit,omitempty" yaml:"allow_merge_commit,omitempty"`
	AllowRebaseMerge              *bool     `json:"allow_rebase_merge,omitempty" yaml:"allow_rebase_merge,omitempty"`
	AllowAutoMerge                *bool     `json:"allow_auto_merge,omitempty" yaml:"allow_auto_merge,omitempty"`
	AllowUpdateBranch             *bool     `json:"allow_update_branch,omitempty" yaml:"allow_update_branch,omitempty"`
	AllowForking                  *bool     `json:"allow_forking,omitempty" yaml:"allow_forking,omitempty"`
	DeleteBranchOnMerge           *bool     `json:"delete_branch_on_merge,omitempty" yaml:"delete_branch_on_merge,omitempty"`
	WebCommitSignoffRequired      *bool     `json:"web_commit_signoff_required,omitempty" yaml:"web_commit_signoff_required,omitempty"`
	SquashMergeCommitTitle        *string   `json:"squash_merge_commit_title,omitempty" yaml:"squash_merge_commit_title,omitempty"`
	SquashMergeCommitMessage      *string   `json:"squash_merge_commit_message,omitempty" yaml:"squash_merge_commit_message,omitempty"`
	MergeCommitTitle              *string   `json:"merge_commit_title,omitempty" yaml:"merge_commit_title,omitempty"`
	MergeCommitMessage            *string   `json:"merge_commit_message,omitempty" yaml:"merge_commit_message,omitempty"`
	Archived                      *bool     `json:"archived,omitempty" yaml:"archived,omitempty"`
	Topics                        *[]string `json:"-" yaml:"topics,omitempty"`
	DependabotSecurityUpdates     *bool     `json:"dependabot_security_updates,omitempty" yaml:"dependabot_security_updates,omitempty"`
	VulnerabilityAlerts           *bool     `json:"vulnerability_alerts,omitempty" yaml:"vulnerability_alerts,omitempty"`
	CodeSecurity                  *bool     `json:"code_security,omitempty" yaml:"code_security,omitempty"`
	SecretScanning                *bool     `json:"secret_scanning,omitempty" yaml:"secret_scanning,omitempty"`
	SecretScanningPushProtection  *bool     `json:"secret_scanning_push_protection,omitempty" yaml:"secret_scanning_push_protection,omitempty"`
	PrivateVulnerabilityReporting *bool     `json:"private_vulnerability_reporting,omitempty" yaml:"private_vulnerability_reporting,omitempty"`
}
