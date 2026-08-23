package staff

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
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

func (s *staffService) GetByIDInClinic(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Staff, error) {
	staff, err := s.repo.FindByIDInClinic(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to get clinic-scoped staff",
			"error", err,
			"id", id,
			"clinic_id", clinicID,
		)
		return nil, apperrors.Wrap(err, "failed to get clinic-scoped staff")
	}
	return staff, nil
}

// validateOccupationOwnership verifies a request-supplied occupation_id belongs to the
// caller's clinic before it is persisted (X-14 master FK guard batch U7). OccupationID
// is an optional FK — nil is always allowed. occupationRepo is a mandatory NewStaffService
// dependency wired from repos.Occupation in production, so a nil repo here is fail-closed
// (unlike validateReservationTypeGroup's nil-skip in reservation_type_service_core.go,
// which predates occupationRepo being a required dependency).
func (s *staffService) lockOccupationOwnership(
	ctx context.Context,
	clinicID uint64,
	occupationID *uint64,
) error {
	if occupationID == nil {
		return nil
	}
	if s.occupationRepo == nil {
		slog.ErrorContext(ctx, "occupation repository not configured", "clinic_id", clinicID)
		return apperrors.WrapInternalServerError("failed to verify occupation ownership: occupation repository not configured")
	}
	if _, err := s.occupationRepo.LockActiveByIDForShare(ctx, clinicID, *occupationID); err != nil {
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
	if err := validateStaffType(input.StaffType); err != nil {
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
		if err := s.lockOccupationOwnership(ctx, input.ClinicID, input.OccupationID); err != nil {
			return err
		}
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
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if clinicID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_id is required")
	}
	if id == 0 {
		return nil, apperrors.WrapInvalidInput("staff_id is required")
	}
	if s.repo == nil || s.assignmentRepo == nil || s.tx == nil {
		return nil, apperrors.WrapInternalServerError("staff update dependencies are not configured")
	}
	if input.Name != nil {
		if err := validateRequiredName(*input.Name); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate required name")
		}
		trimmed := strings.TrimSpace(*input.Name)
		input.Name = &trimmed
	}
	if input.StaffType != nil {
		if err := validateStaffType(*input.StaffType); err != nil {
			return nil, err
		}
	}

	hasPasswordUpdate := input.Password != nil && *input.Password != ""
	hasProfileUpdate := input.Name != nil || input.LicenseNumber != nil || input.OccupationID != nil ||
		input.SortOrder != nil || input.IsActive != nil || input.StaffType != nil ||
		input.ReservationDisplayName != nil || input.ReservationVisible != nil ||
		input.ReservationComment != nil || input.ReservationImageURL != nil

	if !hasProfileUpdate && !hasPasswordUpdate {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}

	var passwordHash string
	if hasPasswordUpdate {
		if err := validatePassword(*input.Password); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate password")
		}
		if s.accountRepo == nil || s.credentialAudit == nil {
			return nil, apperrors.WrapInternalServerError(
				"staff credential mutation dependencies are not configured",
			)
		}
		if err := validateStaffCredentialAudit(
			clinicID,
			id,
			input.CredentialAudit,
		); err != nil {
			return nil, err
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(*input.Password), config.BcryptCost)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to hash password")
		}
		passwordHash = string(hashed)
	}

	var result *model.Staff
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		lockedStaff, err := s.lockStaffForMutation(txCtx, clinicID, id, input.IsSystemAdmin)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock staff for update")
		}
		if lockedStaff == nil || lockedStaff.ID != id {
			return apperrors.WrapInternalServerError("staff lock returned an invalid record")
		}
		assignments, err := s.assignmentRepo.LockActiveByStaff(txCtx, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock staff clinic assignments for update")
		}
		if err := authorizeGlobalStaffUpdate(
			id,
			clinicID,
			assignments,
			input.AuthorizedClinicIDs,
			input.IsSystemAdmin,
		); err != nil {
			return err
		}
		writeClinicID, err := mutationClinicID(id, clinicID, assignments, input.IsSystemAdmin)
		if err != nil {
			return err
		}
		if err := s.lockOccupationOwnership(txCtx, writeClinicID, input.OccupationID); err != nil {
			return err
		}
		if hasPasswordUpdate && lockedStaff.AccountID == nil {
			return apperrors.WrapInvalidInput("staff does not have an account")
		}
		if hasProfileUpdate {
			if err := s.repo.Update(txCtx, writeClinicID, id, buildStaffUpdate(input)); err != nil {
				return apperrors.Wrap(err, "failed to update staff")
			}
		}
		if hasPasswordUpdate {
			if err := s.accountRepo.UpdatePasswordHash(
				txCtx,
				*lockedStaff.AccountID,
				passwordHash,
				time.Now(),
			); err != nil {
				return apperrors.Wrap(err, "failed to update staff account password")
			}
			if err := s.accountRepo.DeletePasswordResetTokens(
				txCtx,
				*lockedStaff.AccountID,
			); err != nil {
				return apperrors.Wrap(
					err,
					"failed to revoke staff password reset tokens",
				)
			}
			if auditErr := s.credentialAudit.LogEntryTx(
				txCtx,
				staffCredentialAuditEntry(
					*input.CredentialAudit,
					*lockedStaff.AccountID,
				),
			); auditErr != nil {
				return apperrors.Wrap(
					auditErr,
					"failed to write staff credential audit",
				)
			}
		}
		var updated *model.Staff
		if input.IsSystemAdmin {
			updated, err = s.repo.FindByID(txCtx, id)
		} else {
			updated, err = s.repo.FindByIDInClinic(txCtx, writeClinicID, id)
		}
		if err != nil {
			return apperrors.Wrap(err, "failed to get updated staff")
		}
		result = updated
		return nil
	}); err != nil {
		slog.ErrorContext(
			ctx,
			"failed to update staff",
			"error_type", fmt.Sprintf("%T", err),
			"id", id,
			"clinic_id", clinicID,
		)
		return nil, apperrors.Wrap(err, "failed to update staff")
	}

	slog.InfoContext(
		ctx,
		"staff updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("staff_id", id),
	)
	return result, nil
}

