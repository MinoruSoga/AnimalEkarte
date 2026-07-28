// Package apicontract holds API-contract invariant checks that cross-reference the
// hand-maintained OpenAPI spec (docs/api.yaml) against target-domain HTTP boundaries.
//
// The checks read Go source via os.ReadFile and go/ast, so their results do not depend
// on target package test compilation. See openapi_date_format_drift_test.go for the
// first invariant:
// OpenAPI `format: date` (date-only) declarations must not be served as a Go time.Time
// (which JSON-encodes as an RFC3339 datetime), the drift class behind BE-refactor.md R2-1/R3-3.
//
// See openapi_route_drift_test.go for the second invariant (BE-refactor.md G1-4): every
// (method, path) resolved by statically walking the explicit target route-root registry
// must be documented in docs/api.yaml, and vice versa. New drift in either direction fails
// the gate; the current residual is pinned to knownMissingFromSpec/knownPhantomInSpec.
package apicontract
