package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v93/github"

	"github-repository-settings/internal/settings"
)

func (c *Client) applyTeamRepository(ctx context.Context, desired settings.TeamRepository) error {
	_, err := c.api.Teams.AddTeamRepoBySlug(ctx, c.owner, desired.Slug, c.owner, c.repo, gh.AddTeamRepoRequest{Permission: &desired.Permission})
	if err != nil {
		return fmt.Errorf("grant repository access to team %q: %w", desired.Slug, err)
	}
	return nil
}