func authorizeGlobalStaffUpdate(
	staffID, currentClinicID uint64,
	assignments []model.StaffClinicAssignment,
	authorizedClinicIDs []uint64,
	isSystemAdmin bool,
) error {
	if len(assignments) == 0 {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", staffID))
	}

	authorized := make(map[uint64]struct{}, len(authorizedClinicIDs))
	for _, clinicID := range authorizedClinicIDs {
		if clinicID != 0 {
			authorized[clinicID] = struct{}{}
		}
	}

	hasCurrentClinic := false
	assignedClinics := make(map[uint64]struct{}, len(assignments))
	for _, assignment := range assignments {
		if assignment.StaffID != staffID || assignment.ClinicID == 0 ||
			assignment.DeletedAt.Valid {
			return apperrors.WrapInternalServerError(
				"staff clinic assignment lock returned an invalid record",
			)
		}
		if _, duplicate := assignedClinics[assignment.ClinicID]; duplicate {
			return apperrors.WrapInternalServerError(
				"staff clinic assignment lock returned duplicate active clinics",
			)
		}
		assignedClinics[assignment.ClinicID] = struct{}{}
		if assignment.ClinicID == currentClinicID {
			hasCurrentClinic = true
		}
	}
	if !hasCurrentClinic && !isSystemAdmin {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", staffID))
	}
	if isSystemAdmin {
		return nil
	}
	for clinicID := range assignedClinics {
		if _, ok := authorized[clinicID]; !ok {
			return apperrors.WrapForbidden(
				"staff update requires access to every assigned clinic",
			)
		}
	}
	if len(assignedClinics) > 1 {
		return apperrors.WrapForbidden(
			"multi-clinic staff updates require a system administrator",
		)
	}
	return nil
}

func (s *staffService) Delete(ctx context.Context, clinicID, id uint64, isSystemAdmin bool) error {
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
		staff, lockStaffErr := s.lockStaffForMutation(txCtx, clinicID, id, isSystemAdmin)
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
		if !assignedToClinic && !isSystemAdmin {
			return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
		}
		if len(assignments) > 1 {
			return apperrors.WrapConflict("複数のクリニックに所属しているスタッフは削除できません")
		}
		scopeClinicID, scopeErr := mutationClinicID(id, clinicID, assignments, isSystemAdmin)
		if scopeErr != nil {
			return scopeErr
		}

		reservationExists, reservationErr := s.reservationRepo.ExistsByStaffID(txCtx, scopeClinicID, id)
		if reservationErr != nil {
			slog.ErrorContext(txCtx, "failed to check reservation dependency", "error", reservationErr, "id", id, "clinic_id", scopeClinicID)
			return apperrors.Wrap(reservationErr, "failed to check reservation dependency")
		}
		if reservationExists {
			return apperrors.WrapConflict("このスタッフはシフト・予約データで使用中のため削除できません")
		}
		shiftExists, shiftErr := s.shiftEntryRepo.ExistsByStaffID(txCtx, scopeClinicID, id)
		if shiftErr != nil {
			slog.ErrorContext(txCtx, "failed to check shift dependency", "error", shiftErr, "id", id, "clinic_id", scopeClinicID)
			return apperrors.Wrap(shiftErr, "failed to check shift dependency")
		}
		if shiftExists {
			return apperrors.WrapConflict("このスタッフはシフト・予約データで使用中のため削除できません")
		}
		dependencies, dependencyErr := s.repo.CountBlockingReferencesByStaffID(txCtx, scopeClinicID, id)
		if dependencyErr != nil {
			slog.ErrorContext(txCtx, "failed to check staff dependencies", "error", dependencyErr, "id", id, "clinic_id", scopeClinicID)
			return apperrors.Wrap(dependencyErr, "failed to check staff dependencies")
		}
		if len(dependencies) > 0 {
			return apperrors.WrapConflict(dependencies[0].Label + "を残しているため削除できません")
		}
		if deleteErr := s.repo.Delete(txCtx, scopeClinicID, id); deleteErr != nil {
			slog.ErrorContext(txCtx, "failed to delete staff", "error", deleteErr, "id", id, "clinic_id", scopeClinicID)
			return apperrors.Wrap(deleteErr, "failed to delete staff")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to delete staff")
	}
	slog.InfoContext(ctx, "staff deleted", slog.Uint64("staff_id", id), slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *staffService) lockStaffForMutation(
	ctx context.Context,
	clinicID, id uint64,
	isSystemAdmin bool,
) (*model.Staff, error) {
	if isSystemAdmin {
		return s.repo.LockActiveByIDForUpdate(ctx, id)
	}
	return s.repo.LockActiveByIDForUpdateInClinic(ctx, clinicID, id)
}

func mutationClinicID(
	staffID, requestClinicID uint64,
	assignments []model.StaffClinicAssignment,
	isSystemAdmin bool,
) (uint64, error) {
	for i := range assignments {
		if assignments[i].ClinicID == requestClinicID {
			return requestClinicID, nil
		}
	}
	if !isSystemAdmin || len(assignments) == 0 {
		return 0, apperrors.WrapNotFound("staff", fmt.Sprintf("%d", staffID))
	}
	return assignments[0].ClinicID, nil
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
