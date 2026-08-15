package trimming

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

func parseOptionalDateQueryFilter(value, field string) (*string, error) {
	if value == "" {
		return nil, nil
	}
	if _, err := time.ParseInLocation(time.DateOnly, value, time.Local); err != nil {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("%s は YYYY-MM-DD 形式で入力してください", field))
	}
	return &value, nil
}
