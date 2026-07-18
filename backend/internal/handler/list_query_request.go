package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

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

func parseRequiredUintQueryFilter(value, field string) (uint64, error) {
	id, err := parseOptionalUintQueryFilter(value, field)
	if err != nil || id == nil || *id == 0 {
		return 0, apperrors.WrapInvalidInput("invalid " + field)
	}
	return *id, nil
}

func parseOptionalUintQueryValue(value string) uint64 {
	if value == "" {
		return 0
	}
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func parseOptionalDateQueryFilter(value, field string) (*string, error) {
	if value == "" {
		return nil, nil
	}
	if _, err := time.ParseInLocation(time.DateOnly, value, time.Local); err != nil {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("%s は YYYY-MM-DD 形式で入力してください", field))
	}
	return &value, nil
}

func optionalStringQueryFilter(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func parseOptionalDateTimeQueryFilter(value, field string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation(time.DateOnly, value, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("%s は YYYY-MM-DD 形式で入力してください", field))
	}
	return &parsed, nil
}
