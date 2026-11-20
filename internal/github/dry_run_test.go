package github

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
)

func TestDryRunTransportForwardsOnlyReadMethods(t *testing.T) {
	var mu sync.Mutex
	var forwarded []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		forwarded = append(forwarded, r.Method)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &dryRunTransport{base: http.DefaultTransport}
	client := &http.Client{Transport: transport}
	methods := []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
	for _, method := range methods {
		request, err := http.NewRequest(method, server.URL+"/repos/acme/widget?per_page=1", nil)
		if err != nil {
			t.Fatal(err)
		}
		response, err := client.Do(request)
		if err != nil {
			t.Fatalf("%s request: %v", method, err)
		}
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Errorf("%s status = %d, want %d", method, response.StatusCode, http.StatusOK)
		}
	}

	mu.Lock()
	gotForwarded := append([]string(nil), forwarded...)
	mu.Unlock()
	if want := []string{http.MethodGet, http.MethodHead}; !reflect.DeepEqual(gotForwarded, want) {
		t.Fatalf("forwarded methods = %v, want %v", gotForwarded, want)
	}
	for _, test := range []struct {
		method string
		action string
	}{
		{http.MethodPost, "create"},
		{http.MethodPut, "update"},
		{http.MethodPatch, "update"},
		{http.MethodDelete, "delete"},
	} {
		if got := (PlannedOperation{Method: test.method}).Action(); got != test.action {
			t.Errorf("%s action = %q, want %q", test.method, got, test.action)
		}
	}
	wantOperations := []PlannedOperation{
		{Method: http.MethodPost, Path: "/repos/acme/widget"},
		{Method: http.MethodPut, Path: "/repos/acme/widget"},
		{Method: http.MethodPatch, Path: "/repos/acme/widget"},
		{Method: http.MethodDelete, Path: "/repos/acme/widget"},
	}
	if got := transport.Operations(); !reflect.DeepEqual(got, wantOperations) {
		t.Fatalf("planned operations = %#v, want %#v", got, wantOperations)
	}
}
