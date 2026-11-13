package settings

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// Settings is the complete schema for a repository-settings manifest.
type Settings struct {
	Repository        *Repository        `yaml:"repository,omitempty"`
	Actions           *Actions           `yaml:"actions,omitempty"`
	Collaborators     []Collaborator     `yaml:"collaborators,omitempty"`
	Teams             []TeamRepository   `yaml:"teams,omitempty"`
	Environments      []Environment      `yaml:"environments,omitempty"`
	Rulesets          []Ruleset          `yaml:"rulesets,omitempty"`
	Pages             *Pages             `yaml:"pages,omitempty"`
	InteractionLimits *InteractionLimits `yaml:"interaction_limits,omitempty"`
	Autolinks         []Autolink         `yaml:"autolinks,omitempty"`
	DeployKeys        []DeployKey        `yaml:"deploy_keys,omitempty"`
	CustomProperties  CustomProperties   `yaml:"custom_properties,omitempty"`
	Webhooks          []Webhook          `yaml:"webhooks,omitempty"`
}

// Decode parses one manifest and rejects fields not represented by this schema.
func Decode(data []byte) (Settings, error) {
	var settings Settings
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&settings); err != nil {
		return Settings{}, fmt.Errorf("parse settings manifest: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Settings{}, fmt.Errorf("settings manifest must contain one YAML document")
		}
		return Settings{}, fmt.Errorf("parse trailing YAML document: %w", err)
	}
	if err := settings.Validate(); err != nil {
		return Settings{}, err
	}
	return settings, nil
}

// Validate checks cross-field and enum constraints before any API request is sent.
func (s Settings) Validate() error {
	if s.Repository == nil && s.Actions == nil && s.Collaborators == nil && len(s.Teams) == 0 &&
		s.Environments == nil && s.Rulesets == nil && s.Pages == nil &&
		s.InteractionLimits == nil && len(s.Autolinks) == 0 && len(s.DeployKeys) == 0 &&
		len(s.CustomProperties) == 0 && len(s.Webhooks) == 0 {
		return fmt.Errorf("settings manifest must configure at least one component")
	}
	if err := ValidateAutolinks(s.Autolinks); err != nil {
		return err
	}
	if err := s.CustomProperties.Validate(); err != nil {
		return fmt.Errorf("custom_properties: %w", err)
	}
	if err := ValidateWebhooks(s.Webhooks); err != nil {
		return err
	}
	if err := ValidateDeployKeys(s.DeployKeys); err != nil {
		return fmt.Errorf("deploy_keys: %w", err)
	}
	if s.InteractionLimits != nil {
		if err := s.InteractionLimits.Validate(); err != nil {
			return fmt.Errorf("interaction_limits: %w", err)
		}
	}
	if err := s.validateRepository(); err != nil {
		return err
	}
	if err := s.validateActions(); err != nil {
		return err
	}
	if err := s.validateCollaborators(); err != nil {
		return err
	}
	if err := s.validateTeams(); err != nil {
		return err
	}
	if err := s.validateEnvironments(); err != nil {
		return err
	}
	if err := s.validateRulesets(); err != nil {
		return err
	}
	if err := s.validatePages(); err != nil {
		return err
	}
	return nil
}

func (s Settings) validateRepository() error {
	if s.Repository == nil {
		return nil
	}
	repository := s.Repository
	if repository.Description == nil && repository.Homepage == nil && repository.Visibility == nil &&
		repository.DefaultBranch == nil && repository.IsTemplate == nil && repository.HasPullRequests == nil &&
		repository.PullRequestCreationPolicy == nil && repository.DependabotSecurityUpdates == nil &&
		repository.VulnerabilityAlerts == nil && repository.CodeSecurity == nil &&
		repository.SecretScanning == nil && repository.SecretScanningPushProtection == nil &&
		repository.PrivateVulnerabilityReporting == nil &&
		repository.HasIssues == nil && repository.HasProjects == nil && repository.HasWiki == nil &&
		repository.HasDiscussions == nil && repository.HasDownloads == nil &&
		repository.AllowSquashMerge == nil && repository.AllowMergeCommit == nil &&
		repository.AllowRebaseMerge == nil && repository.AllowAutoMerge == nil &&
		repository.AllowUpdateBranch == nil && repository.AllowForking == nil &&
		repository.DeleteBranchOnMerge == nil && repository.WebCommitSignoffRequired == nil &&
		repository.SquashMergeCommitTitle == nil && repository.SquashMergeCommitMessage == nil &&
		repository.MergeCommitTitle == nil && repository.MergeCommitMessage == nil &&
		repository.Archived == nil && repository.Topics == nil {
		return fmt.Errorf("repository must configure at least one field")
	}
	if repository.Visibility != nil {
		switch *repository.Visibility {
		case "public", "private", "internal":
		default:
			return fmt.Errorf("repository.visibility must be public, private, or internal")
		}
	}
	if repository.PullRequestCreationPolicy != nil &&
		*repository.PullRequestCreationPolicy != "all" &&
		*repository.PullRequestCreationPolicy != "collaborators_only" {
		return fmt.Errorf("repository.pull_request_creation_policy must be all or collaborators_only")
	}
	return nil
}

