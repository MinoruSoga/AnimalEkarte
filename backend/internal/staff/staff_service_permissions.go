package staff

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func (s *staffService) GetPermissionGroupIDs(ctx context.Context, clinicID, staffID uint64) ([]uint64, error) {
	ids, err := s.permissionGroupRepo.FindAllGroupIDsByStaffID(ctx, clinicID, staffID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get permission group ids",
			"error", err, "clinic_id", clinicID, "staff_id", staffID)
		return nil, apperrors.Wrap(err, "failed to get permission group ids")
	}
	return ids, nil
}

// SetPermissionGroupIDs はスタッフの権限グループを全置換する
func (s *staffService) SetPermissionGroupIDs(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error {
	if s.repo == nil || s.permissionGroupRepo == nil || s.tx == nil || s.permissionAudit == nil {
		return apperrors.WrapInternalServerError(
			"staff permission assignment dependencies are not configured",
		)
	}
	audit, err := permissionAssignmentAuditFromContext(ctx, clinicID, staffID)
	if err != nil {
		return err
	}
	requestedGroupIDs := append([]uint64(nil), groupIDs...)

	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		lockedStaff, lockErr := s.repo.LockActiveByIDForUpdateInClinic(
			txCtx,
			clinicID,
			staffID,
		)
		if lockErr != nil {
			return apperrors.Wrap(lockErr, "failed to lock staff for permission assignment")
		}
		if lockedStaff == nil || lockedStaff.ID != staffID {
			return apperrors.WrapInternalServerError("staff lock returned an invalid record")
		}

		oldGroupIDs, findErr := s.permissionGroupRepo.FindAllGroupIDsByStaffID(
			txCtx,
			clinicID,
			staffID,
		)
		if findErr != nil {
			return apperrors.Wrap(findErr, "failed to get permission group ids")
		}
		if updateErr := s.permissionGroupRepo.UpdateStaffGroups(
			txCtx,
			clinicID,
			staffID,
			requestedGroupIDs,
		); updateErr != nil {
			return apperrors.Wrap(updateErr, "failed to set permission group ids")
		}
		if auditErr := s.permissionAudit.LogEntryTx(
			txCtx,
			permissionAssignmentAuditEntry(*audit, oldGroupIDs, requestedGroupIDs),
		); auditErr != nil {
			return apperrors.Wrap(auditErr, "failed to write staff permission assignment audit")
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to set permission group ids", "error", err, "id", staffID)
		return err
	}
	return nil
}

// GetExcludedReservationTypeIDs はスタッフの除外サービス種別IDリストを返す
func (s *staffService) GetExcludedReservationTypeIDs(
	ctx context.Context,
	clinicID, staffID uint64,
) ([]uint64, error) {
	items, err := s.resStaffRepo.FindAllExcludedReservationTypes(ctx, clinicID, staffID)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to get excluded service type ids",
			"error", err,
			"id", staffID,
			"clinic_id", clinicID,
		)
		return nil, apperrors.Wrap(err, "failed to get excluded service type ids")
	}
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ReservationTypeID)
	}
	return ids, nil
}

// SetExcludedReservationTypeIDs はスタッフの除外サービス種別を全置換する
func (s *staffService) SetExcludedReservationTypeIDs(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error {
	if err := s.resStaffRepo.UpdateExcludedReservationTypes(ctx, clinicID, staffID, typeIDs); err != nil {
		slog.ErrorContext(ctx, "failed to set excluded service type ids", "error", err, "id", staffID)
		return apperrors.Wrap(err, "failed to set excluded service type ids")
	}
	return nil
}

// GetCapableReservationTypeIDs はスタッフの対応可能サービス種別IDリストを返す
func (s *staffService) GetCapableReservationTypeIDs(ctx context.Context, clinicID, staffID uint64) ([]uint64, error) {
	items, err := s.resStaffRepo.FindAllReservationCapabilities(ctx, clinicID, staffID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get capable service type ids", "error", err, "id", staffID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get capable service type ids")
	}
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ReservationTypeID)
	}
	return ids, nil
}

// SetCapableReservationTypeIDs はスタッフの対応可能サービス種別を全置換する
func (s *staffService) SetCapableReservationTypeIDs(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error {
	if err := s.resStaffRepo.UpdateReservationCapabilities(ctx, clinicID, staffID, typeIDs); err != nil {
		slog.ErrorContext(ctx, "failed to set capable service type ids", "error", err, "id", staffID, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to set capable service type ids")
	}
	return nil
}

// VerifyClinicMembership はスタッフが指定クリニックに所属しているかを確認する。
// 所属していない場合は ErrNotFound を返す。
