package reservation

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// delegateStaff は指名なし時に no_staff_mode に従ってスタッフを自動割当する。
// 割当できない場合は 0 を返す（エラーではない）。
func (s *liffService) delegateStaff(ctx context.Context, clinicID, typeID uint64, mode string, date time.Time, startTime, endTime string) (uint64, error) {
	staffs, err := s.resolveTargetStaffs(ctx, clinicID, typeID, 0)
	if err != nil || len(staffs) == 0 {
		slog.ErrorContext(ctx, "failed to resolve target staffs liff", "error", err)
		return 0, apperrors.Wrap(err, "failed to resolve target staffs liff")
	}

	switch mode {
	case "top_priority":
		// 表示順1位（sort_order が最小）のスタッフに固定割当
		return staffs[0].ID, nil

	default: // "first_available"
		// 空き枠があるスタッフを表示順に探す
		dayResv, err := s.adminRepo.FindAllByDay(ctx, clinicID, date)
		if err != nil {
			return 0, nil //nolint:nilerr // 意図的フォールバック: 既存予約取得失敗時は空き確認をスキップして指名なしにする
		}
		startMin, err := MinutesSinceMidnight(startTime)
		if err != nil {
			return 0, nil //nolint:nilerr // 意図的フォールバック: 時刻フォーマット不正時は空き確認をスキップして指名なしにする
		}
		endMin, err := MinutesSinceMidnight(endTime)
		if err != nil {
			return 0, nil //nolint:nilerr // 意図的フォールバック: 時刻フォーマット不正時は空き確認をスキップして指名なしにする
		}
		for i := range staffs {
			if isStaffAvailable(staffs[i].ID, startMin, endMin, dayResv) {
				return staffs[i].ID, nil
			}
		}
		return 0, nil
	}
}

// isStaffAvailable はスタッフが指定時間枠で空いているか確認する。
func isStaffAvailable(staffID uint64, startMin, endMin int, dayResv []model.Reservation) bool {
	for i := range dayResv {
		if dayResv[i].Status == model.ReservationStatusCancelled {
			continue
		}
		if dayResv[i].DoctorID == nil || *dayResv[i].DoctorID != staffID {
			continue
		}
		rStart := dayResv[i].StartTime.Hour()*60 + dayResv[i].StartTime.Minute()
		rEnd := dayResv[i].EndTime.Hour()*60 + dayResv[i].EndTime.Minute()
		// 重複チェック: 新枠が既存予約と重なる場合は NG
		if startMin < rEnd && endMin > rStart {
			return false
		}
	}
	return true
}

// isCapable は指定コースIDがスタッフの対応可能リストに含まれるか確認する。
func isCapable(capabilities []model.StaffReservationCapability, typeID uint64) bool {
	for _, capability := range capabilities {
		if capability.ReservationTypeID == typeID {
			return true
		}
	}
	return false
}
