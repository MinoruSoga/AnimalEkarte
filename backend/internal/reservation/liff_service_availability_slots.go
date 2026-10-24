package reservation

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// buildStaffSlotInputsForDate はスタッフ一覧と指定日から StaffSlotInputs を構築する（単日パス
// GetAvailableTimes 専用）。PERF-FOLLOWUP-08: 旧 buildStaffSlotInputs は staff ごとに
// FindAllByDate/FindAllBreaksByEntryID を再取得する N+1 だった。開始日=終了日=当日の1日ウィンドウで
// buildAvailableDatesStaffInputsFn（複数日パス, G7-1）と同型のプリフェッチを行い、
// buildStaffSlotInputsFromWindow に委譲することで解消する。
// 致命度は旧実装を踏襲: 予約(dayResv)prefetch 失敗はエラー返し、シフト/休憩 prefetch 失敗は
// slog.ErrorContext + 続行（override なし）。
// 複数日パスの buildAvailableDatesStaffInputsFn とプリフェッチロジックが重複するが、多日ウィンドウ
// （予約受付期間全体）と単日ウィンドウでは対象範囲もエラーメッセージも異なるため、共通化は
// 本タスクのスコープ外（変更を最小に保つ）。
func (s *liffService) buildStaffSlotInputsForDate(ctx context.Context, clinicID uint64, staffs []model.Staff, date time.Time) ([]StaffSlotInput, error) {
	// 旧 FindAllByDay と同様、JST の日付境界（00:00〜翌日00:00）へ明示的に truncate する。
	// date が既に JST 深夜0時である不変条件（handler の time.ParseInLocation 経由）に依存せず、
	// この関数単体でも正しい1日ウィンドウになるようにする（go-reviewer 指摘・防御的頑健性）。
	dateJST := date.In(config.JST)
	dayStart := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0, 0, 0, config.JST)
	windowEnd := dayStart.AddDate(0, 0, 1)

	staffIDs := make([]uint64, len(staffs))
	for i := range staffs {
		staffIDs[i] = staffs[i].ID
	}

	entriesByDateStaff := map[string]map[uint64]*model.ShiftEntry{}
	breaksByEntry := map[uint64][]model.ShiftEntryBreak{}
	entries, err := s.scheduleRepo.FindAllByStaffIDsAndDateRange(ctx, clinicID, staffIDs, dayStart, windowEnd)
	if err != nil {
		// LIFF-2 と同様: 取得失敗は観測性のため slog 記録。挙動は従来通り（override なしで続行）。
		slog.ErrorContext(ctx, "failed to prefetch schedule entries for available times", "error", err)
	} else {
		entryIDs := make([]uint64, 0, len(entries))
		for i := range entries {
			dateStr := entries[i].Date.Format(time.DateOnly)
			if entriesByDateStaff[dateStr] == nil {
				entriesByDateStaff[dateStr] = map[uint64]*model.ShiftEntry{}
			}
			entriesByDateStaff[dateStr][entries[i].StaffID] = &entries[i]
			entryIDs = append(entryIDs, entries[i].ID)
		}
		breaks, brkErr := s.scheduleRepo.FindAllBreaksByEntryIDs(ctx, clinicID, entryIDs)
		if brkErr != nil {
			// LIFF-1 と同様: 取得失敗は観測性のため slog 記録。挙動は従来通り（breaks 空で続行）。
			slog.ErrorContext(ctx, "failed to prefetch schedule breaks for available times", "error", brkErr)
		} else {
			breaksByEntry = breaks
		}
	}

	// 当日の全予約を一括取得（N+1回避。旧 FindAllByDay の6 Preload を伴わない軽量版）
	reservations, err := s.adminRepo.FindTimeRangesByDateRange(ctx, clinicID, dayStart, windowEnd)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get day reservations")
	}
	reservationsByDate := map[string][]model.Reservation{}
	for i := range reservations {
		dateStr := reservations[i].StartTime.In(config.JST).Format(time.DateOnly)
		reservationsByDate[dateStr] = append(reservationsByDate[dateStr], reservations[i])
	}

	return buildStaffSlotInputsFromWindow(staffs, dayStart, entriesByDateStaff, breaksByEntry, reservationsByDate), nil
}

