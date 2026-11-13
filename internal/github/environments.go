package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v93/github"

	"github-repository-settings/internal/settings"
)

func (c *Client) applyEnvironment(ctx context.Context, desired settings.Environment) error {
	if desired.State == "absent" {
		response, err := c.api.Repositories.DeleteEnvironment(ctx, c.owner, c.repo, desired.Name)
		if isNotFound(response) {
			return nil
		}
		return err
	}

	current, response, err := c.api.Repositories.GetEnvironment(ctx, c.owner, c.repo, desired.Name)
	if err != nil && !isNotFound(response) {
		return err
	}

	waitTimer := 0
	canAdminsBypass := true
	var reviewers []*gh.EnvReviewers
	var branchPolicy *gh.BranchPolicy
	var preventSelfReview *bool
	if current != nil {
		if current.WaitTimer != nil {
			waitTimer = *current.WaitTimer
		}
		if current.CanAdminsBypass != nil {
			canAdminsBypass = *current.CanAdminsBypass
		}
		reviewers = current.Reviewers
		branchPolicy = current.DeploymentBranchPolicy
		for _, rule := range current.ProtectionRules {
			if rule != nil && rule.GetType() == "required_reviewers" {
				preventSelfReview = rule.PreventSelfReview
				break
			}
		}
	}
	if desired.WaitTimer != nil {
		waitTimer = *desired.WaitTimer
	}
	if desired.CanAdminsBypass != nil {
		canAdminsBypass = *desired.CanAdminsBypass
	}
	if desired.Reviewers != nil {
		reviewers = make([]*gh.EnvReviewers, 0, len(*desired.Reviewers))
		for _, reviewer := range *desired.Reviewers {
			reviewerType, reviewerID := reviewer.Type, reviewer.ID
			reviewers = append(reviewers, &gh.EnvReviewers{Type: &reviewerType, ID: &reviewerID})
		}
	}
	if desired.DeploymentBranchPolicy != nil {
		protected, custom := desired.DeploymentBranchPolicy.ProtectedBranches, desired.DeploymentBranchPolicy.CustomBranchPolicies
		branchPolicy = &gh.BranchPolicy{ProtectedBranches: &protected, CustomBranchPolicies: &custom}
	}
	if desired.PreventSelfReview != nil {
		preventSelfReview = desired.PreventSelfReview
	}

	_, _, err = c.api.Repositories.CreateUpdateEnvironment(ctx, c.owner, c.repo, desired.Name, &gh.CreateUpdateEnvironment{
		WaitTimer:              &waitTimer,
		Reviewers:              reviewers,
		CanAdminsBypass:        &canAdminsBypass,
		DeploymentBranchPolicy: branchPolicy,
		PreventSelfReview:      preventSelfReview,
	})
	if err != nil {
		return fmt.Errorf("create or update environment: %w", err)
	}
	return nil
}
