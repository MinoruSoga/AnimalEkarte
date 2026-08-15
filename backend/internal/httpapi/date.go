// date.go — HTTP 境界の柔軟な日付パース（BE9-2C R①: internal/handler/date.go から昇格。
// owner/pet/lstep/inventory/reservation の恒久ドメイン跨ぎ。handler 側は互換 delegate 残置）。
package httpapi

import (
	"errors"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// errFlexibleDateParse は ParseFlexibleDate 内部の sentinel error。
// Go 標準ライブラリの parse エラー文字列を呼び出し元に漏洩させないために使用する。
var errFlexibleDateParse = errors.New("flexible date parse failed")

// FlexibleDateInvalidInputMsg は日付パース失敗時にクライアントへ返す汎用メッセージ。
const FlexibleDateInvalidInputMsg = "日付の形式が正しくありません（YYYY-MM-DD または RFC3339 形式を使用してください）"

// ParseFlexibleDate は YYYY-MM-DD（time.Local）または RFC3339 形式の日付文字列を time.Time に変換する。
func ParseFlexibleDate(s string) (time.Time, error) {
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

// ParseDate は複数フォーマット（YYYY-MM-DD, RFC3339）に対応した日付パーサー。
func ParseDate(dateStr *string) (*time.Time, error) {
	if dateStr == nil {
		return nil, nil
	}
	t, err := ParseFlexibleDate(*dateStr)
	if err != nil {
		// Go内部のエラー文字列・入力値を漏洩させないため、汎用メッセージを返す
		return nil, apperrors.WrapInvalidInput(FlexibleDateInvalidInputMsg)
	}
	return &t, nil
}
