package staff

import (
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// shiftOperationKeyLength は UUID の正規表記（8-4-4-4-12）の文字数。
const shiftOperationKeyLength = 36

// parseShiftOperationKey は operation_key（api.yaml: format uuid）の形式だけを検証し、
// 小文字の正規表記へ揃えた値を返す。
//
// 冪等再送は未実装（別チケットで追跡）。shift の operation key を保存する先が 001_init.sql に無く、
// 保存先の追加は未承認のため、現状は形式検証だけを行う。したがって同 key・同内容の確定済み操作への
// receipt 再返却、同 key で内容が異なる要求の 409 operation_key_reused、結果照会
// （GET /shifts/operations/{operation_key}）は未提供で、同じ key を再送しても重複効果は防げない。
// 冪等再送を実装する際はこの関数を差し替える。
func parseShiftOperationKey(raw string) (string, error) {
	if len(raw) != shiftOperationKeyLength {
		return "", apperrors.WrapInvalidInput("operation_key must be a UUID")
	}
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return "", apperrors.WrapInvalidInput("operation_key must be a UUID")
	}
	return parsed.String(), nil
}
