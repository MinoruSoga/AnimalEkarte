// Package apicontract holds API-contract invariant checks that cross-reference the
// hand-maintained OpenAPI spec (docs/api.yaml) against the Go handler layer.
//
// It intentionally lives in its own package (not internal/handler) so the checks read
// handler *source* via os.ReadFile + go/ast and are unaffected by the handler test
// package's compile state. See openapi_date_format_drift_test.go for the first invariant:
// OpenAPI `format: date` (date-only) declarations must not be served as a Go time.Time
// (which JSON-encodes as an RFC3339 datetime), the drift class behind BE-refactor.md R2-1/R3-3.
package apicontract
