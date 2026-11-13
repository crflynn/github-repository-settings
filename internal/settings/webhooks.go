package settings

import (
	"fmt"
	"net/url"
	"strings"
)

// Webhook declares one repository webhook. Omitted webhooks are unchanged.
type Webhook struct {
	URL         string   `json:"url" yaml:"url"`
	State       string   `json:"-" yaml:"state,omitempty"`
	Events      []string `json:"events,omitempty" yaml:"events,omitempty"`
	Active      *bool    `json:"active,omitempty" yaml:"active,omitempty"`
	ContentType string   `json:"content_type,omitempty" yaml:"content_type,omitempty"`
	InsecureSSL *bool    `json:"insecure_ssl,omitempty" yaml:"insecure_ssl,omitempty"`
	SecretEnv   string   `json:"secret_env,omitempty" yaml:"secret_env,omitempty"`
}

// ValidateWebhooks checks webhook declarations before any API request is sent.
func ValidateWebhooks(webhooks []Webhook) error {
	seen := make(map[string]struct{}, len(webhooks))
	for i, hook := range webhooks {
		parsed, err := url.ParseRequestURI(hook.URL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			return fmt.Errorf("webhooks[%d].url must be an absolute HTTP or HTTPS URL", i)
		}
		if _, ok := seen[hook.URL]; ok {
			return fmt.Errorf("webhooks[%d].url duplicates another declared webhook", i)
		}
		seen[hook.URL] = struct{}{}
		if hook.State != "" && hook.State != "absent" {
			return fmt.Errorf("webhooks[%d].state must be absent when set", i)
		}
		if hook.State == "absent" {
			if len(hook.Events) != 0 || hook.Active != nil || hook.ContentType != "" || hook.InsecureSSL != nil || hook.SecretEnv != "" {
				return fmt.Errorf("webhooks[%d] may only specify url and state when absent", i)
			}
			continue
		}
		if len(hook.Events) == 0 {
			return fmt.Errorf("webhooks[%d].events must contain at least one event", i)
		}
		events := make(map[string]struct{}, len(hook.Events))
		for _, event := range hook.Events {
			if strings.TrimSpace(event) == "" {
				return fmt.Errorf("webhooks[%d].events must not contain empty values", i)
			}
			if _, ok := events[event]; ok {
				return fmt.Errorf("webhooks[%d].events must not contain duplicates", i)
			}
			events[event] = struct{}{}
		}
		if hook.ContentType != "" && hook.ContentType != "json" && hook.ContentType != "form" {
			return fmt.Errorf("webhooks[%d].content_type must be json or form", i)
		}
		if hook.SecretEnv != "" && !validEnvironmentName(hook.SecretEnv) {
			return fmt.Errorf("webhooks[%d].secret_env must be an environment variable name", i)
		}
	}
	return nil
}

func validEnvironmentName(name string) bool {
	if name == "" {
		return false
	}
	for i, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_' || (i > 0 && char >= '0' && char <= '9') {
			continue
		}
		return false
	}
	return true
}
