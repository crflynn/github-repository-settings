package github

import (
	"context"
	"fmt"
	"strings"

	gh "github.com/google/go-github/v93/github"

	"github-repository-settings/internal/settings"
)

func (c *Client) applyRulesets(ctx context.Context, desired []settings.Ruleset) error {
	existing, err := c.listRulesets(ctx)
	if err != nil {
		return err
	}
	byName := make(map[string]*gh.RepositoryRuleset, len(existing))
	for _, ruleset := range existing {
		byName[strings.ToLower(ruleset.Name)] = ruleset
	}
	for _, ruleset := range desired {
		current := byName[strings.ToLower(ruleset.Name)]
		if ruleset.State == "absent" {
			if current == nil || current.ID == nil {
				continue
			}
			response, err := c.api.Repositories.DeleteRuleset(ctx, c.owner, c.repo, *current.ID)
			if err != nil && !isNotFound(response) {
				return fmt.Errorf("delete %q: %w", ruleset.Name, err)
			}
			continue
		}
		request := toGitHubRuleset(ruleset, c.owner+"/"+c.repo)
		if current == nil || current.ID == nil {
			if _, _, err := c.api.Repositories.CreateRuleset(ctx, c.owner, c.repo, request); err != nil {
				return fmt.Errorf("create %q: %w", ruleset.Name, err)
			}
			continue
		}
		if _, _, err := c.api.Repositories.UpdateRuleset(ctx, c.owner, c.repo, *current.ID, request); err != nil {
			return fmt.Errorf("update %q: %w", ruleset.Name, err)
		}
	}
	return nil
}

func (c *Client) listRulesets(ctx context.Context) ([]*gh.RepositoryRuleset, error) {
	options := &gh.RepositoryListRulesetsOptions{
		IncludesParents: pointer(false),
		ListOptions:     gh.ListOptions{PerPage: 100},
	}
	var all []*gh.RepositoryRuleset
	for {
		rulesets, response, err := c.api.Repositories.GetAllRulesets(ctx, c.owner, c.repo, options)
		if err != nil {
			return nil, err
		}
		all = append(all, rulesets...)
		if response == nil || response.NextPage == 0 {
			return all, nil
		}
		options.Page = response.NextPage
	}
}

