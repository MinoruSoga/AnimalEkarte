package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func (s *staffService) GetPermissionGroupIDs(ctx context.Context, staffID uint64) ([]uint64, error) {
	ids, err := s.permissionGroupRepo.FindAllGroupIDsByStaffID(ctx, staffID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get permission group ids", "error", err, "id", staffID)
		return nil, apperrors.Wrap(err, "failed to get permission group ids")
	}
	return ids, nil
}

// SetPermissionGroupIDs はスタッフの権限グループを全置換する
func (s *staffService) SetPermissionGroupIDs(ctx context.Context, staffID uint64, groupIDs []uint64) error {
	if err := s.permissionGroupRepo.UpdateStaffGroups(ctx, staffID, groupIDs); err != nil {
		slog.ErrorContext(ctx, "failed to set permission group ids", "error", err, "id", staffID)
		return apperrors.Wrap(err, "failed to set permission group ids")
	}
	return nil
}

// GetExcludedReservationTypeIDs はスタッフの除外サービス種別IDリストを返す
func (s *staffService) GetExcludedReservationTypeIDs(ctx context.Context, staffID uint64) ([]uint64, error) {
	items, err := s.resStaffRepo.FindAllExcludedReservationTypes(ctx, staffID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get excluded service type ids", "error", err, "id", staffID)
		return nil, apperrors.Wrap(err, "failed to get excluded service type ids")
	}
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ReservationTypeID)
	}
	return ids, nil
}

// SetExcludedReservationTypeIDs はスタッフの除外サービス種別を全置換する
func (s *staffService) SetExcludedReservationTypeIDs(ctx context.Context, staffID uint64, typeIDs []uint64) error {
	if err := s.resStaffRepo.UpdateExcludedReservationTypes(ctx, staffID, typeIDs); err != nil {
		slog.ErrorContext(ctx, "failed to set excluded service type ids", "error", err, "id", staffID)
		return apperrors.Wrap(err, "failed to set excluded service type ids")
	}
	return nil
}

// VerifyClinicMembership はスタッフが指定クリニックに所属しているかを確認する。
// 所属していない場合は ErrNotFound を返す。
