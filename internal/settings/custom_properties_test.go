package settings

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCustomPropertiesValidate(t *testing.T) {
	tests := []struct {
		name       string
		properties CustomProperties
		wantErr    string
	}{
		{name: "valid scalar and list", properties: CustomProperties{
			{Name: "team", Value: StringPropertyValue("platform")},
			{Name: "labels", Value: StringListPropertyValue([]string{"one", "two"})},
		}},
		{name: "missing value", properties: CustomProperties{{Name: "team"}}, wantErr: "value must be provided"},
		{name: "empty name", properties: CustomProperties{{Name: " ", Value: StringPropertyValue("x")}}, wantErr: "name must not be empty"},
		{name: "trimmed name", properties: CustomProperties{{Name: " team ", Value: StringPropertyValue("x")}}, wantErr: "leading or trailing whitespace"},
		{name: "duplicate name", properties: CustomProperties{
			{Name: "team", Value: StringPropertyValue("one")},
			{Name: "team", Value: StringPropertyValue("two")},
		}, wantErr: "is duplicated"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.properties.Validate()
			if tt.wantErr == "" && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestCustomPropertyValueYAMLIsStrict(t *testing.T) {
	var properties CustomProperties
	if err := yaml.Unmarshal([]byte("- name: enabled\n  value: \"false\"\n- name: teams\n  value: [one, two]\n"), &properties); err != nil {
		t.Fatal(err)
	}
	if err := properties.Validate(); err != nil {
		t.Fatal(err)
	}
	var jsonProperties CustomProperties
	if err := json.Unmarshal([]byte(`[{ "name":"status", "value":"active" }, { "name":"groups", "value":["one","two"] }]`), &jsonProperties); err != nil {
		t.Fatal(err)
	}
	if err := jsonProperties.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := jsonProperties[1].Value.APIValue(); !equalStrings(got, []string{"one", "two"}) {
		t.Errorf("JSON list value = %#v, want [one two]", got)
	}
	var invalidJSON CustomProperties
	if err := json.Unmarshal([]byte(`[{ "name":"status", "value":true }]`), &invalidJSON); err == nil {
		t.Error("JSON boolean value unexpectedly succeeded")
	}

	if got := properties[0].Value.APIValue(); got != "false" {
		t.Errorf("scalar value = %#v, want string false", got)
	}
	if got := properties[1].Value.APIValue(); !equalStrings(got, []string{"one", "two"}) {
		t.Errorf("list value = %#v, want [one two]", got)
	}

	for _, input := range []string{
		"- name: enabled\n  value: true\n",
		"- name: enabled\n  value: 42\n",
		"- name: enabled\n  value: [one, 2]\n",
	} {
		var invalid CustomProperties
		if err := yaml.Unmarshal([]byte(input), &invalid); err == nil {
			t.Errorf("yaml.Unmarshal(%q) unexpectedly succeeded", input)
		}
	}
}

func equalStrings(got any, want []string) bool {
	values, ok := got.([]string)
	if !ok || len(values) != len(want) {
		return false
	}
	for i := range values {
		if values[i] != want[i] {
			return false
		}
	}
	return true
}
