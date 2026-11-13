package github

import (
	"context"
	"fmt"
	"reflect"

	gh "github.com/google/go-github/v93/github"

	"github-repository-settings/internal/settings"
)

// The GET endpoint is available to repository readers; PATCH requires repository-level
// "edit custom property values" permission. PATCHing only the declared properties
// preserves values for properties omitted from the manifest.
func (c *Client) applyCustomProperties(ctx context.Context, desired settings.CustomProperties) error {
	current, _, err := c.api.Repositories.GetAllCustomPropertyValues(ctx, c.owner, c.repo)
	if err != nil {
		return fmt.Errorf("get repository custom property values: %w", err)
	}
	currentByName := make(map[string]any, len(current))
	for _, property := range current {
		if property != nil {
			currentByName[property.PropertyName] = property.Value
		}
	}

	updates := make([]*gh.CustomPropertyValue, 0, len(desired))
	for _, property := range desired {
		value := property.Value.APIValue()
		if existing, ok := currentByName[property.Name]; ok && reflect.DeepEqual(existing, value) {
			continue
		}
		updates = append(updates, &gh.CustomPropertyValue{PropertyName: property.Name, Value: value})
	}
	if len(updates) == 0 {
		return nil
	}
	if _, err := c.api.Repositories.CreateOrUpdateCustomProperties(ctx, c.owner, c.repo, updates); err != nil {
		return fmt.Errorf("update repository custom property values: %w", err)
	}
	return nil
}
