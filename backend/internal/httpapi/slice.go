package httpapi

import (
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// MaxReorderIDs is the hard upper bound on reorder ID array cardinality.
// Unbounded arrays fan out one UPDATE per id (ReorderByClinicID) and can DoS the API.
const MaxReorderIDs = 500

// MapSlice は []M をポインタ経由で変換し []R を返す汎用ヘルパー。
// toXxxResponseList の重複実装をまとめるために使用する。
func MapSlice[M, R any](items []M, f func(*M) R) []R {
	result := make([]R, 0, len(items))
	for i := range items {
		result = append(result, f(&items[i]))
	}
	return result
}

// ReorderRequest は全ドメイン共通の Reorder リクエスト struct。
// 各 *_request.go の重複定義をこの1つに集約する。
// binding max mirrors MaxReorderIDs (gin tag requires a literal).
type ReorderRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1,max=500"`
}

// ValidateReorderIDs rejects empty, over-limit, and duplicate reorder ID lists
// before any repository fan-out UPDATE (SEC-CS-F04).
func ValidateReorderIDs(ids []uint64) error {
	n := len(ids)
	if n == 0 {
		return apperrors.WrapInvalidInput("並び順のIDリストが空です")
	}
	if n > MaxReorderIDs {
		return apperrors.WrapInvalidInput(
			fmt.Sprintf("並び順のIDは%d件以内で指定してください", MaxReorderIDs),
		)
	}
	seen := make(map[uint64]struct{}, n)
	for _, id := range ids {
		if _, dup := seen[id]; dup {
			return apperrors.WrapInvalidInput("並び順のIDリストに重複があります")
		}
		seen[id] = struct{}{}
	}
	return nil
}

// NilIfEmpty は空文字列を nil ポインタに変換する汎用ヘルパー。
// オプション文字列フィールドを *string に変換する際の定型コードを排除するために使用する。
func NilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
