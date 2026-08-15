package httpapi

import (
	"time"

	"github.com/animal-ekarte/backend/internal/timeutil"
)

// LocalTime converts t to the server local timezone (Asia/Tokyo). Zero values pass through
// unchanged so JSON omitempty/zero-check callers keep their existing behavior.
func LocalTime(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.In(time.Local)
}

// LocalTimePtr is the *time.Time variant of LocalTime; nil passes through unchanged.
func LocalTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := LocalTime(*t)
	return &v
}

// LocalTimeRFC3339 はローカルタイムゾーンの RFC3339 文字列を返す。
// レスポンス DTO が string 型のタイムスタンプフィールドを持つ場合に使う
// （time.Time 型フィールドは LocalTime/LocalTimePtr + JSON 自動 marshal を使う）。
// C-6: ローカルタイムゾーンの RFC3339 フォーマットのインライン再実装30箇所を集約する。
// BE-refactor.md C-1: 本体は timeutil.LocalRFC3339 の1行ラッパ（handler 30箇所の再置換はしない）。
func LocalTimeRFC3339(t time.Time) string {
	return timeutil.LocalRFC3339(t)
}
