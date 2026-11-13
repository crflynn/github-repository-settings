package settings

import (
	"strings"
	"testing"
)

func TestValidateAutolinks(t *testing.T) {
	falseValue := false
	tests := []struct {
		name      string
		autolinks []Autolink
		wantError string
	}{
		{name: "present", autolinks: []Autolink{{KeyPrefix: "JIRA-", URLTemplate: "https://issues.example/<num>/"}}},
		{name: "explicitly absent", autolinks: []Autolink{{KeyPrefix: "JIRA-", State: "absent"}}},
		{name: "false is an explicit value", autolinks: []Autolink{{KeyPrefix: "JIRA-", URLTemplate: "https://issues.example/<num>/", IsAlphanumeric: &falseValue}}},
		{name: "missing prefix", autolinks: []Autolink{{URLTemplate: "https://issues.example/<num>/"}}, wantError: "key_prefix is required"},
		{name: "missing template", autolinks: []Autolink{{KeyPrefix: "JIRA-"}}, wantError: "url_template is required"},
		{name: "template missing placeholder", autolinks: []Autolink{{KeyPrefix: "JIRA-", URLTemplate: "https://issues.example/"}}, wantError: "must contain <num>"},
		{name: "empty prefix", autolinks: []Autolink{{KeyPrefix: " \t", State: "absent"}}, wantError: "key_prefix is required"},
		{name: "duplicate prefix", autolinks: []Autolink{{KeyPrefix: "JIRA-", URLTemplate: "https://a/<num>"}, {KeyPrefix: "JIRA-", URLTemplate: "https://b/<num>"}}, wantError: "duplicate key_prefix"},
		{name: "invalid state", autolinks: []Autolink{{KeyPrefix: "JIRA-", State: "removed"}}, wantError: "state must be present or absent"},
		{name: "absent with template", autolinks: []Autolink{{KeyPrefix: "JIRA-", URLTemplate: "https://issues.example/<num>", State: "absent"}}, wantError: "cannot include settings when state is absent"},
		{name: "absent with flag", autolinks: []Autolink{{KeyPrefix: "JIRA-", IsAlphanumeric: &falseValue, State: "absent"}}, wantError: "cannot include settings when state is absent"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAutolinks(test.autolinks)
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("ValidateAutolinks() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("ValidateAutolinks() error = %v, want substring %q", err, test.wantError)
			}
		})
	}
}
