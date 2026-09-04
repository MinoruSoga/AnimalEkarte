package reservation

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// resolveTargetStaffs はtypeID・staffIDに基づいて対象スタッフを返す。
func (s *liffService) resolveTargetStaffs(ctx context.Context, clinicID, typeID, staffID uint64) ([]model.Staff, error) {
	if staffID != 0 {
		staff, err := s.staffRepo.FindByID(ctx, clinicID, staffID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get staff")
		}
		if !staff.IsActive || !staff.ReservationVisible {
			return nil, nil
		}
		supports, err := s.staffRepo.SupportsReservationType(ctx, clinicID, staffID, typeID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get staff reservation capabilities")
		}
		if !supports {
			return nil, nil
		}
		return []model.Staff{*staff}, nil
	}

	all, err := s.staffRepo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get staffs")
	}
	return s.filterVisibleStaffsByTypeID(ctx, clinicID, typeID, all)
}

// filterVisibleStaffsByTypeID は is_active=true かつ reservation_visible=true かつ typeID に対応可能なスタッフを返す。
// FindAllReservationCapabilitiesByStaffIDs で一括取得して N+1 クエリを回避する。
func (s *liffService) filterVisibleStaffsByTypeID(ctx context.Context, clinicID, typeID uint64, all []model.Staff) ([]model.Staff, error) {
	visibleIDs := make([]uint64, 0, len(all))
	for i := range all {
		if all[i].IsActive && all[i].ReservationVisible {
			visibleIDs = append(visibleIDs, all[i].ID)
		}
	}
	if len(visibleIDs) == 0 {
		return nil, nil
	}

	allCapabilities, err := s.staffRepo.FindAllReservationCapabilitiesByStaffIDs(ctx, clinicID, visibleIDs)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get staff reservation capabilities")
	}

	capabilityMap := make(map[uint64][]model.StaffReservationCapability, len(allCapabilities))
	for _, capability := range allCapabilities {
		capabilityMap[capability.StaffID] = append(capabilityMap[capability.StaffID], capability)
	}

	result := make([]model.Staff, 0, len(visibleIDs))
	for i := range all {
		if !all[i].IsActive || !all[i].ReservationVisible {
			continue
		}
		if isCapable(capabilityMap[all[i].ID], typeID) {
			result = append(result, all[i])
		}
	}
	return result, nil
}
