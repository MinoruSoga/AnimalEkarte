// Package timeutil はローカルタイムゾーンの時刻フォーマットヘルパーを提供する。
package timeutil

import "time"

// LocalRFC3339 はローカルタイムゾーンの RFC3339 文字列を返す。
// レスポンス DTO が string 型のタイムスタンプフィールドを持つ場合に使う
// （time.Time 型フィールドは JSON 自動 marshal を使う）。
// BE-refactor.md C-1: t.In(time.Local).Format(time.RFC3339) のインライン実装を集約する。
func LocalRFC3339(t time.Time) string {
	return t.In(time.Local).Format(time.RFC3339)
}
