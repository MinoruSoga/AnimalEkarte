package staff

import (
	"context"
	"log/slog"
	"slices"
	"sort"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const maxStaffClinicAssignments = 50

// StaffAssignmentReservationUsage is the staff-owned consumer view of
// reservation data needed to prevent removing an assignment that an
// appointment still references. Implementations must participate in an
// ambient transaction when the context contains one.
type StaffAssignmentReservationUsage interface {
	ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error)
	FindClinicIDsByStaffID(ctx context.Context, clinicIDs []uint64, staffID uint64) ([]uint64, error)
}

type StaffClinicAssignmentService interface {
	FindAllByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error)
	FindByStaffAndClinic(ctx context.Context, staffID, clinicID uint64) (*model.StaffClinicAssignment, error)
	Create(ctx context.Context, assignment *model.StaffClinicAssignment) error
}

type staffClinicAssignmentService struct {
	repo StaffClinicAssignmentRepository
}

func NewStaffClinicAssignmentService(repo StaffClinicAssignmentRepository) StaffClinicAssignmentService {
	return &staffClinicAssignmentService{repo: repo}
}

func (s *staffClinicAssignmentService) FindAllByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
	assignments, err := s.repo.FindByStaffID(ctx, staffID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find clinic assignments for staff", "error", err)
		return nil, apperrors.Wrap(err, "failed to find clinic assignments for staff")
	}
	return assignments, nil
}

func (s *staffClinicAssignmentService) FindByStaffAndClinic(
	ctx context.Context,
	staffID, clinicID uint64,
) (*model.StaffClinicAssignment, error) {
	assignment, err := s.repo.FindByStaffAndClinic(ctx, staffID, clinicID)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to find clinic-scoped assignment for staff",
			"error", err,
			"staff_id", staffID,
			"clinic_id", clinicID,
		)
		return nil, apperrors.Wrap(err, "failed to find clinic-scoped assignment for staff")
	}
	return assignment, nil
}

func (s *staffClinicAssignmentService) Create(ctx context.Context, assignment *model.StaffClinicAssignment) error {
	if err := s.repo.Create(ctx, assignment); err != nil {
		slog.ErrorContext(ctx, "failed to create staff clinic assignment", "error", err, "staff_id", assignment.StaffID, "clinic_id", assignment.ClinicID)
		return apperrors.Wrap(err, "failed to create staff clinic assignment")
	}
	slog.InfoContext(ctx, "staff clinic assignment created",
		slog.Uint64("staff_id", assignment.StaffID),
		slog.Uint64("clinic_id", assignment.ClinicID))
	return nil
}

func validateAndDedupeClinicAssignments(input *SetClinicAssignmentsInput) ([]uint64, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("clinic assignment input is required")
	}
	if input.StaffID == 0 {
		return nil, apperrors.WrapInvalidInput("staff_id is required")
	}
	if len(input.ClinicIDs) == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_ids must not be empty")
	}
	if len(input.ClinicIDs) > maxStaffClinicAssignments {
		return nil, apperrors.WrapInvalidInput("clinic_ids must contain at most 50 ids")
	}

	authorized := make(map[uint64]struct{}, len(input.AuthorizedClinicIDs))
	for _, clinicID := range input.AuthorizedClinicIDs {
		authorized[clinicID] = struct{}{}
	}

	seen := make(map[uint64]struct{}, len(input.ClinicIDs))
	clinicIDs := make([]uint64, 0, len(input.ClinicIDs))
	for _, clinicID := range input.ClinicIDs {
		if clinicID == 0 {
			return nil, apperrors.WrapInvalidInput("clinic_ids must contain positive ids")
		}
		if _, duplicate := seen[clinicID]; duplicate {
			continue
		}
		if !input.IsSystemAdmin {
			if _, ok := authorized[clinicID]; !ok {
				return nil, apperrors.WrapForbidden("cannot assign staff outside authorized clinics")
			}
		}
		seen[clinicID] = struct{}{}
		clinicIDs = append(clinicIDs, clinicID)
	}
	return clinicIDs, nil
}

func authorizeExistingClinicAssignments(
	input *SetClinicAssignmentsInput,
	assignments []model.StaffClinicAssignment,
) error {
	if input.IsSystemAdmin {
		return nil
	}
	authorized := make(map[uint64]struct{}, len(input.AuthorizedClinicIDs))
	for _, clinicID := range input.AuthorizedClinicIDs {
		authorized[clinicID] = struct{}{}
	}
	for _, assignment := range assignments {
		if _, ok := authorized[assignment.ClinicID]; !ok {
			return apperrors.WrapForbidden("cannot replace staff assignments outside authorized clinics")
		}
	}
	return nil
}