func (s Settings) validateActions() error {
	if s.Actions == nil {
		return nil
	}
	if permissions := s.Actions.Permissions; permissions != nil {
		if permissions.Enabled == nil && permissions.AllowedActions == nil &&
			permissions.SelectedActions == nil && permissions.SHAPinningRequired == nil {
			return fmt.Errorf("actions.permissions must configure at least one field")
		}
		if permissions.AllowedActions != nil {
			switch *permissions.AllowedActions {
			case "all", "local_only", "selected":
			default:
				return fmt.Errorf("actions.permissions.allowed_actions must be all, local_only, or selected")
			}
			if *permissions.AllowedActions == "selected" && permissions.SelectedActions == nil {
				return fmt.Errorf("actions.permissions.selected_actions is required when allowed_actions is selected")
			}
		}
		if permissions.SelectedActions != nil && (permissions.AllowedActions == nil || *permissions.AllowedActions != "selected") {
			return fmt.Errorf("actions.permissions.selected_actions requires allowed_actions: selected")
		}
		if selected := permissions.SelectedActions; selected != nil && selected.GitHubOwnedAllowed == nil && selected.VerifiedAllowed == nil && selected.PatternsAllowed == nil {
			return fmt.Errorf("actions.permissions.selected_actions must set at least one policy")
		}
	}
	if workflow := s.Actions.Workflow; workflow != nil {
		if workflow.DefaultWorkflowPermissions == nil && workflow.CanApprovePullRequestReviews == nil {
			return fmt.Errorf("actions.workflow must configure at least one field")
		}
		if workflow.DefaultWorkflowPermissions != nil {
			switch *workflow.DefaultWorkflowPermissions {
			case "read", "write":
			default:
				return fmt.Errorf("actions.workflow.default_workflow_permissions must be read or write")
			}
		}
	}
	if fork := s.Actions.ForkPR; fork != nil &&
		fork.RunWorkflows == nil && fork.SendWriteTokens == nil &&
		fork.SendSecretsAndVariables == nil && fork.RequireApproval == nil {
		return fmt.Errorf("actions.fork_pr must configure at least one field")
	}
	if retention := s.Actions.Retention; retention != nil {
		if retention.Days == nil {
			return fmt.Errorf("actions.retention.days is required")
		}
		if *retention.Days < 1 || *retention.Days > 400 {
			return fmt.Errorf("actions.retention.days must be between 1 and 400")
		}
	}
	if s.Actions.Permissions == nil && s.Actions.Workflow == nil && s.Actions.ForkPR == nil && s.Actions.Retention == nil {
		return fmt.Errorf("actions must configure permissions, workflow, fork_pr, or retention")
	}
	return nil
}

func (s Settings) validateTeams() error {
	seen := make(map[string]struct{}, len(s.Teams))
	for i, team := range s.Teams {
		slug := strings.ToLower(strings.TrimSpace(team.Slug))
		if slug == "" {
			return fmt.Errorf("teams[%d].slug is required", i)
		}
		if _, exists := seen[slug]; exists {
			return fmt.Errorf("teams contains duplicate slug %q", team.Slug)
		}
		seen[slug] = struct{}{}
		switch team.Permission {
		case "pull", "triage", "push", "maintain", "admin":
		default:
			return fmt.Errorf("teams[%d].permission must be pull, triage, push, maintain, or admin", i)
		}
	}
	return nil
}

