package settings

import (
	"fmt"
	"strings"
)

// Autolink declares one repository autolink. Entries omitted from the manifest are unchanged.
// Existing autolinks cannot be updated through GitHub's REST API; a conflicting
// declaration is rejected rather than deleted and recreated.
type Autolink struct {
	KeyPrefix      string `json:"-" yaml:"key_prefix"`
	URLTemplate    string `json:"-" yaml:"url_template,omitempty"`
	IsAlphanumeric *bool  `json:"-" yaml:"is_alphanumeric,omitempty"`
	State          string `json:"-" yaml:"state,omitempty"`
}

// ValidateAutolinks checks the autolink declarations independently of the
// containing manifest.
func ValidateAutolinks(autolinks []Autolink) error {
	seen := make(map[string]struct{}, len(autolinks))
	for i, autolink := range autolinks {
		if strings.TrimSpace(autolink.KeyPrefix) == "" {
			return fmt.Errorf("autolinks[%d].key_prefix is required", i)
		}
		if _, exists := seen[autolink.KeyPrefix]; exists {
			return fmt.Errorf("autolinks contains duplicate key_prefix %q", autolink.KeyPrefix)
		}
		seen[autolink.KeyPrefix] = struct{}{}

		state := autolink.State
		if state == "" {
			state = "present"
		}
		switch state {
		case "present":
			if strings.TrimSpace(autolink.URLTemplate) == "" {
				return fmt.Errorf("autolinks[%d].url_template is required when state is present", i)
			}
			if !strings.Contains(autolink.URLTemplate, "<num>") {
				return fmt.Errorf("autolinks[%d].url_template must contain <num>", i)
			}
		case "absent":
			if autolink.URLTemplate != "" || autolink.IsAlphanumeric != nil {
				return fmt.Errorf("autolinks[%d] cannot include settings when state is absent", i)
			}
		default:
			return fmt.Errorf("autolinks[%d].state must be present or absent", i)
		}
	}
	return nil
}
