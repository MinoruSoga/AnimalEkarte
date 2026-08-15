// day_schedule.go — 締め時間スケジュール DTO（BE9-2C B⑤: service/closing_settings_service.go から
// 昇格。clinic（closing settings）と billing（cash_register）の恒久ドメイン跨ぎ・純データ）。
package sharedkernel

// DaySchedule は指定日の締め時間スケジュール。
type DaySchedule struct {
	AmPmBoundary string `json:"am_pm_boundary"`
	PmEnd        string `json:"pm_end"`
	// AmStart は AM 開始時刻（#215）。空の場合は既定 09:00 として扱う（resolvePeriodRange 側でフォールバック）。
	AmStart   string `json:"am_start"`
	IsHoliday bool   `json:"is_holiday"`
}
