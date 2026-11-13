package github

import (
	"io"
	"net/http"
	"strings"
	"sync"
)

// PlannedOperation describes a write request that Apply would send in dry-run mode.
type PlannedOperation struct {
	Method string
	Path   string
}

// dryRunTransport forwards read requests and locally acknowledges every other
// method so Apply can report its intended writes without sending them.
type dryRunTransport struct {
	base       http.RoundTripper
	mu         sync.Mutex
	operations []PlannedOperation
}

func (t *dryRunTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Method == http.MethodGet || request.Method == http.MethodHead {
		return t.base.RoundTrip(request)
	}

	t.mu.Lock()
	t.operations = append(t.operations, PlannedOperation{
		Method: request.Method,
		Path:   request.URL.EscapedPath(),
	})
	t.mu.Unlock()

	if request.Body != nil {
		_ = request.Body.Close()
	}
	return &http.Response{
		Status:        "200 OK",
		StatusCode:    http.StatusOK,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        http.Header{"Content-Type": []string{"application/json"}},
		Body:          io.NopCloser(strings.NewReader("{}")),
		ContentLength: 2,
		Request:       request,
	}, nil
}

func (t *dryRunTransport) Operations() []PlannedOperation {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]PlannedOperation(nil), t.operations...)
}
