package handler

import (
	"fmt"
	"time"
)

// parseDate は複数フォーマット（YYYY-MM-DD, RFC3339）に対応した日付パーサー。
// Gin の JSON バインダーが RFC3339 のみを期待するため、
// フロントエンドからの「YYYY-MM-DD」形式を処理するために使用。
func parseDate(dateStr *string) (*time.Time, error) {
	if dateStr == nil {
		return nil, nil
	}

	// フォーマット1: YYYY-MM-DD
	t, err := time.ParseInLocation("2006-01-02", *dateStr, time.Local)
	if err == nil {
		return &t, nil
	}

	// フォーマット2: RFC3339
	t, err = time.Parse(time.RFC3339, *dateStr)
	if err == nil {
		return &t, nil
	}

	return nil, fmt.Errorf("invalid date format: %s", *dateStr)
}