func toGitHubRuleset(config settings.Ruleset, source string) gh.RepositoryRuleset {
	target := gh.RulesetTarget(config.Target)
	var conditions *gh.RepositoryRulesetConditions
	if condition := config.Conditions.RefName; condition != nil {
		conditions = &gh.RepositoryRulesetConditions{
			RefName: &gh.RepositoryRulesetRefConditionParameters{
				Include: condition.Include,
				Exclude: condition.Exclude,
			},
		}
	}
	rules := &gh.RepositoryRulesetRules{}
	for _, rule := range config.Rules {
		parameters := rule.Parameters
		switch rule.Type {
		case "creation":
			rules.Creation = &gh.EmptyRuleParameters{}
		case "update":
			rules.Update = &gh.UpdateRuleParameters{UpdateAllowsFetchAndMerge: value(parameters.UpdateAllowsFetchAndMerge)}
		case "deletion":
			rules.Deletion = &gh.EmptyRuleParameters{}
		case "required_linear_history":
			rules.RequiredLinearHistory = &gh.EmptyRuleParameters{}
		case "required_signatures":
			rules.RequiredSignatures = &gh.EmptyRuleParameters{}
		case "non_fast_forward":
			rules.NonFastForward = &gh.EmptyRuleParameters{}
		case "pull_request":
			mergeMethods := make([]gh.PullRequestMergeMethod, 0, len(parameters.AllowedMergeMethods))
			for _, method := range parameters.AllowedMergeMethods {
				mergeMethods = append(mergeMethods, gh.PullRequestMergeMethod(method))
			}
			rules.PullRequest = &gh.PullRequestRuleParameters{
				AllowedMergeMethods:            mergeMethods,
				RequiredApprovingReviewCount:   value(parameters.RequiredApprovingReviewCount),
				DismissStaleReviewsOnPush:      value(parameters.DismissStaleReviewsOnPush),
				RequireCodeOwnerReview:         value(parameters.RequireCodeOwnerReview),
				RequireLastPushApproval:        value(parameters.RequireLastPushApproval),
				RequiredReviewThreadResolution: value(parameters.RequiredReviewThreadResolution),
				DismissalRestriction:           toDismissalRestriction(parameters.DismissalRestriction),
				RequiredReviewers:              toRequiredReviewers(parameters.RequiredReviewers),
			}
		case "workflows":
			workflows := make([]*gh.RuleWorkflow, 0, len(parameters.Workflows))
			for _, workflow := range parameters.Workflows {
				workflows = append(workflows, &gh.RuleWorkflow{
					Path: workflow.Path, Ref: workflow.Ref, RepositoryID: workflow.RepositoryID, SHA: workflow.SHA,
				})
			}
			rules.Workflows = &gh.WorkflowsRuleParameters{
				DoNotEnforceOnCreate: parameters.DoNotEnforceOnCreate,
				Workflows:            workflows,
			}
		case "code_scanning":
			tools := make([]*gh.RuleCodeScanningTool, 0, len(parameters.CodeScanningTools))
			for _, tool := range parameters.CodeScanningTools {
				tools = append(tools, &gh.RuleCodeScanningTool{
					AlertsThreshold:         gh.CodeScanningAlertsThreshold(tool.AlertsThreshold),
					SecurityAlertsThreshold: gh.CodeScanningSecurityAlertsThreshold(tool.SecurityAlertsThreshold),
					Tool:                    tool.Tool,
				})
			}
			rules.CodeScanning = &gh.CodeScanningRuleParameters{CodeScanningTools: tools}
		case "code_coverage":
			rules.CodeCoverage = &gh.CodeCoverageRuleParameters{
				MaxCoverageDrop: parameters.MaxCoverageDrop,
				MinimumCoverage: parameters.MinimumCoverage,
			}
		case "copilot_code_review":
			rules.CopilotCodeReview = &gh.CopilotCodeReviewRuleParameters{
				ReviewOnPush:            value(parameters.ReviewOnPush),
				ReviewDraftPullRequests: value(parameters.ReviewDraftPullRequests),
			}
		case "file_path_restriction":
			rules.FilePathRestriction = &gh.FilePathRestrictionRuleParameters{RestrictedFilePaths: parameters.RestrictedFilePaths}
		case "file_extension_restriction":
			rules.FileExtensionRestriction = &gh.FileExtensionRestrictionRuleParameters{RestrictedFileExtensions: parameters.RestrictedFileExtensions}
		case "max_file_path_length":
			rules.MaxFilePathLength = &gh.MaxFilePathLengthRuleParameters{MaxFilePathLength: value(parameters.MaxFilePathLength)}
		case "max_file_size":
			rules.MaxFileSize = &gh.MaxFileSizeRuleParameters{MaxFileSize: value(parameters.MaxFileSize)}
		case "required_status_checks":
			checks := make([]*gh.RuleStatusCheck, 0, len(parameters.RequiredStatusChecks))
			for _, check := range parameters.RequiredStatusChecks {
				checks = append(checks, &gh.RuleStatusCheck{Context: check.Context, IntegrationID: check.IntegrationID})
			}
			rules.RequiredStatusChecks = &gh.RequiredStatusChecksRuleParameters{
				DoNotEnforceOnCreate:             parameters.DoNotEnforceOnCreate,
				RequiredStatusChecks:             checks,
				StrictRequiredStatusChecksPolicy: value(parameters.StrictRequiredStatusChecksPolicy),
			}
		case "required_deployments":
			rules.RequiredDeployments = &gh.RequiredDeploymentsRuleParameters{
				RequiredDeploymentEnvironments: parameters.RequiredDeploymentEnvironments,
			}
		case "merge_queue":
			rules.MergeQueue = &gh.MergeQueueRuleParameters{
				CheckResponseTimeoutMinutes:  value(parameters.CheckResponseTimeoutMinutes),
				GroupingStrategy:             gh.MergeGroupingStrategy(value(parameters.GroupingStrategy)),
				MaxEntriesToBuild:            value(parameters.MaxEntriesToBuild),
				MaxEntriesToMerge:            value(parameters.MaxEntriesToMerge),
				MergeMethod:                  gh.MergeQueueMergeMethod(value(parameters.MergeMethod)),
				MinEntriesToMerge:            value(parameters.MinEntriesToMerge),
				MinEntriesToMergeWaitMinutes: value(parameters.MinEntriesToMergeWaitMinutes),
			}
		case "commit_message_pattern":
			rules.CommitMessagePattern = toPattern(parameters)
		case "commit_author_email_pattern":
			rules.CommitAuthorEmailPattern = toPattern(parameters)
		case "committer_email_pattern":
			rules.CommitterEmailPattern = toPattern(parameters)
		case "branch_name_pattern":
			rules.BranchNamePattern = toPattern(parameters)
		case "tag_name_pattern":
			rules.TagNamePattern = toPattern(parameters)
		}
	}
	bypassActors := make([]*gh.BypassActor, 0, len(config.BypassActors))
	for _, actor := range config.BypassActors {
		actorType := gh.BypassActorType(actor.ActorType)
		mode := gh.BypassMode(actor.BypassMode)
		bypassActors = append(bypassActors, &gh.BypassActor{ActorID: actor.ActorID, ActorType: &actorType, BypassMode: &mode})
	}
	return gh.RepositoryRuleset{
		Name:         config.Name,
		Target:       &target,
		Source:       source,
		Enforcement:  gh.RulesetEnforcement(config.Enforcement),
		BypassActors: bypassActors,
		Conditions:   conditions,
		Rules:        rules,
	}
}

