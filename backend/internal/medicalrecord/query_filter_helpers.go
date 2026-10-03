package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

func optionalStringQueryFilter(value string) *string {
	return httpapi.OptionalString(value)
}

func parseOptionalUintQueryFilter(value, field string) (*uint64, error) {
	return httpapi.ParseOptionalUint64Field(value, field)
}

func parseOptionalDateQueryFilter(value, field string) (*string, error) {
	return httpapi.ParseOptionalDateOnlyField(value, field)
}

func nilIfEmpty(s string) *string {
	return httpapi.OptionalString(s)
}

// maxColumnTextFilterLen は表示列フィルタ（owner_name / pet_name / chief_complaint）
// の入力上限。owner 一覧の search（255 byte）と同じ上限で、無制限の検索語が
// folded GIN 前方一致走査を長時間占有するのを防ぐ。
const maxColumnTextFilterLen = 255

// validateColumnTextQueryFilter は表示列の部分一致フィルタ値を長さ検証する。
// 値はそのまま返し、正規化・エスケープはリポジトリ側の textsearch に任せる。
func validateColumnTextQueryFilter(value, field string) (string, error) {
	if len(value) > maxColumnTextFilterLen {
		return "", apperrors.WrapInvalidInput(field + " must be at most 255 characters")
	}
	return value, nil
}
