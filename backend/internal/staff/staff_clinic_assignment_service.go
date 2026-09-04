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

func sortedClinicIDs(clinicIDs []uint64) []uint64 {
	sorted := append([]uint64(nil), clinicIDs...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	return sorted
}

func clinicIDSet(clinicIDs []uint64) map[uint64]struct{} {
	set := make(map[uint64]struct{}, len(clinicIDs))
	for _, clinicID := range clinicIDs {
		if clinicID == 0 {
			continue
		}
		set[clinicID] = struct{}{}
	}
	return set
}

func isMutableClinicAssignment(clinicID uint64, isSystemAdmin bool, authorized map[uint64]struct{}) bool {
	if isSystemAdmin {
		return true
	}
	_, ok := authorized[clinicID]
	return ok
}

func authorizeExistingClinicAssignments(input *SetClinicAssignmentsInput, assignments []model.StaffClinicAssignment) error {
	if input.IsSystemAdmin {
		return nil
	}
	authorized := make(map[uint64]struct{}, len(input.AuthorizedClinicIDs))
	for _, clinicID := range input.AuthorizedClinicIDs {
		authorized[clinicID] = struct{}{}
	}
	for i := range assignments {
		if _, ok := authorized[assignments[i].ClinicID]; !ok {
			return apperrors.WrapForbidden("cannot replace staff assignments outside authorized clinics")
		}
	}
	return nil
}

func existingAssignmentClinicIDs(assignments []model.StaffClinicAssignment) map[uint64]struct{} {
	set := make(map[uint64]struct{}, len(assignments))
	for i := range assignments {
		if assignments[i].ClinicID != 0 {
			set[assignments[i].ClinicID] = struct{}{}
		}
	}
	return set
}

func preservedClinicAssignmentIDs(
	existingAssignments []model.StaffClinicAssignment,
	isSystemAdmin bool,
	authorized map[uint64]struct{},
) []uint64 {
	preserved := make([]uint64, 0)
	seen := make(map[uint64]struct{})
	for i := range existingAssignments {
		clinicID := existingAssignments[i].ClinicID
		if clinicID == 0 {
			continue
		}
		if _, duplicate := seen[clinicID]; duplicate {
			continue
		}
		if isMutableClinicAssignment(clinicID, isSystemAdmin, authorized) {
			continue
		}
		seen[clinicID] = struct{}{}
		preserved = append(preserved, clinicID)
	}
	return preserved
}

func removedMutableClinicAssignmentIDs(
	existingAssignments []model.StaffClinicAssignment,
	targetClinicIDs []uint64,
	isSystemAdmin bool,
	authorized map[uint64]struct{},
) []uint64 {
	targets := clinicIDSet(targetClinicIDs)
	removed := make([]uint64, 0, len(existingAssignments))
	for i := range existingAssignments {
		clinicID := existingAssignments[i].ClinicID
		if !isMutableClinicAssignment(clinicID, isSystemAdmin, authorized) {
			continue
		}
		if _, retained := targets[clinicID]; retained {
			continue
		}
		removed = append(removed, clinicID)
	}
	return sortedClinicIDs(removed)
}

func addedClinicAssignmentIDs(existing map[uint64]struct{}, targetClinicIDs []uint64) []uint64 {
	added := make([]uint64, 0)
	for _, clinicID := range targetClinicIDs {
		if _, found := existing[clinicID]; !found {
			added = append(added, clinicID)
		}
	}
	return sortedClinicIDs(added)
}

func finalClinicAssignmentIDs(preserved, requested []uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(preserved)+len(requested))
	final := make([]uint64, 0, len(preserved)+len(requested))
	for _, clinicID := range preserved {
		if clinicID == 0 {
			continue
		}
		if _, duplicate := seen[clinicID]; duplicate {
			continue
		}
		seen[clinicID] = struct{}{}
		final = append(final, clinicID)
	}
	for _, clinicID := range requested {
		if clinicID == 0 {
			continue
		}
		if _, duplicate := seen[clinicID]; duplicate {
			continue
		}
		seen[clinicID] = struct{}{}
		final = append(final, clinicID)
	}
	return final
}

