package sharedkernel

import (
	"errors"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ErrStoredValueMissing は NULL 許容列の保存値が未記録（NULL）なのに、呼び出し側が値を必要とする場合のセンチネル
// （EMR-549 C3）。NULL を既定値で埋めて処理を続けてはならない。判別は errors.Is で行い、業務判定は拒否側へ倒す。
var ErrStoredValueMissing = errors.New("stored value is not recorded")

// RequireStored は NULL 許容列を値として必要とする箇所で使う。nil（未記録）は column 名付きのエラーにする。
func RequireStored[T any](value *T, column string) (T, error) {
	if value == nil {
		var zero T
		return zero, apperrors.Wrap(ErrStoredValueMissing, column)
	}
	return *value, nil
}

// Ptr は value のコピーを指すポインタを返す。NULL 許容列へ明示値を書く writer 用。
func Ptr[T any](value T) *T {
	return &value
}

// PtrUnlessZero は value がゼロ値（未指定）なら nil を返す。nil を書くと GORM は DB default を適用する
// （旧 non-pointer 列のゼロ値省略と同じ挙動）。ゼロ値が正当な明示値になる列には使わない。
func PtrUnlessZero[T comparable](value T) *T {
	var zero T
	if value == zero {
		return nil
	}
	return &value
}

// PtrEquals は保存値が want と一致する場合だけ true を返す。nil（未記録）は常に false。
// 「一致しない」側が安全な判定（取消済みでない・仮予約でない・受付済みでない等）にだけ使う。
// 一致しないことが許可側に倒れる判定は RequireStored でエラーにすること。
func PtrEquals[T comparable](value *T, want T) bool {
	return value != nil && *value == want
}

// TextOrEmpty は NULL 許容の文字列系 enum を応答用の文字列へ写す。nil（未記録）は空文字（有効な enum 値ではない unknown）。
// pending 等の既定値で補完しない。
func TextOrEmpty[T ~string](value *T) string {
	if value == nil {
		return ""
	}
	return string(*value)
}

// SortOrderValue は NULL 許容の sort_order を表示用の int へ写す。
// 表示専用で NULL は 0 とする（EMR-549 C3 の唯一の既定値例外。staff.StaffSortOrderValue と同じ規則）。
func SortOrderValue(sortOrder *int) int {
	if sortOrder == nil {
		return 0
	}
	return *sortOrder
}