func toPattern(parameters *settings.RuleParameters) *gh.PatternRuleParameters {
	return &gh.PatternRuleParameters{
		Name:     parameters.Name,
		Negate:   parameters.Negate,
		Operator: gh.PatternRuleOperator(value(parameters.Operator)),
		Pattern:  value(parameters.Pattern),
	}
}

func pointer[T any](value T) *T {
	return &value
}

func value[T any](pointer *T) T {
	if pointer == nil {
		var zero T
		return zero
	}
	return *pointer
}

func toDismissalRestriction(config *settings.DismissalRestriction) *gh.DismissalRestriction {
	if config == nil {
		return nil
	}
	actors := make([]*gh.DismissalRestrictionActor, 0, len(config.AllowedActors))
	for _, actor := range config.AllowedActors {
		actors = append(actors, &gh.DismissalRestrictionActor{
			ID: actor.ID, Type: gh.DismissalRestrictionActorType(actor.Type),
		})
	}
	return &gh.DismissalRestriction{Enabled: value(config.Enabled), AllowedActors: actors}
}

func toRequiredReviewers(config []settings.RequiredReviewer) []*gh.RulesetRequiredReviewer {
	reviewers := make([]*gh.RulesetRequiredReviewer, 0, len(config))
	for _, reviewer := range config {
		var selection *gh.RulesetReviewer
		if reviewer.Reviewer != nil {
			var reviewerType *gh.RulesetReviewerType
			if reviewer.Reviewer.Type != nil {
				reviewerType = pointer(gh.RulesetReviewerType(*reviewer.Reviewer.Type))
			}
			selection = &gh.RulesetReviewer{ID: reviewer.Reviewer.ID, Type: reviewerType}
		}
		reviewers = append(reviewers, &gh.RulesetRequiredReviewer{
			MinimumApprovals: reviewer.MinimumApprovals,
			FilePatterns:     reviewer.FilePatterns,
			Reviewer:         selection,
		})
	}
	return reviewers
}
