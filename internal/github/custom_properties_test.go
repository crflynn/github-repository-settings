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

func TestApplyCustomPropertiesReadsAndPatchesDeclaredValues(t *testing.T) {
	var requests []string
	var patch map[string][]map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch r.Method + " " + r.URL.Path {
		case "GET /repos/acme/widget/properties/values":
			_, _ = w.Write([]byte(`[{"property_name":"team","value":"platform"},{"property_name":"region","value":"west"},{"property_name":"untouched","value":"keep"}]`))
		case "PATCH /repos/acme/widget/properties/values":
			if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
				t.Errorf("decode PATCH body: %v", err)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test-token", "acme/widget", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	properties := settings.CustomProperties{
		{Name: "team", Value: settings.StringPropertyValue("platform")},
		{Name: "region", Value: settings.StringPropertyValue("east")},
		{Name: "services", Value: settings.StringListPropertyValue([]string{"api", "web"})},
	}
	if err := client.applyCustomProperties(context.Background(), properties); err != nil {
		t.Fatal(err)
	}

	if want := []string{"GET /repos/acme/widget/properties/values", "PATCH /repos/acme/widget/properties/values"}; !reflect.DeepEqual(requests, want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	wantPatch := map[string][]map[string]any{"properties": {
		{"property_name": "region", "value": "east"},
		{"property_name": "services", "value": []any{"api", "web"}},
	}}
	if !reflect.DeepEqual(patch, wantPatch) {
		t.Errorf("PATCH body = %#v, want %#v", patch, wantPatch)
	}
}

func TestApplyCustomPropertiesDoesNotPatchWhenValuesMatch(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodGet || r.URL.Path != "/repos/acme/widget/properties/values" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"property_name":"team","value":"platform"}]`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test-token", "acme/widget", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	properties := settings.CustomProperties{{Name: "team", Value: settings.StringPropertyValue("platform")}}
	if err := client.applyCustomProperties(context.Background(), properties); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("request count = %d, want GET only", calls)
	}
}
