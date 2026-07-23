package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *staffService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error) {
	staff, total, err := s.repo.FindAll(ctx, clinicID, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list staff", "error", err, "clinic_id", clinicID)
		return nil, 0, apperrors.Wrap(err, "failed to list staff")
	}
	return staff, total, nil
}

func (s *staffService) GetByID(ctx context.Context, id uint64) (*model.Staff, error) {
	staff, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get staff", "error", err, "id", id)
		return nil, apperrors.Wrap(err, "failed to get staff")
	}
	return staff, nil
}

// validateOccupationOwnership verifies a request-supplied occupation_id belongs to the
// caller's clinic before it is persisted (X-14 master FK guard batch U7). OccupationID
// is an optional FK — nil is always allowed. occupationRepo is a mandatory NewStaffService
// dependency wired from repos.Occupation in production, so a nil repo here is fail-closed
// (unlike validateReservationTypeGroup's nil-skip in reservation_type_service_core.go,
// which predates occupationRepo being a required dependency).
func (s *staffService) validateOccupationOwnership(ctx context.Context, clinicID uint64, occupationID *uint64) error {
	if occupationID == nil {
		return nil
	}
	if s.occupationRepo == nil {
		slog.ErrorContext(ctx, "occupation repository not configured", "clinic_id", clinicID)
		return apperrors.WrapInternalServerError("failed to verify occupation ownership: occupation repository not configured")
	}
	if _, err := s.occupationRepo.FindByID(ctx, clinicID, *occupationID); err != nil {
		slog.ErrorContext(ctx, "failed to verify occupation ownership", "error", err, "occupation_id", *occupationID, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to verify occupation ownership")
	}
	return nil
}

func (s *staffService) Create(ctx context.Context, input *CreateStaffInput) (*model.Staff, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if input.ClinicID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_id is required")
	}
	if err := s.validateOccupationOwnership(ctx, input.ClinicID, input.OccupationID); err != nil {
		return nil, err
	}
	input.Name = strings.TrimSpace(input.Name)

	staffType := model.StaffType(input.StaffType)
	if staffType == "" {
		staffType = model.StaffTypeDoctor
	}

	reservationVisible := true
	if input.ReservationVisible != nil {
		reservationVisible = *input.ReservationVisible
	}

	var staff *model.Staff
	if err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		staff = &model.Staff{
			ClinicID:               input.ClinicID,
			Name:                   input.Name,
			LicenseNumber:          input.LicenseNumber,
			OccupationID:           input.OccupationID,
			SortOrder:              input.SortOrder,
			IsActive:               true,
			AccountID:              input.AccountID,
			StaffType:              staffType,
			ReservationDisplayName: input.ReservationDisplayName,
			ReservationVisible:     reservationVisible,
			ReservationComment:     input.ReservationComment,
			ReservationImageURL:    input.ReservationImageURL,
		}
		if err := s.repo.Create(ctx, staff); err != nil {
			slog.ErrorContext(ctx, "failed to create staff", "error", err)
			return apperrors.Wrap(err, "failed to create staff")
		}
		if input.ClinicID != 0 {
			if err := s.assignmentRepo.Create(ctx, &model.StaffClinicAssignment{
				StaffID:  staff.ID,
				ClinicID: input.ClinicID,
				IsMain:   true,
			}); err != nil {
				slog.ErrorContext(ctx, "failed to assign staff to clinic", "error", err, "clinic_id", input.ClinicID)
				return apperrors.Wrap(err, "failed to assign staff to clinic")
			}
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to create staff", "error", err)
		return nil, apperrors.Wrap(err, "failed to create staff")
	}

	slog.InfoContext(ctx, "staff created", slog.Uint64("clinic_id", input.ClinicID), slog.Uint64("staff_id", staff.ID))
	return staff, nil
}