// buildAvailableDatesStaffInputsFn は GetAvailableDates の予約受付期間全体ぶんのシフト/休憩/当日予約を
// まとめて3クエリでプリフェッチし、日付ごとの StaffInputsFn クロージャ（DB問い合わせなし）を返す(G7-1)。
//
// エラー処理は buildStaffSlotInputsForDate（旧 buildStaffSlotInputs）の per-day 実装と同じ致命度を踏襲する:
//   - シフトエントリ・休憩の取得失敗は非致命的（LIFF-1/LIFF-2 と同様 slog 記録のみ。override なしで続行）。
//   - 当日予約(reservations)の取得失敗は致命的（元実装が dayResv 取得失敗で即エラーを返していたことに対応）。
//
// 影響範囲の変化に注意: 旧実装はスタッフ×日単位でシフト/休憩を取得していたため、1件の失敗は
// 「その日・そのスタッフのみ override なし」に縮退した。本実装は期間全体を1クエリでまとめて取得するため、
// 失敗時は期間内の全日・全スタッフが override なし（＝シフト制約を無視して空き扱い）に縮退する
// （blast radius が per-staff-day → window-wide に拡大）。挙動保存(致命度不変)の範囲内の意図的トレードオフだが、
// 患者向け予約カレンダーで「本来休みのスタッフが空きに見える」広域劣化に繋がりうるため、この slog.ErrorContext
// を監視・アラート対象に含めることを推奨する。
func (s *liffService) buildAvailableDatesStaffInputsFn(
	ctx context.Context, clinicID uint64, staffs []model.Staff, settings AvailableDatesSettings,
) (func(context.Context, time.Time, uint64, uint64) ([]StaffSlotInput, error), error) {
	minDate, maxDate := BookingWindowDates(settings)
	windowEnd := maxDate.AddDate(0, 0, 1) // 半開区間の排他的上限

	staffIDs := make([]uint64, len(staffs))
	for i := range staffs {
		staffIDs[i] = staffs[i].ID
	}

	entriesByDateStaff := map[string]map[uint64]*model.ShiftEntry{}
	breaksByEntry := map[uint64][]model.ShiftEntryBreak{}
	entries, err := s.scheduleRepo.FindAllByStaffIDsAndDateRange(ctx, clinicID, staffIDs, minDate, windowEnd)
	if err != nil {
		// LIFF-2 と同様: 取得失敗は観測性のため slog 記録。挙動は従来通り（override なしで続行）。
		slog.ErrorContext(ctx, "failed to prefetch schedule entries for available dates", "error", err)
	} else {
		entryIDs := make([]uint64, 0, len(entries))
		for i := range entries {
			dateStr := entries[i].Date.Format(time.DateOnly)
			if entriesByDateStaff[dateStr] == nil {
				entriesByDateStaff[dateStr] = map[uint64]*model.ShiftEntry{}
			}
			entriesByDateStaff[dateStr][entries[i].StaffID] = &entries[i]
			entryIDs = append(entryIDs, entries[i].ID)
		}
		breaks, brkErr := s.scheduleRepo.FindAllBreaksByEntryIDs(ctx, clinicID, entryIDs)
		if brkErr != nil {
			// LIFF-1 と同様: 取得失敗は観測性のため slog 記録。挙動は従来通り（breaks 空で続行）。
			slog.ErrorContext(ctx, "failed to prefetch schedule breaks for available dates", "error", brkErr)
		} else {
			breaksByEntry = breaks
		}
	}

	reservations, err := s.adminRepo.FindTimeRangesByDateRange(ctx, clinicID, minDate, windowEnd)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get day reservations")
	}
	reservationsByDate := map[string][]model.Reservation{}
	for i := range reservations {
		dateStr := reservations[i].StartTime.In(config.JST).Format(time.DateOnly)
		reservationsByDate[dateStr] = append(reservationsByDate[dateStr], reservations[i])
	}

	return func(_ context.Context, date time.Time, _ uint64, _ uint64) ([]StaffSlotInput, error) {
		return buildStaffSlotInputsFromWindow(staffs, date, entriesByDateStaff, breaksByEntry, reservationsByDate), nil
	}, nil
}

