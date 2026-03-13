package handler

import (
	"fmt"
	"strings"
	"time"
)

// jsonDate は YYYY-MM-DD または RFC3339 両形式を受け付ける time.Time ラッパー。
// リクエストボディの日付フィールド（birth_date, neutered_date, last_visit 等）に使用する。
type jsonDate struct {
	time.Time
}

func (d *jsonDate) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "null" || s == "" {
		return nil
	}
	// YYYY-MM-DD を優先
	if t, err := time.Parse("2006-01-02", s); err == nil {
		d.Time = t
		return nil
	}
	// フォールバック: RFC3339
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}
	d.Time = t
	return nil
}

// jsonDatePtr は *jsonDate を *time.Time に変換する。nil の場合は nil を返す。
func jsonDatePtr(d *jsonDate) *time.Time {
	if d == nil {
		return nil
	}
	t := d.Time
	return &t
}
