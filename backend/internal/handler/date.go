package handler

import (
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// flexibleDateInvalidInputMsg は日付パース失敗時にクライアントへ返す汎用メッセージ
// （BE9-2C R①: 実装は httpapi へ昇格・本定数と以下の関数は既存呼び出し面互換の delegate）。
const flexibleDateInvalidInputMsg = httpapi.FlexibleDateInvalidInputMsg

// parseFlexibleDate は httpapi.ParseFlexibleDate への delegate。
func parseFlexibleDate(s string) (time.Time, error) {
	return httpapi.ParseFlexibleDate(s)
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
	return httpapi.ParseDate(dateStr)
}
