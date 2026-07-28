package medicalrecord

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// query_filter_helpers.go — documented local copies (BE9-2D Batch C) of the small,
// string-based list-query filter helpers from internal/handler
// (list_query_request.go / slice_helpers.go). The checkup / vaccine / vaccination request
// binders moved into this package build their service filters from url.Values-derived
// strings via these helpers. The originals stay in internal/handler (still used by ~a dozen
// not-yet-migrated handlers: accounting/estimate/examination/hospitalization/medical_record/
// pet/shift/trimming), and medicalrecord cannot import internal/handler (import cycle —
// internal/handler already imports internal/medicalrecord). Behavior-identical unexported
// copies (same error messages), following the validators.go local-copy precedent.
//
// Context-based query parsing (e.g. checkup_field_handler.go's pet_id) uses
// httpapi.ParseOptionalUint64Query directly — only these string/struct-field-based variants
// need copies here.

// optionalStringQueryFilter mirrors internal/handler.optionalStringQueryFilter.
func optionalStringQueryFilter(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// parseOptionalUintQueryFilter mirrors internal/handler.parseOptionalUintQueryFilter.
func parseOptionalUintQueryFilter(value, field string) (*uint64, error) {
	if value == "" {
		return nil, nil
	}
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid " + field)
	}
	return &id, nil
}

// parseOptionalDateQueryFilter mirrors internal/handler.parseOptionalDateQueryFilter.
func parseOptionalDateQueryFilter(value, field string) (*string, error) {
	if value == "" {
		return nil, nil
	}
	if _, err := time.ParseInLocation(time.DateOnly, value, time.Local); err != nil {
		return nil, apperrors.WrapInvalidInput(field + " は YYYY-MM-DD 形式で入力してください")
	}
	return &value, nil
}

// nilIfEmpty mirrors internal/handler.nilIfEmpty (slice_helpers.go).
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