func sortedClinicIDs(clinicIDs []uint64) []uint64 {
	sorted := append([]uint64(nil), clinicIDs...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	return sorted
}

func removedClinicAssignmentIDs(
	existingAssignments []model.StaffClinicAssignment,
	targetClinicIDs []uint64,
) []uint64 {
	targets := make(map[uint64]struct{}, len(targetClinicIDs))
	for _, clinicID := range targetClinicIDs {
		targets[clinicID] = struct{}{}
	}

	removed := make([]uint64, 0, len(existingAssignments))
	for _, assignment := range existingAssignments {
		if _, retained := targets[assignment.ClinicID]; !retained {
			removed = append(removed, assignment.ClinicID)
		}
	}
	return sortedClinicIDs(removed)
}

func (s *staffService) ensureRemovedClinicAssignmentsUnused(
	ctx context.Context,
	staffID uint64,
	clinicIDs []uint64,
) error {
	if len(clinicIDs) == 0 {
		return nil
	}

	reservationClinicIDs, dependencyErr := s.reservationRepo.FindClinicIDsByStaffID(
		ctx,
		clinicIDs,
		staffID,
	)
	if dependencyErr != nil {
		slog.ErrorContext(
			ctx,
			"failed to check reservation dependency before removing clinic assignment",
			"error", dependencyErr,
			"staff_id", staffID,
			"clinic_ids", clinicIDs,
		)
		return apperrors.Wrap(
			dependencyErr,
			"failed to check reservation dependency before removing clinic assignment",
		)
	}
	if len(reservationClinicIDs) > 0 {
		return apperrors.WrapConflict("予約データがあるクリニック所属は解除できません")
	}

	shiftClinicIDs, dependencyErr := s.shiftEntryRepo.FindClinicIDsByStaffID(ctx, clinicIDs, staffID)
	if dependencyErr != nil {
		slog.ErrorContext(
			ctx,
			"failed to check shift dependency before removing clinic assignment",
			"error", dependencyErr,
			"staff_id", staffID,
			"clinic_ids", clinicIDs,
		)
		return apperrors.Wrap(
			dependencyErr,
			"failed to check shift dependency before removing clinic assignment",
		)
	}
	if len(shiftClinicIDs) > 0 {
		return apperrors.WrapConflict("シフトデータがあるクリニック所属は解除できません")
	}
	return nil
}

// SetClinicAssignments replaces a staff member's clinic assignments while
// preserving the canonical lock order shared with reservation writes:
// staff row, active assignment rows, then dependency checks and mutation.
// AUS-01: fail-closed audit of old/new clinic_ids when production audit is wired
// (permissionAudit + attachPermissionAssignmentAudit). Missing either half fails
// closed; both absent (legacy NewStaffService unit constructors) skips audit.
func (s *staffService) SetClinicAssignments(ctx context.Context, input *SetClinicAssignmentsInput) error {
	clinicIDs, err := validateAndDedupeClinicAssignments(input)
	if err != nil {
		return err
	}
	if s.repo == nil || s.assignmentRepo == nil || s.reservationRepo == nil ||
		s.shiftEntryRepo == nil || s.clinicRepo == nil || s.tx == nil {
		return apperrors.WrapInternalServerError("clinic assignment dependencies are not configured")
	}
	auditMeta, hasAuditMeta := clinicAssignmentAuditFromContext(ctx, input.StaffID)
	if s.permissionAudit == nil && hasAuditMeta {
		return apperrors.WrapInternalServerError("clinic assignment audit logger is not configured")
	}
	if s.permissionAudit != nil && !hasAuditMeta {
		return apperrors.WrapInternalServerError("staff clinic assignment audit metadata is invalid")
	}

	if err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		staff, lockStaffErr := s.repo.LockActiveByIDForUpdate(ctx, input.StaffID)
		if lockStaffErr != nil {
			return apperrors.Wrap(lockStaffErr, "failed to lock staff for clinic assignment replacement")
		}
		if staff == nil || staff.ID != input.StaffID {
			return apperrors.WrapInternalServerError("staff lock returned an invalid record")
		}

		existingAssignments, lockAssignmentsErr := s.assignmentRepo.LockActiveByStaff(ctx, input.StaffID)
		if lockAssignmentsErr != nil {
			return apperrors.Wrap(lockAssignmentsErr, "failed to lock existing clinic assignments")
		}
		for _, assignment := range existingAssignments {
			if assignment.StaffID != input.StaffID || assignment.ClinicID == 0 {
				return apperrors.WrapInternalServerError("clinic assignment lock returned an invalid record")
			}
		}
		if authorizeErr := authorizeExistingClinicAssignments(input, existingAssignments); authorizeErr != nil {
			return authorizeErr
		}

		// Multiple staff replacements can target the same clinic set. Locking
		// these rows by ID avoids an input-order-dependent deadlock while
		// retaining the caller's first clinic as the primary assignment.
		for _, clinicID := range sortedClinicIDs(clinicIDs) {
			clinic, lockClinicErr := s.clinicRepo.LockActiveByID(ctx, clinicID)
			if lockClinicErr != nil {
				return apperrors.Wrap(lockClinicErr, "failed to lock clinic assignment target")
			}
			if clinic == nil || clinic.ID != clinicID {
				return apperrors.WrapInternalServerError("clinic lock returned an invalid record")
			}
		}

		removedClinicIDs := removedClinicAssignmentIDs(existingAssignments, clinicIDs)
		if dependencyErr := s.ensureRemovedClinicAssignmentsUnused(ctx, input.StaffID, removedClinicIDs); dependencyErr != nil {
			return dependencyErr
		}

		oldClinicIDs := make([]uint64, 0, len(existingAssignments))
		for _, a := range existingAssignments {
			oldClinicIDs = append(oldClinicIDs, a.ClinicID)
		}

		if err := s.assignmentRepo.Delete(ctx, input.StaffID); err != nil {
			slog.ErrorContext(ctx, "failed to delete existing clinic assignments", "error", err, "staff_id", input.StaffID)
			return apperrors.Wrap(err, "failed to delete existing clinic assignments")
		}
		for i, clinicID := range clinicIDs {
			assignment := &model.StaffClinicAssignment{
				StaffID:  input.StaffID,
				ClinicID: clinicID,
				IsMain:   i == 0,
			}
			if err := s.assignmentRepo.RestoreOrCreate(ctx, assignment); err != nil {
				slog.ErrorContext(ctx, "failed to restore or create clinic assignment", "error", err, "staff_id", input.StaffID, "clinic_id", clinicID)
				return apperrors.Wrap(err, "failed to restore or create clinic assignment")
			}
		}
		if err := s.repo.UpdatePrimaryClinicID(ctx, input.StaffID, clinicIDs[0]); err != nil {
			slog.ErrorContext(ctx, "failed to update staff primary clinic", "error", err, "staff_id", input.StaffID, "clinic_id", clinicIDs[0])
			return apperrors.Wrap(err, "failed to update staff primary clinic")
		}
		if s.permissionAudit != nil && auditMeta != nil {
			if auditWriteErr := s.permissionAudit.LogEntryTx(
				ctx,
				clinicAssignmentAuditEntry(*auditMeta, oldClinicIDs, clinicIDs),
			); auditWriteErr != nil {
				return apperrors.Wrap(auditWriteErr, "failed to write staff clinic assignment audit")
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to set clinic assignments")
	}
	slog.InfoContext(ctx, "clinic assignments updated", slog.Uint64("staff_id", input.StaffID), slog.Int("count", len(clinicIDs)))
	return nil
}

// clinicAssignmentAuditFromContext reuses attachPermissionAssignmentAudit metadata (AUS-01).
// ok=false when middleware did not attach metadata (legacy unit-test constructors).
func clinicAssignmentAuditFromContext(
	ctx context.Context,
	targetStaffID uint64,
) (*PermissionAssignmentAudit, bool) {
	audit, ok := ctx.Value(permissionAssignmentAuditContextKey{}).(PermissionAssignmentAudit)
	if !ok ||
		audit.ClinicID == 0 ||
		audit.ActorStaffID == 0 ||
		audit.TargetStaffID == 0 ||
		audit.TargetStaffID != targetStaffID {
		return nil, false
	}
	return &audit, true
}

func clinicAssignmentAuditEntry(
	audit PermissionAssignmentAudit,
	oldClinicIDs, newClinicIDs []uint64,
) *PermissionAssignmentAuditEntry {
	clinicID := audit.ClinicID
	actorID := audit.ActorStaffID
	resourceID := audit.TargetStaffID
	sortedOld := append([]uint64(nil), oldClinicIDs...)
	sortedNew := append([]uint64(nil), newClinicIDs...)
	slices.Sort(sortedOld)
	slices.Sort(sortedNew)
	return &PermissionAssignmentAuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     "staff.clinic_assignments.replace",
		Resource:   model.AuditResourceStaff,
		ResourceID: &resourceID,
		OldValue: map[string]any{
			"staff_id":   audit.TargetStaffID,
			"clinic_ids": sortedOld,
		},
		NewValue: map[string]any{
			"staff_id":   audit.TargetStaffID,
			"clinic_ids": sortedNew,
		},
		IPAddress: audit.IPAddress,
		UserAgent: audit.UserAgent,
	}
}
