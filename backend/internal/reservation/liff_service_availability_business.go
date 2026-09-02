package reservation

// BE9-2C R③: liff系(R⑤)から前倒し移動（timeslot_engineのBusinessHours/BreakPeriod型と同居が自然な葉helper）。

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// parseBusinessHoursForDate はパッケージ内共通のヘルパー（liffService・validator から呼ばれる）。
//
// 戻り値の error は break_hours の JSON unmarshal 失敗時のみ非 nil になる（BE-refactor.md D10/F-2）。
// business_hours の unmarshal 失敗は運用上のデフォルト値（09:00-19:00）へのフォールバックを維持する
// （既存挙動・本関数の責務外・error 対象外）。
//
// 呼び出し元の使い分け:
//   - 予約作成の検証パス（reservation_validators.go の validateBusinessRules）: エラーを
//     fail-closed で伝播する。休憩時間との重複を検証できない状態で予約を通すと、
//     休憩時間帯の予約が誤って確定しうる（従来は breaks=nil にフォールバックし検証がスキップされていた）。
//   - 空き時間一覧表示パス（liff_service_availability.go）: エラーを無視し breaks=nil で継続する
//     既存挙動を維持する（意図的・スコープ外。表示系の劣化は LIFF-1/LIFF-2 で別途観測性を確保済み）。
func ParseBusinessHoursForDate(ctx context.Context, setting *model.LineReservationSetting, date time.Time) (BusinessHours, []BreakPeriod, error) {
	var bh BusinessHours
	if err := json.Unmarshal(setting.BusinessHours, &bh); err != nil {
		bh = BusinessHours{Start: "0900", End: "1900"}
	}
	var breaks []BreakPeriod
	var breaksErr error
	if err := json.Unmarshal(setting.BreakHours, &breaks); err != nil {
		breaks = nil
		breaksErr = apperrors.Wrap(err, "invalid break_hours json")
	}

	// 曜日別営業時間があれば上書き（例: 土曜だけ短縮営業）
	if len(setting.BusinessHoursByWeekday) > 0 {
		bh = overlayWeekdayHours(ctx, setting, date, bh)
	}

	return bh, breaks, breaksErr
}

func overlayWeekdayHours(
	ctx context.Context,
	setting *model.LineReservationSetting,
	date time.Time,
	bh BusinessHours,
) BusinessHours {
	var byWeekday map[string]BusinessHours
	if err := json.Unmarshal(setting.BusinessHoursByWeekday, &byWeekday); err != nil {
		// A-3: fail-closed にはしない（表示系で、デフォルト営業時間へのフォールバックが妥当）。
		slog.WarnContext(ctx, "invalid business_hours_by_weekday json; falling back to default hours", "error", err, "clinic_id", setting.ClinicID)
		return bh
	}
	key := strconv.Itoa(int(date.In(config.JST).Weekday()))
	if wdBH, ok := byWeekday[key]; ok {
		return wdBH
	}
	return bh
}
