package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// auditVitalTx は vital create/update/delete の監査を ambient transaction へ参加させる（BUG-015）。
// LogVitalChange（base conn Create）は使わず、LogEntryTx → CreateTx と同型にする。
func (s *vitalService) auditVitalTx(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	action string,
	vitalID, medicalRecordID uint64,
	oldValue, newValue map[string]any,
) error {
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("vital audit dependency is required")
	}
	resourceID := vitalID
	entry := &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  sharedkernel.AuditActorTypeFor(actorID),
		Action:     action,
		Resource:   "vital",
		ResourceID: &resourceID,
		OldValue:   oldValue,
		NewValue:   newValue,
		Metadata: map[string]any{
			"medical_record_id": medicalRecordID,
		},
	}
	if err := s.auditTx.LogEntryTx(ctx, entry); err != nil {
		slog.ErrorContext(ctx, "failed to write vital audit",
			"error", err,
			"action", action,
			"vital_id", vitalID,
			"medical_record_id", medicalRecordID,
		)
		return apperrors.Wrap(err, "failed to write vital "+action+" audit")
	}
	return nil
}
