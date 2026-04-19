package handler

// mapSlice は []M をポインタ経由で変換し []R を返す汎用ヘルパー。
// toXxxResponseList の重複実装をまとめるために使用する。
func mapSlice[M, R any](items []M, f func(*M) R) []R {
	result := make([]R, 0, len(items))
	for i := range items {
		result = append(result, f(&items[i]))
	}
	return result
}

// reorderRequest は全ドメイン共通の Reorder リクエスト struct。
// 各 *_request.go の重複定義をこの1つに集約する。
type reorderRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}

// nilIfEmpty は空文字列を nil ポインタに変換する汎用ヘルパー。
// オプション文字列フィールドを *string に変換する際の定型コードを排除するために使用する。
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
