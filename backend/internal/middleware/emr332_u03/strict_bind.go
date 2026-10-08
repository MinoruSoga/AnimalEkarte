// Package emr332_u03 holds the U03-INPUT-OUTPUT scoped defensive helpers and
// their memory-only verification (EMR-332 security-defense campaign).
package emr332_u03

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// DefaultStrictBindMaxBytes bounds the request body when callers pass a
// non-positive maxBytes to BindStrictJSON. It mirrors the global
// middleware.DefaultJSONBodyMaxBytes ceiling for non-binary write traffic.
const DefaultStrictBindMaxBytes int64 = 1 << 20 // 1 MiB

// BindStrictJSON is the strict-decoding variant of c.ShouldBindJSON for JSON
// write endpoints.
//
// U03/C05 evidence (docs/work/emr-332-security/u03/c05-input-validation.md):
// neither ShouldBindJSON nor the per-package bindStaffJSON / bindAuthJSON
// wrappers reject unknown request fields, because Gin's JSON decoder never
// enables Decoder.DisallowUnknownFields (EnableDecoderDisallowUnknownFields is
// unset). Unknown/disallowed fields are silently dropped on every JSON write
// path that uses them.
//
// This helper keeps the same boundary contract as the existing per-package
// bind helpers — bounded read to EOF, exactly one JSON value, `binding` tag
// validation via Gin's validator engine, sanitized error messages via
// httpapi.ParseBindError — and additionally rejects unknown fields. It decodes
// into dst directly from the bounded buffer so the strict decoder applies;
// `binding` tags are still enforced afterwards through binding.Validator, so
// ParseBindError produces messages identical to the ShouldBindJSON path.
//
// Wiring it into handlers is intentionally NOT done in this unit — in-place
// handler edits live outside the U03 owned paths and are recorded as a
// remediation candidate in the C05 verdict record.
func BindStrictJSON(c *gin.Context, dst any, maxBytes int64) error {
	if maxBytes <= 0 {
		maxBytes = DefaultStrictBindMaxBytes
	}
	if c.Request.ContentLength > maxBytes {
		return apperrors.WrapPayloadTooLarge("request body exceeds size limit")
	}

	bounded := http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
	defer func() {
		_ = bounded.Close()
	}()
	body, err := io.ReadAll(bounded)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return apperrors.WrapPayloadTooLarge("request body exceeds size limit")
		}
		return apperrors.WrapInvalidInput("invalid request body")
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return apperrors.WrapInvalidInput(httpapi.ParseBindError(err))
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return apperrors.WrapInvalidInput("request body must contain one JSON value")
	}

	if binding.Validator != nil {
		if err := binding.Validator.ValidateStruct(dst); err != nil {
			return apperrors.WrapInvalidInput(httpapi.ParseBindError(err))
		}
	}
	return nil
}
