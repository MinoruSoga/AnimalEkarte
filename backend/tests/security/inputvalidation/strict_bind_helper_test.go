// Package inputvalidation verifies request binding with memory-only tests.
// This helper is a test-only comparison decoder; no handler uses it.
package inputvalidation

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

// BindStrictJSON is a test-only comparison to c.ShouldBindJSON.
// It demonstrates strict decoding without changing any write endpoint.
//
// U03/C05 evidence (docs/work/emr-332-security/input-output/c05-input-validation.md):
// neither ShouldBindJSON nor the per-package bindStaffJSON / bindAuthJSON
// wrappers reject unknown request fields, because Gin's JSON decoder never
// enables Decoder.DisallowUnknownFields (EnableDecoderDisallowUnknownFields is
// unset). Unknown/disallowed fields are silently dropped on every JSON write
// path that uses them.
//
// This helper shares the core boundary contract of the existing per-package
// bind helpers — bounded read to EOF, exactly one JSON value, `binding` tag
// validation via Gin's validator engine, sanitized error messages via
// httpapi.ParseBindError — and additionally rejects unknown fields. It decodes
// into dst directly from the bounded buffer so the strict decoder applies;
// `binding` tags are still enforced afterwards through binding.Validator, so
// ParseBindError produces messages identical to the ShouldBindJSON path.
//
// Contract difference: unlike bindStaffJSON / bindAuthJSON it does NOT run
// per-type byte-length checks (validateStaffJSONByteLengths /
// validateAuthJSONByteLengths — e.g. email <=254 bytes, password <=72 bytes,
// reset token <=128 bytes). Callers migrating from those wrappers must keep
// or port the equivalent checks; this helper is not a drop-in equivalent.
//
// This helper is compiled only during tests. Any production adoption requires
// a separate endpoint compatibility decision and the per-type checks above.
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
