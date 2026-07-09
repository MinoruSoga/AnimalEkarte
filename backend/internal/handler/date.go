package handler

import (
	"errors"
	"fmt"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// errFlexibleDateParse は parseFlexibleDate 内部の sentinel error。
// Go 標準ライブラリの parse エラー文字列を呼び出し元に漏洩させないために使用する。
var errFlexibleDateParse = errors.New("flexible date parse failed")

// parseFlexibleDate は YYYY-MM-DD（time.Local）または RFC3339 形式の日付文字列を
// time.Time に変換する共通コア。jsonDate.UnmarshalJSON と parseDate の両方から使用される。
func parseFlexibleDate(s string) (time.Time, error) {
	// YYYY-MM-DD を優先
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return t, nil
	}
	// フォールバック: RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, errFlexibleDateParse
}

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
	t, err := parseFlexibleDate(s)
	if err != nil {
		// Go内部のエラー文字列を漏洩させないため、汎用メッセージを返す
		return apperrors.WrapInvalidInput("日付の形式が正しくありません（YYYY-MM-DD または RFC3339 形式を使用してください）")
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

// parseDate は複数フォーマット（YYYY-MM-DD, RFC3339）に対応した日付パーサー。
// Gin の JSON バインダーが RFC3339 のみを期待するため、
// フロントエンドからの「YYYY-MM-DD」形式を処理するために使用。
func parseDate(dateStr *string) (*time.Time, error) {
	if dateStr == nil {
		return nil, nil
	}

	t, err := parseFlexibleDate(*dateStr)
	if err != nil {
		// 呼び出し側の既存契約を維持するため、入力値をエコーする
		return nil, fmt.Errorf("invalid date format: %s", *dateStr)
	}
	return &t, nil
}
