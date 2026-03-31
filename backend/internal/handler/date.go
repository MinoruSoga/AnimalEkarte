package handler

import (
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
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
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		d.Time = t
		return nil
	}
	// Go内部のエラー文字列を漏洩させないため、汎用メッセージを返す
	return apperrors.WrapInvalidInput("日付の形式が正しくありません（YYYY-MM-DD または RFC3339 形式を使用してください）")
}

// jsonDatePtr は *jsonDate を *time.Time に変換する。nil の場合は nil を返す。
func jsonDatePtr(d *jsonDate) *time.Time {
	if d == nil {
		return nil
	}
	t := d.Time
	return &t
}
