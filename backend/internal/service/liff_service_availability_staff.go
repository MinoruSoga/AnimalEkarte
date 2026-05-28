package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// resolveTargetStaffs はtypeID・staffIDに基づいて対象スタッフを返す。
func (s *liffService) resolveTargetStaffs(ctx context.Context, clinicID, typeID, staffID uint64) ([]model.Staff, error) {
	if staffID != 0 {
		staff, err := s.staffRepo.FindByID(ctx, clinicID, staffID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get staff", "error", err)
			return nil, apperrors.Wrap(err, "failed to get staff")
		}
		if !staff.ReservationVisible {
			return nil, nil
		}
		return []model.Staff{*staff}, nil
	}

	all, err := s.staffRepo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get staffs", "error", err)
		return nil, apperrors.Wrap(err, "failed to get staffs")
	}
	return s.filterVisibleStaffsByTypeID(ctx, typeID, all)
}

// filterVisibleStaffsByTypeID は reservation_visible=true かつ typeID を除外していないスタッフを返す。
// FindAllExcludedReservationTypesByStaffIDs で一括取得して N+1 クエリを回避する。
func (s *liffService) filterVisibleStaffsByTypeID(ctx context.Context, typeID uint64, all []model.Staff) ([]model.Staff, error) {
	visibleIDs := make([]uint64, 0, len(all))
	for i := range all {
		if all[i].ReservationVisible {
			visibleIDs = append(visibleIDs, all[i].ID)
		}
	}
	if len(visibleIDs) == 0 {
		return nil, nil
	}

	allExclusions, err := s.staffRepo.FindAllExcludedReservationTypesByStaffIDs(ctx, visibleIDs)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get excluded service types", "error", err)
		return nil, apperrors.Wrap(err, "failed to get excluded service types")
	}

	// staffID → 除外予約区分 のマップを構築（O(N) で済む）
	excludeMap := make(map[uint64][]model.StaffReservationExclusion, len(allExclusions))
	for _, ex := range allExclusions {
		excludeMap[ex.StaffID] = append(excludeMap[ex.StaffID], ex)
	}

	result := make([]model.Staff, 0, len(visibleIDs))
	for i := range all {
		if !all[i].ReservationVisible {
			continue
		}
		if !isExcluded(excludeMap[all[i].ID], typeID) {
			result = append(result, all[i])
		}
	}
	return result, nil
}
