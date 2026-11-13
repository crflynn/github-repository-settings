package settings

import "testing"

func TestValidateWebhooks(t *testing.T) {
	valid := []Webhook{{URL: "https://hooks.example.test/events", Events: []string{"push"}}}
	if err := ValidateWebhooks(valid); err != nil {
		t.Fatalf("ValidateWebhooks rejected valid webhook: %v", err)
	}

	tests := []struct {
		name    string
		webhook []Webhook
	}{
		{name: "relative URL", webhook: []Webhook{{URL: "/events", Events: []string{"push"}}}},
		{name: "unsupported scheme", webhook: []Webhook{{URL: "ftp://hooks.example.test/events", Events: []string{"push"}}}},
		{name: "missing events", webhook: []Webhook{{URL: "https://hooks.example.test/events"}}},
		{name: "empty event", webhook: []Webhook{{URL: "https://hooks.example.test/events", Events: []string{""}}}},
		{name: "duplicate event", webhook: []Webhook{{URL: "https://hooks.example.test/events", Events: []string{"push", "push"}}}},
		{name: "invalid content type", webhook: []Webhook{{URL: "https://hooks.example.test/events", Events: []string{"push"}, ContentType: "xml"}}},
		{name: "duplicate URL", webhook: []Webhook{
			{URL: "https://hooks.example.test/events", Events: []string{"push"}},
			{URL: "https://hooks.example.test/events", Events: []string{"issues"}},
		}},
		{name: "absent with configuration", webhook: []Webhook{{URL: "https://hooks.example.test/events", State: "absent", Events: []string{"push"}}}},
		{name: "unknown state", webhook: []Webhook{{URL: "https://hooks.example.test/events", State: "present", Events: []string{"push"}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateWebhooks(test.webhook); err == nil {
				t.Fatal("ValidateWebhooks accepted invalid webhook")
			}
		})
	}
}

func TestValidateWebhookAbsentDeclaration(t *testing.T) {
	if err := ValidateWebhooks([]Webhook{{URL: "https://hooks.example.test/events", State: "absent"}}); err != nil {
		t.Fatalf("ValidateWebhooks rejected URL-only absent declaration: %v", err)
	}
}
