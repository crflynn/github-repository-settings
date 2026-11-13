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

const testDeployPublicKey = "ssh-ed25519 YWJj test-key"

func TestApplyDeployKeysCreatesOnlyMissingDeclaredKey(t *testing.T) {
	var methods []string
	var created map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[{"id":7,"key":"ssh-ed25519 ZGVm existing-key","title":"existing","read_only":true}]`))
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&created); err != nil {
			t.Errorf("decode create request: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":8}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "token", "acme/widget", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	readOnly := false
	if err := client.applyDeployKeys(context.Background(), []settings.DeployKey{{Key: testDeployPublicKey, Title: "ci", ReadOnly: &readOnly}}); err != nil {
		t.Fatal(err)
	}
	if want := []string{"GET /repos/acme/widget/keys", "POST /repos/acme/widget/keys"}; !reflect.DeepEqual(methods, want) {
		t.Fatalf("requests = %#v, want %#v", methods, want)
	}
	wantBody := map[string]any{"key": testDeployPublicKey, "title": "ci", "read_only": false}
	if !reflect.DeepEqual(created, wantBody) {
		t.Errorf("create body = %#v, want %#v", created, wantBody)
	}
}

func TestApplyDeployKeysDeletesOnlyExplicitAbsentKey(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[{"id":42,"key":"ssh-ed25519 YWJj old","title":"obsolete","read_only":true}]`))
		}
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "token", "acme/widget", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if err := client.applyDeployKeys(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if want := []string{"GET /repos/acme/widget/keys"}; !reflect.DeepEqual(methods, want) {
		t.Fatalf("undeclared key requests = %#v, want %#v", methods, want)
	}
	methods = nil
	if err := client.applyDeployKeys(context.Background(), []settings.DeployKey{{Key: "ssh-ed25519 YWJj old", State: "absent"}}); err != nil {
		t.Fatal(err)
	}
	if want := []string{"GET /repos/acme/widget/keys", "DELETE /repos/acme/widget/keys/42"}; !reflect.DeepEqual(methods, want) {
		t.Fatalf("explicit deletion requests = %#v, want %#v", methods, want)
	}
}
