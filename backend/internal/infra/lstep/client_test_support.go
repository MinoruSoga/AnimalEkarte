package lstep

import (
	"net/http"
	"testing"
)

// Shared test support is compiled as normal Go for cross-package tests.
// Moving this constructor to _test.go would hide it from those consumers.
// The runtime test guard below remains mandatory.

// NewInsecureTestClient talks to httptest.Server (loopback). It panics outside
// `go test` so production binaries cannot use it to bypass hardenedDialContext.
func NewInsecureTestClient(apiKey, baseURL string) Client {
	if !testing.Testing() {
		panic("lstep.NewInsecureTestClient is test-only")
	}
	return &httpLstepClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		http:    &http.Client{Timeout: defaultTimeout},
	}
}
