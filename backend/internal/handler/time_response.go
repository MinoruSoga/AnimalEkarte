package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/timeutil"
)

func localTime(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.In(time.Local)
}

func localTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := localTime(*t)
	return &v
}

// localTimeRFC3339 はローカルタイムゾーンの RFC3339 文字列を返す。
// レスポンス DTO が string 型のタイムスタンプフィールドを持つ場合に使う
// （time.Time 型フィールドは localTime/localTimePtr + JSON 自動 marshal を使う）。
// C-6: ローカルタイムゾーンの RFC3339 フォーマットのインライン再実装30箇所を集約する。
// BE-refactor.md C-1: 本体は timeutil.LocalRFC3339 の1行ラッパ（handler 30箇所の再置換はしない）。
func localTimeRFC3339(t time.Time) string {
	return timeutil.LocalRFC3339(t)
}
