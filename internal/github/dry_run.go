package github

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
)

// PlannedOperation describes a write request that Apply would send in dry-run mode.
type PlannedOperation struct {
	Method  string
	Path    string
	Details json.RawMessage
}

// Action describes the operation represented by an intercepted write request.
func (operation PlannedOperation) Action() string {
	switch operation.Method {
	case http.MethodPost:
		return "create"
	case http.MethodPut, http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "write"
	}
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

	var body []byte
	if request.Body != nil {
		var err error
		body, err = io.ReadAll(request.Body)
		_ = request.Body.Close()
		if err != nil {
			return nil, err
		}
	}
	t.mu.Lock()
	t.operations = append(t.operations, PlannedOperation{
		Method:  request.Method,
		Path:    request.URL.EscapedPath(),
		Details: append(json.RawMessage(nil), body...),
	})
	t.mu.Unlock()

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
