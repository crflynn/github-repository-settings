package settings

// Ruleset declares one repository ruleset. Omitted rulesets are unchanged.
type Ruleset struct {
	Name         string            `json:"name" yaml:"name"`
	Target       string            `json:"target" yaml:"target"`
	Enforcement  string            `json:"enforcement" yaml:"enforcement"`
	State        string            `json:"-" yaml:"state,omitempty"`
	Conditions   RulesetConditions `json:"conditions" yaml:"conditions"`
	Rules        []Rule            `json:"rules" yaml:"rules"`
	BypassActors []BypassActor     `json:"bypass_actors,omitempty" yaml:"bypass_actors,omitempty"`
}

type RulesetConditions struct {
	RefName *RefNameCondition `json:"ref_name,omitempty" yaml:"ref_name,omitempty"`
}

type RefNameCondition struct {
	Include []string `json:"include" yaml:"include"`
	Exclude []string `json:"exclude" yaml:"exclude"`
}

type BypassActor struct {
	ActorID    *int64 `json:"actor_id,omitempty" yaml:"actor_id,omitempty"`
	ActorType  string `json:"actor_type" yaml:"actor_type"`
	BypassMode string `json:"bypass_mode" yaml:"bypass_mode"`
}

type Rule struct {
	Type       string          `json:"type" yaml:"type"`
	Parameters *RuleParameters `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

type RuleParameters struct {
	UpdateAllowsFetchAndMerge        *bool                 `json:"update_allows_fetch_and_merge,omitempty" yaml:"update_allows_fetch_and_merge,omitempty"`
	RequiredApprovingReviewCount     *int                  `json:"required_approving_review_count,omitempty" yaml:"required_approving_review_count,omitempty"`
	AllowedMergeMethods              []string              `json:"allowed_merge_methods,omitempty" yaml:"allowed_merge_methods,omitempty"`
	DismissStaleReviewsOnPush        *bool                 `json:"dismiss_stale_reviews_on_push,omitempty" yaml:"dismiss_stale_reviews_on_push,omitempty"`
	RequireCodeOwnerReview           *bool                 `json:"require_code_owner_review,omitempty" yaml:"require_code_owner_review,omitempty"`
	RequireLastPushApproval          *bool                 `json:"require_last_push_approval,omitempty" yaml:"require_last_push_approval,omitempty"`
	RequiredReviewThreadResolution   *bool                 `json:"required_review_thread_resolution,omitempty" yaml:"required_review_thread_resolution,omitempty"`
	RequiredStatusChecks             []RequiredStatusCheck `json:"required_status_checks,omitempty" yaml:"required_status_checks,omitempty"`
	StrictRequiredStatusChecksPolicy *bool                 `json:"strict_required_status_checks_policy,omitempty" yaml:"strict_required_status_checks_policy,omitempty"`
	DoNotEnforceOnCreate             *bool                 `json:"do_not_enforce_on_create,omitempty" yaml:"do_not_enforce_on_create,omitempty"`
	RequiredDeploymentEnvironments   []string              `json:"required_deployment_environments,omitempty" yaml:"required_deployment_environments,omitempty"`
	CheckResponseTimeoutMinutes      *int                  `json:"check_response_timeout_minutes,omitempty" yaml:"check_response_timeout_minutes,omitempty"`
	GroupingStrategy                 *string               `json:"grouping_strategy,omitempty" yaml:"grouping_strategy,omitempty"`
	MaxEntriesToBuild                *int                  `json:"max_entries_to_build,omitempty" yaml:"max_entries_to_build,omitempty"`
	MaxEntriesToMerge                *int                  `json:"max_entries_to_merge,omitempty" yaml:"max_entries_to_merge,omitempty"`
	MergeMethod                      *string               `json:"merge_method,omitempty" yaml:"merge_method,omitempty"`
	MinEntriesToMerge                *int                  `json:"min_entries_to_merge,omitempty" yaml:"min_entries_to_merge,omitempty"`
	MinEntriesToMergeWaitMinutes     *int                  `json:"min_entries_to_merge_wait_minutes,omitempty" yaml:"min_entries_to_merge_wait_minutes,omitempty"`
	Pattern                          *string               `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Name                             *string               `json:"name,omitempty" yaml:"name,omitempty"`
	Negate                           *bool                 `json:"negate,omitempty" yaml:"negate,omitempty"`
	Operator                         *string               `json:"operator,omitempty" yaml:"operator,omitempty"`
	DismissalRestriction             *DismissalRestriction `json:"dismissal_restriction,omitempty" yaml:"dismissal_restriction,omitempty"`
	RequiredReviewers                []RequiredReviewer    `json:"required_reviewers,omitempty" yaml:"required_reviewers,omitempty"`
	RestrictedFilePaths              []string              `json:"restricted_file_paths,omitempty" yaml:"restricted_file_paths,omitempty"`
	RestrictedFileExtensions         []string              `json:"restricted_file_extensions,omitempty" yaml:"restricted_file_extensions,omitempty"`
	MaxFilePathLength                *int                  `json:"max_file_path_length,omitempty" yaml:"max_file_path_length,omitempty"`
	MaxFileSize                      *int64                `json:"max_file_size,omitempty" yaml:"max_file_size,omitempty"`
	Workflows                        []RuleWorkflow        `json:"workflows,omitempty" yaml:"workflows,omitempty"`
	CodeScanningTools                []CodeScanningTool    `json:"code_scanning_tools,omitempty" yaml:"code_scanning_tools,omitempty"`
	MaxCoverageDrop                  *float64              `json:"max_coverage_drop,omitempty" yaml:"max_coverage_drop,omitempty"`
	MinimumCoverage                  *float64              `json:"minimum_coverage,omitempty" yaml:"minimum_coverage,omitempty"`
	ReviewOnPush                     *bool                 `json:"review_on_push,omitempty" yaml:"review_on_push,omitempty"`
	ReviewDraftPullRequests          *bool                 `json:"review_draft_pull_requests,omitempty" yaml:"review_draft_pull_requests,omitempty"`
}

