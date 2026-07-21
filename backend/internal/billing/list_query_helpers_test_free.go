package billing

// list_query_helpers_test_free.go — internal/handler/list_query_request.go の同名ヘルパーの複製
// （残留consumer多数のため原本残置・reservation 先例・B④〜⑥で統合判断）。

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

func optionalStringQueryFilter(value string) *string {
	if value == "" {
		return nil
	}
	return &value
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
