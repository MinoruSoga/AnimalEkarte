package handler

import (
	"fmt"
	"slices"
)

// validateEnum はstring値vが許可されたenum値のいずれかであることを検証する。
// 有効な場合は型付きenum値とnilエラーを返す。無効な場合はゼロ値とエラーを返す。
func validateEnum[T ~string](v string, allowed ...T) (T, error) {
	if slices.Contains(allowed, T(v)) {
		return T(v), nil
	}
	var zero T
	return zero, fmt.Errorf("invalid value %q", v)
}
