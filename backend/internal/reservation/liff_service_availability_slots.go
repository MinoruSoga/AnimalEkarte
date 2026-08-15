package reservation

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
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
		slog.ErrorContext(ctx, "failed to get day reservations", "error", err)
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
		slog.ErrorContext(ctx, "failed to prefetch day reservations for available dates", "error", err)
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

	inputs := make([]StaffSlotInput, 0, len(staffs))
	for i := range staffs {
		si := StaffSlotInput{StaffID: staffs[i].ID}

		if entry, ok := staffEntries[staffs[i].ID]; ok && entry != nil {
			breaks := breaksByEntry[entry.ID]
			override := &StaffScheduleOverride{
				ShiftType: string(entry.ShiftType),
				WorkStart: ptrTimeToHHMM(entry.StartTime), // PostgreSQL "HH:MM:SS" → "HHMM"
				WorkEnd:   ptrTimeToHHMM(entry.EndTime),   // PostgreSQL "HH:MM:SS" → "HHMM"
			}
			for _, b := range breaks {
				override.Breaks = append(override.Breaks, BreakPeriod{
					Start: timeToHHMM(b.BreakStart), // PostgreSQL "HH:MM:SS" → "HHMM"
					End:   timeToHHMM(b.BreakEnd),   // PostgreSQL "HH:MM:SS" → "HHMM"
				})
			}
			si.ScheduleOverride = override
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