// chooseStaffAssignmentPrimaryClinicID records the actor-scoped primary rule:
// if the resulting set is non-empty, primary is the first request ID when that
// ID is in the final set, else the existing primary when it remains, else the
// first remaining ID. An immutable/preserved existing primary is kept unless it
// was in the removed mutable set.
func chooseStaffAssignmentPrimaryClinicID(
	requestClinicIDs []uint64,
	finalSet map[uint64]struct{},
	existingPrimary uint64,
	isSystemAdmin bool,
	authorized map[uint64]struct{},
	remaining []uint64,
) uint64 {
	if existingPrimary != 0 {
		if _, ok := finalSet[existingPrimary]; ok &&
			!isMutableClinicAssignment(existingPrimary, isSystemAdmin, authorized) {
			return existingPrimary
		}
	}
	if len(requestClinicIDs) > 0 {
		if _, ok := finalSet[requestClinicIDs[0]]; ok {
			return requestClinicIDs[0]
		}
	}
	if existingPrimary != 0 {
		if _, ok := finalSet[existingPrimary]; ok {
			return existingPrimary
		}
	}
	if len(remaining) > 0 {
		return remaining[0]
	}
	return 0
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

// SetClinicAssignments applies an actor-scoped delta to a staff member's
// clinic assignments while preserving the canonical lock order shared with
// reservation writes: staff row, active assignment rows, then dependency
// checks and mutation. Mutable scope is every clinic for a system admin and
// AuthorizedClinicIDs otherwise. Non-admin actors fail closed when any
// existing assignment is outside AuthorizedClinicIDs. System-admin writes
// still preserve assignments outside the requested delta (including inactive
// clinics). AUS-01: fail-closed audit of old/new clinic_ids when production
// audit is wired (permissionAudit + attachPermissionAssignmentAudit).
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
		return s.setClinicAssignmentsInTx(ctx, input, clinicIDs, auditMeta)
	}); err != nil {
		return apperrors.Wrap(err, "failed to set clinic assignments")
	}
	slog.InfoContext(ctx, "clinic assignments updated", slog.Uint64("staff_id", input.StaffID), slog.Int("count", len(clinicIDs)))
	return nil
}

func (s *staffService) setClinicAssignmentsInTx(
	ctx context.Context,
	input *SetClinicAssignmentsInput,
	clinicIDs []uint64,
	auditMeta *PermissionAssignmentAudit,
) error {
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
	for i := range existingAssignments {
		if existingAssignments[i].StaffID != input.StaffID || existingAssignments[i].ClinicID == 0 {
			return apperrors.WrapInternalServerError("clinic assignment lock returned an invalid record")
		}
	}

	if err := authorizeExistingClinicAssignments(input, existingAssignments); err != nil {
		return err
	}

	authorized := clinicIDSet(input.AuthorizedClinicIDs)
	existingSet := existingAssignmentClinicIDs(existingAssignments)
	preservedIDs := preservedClinicAssignmentIDs(existingAssignments, input.IsSystemAdmin, authorized)
	removedClinicIDs := removedMutableClinicAssignmentIDs(
		existingAssignments,
		clinicIDs,
		input.IsSystemAdmin,
		authorized,
	)
	finalIDs := finalClinicAssignmentIDs(preservedIDs, clinicIDs)
	finalSet := clinicIDSet(finalIDs)

	// LockActiveByID only newly added clinics. Existing rows may belong to
	// inactive clinics (especially system-admin GET round-trips) and must
	// not be re-validated as active.
	for _, clinicID := range addedClinicAssignmentIDs(existingSet, clinicIDs) {
		clinic, lockClinicErr := s.clinicRepo.LockActiveByID(ctx, clinicID)
		if lockClinicErr != nil {
			return apperrors.Wrap(lockClinicErr, "failed to lock clinic assignment target")
		}
		if clinic == nil || clinic.ID != clinicID {
			return apperrors.WrapInternalServerError("clinic lock returned an invalid record")
		}
	}

	if dependencyErr := s.ensureRemovedClinicAssignmentsUnused(ctx, input.StaffID, removedClinicIDs); dependencyErr != nil {
		return dependencyErr
	}

	oldClinicIDs := make([]uint64, 0, len(existingAssignments))
	for i := range existingAssignments {
		oldClinicIDs = append(oldClinicIDs, existingAssignments[i].ClinicID)
	}

	if err := s.assignmentRepo.DeleteByStaffAndClinicIDs(ctx, input.StaffID, removedClinicIDs); err != nil {
		slog.ErrorContext(ctx, "failed to delete existing clinic assignments", "error", err, "staff_id", input.StaffID)
		return apperrors.Wrap(err, "failed to delete existing clinic assignments")
	}

	primaryClinicID := chooseStaffAssignmentPrimaryClinicID(
		clinicIDs,
		finalSet,
		staff.ClinicID,
		input.IsSystemAdmin,
		authorized,
		finalIDs,
	)
	if primaryClinicID == 0 {
		return apperrors.WrapInternalServerError("clinic assignment primary is missing")
	}

	for _, clinicID := range clinicIDs {
		assignment := &model.StaffClinicAssignment{
			StaffID:  input.StaffID,
			ClinicID: clinicID,
			IsMain:   clinicID == primaryClinicID,
		}
		if err := s.assignmentRepo.RestoreOrCreate(ctx, assignment); err != nil {
			slog.ErrorContext(ctx, "failed to restore or create clinic assignment", "error", err, "staff_id", input.StaffID, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to restore or create clinic assignment")
		}
	}
	if err := s.repo.UpdatePrimaryClinicID(ctx, input.StaffID, primaryClinicID); err != nil {
		slog.ErrorContext(ctx, "failed to update staff primary clinic", "error", err, "staff_id", input.StaffID, "clinic_id", primaryClinicID)
		return apperrors.Wrap(err, "failed to update staff primary clinic")
	}
	if s.permissionAudit != nil && auditMeta != nil {
		if auditWriteErr := s.permissionAudit.LogEntryTx(
			ctx,
			clinicAssignmentAuditEntry(*auditMeta, oldClinicIDs, finalIDs),
		); auditWriteErr != nil {
			return apperrors.Wrap(auditWriteErr, "failed to write staff clinic assignment audit")
		}
	}
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
