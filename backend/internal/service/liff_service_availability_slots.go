package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/config"
	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// buildStaffSlotInputs はスタッフ一覧と指定日からStaffSlotInputsを構築する。
func (s *liffService) buildStaffSlotInputs(ctx context.Context, clinicID uint64, staffs []model.Staff, date time.Time) ([]StaffSlotInput, error) {
	// 当日の全予約を一括取得（N+1回避）
	dayResv, err := s.adminRepo.FindAllByDay(ctx, clinicID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get day reservations", "error", err)
		return nil, apperrors.Wrap(err, "failed to get day reservations")
	}

	inputs := make([]StaffSlotInput, 0, len(staffs))
	for i := range staffs {
		si := StaffSlotInput{StaffID: staffs[i].ID}

		// シフトエントリを取得
		entry, err := s.scheduleRepo.FindAllByDate(ctx, clinicID, staffs[i].ID, date)
		if err != nil {
			// LIFF-2: 取得失敗は観測性のため slog 記録（P11）。挙動は従来通り（entry を使わずスキップ）。
			slog.ErrorContext(ctx, "failed to get schedule entry", "error", err, "staff_id", staffs[i].ID)
		}
		if err == nil && entry != nil {
			breaks, brkErr := s.scheduleRepo.FindAllBreaksByEntryID(ctx, entry.ID)
			if brkErr != nil {
				// LIFF-1: 休憩取得失敗は観測性のため slog 記録（P11）。挙動は従来通り（breaks 空で続行）。
				slog.ErrorContext(ctx, "failed to get breaks", "error", brkErr, "entry_id", entry.ID)
			}
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
	return inputs, nil
}

// buildAvailableDatesStaffInputsFn は GetAvailableDates の予約受付期間全体ぶんのシフト/休憩/当日予約を
// まとめて3クエリでプリフェッチし、日付ごとの StaffInputsFn クロージャ（DB問い合わせなし）を返す(G7-1)。
//
// エラー処理は元の buildStaffSlotInputs の per-day 実装と同じ致命度を踏襲する:
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
	minDate, maxDate := bookingWindowDates(settings)
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
			dateStr := entries[i].Date.Format("2006-01-02")
			if entriesByDateStaff[dateStr] == nil {
				entriesByDateStaff[dateStr] = map[uint64]*model.ShiftEntry{}
			}
			entriesByDateStaff[dateStr][entries[i].StaffID] = &entries[i]
			entryIDs = append(entryIDs, entries[i].ID)
		}
		breaks, brkErr := s.scheduleRepo.FindAllBreaksByEntryIDs(ctx, entryIDs)
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
		dateStr := reservations[i].StartTime.In(config.JST).Format("2006-01-02")
		reservationsByDate[dateStr] = append(reservationsByDate[dateStr], reservations[i])
	}

	return func(_ context.Context, date time.Time, _ uint64, _ uint64) ([]StaffSlotInput, error) {
		return buildStaffSlotInputsFromWindow(staffs, date, entriesByDateStaff, breaksByEntry, reservationsByDate), nil
	}, nil
}

// buildStaffSlotInputsFromWindow は buildStaffSlotInputs と同じ StaffSlotInput 構築ロジックを、
// 事前取得済みの期間データ（シフト・休憩・当日予約）から DB 問い合わせなしで行う（G7-1: 日付ループ N+1 解消）。
// GetAvailableDates の複数日プレビューでのみ使用する。単日パスの GetAvailableTimes は
// buildStaffSlotInputs をそのまま使い続ける（挙動不変）。
func buildStaffSlotInputsFromWindow(
	staffs []model.Staff,
	date time.Time,
	entriesByDateStaff map[string]map[uint64]*model.ShiftEntry,
	breaksByEntry map[uint64][]model.ShiftEntryBreak,
	reservationsByDate map[string][]model.Reservation,
) []StaffSlotInput {
	dateStr := date.Format("2006-01-02")
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
