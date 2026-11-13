package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github-repository-settings/internal/settings"
)

func TestApplyInteractionLimits(t *testing.T) {
	for _, test := range []struct {
		name    string
		desired settings.InteractionLimits
		method  string
		body    map[string]any
	}{
		{"expiry omitted", settings.InteractionLimits{Limit: "contributors_only"}, "PUT", map[string]any{"limit": "contributors_only"}},
		{"expiry preserved", settings.InteractionLimits{Limit: "collaborators_only", Expiry: "three_days"}, "PUT", map[string]any{"limit": "collaborators_only", "expiry": "three_days"}},
		{"delete only when absent", settings.InteractionLimits{State: "absent"}, "DELETE", nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != test.method || r.URL.Path != "/repos/acme/widget/interaction-limits" {
					t.Errorf("request = %s %s", r.Method, r.URL.Path)
				}
				if got := r.Header.Get("Accept"); got != "application/vnd.github.sombra-preview+json" {
					t.Errorf("Accept = %q", got)
				}
				if test.body != nil {
					var got map[string]any
					if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
						t.Errorf("decode body: %v", err)
					}
					if !reflect.DeepEqual(got, test.body) {
						t.Errorf("body = %#v, want %#v", got, test.body)
					}
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()
			client, err := NewClient(server.URL, "test-token", "acme/widget", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			if err := client.applyInteractionLimits(context.Background(), test.desired); err != nil {
				t.Fatal(err)
			}
		})
	}
}
