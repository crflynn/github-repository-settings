package github

import (
	"context"

	gh "github.com/google/go-github/v93/github"

	"github-repository-settings/internal/settings"
)

func (c *Client) applyPages(ctx context.Context, desired settings.Pages) error {
	current, response, err := c.api.Repositories.GetPagesInfo(ctx, c.owner, c.repo)
	if err != nil && !isNotFound(response) {
		return err
	}
	if !*desired.Enabled {
		if current == nil {
			return nil
		}
		response, err := c.api.Repositories.DisablePages(ctx, c.owner, c.repo)
		if isNotFound(response) {
			return nil
		}
		return err
	}
	if current == nil {
		buildType := desired.BuildType
		pages := &gh.Pages{BuildType: &buildType}
		if desired.Source != nil {
			pages.Source = toGitHubPagesSource(desired.Source)
		}
		created, _, err := c.api.Repositories.EnablePages(ctx, c.owner, c.repo, pages)
		if err != nil {
			return err
		}
		if created != nil {
			current = created
		} else {
			current = pages
		}
		if desired.CNAME == nil && desired.Public == nil && desired.HTTPSEnforced == nil {
			return nil
		}
	} else if desired.BuildType == "" && desired.Source == nil && desired.CNAME == nil && desired.Public == nil && desired.HTTPSEnforced == nil {
		return nil
	}
	buildType := desired.BuildType
	if buildType == "" && current != nil {
		buildType = value(current.BuildType)
	}
	source := current.Source
	if desired.Source != nil {
		source = toGitHubPagesSource(desired.Source)
	}
	cname := desired.CNAME
	if cname == nil {
		cname = current.CNAME
	}
	isPublic := desired.Public
	if isPublic == nil {
		isPublic = current.Public
	}
	httpsEnforced := desired.HTTPSEnforced
	if httpsEnforced == nil {
		httpsEnforced = current.HTTPSEnforced
	}
	_, err = c.api.Repositories.UpdatePages(ctx, c.owner, c.repo, &gh.PagesUpdate{
		CNAME:         cname,
		BuildType:     &buildType,
		Source:        source,
		Public:        isPublic,
		HTTPSEnforced: httpsEnforced,
	})
	return err
}

func toGitHubPagesSource(source *settings.PagesSource) *gh.PagesSource {
	return &gh.PagesSource{Branch: &source.Branch, Path: &source.Path}
}
