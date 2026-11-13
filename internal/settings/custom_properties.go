package settings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// CustomProperties declares organization-defined property values for a repository.
// Only listed properties are updated; omitted properties are preserved.
type CustomProperties []CustomProperty

// CustomProperty declares one organization-defined property value.
type CustomProperty struct {
	Name  string              `json:"name" yaml:"name"`
	Value CustomPropertyValue `json:"value" yaml:"value"`
}

// CustomPropertyValue is either one string or a list of strings, matching the
// values accepted by GitHub's custom property values API.
type CustomPropertyValue struct {
	stringValue  string
	stringValues []string
	isList       bool
	set          bool
}

// StringPropertyValue constructs a scalar custom property value.
func StringPropertyValue(value string) CustomPropertyValue {
	return CustomPropertyValue{stringValue: value, set: true}
}

// StringListPropertyValue constructs a multi-value custom property value.
func StringListPropertyValue(values []string) CustomPropertyValue {
	if values == nil {
		values = []string{}
	}
	return CustomPropertyValue{stringValues: values, isList: true, set: true}
}

// APIValue returns the string or string slice represented by the value.
func (v CustomPropertyValue) APIValue() any {
	if v.isList {
		return v.stringValues
	}
	return v.stringValue
}

func (v *CustomPropertyValue) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode && node.Tag == "!!str" {
		v.stringValue = node.Value
		v.stringValues = nil
		v.isList = false
		v.set = true
		return nil
	}
	if node.Kind != yaml.SequenceNode {
		return fmt.Errorf("custom property value must be a string or list of strings")
	}
	values := make([]string, 0, len(node.Content))
	for i, child := range node.Content {
		if child.Kind != yaml.ScalarNode || child.Tag != "!!str" {
			return fmt.Errorf("custom property value[%d] must be a string", i)
		}
		values = append(values, child.Value)
	}
	v.stringValue = ""
	v.stringValues = values
	v.isList = true
	v.set = true
	return nil
}

func (v *CustomPropertyValue) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return fmt.Errorf("custom property value must be a string or list of strings")
	}
	if data[0] == '[' {
		var values []string
		if err := json.Unmarshal(data, &values); err != nil {
			return fmt.Errorf("custom property value must be a list of strings: %w", err)
		}
		if values == nil {
			values = []string{}
		}
		v.stringValue = ""
		v.stringValues = values
		v.isList = true
		v.set = true
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("custom property value must be a string or list of strings: %w", err)
	}
	v.stringValue = value
	v.stringValues = nil
	v.isList = false
	v.set = true
	return nil
}

// Validate checks property names and duplicate declarations before making API calls.
func (properties CustomProperties) Validate() error {
	seen := make(map[string]struct{}, len(properties))
	for i, property := range properties {
		if !property.Value.set {
			return fmt.Errorf("custom_properties[%d].value must be provided", i)
		}
		name := strings.TrimSpace(property.Name)
		if name == "" {
			return fmt.Errorf("custom_properties[%d].name must not be empty", i)
		}
		if name != property.Name {
			return fmt.Errorf("custom_properties[%d].name must not have leading or trailing whitespace", i)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("custom_properties[%d].name %q is duplicated", i, name)
		}
		seen[name] = struct{}{}
	}
	return nil
}