// buildStaffSlotInputsFromWindow は StaffSlotInput 構築ロジックを、事前取得済みの期間データ
// （シフト・休憩・当日予約）から DB 問い合わせなしで行う（G7-1: 日付ループ N+1 解消）。
// GetAvailableDates の複数日プレビュー（buildAvailableDatesStaffInputsFn 経由）と、
// GetAvailableTimes の単日パス（buildStaffSlotInputsForDate 経由、PERF-FOLLOWUP-08 で1日ウィンドウに
// 一本化）の両方から呼ばれる共通ロジック。
func buildStaffSlotInputsFromWindow(
	staffs []model.Staff,
	date time.Time,
	entriesByDateStaff map[string]map[uint64]*model.ShiftEntry,
	breaksByEntry map[uint64][]model.ShiftEntryBreak,
	reservationsByDate map[string][]model.Reservation,
) []StaffSlotInput {
	dateStr := date.Format(time.DateOnly)
	dayResv := reservationsByDate[dateStr]
	staffEntries := entriesByDateStaff[dateStr]
	// 評価日（JST）の 00:00〜翌 00:00。勤務・休憩の明示日時（starts_at / ends_at）をこの窓へ切り出して
	// "HHMM" にする。時刻文字列の勤務日基準コンバータは使わない。
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, config.JST)
	dayEnd := dayStart.AddDate(0, 0, 1)

	inputs := make([]StaffSlotInput, 0, len(staffs))
	for i := range staffs {
		si := StaffSlotInput{StaffID: staffs[i].ID}

		if entry, ok := staffEntries[staffs[i].ID]; ok && entry != nil {
			breaks := breaksByEntry[entry.ID]
			// interval_state=unknown（勤務の starts_at / ends_at が NULL、いずれかの休憩が区間不明、または
			// off / paid_leave に矛盾する日時・休憩が残る行）は勤務可能と扱わない（api.yaml ShiftEntry）。
			// 営業時間や shift_type の既定時間帯へ戻さず（fail-open にせず）、その日の候補から当該スタッフを外す。
			if sharedkernel.ShiftEntryIntervalState(entry, breaks) == sharedkernel.ShiftIntervalStateUnknown {
				continue
			}
			si.ScheduleOverride = buildScheduleOverride(entry, breaks, dayStart, dayEnd)
		}

		// 当日の既存予約を絞り込み
		for j := range dayResv {
			if dayResv[j].Status == model.ReservationStatusCancelled {
				continue
			}
			if dayResv[j].DoctorID != nil && *dayResv[j].DoctorID == staffs[i].ID {
				si.ExistingResvs = append(si.ExistingResvs, ExistingReservation{
					StaffID:   staffs[i].ID,
					StartTime: dayResv[j].StartTime.Format("1504"),
					EndTime:   dayResv[j].EndTime.Format("1504"),
				})
			}
		}

		inputs = append(inputs, si)
	}
	return inputs
}

