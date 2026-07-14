package timeutil

import "time"

var weekdaysJP = [...]string{"日", "月", "火", "水", "木", "金", "土"}

// WeekdayJP は t の曜日を日本語1文字（日〜土）で返す。
func WeekdayJP(t time.Time) string {
	return weekdaysJP[t.Weekday()]
}