type DismissalRestriction struct {
	Enabled       *bool                       `json:"enabled" yaml:"enabled"`
	AllowedActors []DismissalRestrictionActor `json:"allowed_actors,omitempty" yaml:"allowed_actors,omitempty"`
}

type DismissalRestrictionActor struct {
	ID   int64  `json:"id" yaml:"id"`
	Type string `json:"type" yaml:"type"`
}

type RequiredReviewer struct {
	MinimumApprovals *int             `json:"minimum_approvals,omitempty" yaml:"minimum_approvals,omitempty"`
	FilePatterns     []string         `json:"file_patterns,omitempty" yaml:"file_patterns,omitempty"`
	Reviewer         *RulesetReviewer `json:"reviewer,omitempty" yaml:"reviewer,omitempty"`
}

type RulesetReviewer struct {
	ID   *int64  `json:"id,omitempty" yaml:"id,omitempty"`
	Type *string `json:"type,omitempty" yaml:"type,omitempty"`
}

type RuleWorkflow struct {
	Path         string  `json:"path" yaml:"path"`
	Ref          *string `json:"ref,omitempty" yaml:"ref,omitempty"`
	RepositoryID *int64  `json:"repository_id,omitempty" yaml:"repository_id,omitempty"`
	SHA          *string `json:"sha,omitempty" yaml:"sha,omitempty"`
}

type CodeScanningTool struct {
	AlertsThreshold         string `json:"alerts_threshold" yaml:"alerts_threshold"`
	SecurityAlertsThreshold string `json:"security_alerts_threshold" yaml:"security_alerts_threshold"`
	Tool                    string `json:"tool" yaml:"tool"`
}

type RequiredStatusCheck struct {
	Context       string `json:"context" yaml:"context"`
	IntegrationID *int64 `json:"integration_id,omitempty" yaml:"integration_id,omitempty"`
}