// buildScheduleOverride は interval_state=known の勤務行を、評価日 [dayStart, dayEnd) に切り出した
// ScheduleOverride へ変換する。勤務・休憩とも starts_at / ends_at の明示日時から直接 "HHMM" を求め、
// shift_type の名称や営業時間の既定時間帯は使わない。日跨ぎ勤務は当日分だけを切り出す。
// 呼出し側で勤務種別が勤務なら両日時が非 NULL であること（interval_state=known）を確認済みであること。
func buildScheduleOverride(entry *model.ShiftEntry, breaks []model.ShiftEntryBreak, dayStart, dayEnd time.Time) *StaffScheduleOverride {
	override := &StaffScheduleOverride{ShiftType: string(entry.ShiftType)}
	if !sharedkernel.RequiresTimeSlot(entry.ShiftType) || entry.StartsAt == nil || entry.EndsAt == nil {
		return override
	}
	workStart, workEnd, onDay := clipToDayHHMM(*entry.StartsAt, *entry.EndsAt, dayStart, dayEnd, roundMinuteUp, roundMinuteDown)
	if !onDay {
		// 当日に掛からない勤務行（勤務日と starts_at の暦日が食い違う旧データ等）は空の勤務窓にし、営業時間へ戻さない。
		workStart, workEnd = "0000", "0000"
	}
	override.WorkStart = &workStart
	override.WorkEnd = &workEnd
	for _, b := range breaks {
		if b.StartsAt == nil || b.EndsAt == nil {
			continue
		}
		if start, end, overlaps := clipToDayHHMM(*b.StartsAt, *b.EndsAt, dayStart, dayEnd, roundMinuteDown, roundMinuteUp); overlaps {
			override.Breaks = append(override.Breaks, BreakPeriod{Start: start, End: end})
		}
	}
	return override
}

// minuteRounding は秒・秒未満を含む時刻を "HHMM"（分単位）へ丸めるときの方向を表す。
type minuteRounding int

const (
	// roundMinuteDown は秒・秒未満を切り捨てる（ちょうどの分はそのまま）。
	roundMinuteDown minuteRounding = iota
	// roundMinuteUp は秒・秒未満があるときだけ次の分へ切り上げる（ちょうどの分はそのまま）。
	roundMinuteUp
)

// roundToMinute は t を JST の分境界へ丸める。ちょうどの分の時刻は変えない。
func roundToMinute(t time.Time, mode minuteRounding) time.Time {
	j := t.In(config.JST)
	floor := time.Date(j.Year(), j.Month(), j.Day(), j.Hour(), j.Minute(), 0, 0, config.JST)
	if mode == roundMinuteUp && floor.Before(j) {
		return floor.Add(time.Minute)
	}
	return floor
}

// clipToDayHHMM は [startsAt, endsAt) を評価日 [dayStart, dayEnd) へ切り出し、JST の "HHMM" の組で返す。
// 秒・秒未満を含む境界は、使える時間を広げない向きに分単位へ丸める（startRound / endRound で指定する）。
//   - 勤務: 開始は切り上げ、終了は切り捨て（勤務窓を狭める）
//   - 休憩: 開始は切り捨て、終了は切り上げ（休憩を広げる）
//
// ちょうどの分の境界は変えない。丸めた結果、切り出し後の区間が空になったときも ok=false とする。
// 当日に掛からないときは ok=false。翌 0:00 は "HHMM" で表せない（2400 は下流の解析で拒否される）ため "2359" とする。
// 末尾 1 分の丸めでコース所要時間を収める枠が増えることはない。
func clipToDayHHMM(startsAt, endsAt, dayStart, dayEnd time.Time, startRound, endRound minuteRounding) (start, end string, ok bool) {
	if !startsAt.Before(dayEnd) || !endsAt.After(dayStart) {
		return "", "", false
	}
	clippedStart := startsAt
	if clippedStart.Before(dayStart) {
		clippedStart = dayStart
	}
	roundedStart := roundToMinute(clippedStart, startRound)
	if !roundedStart.Before(dayEnd) {
		// 開始の切り上げが翌 0:00 に達した = 当日に使える区間が残らない。
		return "", "", false
	}
	start = roundedStart.In(config.JST).Format("1504")
	if !endsAt.Before(dayEnd) {
		return start, "2359", true
	}
	roundedEnd := roundToMinute(endsAt, endRound)
	if clippedStart.Before(endsAt) && !roundedStart.Before(roundedEnd) {
		// 丸めで空になった区間（例: 勤務 09:00:10〜09:00:50）は可用時間を足さない。
		return "", "", false
	}
	if !roundedEnd.Before(dayEnd) {
		// 終了の切り上げが翌 0:00 に達した場合も、既存の末尾規約どおり "2359" とする。
		return start, "2359", true
	}
	return start, roundedEnd.In(config.JST).Format("1504"), true
}
