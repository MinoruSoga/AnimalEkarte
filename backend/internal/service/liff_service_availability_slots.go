package service

import (
	"context"
	"log/slog"
	"time"

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
