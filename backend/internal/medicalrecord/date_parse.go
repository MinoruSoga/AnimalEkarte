package medicalrecord

import (
	"errors"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// date_parse.go — documented local copies (BE9-2D Batch C) of the flexible date parsing
// helpers from internal/handler/date.go, needed by the vaccination request binder that moved
// into this package (parseDate over YYYY-MM-DD / RFC3339). internal/handler/date.go stays
// (still used by many not-yet-migrated handlers), and medicalrecord cannot import
// internal/handler (import cycle). Behavior-identical copies (same generic error message that
// deliberately hides Go's internal parse error / raw input).
//
// The original date.go also defines the jsonDate JSON wrapper (+ jsonDatePtr); those are NOT
// copied here because no file moved in this batch (handlers/requests/responses/tests) uses
// them — copying them would introduce unused code that golangci-lint's unused check flags.
// This is a deliberate, narrower copy than the batch spec's "parseDate/parseFlexibleDate/
// jsonDate" wording (recorded as a spec deviation in the batch report).

// errFlexibleDateParse mirrors internal/handler.errFlexibleDateParse — an internal sentinel
// so Go's stdlib parse-error strings never leak to callers.
var errFlexibleDateParse = errors.New("flexible date parse failed")

// flexibleDateInvalidInputMsg mirrors internal/handler.flexibleDateInvalidInputMsg — the
// generic message returned on parse failure (never the raw input or Go's internal error).
const flexibleDateInvalidInputMsg = "日付の形式が正しくありません（YYYY-MM-DD または RFC3339 形式を使用してください）"

// parseFlexibleDate mirrors internal/handler.parseFlexibleDate: YYYY-MM-DD (time.Local)
// preferred, RFC3339 fallback.
func parseFlexibleDate(s string) (time.Time, error) {
	if t, err := time.ParseInLocation(time.DateOnly, s, time.Local); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, errFlexibleDateParse
}

// parseDate mirrors internal/handler.parseDate: nil-in → nil-out, otherwise flexible parse.
func parseDate(dateStr *string) (*time.Time, error) {
	if dateStr == nil {
		return nil, nil
	}
	t, err := parseFlexibleDate(*dateStr)
	if err != nil {
		return nil, apperrors.WrapInvalidInput(flexibleDateInvalidInputMsg)
	}
	return &t, nil
}
