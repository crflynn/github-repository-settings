package settings

import (
	"fmt"
)

// InteractionLimits configures repository interaction restrictions. Omitted
// components are unchanged; state: absent explicitly removes the restriction.
type InteractionLimits struct {
	State  string `json:"-" yaml:"state,omitempty"`
	Limit  string `json:"limit,omitempty" yaml:"limit,omitempty"`
	Expiry string `json:"expiry,omitempty" yaml:"expiry,omitempty"`
}

// Validate checks the interaction restriction enums and state-specific fields.
func (i InteractionLimits) Validate() error {
	if i.State != "" && i.State != "present" && i.State != "absent" {
		return fmt.Errorf("state must be present or absent")
	}
	if i.State == "absent" {
		if i.Limit != "" || i.Expiry != "" {
			return fmt.Errorf("limit and expiry cannot be set when state is absent")
		}
		return nil
	}
	if !oneOf(i.Limit, "existing_users", "contributors_only", "collaborators_only") {
		return fmt.Errorf("limit must be existing_users, contributors_only, or collaborators_only")
	}
	if i.Expiry != "" && !oneOf(i.Expiry, "one_day", "three_days", "one_week", "one_month", "six_months") {
		return fmt.Errorf("expiry must be one_day, three_days, one_week, one_month, or six_months")
	}
	return nil
}
