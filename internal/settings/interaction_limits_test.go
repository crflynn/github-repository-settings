package settings

import "testing"

func TestInteractionLimitsValidate(t *testing.T) {
	tests := []struct {
		name  string
		input InteractionLimits
		valid bool
	}{
		{"all limits", InteractionLimits{Limit: "existing_users"}, true},
		{"contributors", InteractionLimits{Limit: "contributors_only", Expiry: "six_months"}, true},
		{"collaborators", InteractionLimits{Limit: "collaborators_only", Expiry: "one_day"}, true},
		{"all expiries", InteractionLimits{Limit: "existing_users", Expiry: "three_days"}, true},
		{"one week", InteractionLimits{Limit: "existing_users", Expiry: "one_week"}, true},
		{"one month", InteractionLimits{Limit: "existing_users", Expiry: "one_month"}, true},
		{"default expiry", InteractionLimits{Limit: "existing_users"}, true},
		{"absence", InteractionLimits{State: "absent"}, true},
		{"missing limit", InteractionLimits{}, false},
		{"bad limit", InteractionLimits{Limit: "everyone"}, false},
		{"bad expiry", InteractionLimits{Limit: "existing_users", Expiry: "forever"}, false},
		{"unknown state", InteractionLimits{State: "remove"}, false},
		{"absent with limit", InteractionLimits{State: "absent", Limit: "existing_users"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err == nil) != tt.valid {
				t.Fatalf("Validate() error = %v, valid = %v", err, tt.valid)
			}
		})
	}
}