func (s Settings) validateCollaborators() error {
	seen := make(map[string]struct{}, len(s.Collaborators))
	for i, collaborator := range s.Collaborators {
		if strings.TrimSpace(collaborator.Username) == "" {
			return fmt.Errorf("collaborators[%d].username is required", i)
		}
		if _, exists := seen[strings.ToLower(collaborator.Username)]; exists {
			return fmt.Errorf("collaborators contains duplicate username %q", collaborator.Username)
		}
		seen[strings.ToLower(collaborator.Username)] = struct{}{}
		state := collaborator.State
		if state == "" {
			state = "present"
		}
		switch state {
		case "present":
			switch collaborator.Permission {
			case "pull", "triage", "push", "maintain", "admin":
			default:
				return fmt.Errorf("collaborators[%d].permission must be pull, triage, push, maintain, or admin", i)
			}
		case "absent":
			if collaborator.Permission != "" {
				return fmt.Errorf("collaborators[%d].permission is not valid when state is absent", i)
			}
		default:
			return fmt.Errorf("collaborators[%d].state must be present or absent", i)
		}
	}
	return nil
}

func (s Settings) validateEnvironments() error {
	seen := make(map[string]struct{}, len(s.Environments))
	for i, environment := range s.Environments {
		if strings.TrimSpace(environment.Name) == "" {
			return fmt.Errorf("environments[%d].name is required", i)
		}
		if _, exists := seen[strings.ToLower(environment.Name)]; exists {
			return fmt.Errorf("environments contains duplicate name %q", environment.Name)
		}
		seen[strings.ToLower(environment.Name)] = struct{}{}
		state := environment.State
		if state == "" {
			state = "present"
		}
		switch state {
		case "present":
		case "absent":
			if environment.WaitTimer != nil || environment.PreventSelfReview != nil || environment.CanAdminsBypass != nil || environment.Reviewers != nil || environment.DeploymentBranchPolicy != nil {
				return fmt.Errorf("environments[%d] cannot include settings when state is absent", i)
			}
			continue
		default:
			return fmt.Errorf("environments[%d].state must be present or absent", i)
		}
		if environment.WaitTimer != nil && (*environment.WaitTimer < 0 || *environment.WaitTimer > 43200) {
			return fmt.Errorf("environments[%d].wait_timer must be between 0 and 43200", i)
		}
		if environment.Reviewers != nil {
			for j, reviewer := range *environment.Reviewers {
				if (reviewer.Type != "User" && reviewer.Type != "Team") || reviewer.ID < 1 {
					return fmt.Errorf("environments[%d].reviewers[%d] requires type User or Team and a positive ID", i, j)
				}
			}
		}
	}
	return nil
}

func (s Settings) validateRulesets() error {
	seen := make(map[string]struct{}, len(s.Rulesets))
	for i, ruleset := range s.Rulesets {
		if strings.TrimSpace(ruleset.Name) == "" {
			return fmt.Errorf("rulesets[%d].name is required", i)
		}
		key := strings.ToLower(ruleset.Name)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("rulesets contains duplicate name %q", ruleset.Name)
		}
		seen[key] = struct{}{}

		state := ruleset.State
		if state == "" {
			state = "present"
		}
		if state == "absent" {
			if ruleset.Target != "" || ruleset.Enforcement != "" || ruleset.Conditions.RefName != nil || len(ruleset.Rules) != 0 || len(ruleset.BypassActors) != 0 {
				return fmt.Errorf("rulesets[%d] cannot include settings when state is absent", i)
			}
			continue
		}
		if state != "present" {
			return fmt.Errorf("rulesets[%d].state must be present or absent", i)
		}
		if ruleset.Target != "branch" && ruleset.Target != "tag" && ruleset.Target != "push" {
			return fmt.Errorf("rulesets[%d].target must be branch, tag, or push", i)
		}
		if ruleset.Enforcement != "active" && ruleset.Enforcement != "disabled" && ruleset.Enforcement != "evaluate" {
			return fmt.Errorf("rulesets[%d].enforcement must be active, disabled, or evaluate", i)
		}
		if ruleset.Target != "push" && ruleset.Conditions.RefName == nil {
			return fmt.Errorf("rulesets[%d].conditions.ref_name is required for branch and tag targets", i)
		}
		if ruleset.Target == "push" && ruleset.Conditions.RefName != nil {
			return fmt.Errorf("rulesets[%d].conditions.ref_name is not valid for push targets", i)
		}
		if ruleset.Target != "push" && (len(ruleset.Conditions.RefName.Include) == 0 || ruleset.Conditions.RefName.Exclude == nil) {
			return fmt.Errorf("rulesets[%d].conditions.ref_name requires include and exclude lists, with at least one include", i)
		}
		if len(ruleset.Rules) == 0 {
			return fmt.Errorf("rulesets[%d].rules must contain at least one rule", i)
		}
		for j, actor := range ruleset.BypassActors {
			if (actor.ActorID != nil && *actor.ActorID <= 0) || (actor.ActorType != "Integration" && actor.ActorType != "OrganizationAdmin" && actor.ActorType != "RepositoryRole" && actor.ActorType != "Team" && actor.ActorType != "DeployKey") {
				return fmt.Errorf("rulesets[%d].bypass_actors[%d] has an invalid actor", i, j)
			}
			if actor.BypassMode != "always" && actor.BypassMode != "exempt" && actor.BypassMode != "pull_request" && actor.BypassMode != "never" {
				return fmt.Errorf("rulesets[%d].bypass_actors[%d].bypass_mode is invalid", i, j)
			}
		}
		for j, rule := range ruleset.Rules {
			if err := validateRule(ruleset.Target, rule); err != nil {
				return fmt.Errorf("rulesets[%d].rules[%d]: %w", i, j, err)
			}
		}
	}
	return nil
}

