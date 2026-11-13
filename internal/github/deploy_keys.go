package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v93/github"

	"github-repository-settings/internal/settings"
)

func (c *Client) applyDeployKeys(ctx context.Context, desired []settings.DeployKey) error {
	existing, err := c.listDeployKeys(ctx)
	if err != nil {
		return err
	}
	byIdentity := make(map[string]*gh.Key, len(existing))
	for _, key := range existing {
		if key == nil || key.Key == nil {
			continue
		}
		identity, identityErr := settings.DeployKeyIdentity(*key.Key)
		if identityErr == nil {
			byIdentity[identity] = key
		}
	}
	for _, declaration := range desired {
		identity, _ := settings.DeployKeyIdentity(declaration.Key) // validated by Client.Apply before any API request.
		current := byIdentity[identity]
		if declaration.State == "absent" {
			if current == nil {
				continue
			}
			if current.ID == nil {
				return fmt.Errorf("cannot delete deploy key because GitHub omitted its ID")
			}
			if _, err := c.api.Repositories.DeleteKey(ctx, c.owner, c.repo, *current.ID); err != nil {
				return fmt.Errorf("delete deploy key with title %q: %w", value(current.Title), err)
			}
			continue
		}
		if current != nil {
			if current.Title != nil && *current.Title == declaration.Title && current.ReadOnly != nil && *current.ReadOnly == *declaration.ReadOnly {
				continue
			}
			return fmt.Errorf("deploy key %q already exists with different settings; delete it explicitly before recreating", declaration.Title)
		}
		_, _, err := c.api.Repositories.CreateKey(ctx, c.owner, c.repo, gh.CreateDeployKeyRequest{
			Key: declaration.Key, Title: &declaration.Title, ReadOnly: declaration.ReadOnly,
		})
		if err != nil {
			return fmt.Errorf("create deploy key %q: %w", declaration.Title, err)
		}
	}
	return nil
}

func (c *Client) listDeployKeys(ctx context.Context) ([]*gh.Key, error) {
	options := &gh.ListOptions{PerPage: 100}
	var all []*gh.Key
	for {
		keys, response, err := c.api.Repositories.ListKeys(ctx, c.owner, c.repo, options)
		if err != nil {
			return nil, fmt.Errorf("list deploy keys: %w", err)
		}
		all = append(all, keys...)
		if response == nil || response.NextPage == 0 {
			return all, nil
		}
		options.Page = response.NextPage
	}
}
