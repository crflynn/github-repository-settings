package github

import (
	"context"

	gh "github.com/google/go-github/v93/github"

	"github-repository-settings/internal/settings"
)

func (c *Client) applyRepository(ctx context.Context, desired settings.Repository) error {
	update := &gh.Repository{}
	hasUpdate := false
	hasUpdate = setIfPresent(&update.Description, desired.Description) || hasUpdate
	hasUpdate = setIfPresent(&update.Homepage, desired.Homepage) || hasUpdate
	hasUpdate = setIfPresent(&update.Visibility, desired.Visibility) || hasUpdate
	hasUpdate = setIfPresent(&update.DefaultBranch, desired.DefaultBranch) || hasUpdate
	hasUpdate = setIfPresent(&update.IsTemplate, desired.IsTemplate) || hasUpdate
	hasUpdate = setIfPresent(&update.HasPullRequests, desired.HasPullRequests) || hasUpdate
	hasUpdate = setIfPresent(&update.PullRequestCreationPolicy, desired.PullRequestCreationPolicy) || hasUpdate
	hasUpdate = setIfPresent(&update.HasIssues, desired.HasIssues) || hasUpdate
	hasUpdate = setIfPresent(&update.HasProjects, desired.HasProjects) || hasUpdate
	hasUpdate = setIfPresent(&update.HasWiki, desired.HasWiki) || hasUpdate
	hasUpdate = setIfPresent(&update.HasDiscussions, desired.HasDiscussions) || hasUpdate
	hasUpdate = setIfPresent(&update.HasDownloads, desired.HasDownloads) || hasUpdate
	hasUpdate = setIfPresent(&update.AllowSquashMerge, desired.AllowSquashMerge) || hasUpdate
	hasUpdate = setIfPresent(&update.AllowMergeCommit, desired.AllowMergeCommit) || hasUpdate
	hasUpdate = setIfPresent(&update.AllowRebaseMerge, desired.AllowRebaseMerge) || hasUpdate
	hasUpdate = setIfPresent(&update.AllowAutoMerge, desired.AllowAutoMerge) || hasUpdate
	hasUpdate = setIfPresent(&update.AllowUpdateBranch, desired.AllowUpdateBranch) || hasUpdate
	hasUpdate = setIfPresent(&update.AllowForking, desired.AllowForking) || hasUpdate
	hasUpdate = setIfPresent(&update.DeleteBranchOnMerge, desired.DeleteBranchOnMerge) || hasUpdate
	hasUpdate = setIfPresent(&update.WebCommitSignoffRequired, desired.WebCommitSignoffRequired) || hasUpdate
	hasUpdate = setIfPresent(&update.SquashMergeCommitTitle, desired.SquashMergeCommitTitle) || hasUpdate
	hasUpdate = setIfPresent(&update.SquashMergeCommitMessage, desired.SquashMergeCommitMessage) || hasUpdate
	hasUpdate = setIfPresent(&update.MergeCommitTitle, desired.MergeCommitTitle) || hasUpdate
	hasUpdate = setIfPresent(&update.MergeCommitMessage, desired.MergeCommitMessage) || hasUpdate
	hasUpdate = setIfPresent(&update.Archived, desired.Archived) || hasUpdate
	security := &gh.SecurityAndAnalysis{}
	if desired.DependabotSecurityUpdates != nil {
		security.DependabotSecurityUpdates = &gh.DependabotSecurityUpdates{Status: statePtr(*desired.DependabotSecurityUpdates)}
	}
	if desired.CodeSecurity != nil {
		security.CodeSecurity = &gh.CodeSecurity{Status: statePtr(*desired.CodeSecurity)}
	}
	if desired.SecretScanning != nil {
		security.SecretScanning = &gh.SecretScanning{Status: statePtr(*desired.SecretScanning)}
	}
	if desired.SecretScanningPushProtection != nil {
		security.SecretScanningPushProtection = &gh.SecretScanningPushProtection{Status: statePtr(*desired.SecretScanningPushProtection)}
	}
	if desired.DependabotSecurityUpdates != nil || desired.CodeSecurity != nil || desired.SecretScanning != nil || desired.SecretScanningPushProtection != nil {
		update.SecurityAndAnalysis = security
		hasUpdate = true
	}
	if hasUpdate {
		if _, _, err := c.api.Repositories.Edit(ctx, c.owner, c.repo, update); err != nil {
			return err
		}
	}
	if desired.Topics != nil {
		if _, _, err := c.api.Repositories.ReplaceAllTopics(ctx, c.owner, c.repo, *desired.Topics); err != nil {
			return err
		}
	}
	if desired.VulnerabilityAlerts != nil {
		if *desired.VulnerabilityAlerts {
			if _, err := c.api.Repositories.EnableVulnerabilityAlerts(ctx, c.owner, c.repo); err != nil {
				return err
			}
		} else if _, err := c.api.Repositories.DisableVulnerabilityAlerts(ctx, c.owner, c.repo); err != nil {
			return err
		}
	}
	if desired.PrivateVulnerabilityReporting != nil {
		if *desired.PrivateVulnerabilityReporting {
			if _, err := c.api.Repositories.EnablePrivateReporting(ctx, c.owner, c.repo); err != nil {
				return err
			}
		} else if _, err := c.api.Repositories.DisablePrivateReporting(ctx, c.owner, c.repo); err != nil {
			return err
		}
	}
	return nil
}

func statePtr(enabled bool) *string {
	if enabled {
		return gh.String("enabled")
	}
	return gh.String("disabled")
}

func setIfPresent[T any](destination **T, value *T) bool {
	if value == nil {
		return false
	}
	*destination = value
	return true
}