var pushRuleTypes = map[string]struct{}{
	"file_path_restriction":      {},
	"file_extension_restriction": {},
	"max_file_path_length":       {},
	"max_file_size":              {},
}

var parameterlessRuleTypes = map[string]struct{}{
	"creation":                {},
	"deletion":                {},
	"required_linear_history": {},
	"required_signatures":     {},
	"non_fast_forward":        {},
}

var allowedRuleParameters = map[string]map[string]struct{}{
	"update": {"update_allows_fetch_and_merge": {}},
	"pull_request": {
		"required_approving_review_count": {}, "allowed_merge_methods": {},
		"dismiss_stale_reviews_on_push": {}, "require_code_owner_review": {},
		"require_last_push_approval": {}, "required_review_thread_resolution": {},
		"dismissal_restriction": {}, "required_reviewers": {},
	},
	"required_status_checks": {
		"required_status_checks": {}, "strict_required_status_checks_policy": {},
		"do_not_enforce_on_create": {},
	},
	"required_deployments": {"required_deployment_environments": {}},
	"merge_queue": {
		"check_response_timeout_minutes": {}, "grouping_strategy": {},
		"max_entries_to_build": {}, "max_entries_to_merge": {}, "merge_method": {},
		"min_entries_to_merge": {}, "min_entries_to_merge_wait_minutes": {},
	},
	"commit_message_pattern": {
		"name": {}, "pattern": {}, "negate": {}, "operator": {},
	},
	"commit_author_email_pattern": {
		"name": {}, "pattern": {}, "negate": {}, "operator": {},
	},
	"committer_email_pattern": {
		"name": {}, "pattern": {}, "negate": {}, "operator": {},
	},
	"branch_name_pattern": {
		"name": {}, "pattern": {}, "negate": {}, "operator": {},
	},
	"tag_name_pattern": {
		"name": {}, "pattern": {}, "negate": {}, "operator": {},
	},
	"workflows":                  {"do_not_enforce_on_create": {}, "workflows": {}},
	"code_scanning":              {"code_scanning_tools": {}},
	"code_coverage":              {"max_coverage_drop": {}, "minimum_coverage": {}},
	"copilot_code_review":        {"review_on_push": {}, "review_draft_pull_requests": {}},
	"file_path_restriction":      {"restricted_file_paths": {}},
	"file_extension_restriction": {"restricted_file_extensions": {}},
	"max_file_path_length":       {"max_file_path_length": {}},
	"max_file_size":              {"max_file_size": {}},
}

