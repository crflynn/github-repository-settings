package github

import (
	"context"
	"fmt"
	"os"

	gh "github.com/google/go-github/v93/github"

	"github-repository-settings/internal/settings"
)

func (c *Client) applyWebhooks(ctx context.Context, desired []settings.Webhook) error {
	if err := settings.ValidateWebhooks(desired); err != nil {
		return err
	}
	existing, err := c.listHooks(ctx)
	if err != nil {
		return fmt.Errorf("list repository webhooks: %w", err)
	}
	byURL := make(map[string]*gh.Hook, len(existing))
	for _, hook := range existing {
		if hook.Config != nil && hook.Config.URL != nil {
			byURL[*hook.Config.URL] = hook
		}
	}

	// PATCH /hooks/{id} removes an existing secret if the request does not include it.
	// Resolve every secret reference and reject unsafe event/active changes before mutation.
	secrets := make(map[string]*string, len(desired))
	for _, config := range desired {
		if config.State == "absent" {
			continue
		}
		secret, err := webhookSecret(config.SecretEnv)
		if err != nil {
			return fmt.Errorf("webhook %q: %w", config.URL, err)
		}
		secrets[config.URL] = secret
		if secret != nil {
			continue
		}
		hook := byURL[config.URL]
		if hook != nil && (eventsDiffer(hook.Events, config.Events) || (config.Active != nil && (hook.Active == nil || *config.Active != *hook.Active))) {
			return fmt.Errorf("webhook %q requires secret_env to safely update events or active state", config.URL)
		}
	}

	for _, config := range desired {
		hook := byURL[config.URL]
		if config.State == "absent" {
			if hook == nil {
				continue
			}
			if hook.ID == nil {
				return fmt.Errorf("delete webhook %q: GitHub response has no hook id", config.URL)
			}
			if _, err := c.api.Repositories.DeleteHook(ctx, c.owner, c.repo, *hook.ID); err != nil {
				return fmt.Errorf("delete webhook %q: %w", config.URL, err)
			}
			continue
		}

		secret := secrets[config.URL]
		if hook == nil {
			active := true
			if config.Active != nil {
				active = *config.Active
			}
			created := &gh.Hook{
				Config: webhookConfig(config, nil, secret),
				Events: config.Events,
				Active: gh.Bool(active),
			}
			if _, _, err := c.api.Repositories.CreateHook(ctx, c.owner, c.repo, created); err != nil {
				if secret != nil {
					return fmt.Errorf("create webhook %q failed", config.URL)
				}
				return fmt.Errorf("create webhook %q: %w", config.URL, err)
			}
			continue
		}
		if hook.ID == nil {
			return fmt.Errorf("update webhook %q: GitHub response has no hook id", config.URL)
		}

		eventsOrActiveDiffer := eventsDiffer(hook.Events, config.Events) || (config.Active != nil && (hook.Active == nil || *config.Active != *hook.Active))
		if secret != nil || config.ContentType != "" || config.InsecureSSL != nil {
			current := hook.Config
			if current == nil || current.URL == nil || (eventsOrActiveDiffer && secret != nil) {
				current, _, err = c.api.Repositories.GetHookConfiguration(ctx, c.owner, c.repo, *hook.ID)
				if err != nil {
					return fmt.Errorf("get webhook configuration %q: %w", config.URL, err)
				}
			}
			if current == nil {
				return fmt.Errorf("get webhook configuration %q: GitHub returned no configuration", config.URL)
			}
			if secret != nil && eventsOrActiveDiffer {
				update := &gh.Hook{Config: webhookConfig(config, current, secret), Events: config.Events}
				if config.Active != nil {
					update.Active = config.Active
				}
				if _, _, err := c.api.Repositories.EditHook(ctx, c.owner, c.repo, *hook.ID, update); err != nil {
					return fmt.Errorf("update webhook %q failed", config.URL)
				}
			} else if secret != nil || webhookConfigDiffers(config, current) {
				update := webhookConfig(config, current, secret)
				if _, _, err := c.api.Repositories.UpdateHookConfiguration(ctx, c.owner, c.repo, *hook.ID, *update); err != nil {
					if secret != nil {
						return fmt.Errorf("update webhook configuration %q failed", config.URL)
					}
					return fmt.Errorf("update webhook configuration %q: %w", config.URL, err)
				}
			}
		}
		if secret != nil && eventsOrActiveDiffer {
			continue
		}
		if eventsOrActiveDiffer {
			return fmt.Errorf("webhook %q requires secret_env to safely update events or active state", config.URL)
		}
	}
	return nil
}

func (c *Client) listHooks(ctx context.Context) ([]*gh.Hook, error) {
	options := &gh.ListOptions{PerPage: 100}
	var all []*gh.Hook
	for {
		hooks, response, err := c.api.Repositories.ListHooks(ctx, c.owner, c.repo, options)
		if err != nil {
			return nil, err
		}
		all = append(all, hooks...)
		if response == nil || response.NextPage == 0 {
			return all, nil
		}
		options.Page = response.NextPage
	}
}

func webhookSecret(environment string) (*string, error) {
	if environment == "" {
		return nil, nil
	}
	secret, ok := os.LookupEnv(environment)
	if !ok || secret == "" {
		return nil, fmt.Errorf("environment variable %q is unset or empty", environment)
	}
	return &secret, nil
}

func webhookConfig(desired settings.Webhook, current *gh.HookConfig, secret *string) *gh.HookConfig {
	config := &gh.HookConfig{URL: gh.String(desired.URL)}
	if current != nil {
		config.ContentType = current.ContentType
		config.InsecureSSL = current.InsecureSSL
	}
	if desired.ContentType != "" {
		config.ContentType = gh.String(desired.ContentType)
	}
	if desired.InsecureSSL != nil {
		value := "0"
		if *desired.InsecureSSL {
			value = "1"
		}
		config.InsecureSSL = &value
	}
	config.Secret = secret
	return config
}

func webhookConfigDiffers(desired settings.Webhook, current *gh.HookConfig) bool {
	if current.URL == nil || *current.URL != desired.URL {
		return true
	}
	if desired.ContentType != "" && (current.ContentType == nil || *current.ContentType != desired.ContentType) {
		return true
	}
	if desired.InsecureSSL != nil {
		want := "0"
		if *desired.InsecureSSL {
			want = "1"
		}
		return current.InsecureSSL == nil || *current.InsecureSSL != want
	}
	return false
}

func eventsDiffer(current, desired []string) bool {
	if len(current) != len(desired) {
		return true
	}
	for _, event := range desired {
		found := false
		for _, currentEvent := range current {
			if event == currentEvent {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	return false
}
