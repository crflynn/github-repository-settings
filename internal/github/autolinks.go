package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v93/github"

	"github-repository-settings/internal/settings"
)

func (c *Client) applyAutolinks(ctx context.Context, desired []settings.Autolink) error {
	if err := settings.ValidateAutolinks(desired); err != nil {
		return err
	}
	if len(desired) == 0 {
		return nil
	}
	current, _, err := c.api.Repositories.ListAutolinks(ctx, c.owner, c.repo)
	if err != nil {
		return fmt.Errorf("list repository autolinks: %w", err)
	}

	for _, declaration := range desired {
		var match *gh.Autolink
		for _, existing := range current {
			if existing == nil || value(existing.KeyPrefix) != declaration.KeyPrefix {
				continue
			}
			if match != nil {
				return fmt.Errorf("repository has multiple autolinks with key_prefix %q", declaration.KeyPrefix)
			}
			match = existing
		}

		state := declaration.State
		if state == "" {
			state = "present"
		}
		if state == "absent" {
			if match == nil {
				continue
			}
			if match.ID == nil {
				return fmt.Errorf("delete autolink %q: GitHub response omitted its ID", declaration.KeyPrefix)
			}
			if _, err := c.api.Repositories.DeleteAutolink(ctx, c.owner, c.repo, *match.ID); err != nil {
				return fmt.Errorf("delete autolink %q: %w", declaration.KeyPrefix, err)
			}
			continue
		}

		if match == nil {
			_, _, err := c.api.Repositories.CreateAutolink(ctx, c.owner, c.repo, gh.CreateAutolinkRequest{
				KeyPrefix:      declaration.KeyPrefix,
				URLTemplate:    declaration.URLTemplate,
				IsAlphanumeric: declaration.IsAlphanumeric,
			})
			if err != nil {
				return fmt.Errorf("create autolink %q: %w", declaration.KeyPrefix, err)
			}
			continue
		}

		if value(match.URLTemplate) != declaration.URLTemplate {
			return fmt.Errorf("autolink %q already exists with a different url_template; GitHub does not support updating autolinks", declaration.KeyPrefix)
		}
		if declaration.IsAlphanumeric != nil && (match.IsAlphanumeric == nil || *match.IsAlphanumeric != *declaration.IsAlphanumeric) {
			return fmt.Errorf("autolink %q already exists with a different is_alphanumeric value; GitHub does not support updating autolinks", declaration.KeyPrefix)
		}
	}
	return nil
}
