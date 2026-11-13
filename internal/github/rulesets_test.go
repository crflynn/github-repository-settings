package github

import (
	"encoding/json"
	"reflect"
	"testing"

	"github-repository-settings/internal/settings"
)

func TestToGitHubRulesetMapsExpandedRules(t *testing.T) {
	config := settings.Ruleset{
		Name: "push-protection", Target: "push", Enforcement: "active",
		BypassActors: []settings.BypassActor{{ActorType: "OrganizationAdmin", BypassMode: "always"}},
		Rules: []settings.Rule{
			{Type: "file_path_restriction", Parameters: &settings.RuleParameters{RestrictedFilePaths: []string{"*.pem"}}},
			{Type: "file_extension_restriction", Parameters: &settings.RuleParameters{RestrictedFileExtensions: []string{".exe"}}},
			{Type: "max_file_path_length", Parameters: &settings.RuleParameters{MaxFilePathLength: new(200)}},
			{Type: "max_file_size", Parameters: &settings.RuleParameters{MaxFileSize: new(int64(1000))}},
		},
	}
	data, err := json.Marshal(toGitHubRuleset(config, "acme/widget"))
	if err != nil {
		t.Fatalf("marshal ruleset: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal ruleset: %v", err)
	}
	actors := got["bypass_actors"].([]any)
	if _, exists := actors[0].(map[string]any)["actor_id"]; exists {
		t.Fatalf("optional actor_id was emitted: %s", data)
	}
	rules := got["rules"].([]any)
	want := map[string]map[string]any{
		"file_path_restriction":      {"restricted_file_paths": []any{"*.pem"}},
		"file_extension_restriction": {"restricted_file_extensions": []any{".exe"}},
		"max_file_path_length":       {"max_file_path_length": float64(200)},
		"max_file_size":              {"max_file_size": float64(1000)},
	}
	for _, raw := range rules {
		rule := raw.(map[string]any)
		typeName := rule["type"].(string)
		if expected, exists := want[typeName]; exists {
			parameters := rule["parameters"].(map[string]any)
			for key, value := range expected {
				if !reflect.DeepEqual(parameters[key], value) {
					t.Errorf("%s parameter %s = %#v, want %#v", typeName, key, parameters[key], value)
				}
			}
			delete(want, typeName)
		}
	}
	if len(want) != 0 {
		t.Errorf("missing rules: %#v", want)
	}
}

func TestToGitHubRulesetMapsPullRequestParameters(t *testing.T) {
	config := settings.Ruleset{
		Name: "protect-main", Target: "branch", Enforcement: "active",
		Conditions: settings.RulesetConditions{RefName: &settings.RefNameCondition{Include: []string{"main"}, Exclude: []string{}}},
		Rules: []settings.Rule{{Type: "pull_request", Parameters: &settings.RuleParameters{
			RequiredApprovingReviewCount: new(2),
			DismissalRestriction:         &settings.DismissalRestriction{Enabled: new(true), AllowedActors: []settings.DismissalRestrictionActor{{ID: 11, Type: "Team"}}},
			RequiredReviewers:            []settings.RequiredReviewer{{MinimumApprovals: new(1), FilePatterns: []string{"*.go"}, Reviewer: &settings.RulesetReviewer{ID: new(int64(12)), Type: new("Team")}}},
		}}},
	}
	data, err := json.Marshal(toGitHubRuleset(config, "acme/widget"))
	if err != nil {
		t.Fatalf("marshal ruleset: %v", err)
	}
	var got struct {
		Rules []struct {
			Parameters struct {
				DismissalRestriction map[string]any `json:"dismissal_restriction"`
				RequiredReviewers    []any          `json:"required_reviewers"`
			} `json:"parameters"`
		} `json:"rules"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal ruleset: %v", err)
	}
	parameters := got.Rules[0].Parameters
	if parameters.DismissalRestriction["enabled"] != true || len(parameters.DismissalRestriction["allowed_actors"].([]any)) != 1 {
		t.Errorf("dismissal restriction mapping = %#v", parameters.DismissalRestriction)
	}
	if len(parameters.RequiredReviewers) != 1 {
		t.Errorf("required reviewers mapping = %#v", parameters.RequiredReviewers)
	}
}
