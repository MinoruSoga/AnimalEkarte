package staff

import (
	"context"
	"slices"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type permissionAssignmentAuditContextKey struct{}

// PermissionAssignmentAudit is trusted request metadata for a staff permission
// group replacement.
type PermissionAssignmentAudit struct {
	ClinicID      uint64
	ActorStaffID  uint64
	TargetStaffID uint64
	IPAddress     string
	UserAgent     string
}

// PermissionAssignmentAuditEntry is staff's consumer-owned audit entry.
type PermissionAssignmentAuditEntry struct {
	ClinicID   *uint64
	ActorID    *uint64
	ActorType  string
	Action     string
	Resource   string
	ResourceID *uint64
	OldValue   any
	NewValue   any
	IPAddress  string
	UserAgent  string
}

// PermissionAssignmentAuditTxLogger persists an entry in the caller's ambient
// transaction.
type PermissionAssignmentAuditTxLogger interface {
	LogEntryTx(ctx context.Context, entry *PermissionAssignmentAuditEntry) error
}

func withPermissionAssignmentAudit(
	ctx context.Context,
	audit PermissionAssignmentAudit,
) context.Context {
	return context.WithValue(ctx, permissionAssignmentAuditContextKey{}, audit)
}

func permissionAssignmentAuditFromContext(
	ctx context.Context,
	clinicID, targetStaffID uint64,
) (*PermissionAssignmentAudit, error) {
	audit, ok := ctx.Value(permissionAssignmentAuditContextKey{}).(PermissionAssignmentAudit)
	if !ok ||
		audit.ClinicID == 0 ||
		audit.ClinicID != clinicID ||
		audit.ActorStaffID == 0 ||
		audit.TargetStaffID == 0 ||
		audit.TargetStaffID != targetStaffID {
		return nil, apperrors.WrapInternalServerError(
			"staff permission assignment audit metadata is invalid",
		)
	}
	return &audit, nil
}

func permissionAssignmentAuditEntry(
	audit PermissionAssignmentAudit,
	oldGroupIDs, newGroupIDs []uint64,
) *PermissionAssignmentAuditEntry {
	clinicID := audit.ClinicID
	actorID := audit.ActorStaffID
	resourceID := audit.TargetStaffID
	sortedOldGroupIDs := append([]uint64(nil), oldGroupIDs...)
	sortedNewGroupIDs := append([]uint64(nil), newGroupIDs...)
	slices.Sort(sortedOldGroupIDs)
	slices.Sort(sortedNewGroupIDs)
	return &PermissionAssignmentAuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     model.AuditActionStaffPermissionGroupsReplace,
		Resource:   model.AuditResourceStaff,
		ResourceID: &resourceID,
		OldValue: map[string]any{
			"staff_id":  audit.TargetStaffID,
			"group_ids": sortedOldGroupIDs,
		},
		NewValue: map[string]any{
			"staff_id":  audit.TargetStaffID,
			"group_ids": sortedNewGroupIDs,
		},
		IPAddress: audit.IPAddress,
		UserAgent: audit.UserAgent,
	}
}
