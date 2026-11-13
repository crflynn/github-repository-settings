package github

import (
	"context"

	gh "github.com/google/go-github/v93/github"

	"github-repository-settings/internal/settings"
)

type interactionLimitRequest struct {
	Limit  string `json:"limit"`
	Expiry string `json:"expiry,omitempty"`
}

// applyInteractionLimits reconciles the repository's interaction restriction.
func (c *Client) applyInteractionLimits(ctx context.Context, desired settings.InteractionLimits) error {
	if desired.State == "absent" {
		_, err := c.api.Interactions.RemoveRestrictionsFromRepo(ctx, c.owner, c.repo)
		return err
	}
	// go-github v93's UpdateRestrictionsForRepo only accepts limit and therefore
	// cannot pass the API's optional expiry. Use its request/response machinery
	// directly so the configured duration is not silently discarded.
	request := &interactionLimitRequest{Limit: desired.Limit, Expiry: desired.Expiry}
	req, err := c.api.NewRequest(ctx, "PUT", "repos/"+c.owner+"/"+c.repo+"/interaction-limits", request)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github.sombra-preview+json")
	var response *gh.InteractionRestriction
	_, err = c.api.Do(req, &response)
	return err
}
