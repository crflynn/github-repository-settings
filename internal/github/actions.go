package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v93/github"

	"github-repository-settings/internal/settings"
)

func (c *Client) applyActions(ctx context.Context, desired settings.Actions) error {
	if permissions := desired.Permissions; permissions != nil {
		current, _, err := c.api.Repositories.GetActionsPermissions(ctx, c.owner, c.repo)
		if err != nil {
			return fmt.Errorf("get repository Actions permissions: %w", err)
		}
		update := gh.ActionsPermissionsRepository{
			Enabled:            current.Enabled,
			AllowedActions:     current.AllowedActions,
			SHAPinningRequired: current.SHAPinningRequired,
		}
		if permissions.Enabled != nil {
			update.Enabled = *permissions.Enabled
		}
		if permissions.AllowedActions != nil {
			update.AllowedActions = permissions.AllowedActions
		}
		if permissions.SHAPinningRequired != nil {
			update.SHAPinningRequired = permissions.SHAPinningRequired
		}
		if _, _, err := c.api.Repositories.UpdateActionsPermissions(ctx, c.owner, c.repo, update); err != nil {
			return fmt.Errorf("update repository Actions permissions: %w", err)
		}
		if selected := permissions.SelectedActions; selected != nil {
			currentSelected, _, err := c.api.Repositories.GetActionsAllowed(ctx, c.owner, c.repo)
			if err != nil {
				return fmt.Errorf("get allowed Actions: %w", err)
			}
			selectedUpdate := gh.ActionsAllowed{
				GithubOwnedAllowed: currentSelected.GithubOwnedAllowed,
				VerifiedAllowed:    currentSelected.VerifiedAllowed,
				PatternsAllowed:    currentSelected.PatternsAllowed,
			}
			if selected.GitHubOwnedAllowed != nil {
				selectedUpdate.GithubOwnedAllowed = selected.GitHubOwnedAllowed
			}
			if selected.VerifiedAllowed != nil {
				selectedUpdate.VerifiedAllowed = selected.VerifiedAllowed
			}
			if selected.PatternsAllowed != nil {
				selectedUpdate.PatternsAllowed = *selected.PatternsAllowed
			}
			if _, _, err := c.api.Repositories.EditActionsAllowed(ctx, c.owner, c.repo, selectedUpdate); err != nil {
				return fmt.Errorf("update allowed Actions: %w", err)
			}
		}
	}
	if workflow := desired.Workflow; workflow != nil {
		current, _, err := c.api.Repositories.GetDefaultWorkflowPermissions(ctx, c.owner, c.repo)
		if err != nil {
			return fmt.Errorf("get default workflow permissions: %w", err)
		}
		update := gh.DefaultWorkflowPermissionRepository{
			DefaultWorkflowPermissions:   current.DefaultWorkflowPermissions,
			CanApprovePullRequestReviews: current.CanApprovePullRequestReviews,
		}
		if workflow.DefaultWorkflowPermissions != nil {
			update.DefaultWorkflowPermissions = workflow.DefaultWorkflowPermissions
		}
		if workflow.CanApprovePullRequestReviews != nil {
			update.CanApprovePullRequestReviews = workflow.CanApprovePullRequestReviews
		}
		if _, _, err := c.api.Repositories.UpdateDefaultWorkflowPermissions(ctx, c.owner, c.repo, update); err != nil {
			return fmt.Errorf("update default workflow permissions: %w", err)
		}
	}
	if fork := desired.ForkPR; fork != nil {
		current, _, err := c.api.Repositories.GetPrivateRepoForkPRWorkflowSettings(ctx, c.owner, c.repo)
		if err != nil {
			return fmt.Errorf("get private repository fork PR workflow settings: %w", err)
		}
		update := &gh.WorkflowsPermissionsOpt{
			RunWorkflowsFromForkPullRequests:  current.GetRunWorkflowsFromForkPullRequests(),
			SendWriteTokensToWorkflows:        current.SendWriteTokensToWorkflows,
			SendSecretsAndVariables:           current.SendSecretsAndVariables,
			RequireApprovalForForkPRWorkflows: current.RequireApprovalForForkPRWorkflows,
		}
		if fork.RunWorkflows != nil {
			update.RunWorkflowsFromForkPullRequests = *fork.RunWorkflows
		}
		if fork.SendWriteTokens != nil {
			update.SendWriteTokensToWorkflows = fork.SendWriteTokens
		}
		if fork.SendSecretsAndVariables != nil {
			update.SendSecretsAndVariables = fork.SendSecretsAndVariables
		}
		if fork.RequireApproval != nil {
			update.RequireApprovalForForkPRWorkflows = fork.RequireApproval
		}
		if _, err := c.api.Repositories.UpdatePrivateRepoForkPRWorkflowSettings(ctx, c.owner, c.repo, update); err != nil {
			return fmt.Errorf("update private repository fork PR workflow settings: %w", err)
		}
	}
	if retention := desired.Retention; retention != nil && retention.Days != nil {
		if _, err := c.api.Repositories.UpdateArtifactAndLogRetentionPeriod(ctx, c.owner, c.repo, gh.ArtifactPeriodOpt{Days: *retention.Days}); err != nil {
			return fmt.Errorf("update artifact and log retention period: %w", err)
		}
	}
	return nil
}
