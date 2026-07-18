package handler

import (
	"errors"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// errFlexibleDateParse は parseFlexibleDate 内部の sentinel error。
// Go 標準ライブラリの parse エラー文字列を呼び出し元に漏洩させないために使用する。
var errFlexibleDateParse = errors.New("flexible date parse failed")

// flexibleDateInvalidInputMsg は日付パース失敗時にクライアントへ返す汎用メッセージ。
// 不正な入力値そのもの（Go内部のparseエラー文字列や生の入力）を露出させないため、
// jsonDate.UnmarshalJSON と parseDate の両方がこの定数を共有する。
const flexibleDateInvalidInputMsg = "日付の形式が正しくありません（YYYY-MM-DD または RFC3339 形式を使用してください）"

// parseFlexibleDate は YYYY-MM-DD（time.Local）または RFC3339 形式の日付文字列を
// time.Time に変換する共通コア。jsonDate.UnmarshalJSON と parseDate の両方から使用される。
func parseFlexibleDate(s string) (time.Time, error) {
	// YYYY-MM-DD を優先
	if t, err := time.ParseInLocation(time.DateOnly, s, time.Local); err == nil {
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
		return apperrors.WrapInvalidInput(flexibleDateInvalidInputMsg)
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
		// Go内部のエラー文字列・入力値を漏洩させないため、jsonDate と同一の汎用メッセージを返す
		return nil, apperrors.WrapInvalidInput(flexibleDateInvalidInputMsg)
	}
	return &t, nil
}
