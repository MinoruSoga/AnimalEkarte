package httpapi

import (
	"fmt"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// OptionalString returns nil for empty input and a pointer otherwise.
func OptionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// ParseOptionalUint64Field parses an optional uint64 from a query/form string.
func ParseOptionalUint64Field(value, field string) (*uint64, error) {
	if value == "" {
		return nil, nil
	}
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid " + field)
	}
	return &id, nil
}

// ParseOptionalDateOnlyField accepts empty or YYYY-MM-DD and returns the original string.
func ParseOptionalDateOnlyField(value, field string) (*string, error) {
	if value == "" {
		return nil, nil
	}
	if _, err := time.ParseInLocation(time.DateOnly, value, time.Local); err != nil {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("%s は YYYY-MM-DD 形式で入力してください", field))
	}
	return &value, nil
}

// ParseOptionalDateTimeField accepts empty or YYYY-MM-DD and returns midnight in Local.
func ParseOptionalDateTimeField(value, field string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation(time.DateOnly, value, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("%s は YYYY-MM-DD 形式で入力してください", field))
	}
	return &parsed, nil
}
