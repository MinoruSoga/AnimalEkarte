package handler

import "time"

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
// C-6: t.In(time.Local).Format(time.RFC3339) のインライン再実装30箇所を集約する。
func localTimeRFC3339(t time.Time) string {
	return t.In(time.Local).Format(time.RFC3339)
}