// CreateWithAccount はメール・パスワード付きでスタッフを作成する。
// email 重複チェック・パスワードバリデーション・bcrypt ハッシュ化・Account 作成・Staff 作成をトランザクション内で一括で行う。
func (s *staffService) Update(ctx context.Context, clinicID, id uint64, input *UpdateStaffInput) (*model.Staff, error) {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		slog.ErrorContext(ctx, "failed to get staff", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get staff")
	}
	if err := s.validateOccupationOwnership(ctx, clinicID, input.OccupationID); err != nil {
		return nil, err
	}
	if input.Name != nil {
		if err := validateRequiredName(*input.Name); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate required name")
		}
		trimmed := strings.TrimSpace(*input.Name)
		input.Name = &trimmed
	}

	hasPasswordUpdate := input.Password != nil && *input.Password != ""
	hasProfileUpdate := input.Name != nil || input.LicenseNumber != nil || input.OccupationID != nil ||
		input.SortOrder != nil || input.IsActive != nil || input.StaffType != nil ||
		input.ReservationDisplayName != nil || input.ReservationVisible != nil ||
		input.ReservationComment != nil || input.ReservationImageURL != nil

	if !hasProfileUpdate && !hasPasswordUpdate {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}

	var staff *model.Staff
	if hasProfileUpdate {
		fields := buildStaffUpdate(input)
		if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
			slog.ErrorContext(ctx, "failed to update staff", "error", err, "id", id, "clinic_id", clinicID)
			return nil, apperrors.Wrap(err, "failed to update staff")
		}
		slog.InfoContext(ctx, "staff updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("staff_id", id))
	}

	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get updated staff", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get updated staff")
	}
	staff = updated

	if hasPasswordUpdate {
		if err := validatePassword(*input.Password); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate password")
		}
		if staff.AccountID != nil {
			if err := s.UpdatePassword(ctx, *staff.AccountID, *input.Password); err != nil {
				slog.ErrorContext(ctx, "failed to update password staff", "error", err)
				return nil, apperrors.Wrap(err, "failed to update password staff")
			}
		}
	}

	return staff, nil
}

func (s *staffService) Delete(ctx context.Context, clinicID, id uint64) error {
	if clinicID == 0 {
		return apperrors.WrapInvalidInput("clinic_id is required")
	}
	if id == 0 {
		return apperrors.WrapInvalidInput("staff_id is required")
	}
	if s.repo == nil || s.assignmentRepo == nil || s.reservationRepo == nil ||
		s.shiftEntryRepo == nil || s.tx == nil {
		return apperrors.WrapInternalServerError("staff deletion dependencies are not configured")
	}

	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		staff, lockStaffErr := s.repo.LockActiveByIDForUpdateInClinic(txCtx, clinicID, id)
		if lockStaffErr != nil {
			return apperrors.Wrap(lockStaffErr, "failed to lock staff before delete")
		}
		if staff == nil || staff.ID != id {
			return apperrors.WrapInternalServerError("staff lock returned an invalid record")
		}

		assignments, lockAssignmentsErr := s.assignmentRepo.LockActiveByStaff(txCtx, id)
		if lockAssignmentsErr != nil {
			return apperrors.Wrap(lockAssignmentsErr, "failed to lock staff clinic assignments before delete")
		}
		if len(assignments) == 0 {
			return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
		}
		assignedToClinic := false
		for _, assignment := range assignments {
			if assignment.StaffID != id || assignment.ClinicID == 0 {
				return apperrors.WrapInternalServerError("staff clinic assignment lock returned an invalid record")
			}
			if assignment.ClinicID == clinicID {
				assignedToClinic = true
			}
		}
		if !assignedToClinic {
			return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
		}
		if len(assignments) > 1 {
			return apperrors.WrapConflict("複数のクリニックに所属しているスタッフは削除できません")
		}

		reservationExists, reservationErr := s.reservationRepo.ExistsByStaffID(txCtx, clinicID, id)
		if reservationErr != nil {
			slog.ErrorContext(txCtx, "failed to check reservation dependency", "error", reservationErr, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(reservationErr, "failed to check reservation dependency")
		}
		if reservationExists {
			return apperrors.WrapConflict("このスタッフはシフト・予約データで使用中のため削除できません")
		}
		shiftExists, shiftErr := s.shiftEntryRepo.ExistsByStaffID(txCtx, clinicID, id)
		if shiftErr != nil {
			slog.ErrorContext(txCtx, "failed to check shift dependency", "error", shiftErr, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(shiftErr, "failed to check shift dependency")
		}
		if shiftExists {
			return apperrors.WrapConflict("このスタッフはシフト・予約データで使用中のため削除できません")
		}
		dependencies, dependencyErr := s.repo.CountBlockingReferencesByStaffID(txCtx, clinicID, id)
		if dependencyErr != nil {
			slog.ErrorContext(txCtx, "failed to check staff dependencies", "error", dependencyErr, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(dependencyErr, "failed to check staff dependencies")
		}
		if len(dependencies) > 0 {
			return apperrors.WrapConflict(dependencies[0].Label + "を残しているため削除できません")
		}
		if deleteErr := s.repo.Delete(txCtx, clinicID, id); deleteErr != nil {
			slog.ErrorContext(txCtx, "failed to delete staff", "error", deleteErr, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(deleteErr, "failed to delete staff")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to delete staff")
	}
	slog.InfoContext(ctx, "staff deleted", slog.Uint64("staff_id", id), slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *staffService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder staff", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder staff")
	}
	slog.InfoContext(ctx, "staff reordered", slog.Uint64("clinic_id", clinicID))
	return nil
}

// GetPermissionGroupIDs はスタッフが所属する権限グループIDリストを返す
