package github

import (
	"context"

	gh "github.com/google/go-github/v93/github"

	"github-repository-settings/internal/settings"
)

func (c *Client) applyCollaborator(ctx context.Context, desired settings.Collaborator) error {
	if desired.State == "absent" {
		_, err := c.api.Repositories.RemoveCollaborator(ctx, c.owner, c.repo, desired.Username)
		return err
	}
	_, _, err := c.api.Repositories.AddCollaborator(ctx, c.owner, c.repo, desired.Username, gh.AddCollaboratorRequest{Permission: &desired.Permission})
	return err
}