func validateRule(target string, rule Rule) error {
	_, isPushRule := pushRuleTypes[rule.Type]
	if isPushRule != (target == "push") {
		targetType := "branch and tag"
		if isPushRule {
			targetType = "push"
		}
		return fmt.Errorf("rule %q is only valid for %s targets", rule.Type, targetType)
	}
	if _, parameterless := parameterlessRuleTypes[rule.Type]; parameterless {
		if rule.Parameters != nil {
			return fmt.Errorf("rule %q does not accept parameters", rule.Type)
		}
		return nil
	}
	if rule.Parameters == nil {
		return fmt.Errorf("rule %q requires parameters", rule.Type)
	}
	allowedFields, supported := allowedRuleParameters[rule.Type]
	if !supported {
		return fmt.Errorf("unsupported rule type %q", rule.Type)
	}
	parameters := reflect.ValueOf(rule.Parameters).Elem()
	parameterType := parameters.Type()
	for i := range parameters.NumField() {
		fieldName, _, _ := strings.Cut(parameterType.Field(i).Tag.Get("yaml"), ",")
		if fieldName == "-" || parameters.Field(i).IsZero() {
			continue
		}
		if _, allowed := allowedFields[fieldName]; !allowed {
			return fmt.Errorf("parameter %q is not valid for rule %q", fieldName, rule.Type)
		}
	}
	switch rule.Type {
	case "update":
		if rule.Parameters.UpdateAllowsFetchAndMerge == nil {
			return fmt.Errorf("update requires update_allows_fetch_and_merge")
		}
	case "pull_request":
		if rule.Parameters.RequiredApprovingReviewCount == nil || *rule.Parameters.RequiredApprovingReviewCount < 0 {
			return fmt.Errorf("pull_request requires a non-negative required_approving_review_count")
		}
		for _, method := range rule.Parameters.AllowedMergeMethods {
			if method != "merge" && method != "rebase" && method != "squash" {
				return fmt.Errorf("pull_request allowed_merge_methods contains unsupported method %q", method)
			}
		}
		if restriction := rule.Parameters.DismissalRestriction; restriction != nil {
			if restriction.Enabled == nil {
				return fmt.Errorf("pull_request.dismissal_restriction requires enabled")
			}
			for i, actor := range restriction.AllowedActors {
				if actor.ID <= 0 || !oneOf(actor.Type, "User", "Team", "IntegrationInstallation", "RepositoryRole") {
					return fmt.Errorf("pull_request.dismissal_restriction.allowed_actors[%d] is invalid", i)
				}
			}
		}
		for i, reviewer := range rule.Parameters.RequiredReviewers {
			if reviewer.MinimumApprovals != nil && *reviewer.MinimumApprovals < 1 {
				return fmt.Errorf("pull_request.required_reviewers[%d].minimum_approvals must be positive", i)
			}
			if reviewer.Reviewer != nil {
				if reviewer.Reviewer.ID != nil && *reviewer.Reviewer.ID <= 0 {
					return fmt.Errorf("pull_request.required_reviewers[%d].reviewer.id must be positive", i)
				}
				if reviewer.Reviewer.Type != nil && *reviewer.Reviewer.Type != "Team" {
					return fmt.Errorf("pull_request.required_reviewers[%d].reviewer.type must be Team", i)
				}
			}
		}
	case "required_status_checks":
		if len(rule.Parameters.RequiredStatusChecks) == 0 {
			return fmt.Errorf("required_status_checks requires at least one required_status_check")
		}
		for i, check := range rule.Parameters.RequiredStatusChecks {
			if strings.TrimSpace(check.Context) == "" {
				return fmt.Errorf("required_status_checks.required_status_checks[%d].context is required", i)
			}
		}
	case "required_deployments":
		if len(rule.Parameters.RequiredDeploymentEnvironments) == 0 {
			return fmt.Errorf("required_deployments requires at least one environment")
		}
	case "merge_queue":
		parameters := rule.Parameters
		if parameters.CheckResponseTimeoutMinutes == nil || parameters.GroupingStrategy == nil || parameters.MaxEntriesToBuild == nil || parameters.MaxEntriesToMerge == nil || parameters.MergeMethod == nil || parameters.MinEntriesToMerge == nil || parameters.MinEntriesToMergeWaitMinutes == nil {
			return fmt.Errorf("merge_queue requires all queue parameters")
		}
		if (*parameters.GroupingStrategy != "ALLGREEN" && *parameters.GroupingStrategy != "HEADGREEN") || (*parameters.MergeMethod != "MERGE" && *parameters.MergeMethod != "REBASE" && *parameters.MergeMethod != "SQUASH") {
			return fmt.Errorf("merge_queue grouping_strategy or merge_method is invalid")
		}
	case "commit_message_pattern", "commit_author_email_pattern", "committer_email_pattern", "branch_name_pattern", "tag_name_pattern":
		if rule.Parameters.Name == nil || strings.TrimSpace(*rule.Parameters.Name) == "" || rule.Parameters.Pattern == nil || rule.Parameters.Operator == nil {
			return fmt.Errorf("%s requires non-empty name, pattern, and operator", rule.Type)
		}
		switch *rule.Parameters.Operator {
		case "starts_with", "ends_with", "contains", "regex":
		default:
			return fmt.Errorf("%s operator is invalid", rule.Type)
		}
	case "workflows":
		if len(rule.Parameters.Workflows) == 0 {
			return fmt.Errorf("workflows requires at least one workflow")
		}
		for i, workflow := range rule.Parameters.Workflows {
			if strings.TrimSpace(workflow.Path) == "" || (workflow.SHA == nil && workflow.Ref == nil) {
				return fmt.Errorf("workflows.workflows[%d] requires path and either sha or ref", i)
			}
		}
	case "code_scanning":
		if len(rule.Parameters.CodeScanningTools) == 0 {
			return fmt.Errorf("code_scanning requires at least one tool")
		}
		for i, tool := range rule.Parameters.CodeScanningTools {
			if strings.TrimSpace(tool.Tool) == "" || !oneOf(tool.AlertsThreshold, "none", "errors", "errors_and_warnings", "all") ||
				!oneOf(tool.SecurityAlertsThreshold, "none", "critical", "high_or_higher", "medium_or_higher", "all") {
				return fmt.Errorf("code_scanning.code_scanning_tools[%d] is invalid", i)
			}
		}
	case "code_coverage":
		for _, coverage := range []*float64{rule.Parameters.MaxCoverageDrop, rule.Parameters.MinimumCoverage} {
			if coverage != nil && (*coverage < 0 || *coverage > 100) {
				return fmt.Errorf("code_coverage values must be between 0 and 100")
			}
		}
		if rule.Parameters.MaxCoverageDrop == nil && rule.Parameters.MinimumCoverage == nil {
			return fmt.Errorf("code_coverage requires max_coverage_drop or minimum_coverage")
		}
	case "copilot_code_review":
		if rule.Parameters.ReviewOnPush == nil && rule.Parameters.ReviewDraftPullRequests == nil {
			return fmt.Errorf("copilot_code_review requires review_on_push or review_draft_pull_requests")
		}
	case "file_path_restriction":
		if len(rule.Parameters.RestrictedFilePaths) == 0 {
			return fmt.Errorf("file_path_restriction requires at least one restricted_file_path")
		}
	case "file_extension_restriction":
		if len(rule.Parameters.RestrictedFileExtensions) == 0 {
			return fmt.Errorf("file_extension_restriction requires at least one restricted_file_extension")
		}
	case "max_file_path_length":
		if rule.Parameters.MaxFilePathLength == nil || *rule.Parameters.MaxFilePathLength <= 0 {
			return fmt.Errorf("max_file_path_length requires a positive limit")
		}
	case "max_file_size":
		if rule.Parameters.MaxFileSize == nil || *rule.Parameters.MaxFileSize <= 0 {
			return fmt.Errorf("max_file_size requires a positive limit")
		}
	}
	return nil
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func (s Settings) validatePages() error {
	if s.Pages == nil {
		return nil
	}
	if s.Pages.Enabled == nil {
		return fmt.Errorf("pages.enabled is required")
	}
	if !*s.Pages.Enabled {
		if s.Pages.BuildType != "" || s.Pages.Source != nil || s.Pages.CNAME != nil || s.Pages.Public != nil || s.Pages.HTTPSEnforced != nil {
			return fmt.Errorf("pages settings cannot be set when pages.enabled is false")
		}
		return nil
	}
	if s.Pages.BuildType != "legacy" && s.Pages.BuildType != "workflow" {
		return fmt.Errorf("pages.build_type must be legacy or workflow when Pages is enabled")
	}
	if s.Pages.BuildType == "legacy" && s.Pages.Source == nil {
		return fmt.Errorf("pages.source is required for legacy Pages builds")
	}
	if s.Pages.Source != nil && (s.Pages.Source.Branch == "" || (s.Pages.Source.Path != "/" && s.Pages.Source.Path != "/docs")) {
		return fmt.Errorf("pages.source requires a branch and path / or /docs")
	}
	return nil
}
