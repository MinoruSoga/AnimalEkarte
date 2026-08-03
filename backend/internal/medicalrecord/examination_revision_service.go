package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ExaminationOfficialProjection keeps the immutable revision snapshot separate from
// the mutable parent row's current revision pointer.
type ExaminationOfficialProjection struct {
	model.Examination
	OfficialVersion uint64
}

// ExaminationOfficialReader is the separate Slice-A read capability for immutable
// official snapshots. The legacy general examination service interface remains unchanged.
type ExaminationOfficialReader interface {
	GetOfficialByID(ctx context.Context, clinicID, examinationID uint64) (*ExaminationOfficialProjection, error)
}

func (s *examinationService) confirmFirstRevisionTx(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	current, before *model.Examination,
	action, operation string,
) (*model.Examination, error) {
	if s.revisions == nil {
		return nil, apperrors.WrapInternalServerError("examination revision repository capability is required")
	}
	if actorID == nil || *actorID == 0 {
		return nil, apperrors.WrapInvalidInput("authenticated staff actor is required for examination confirmation")
	}
	if current == nil {
		return nil, apperrors.WrapInternalServerError("locked examination is required for confirmation")
	}
	if current.CurrentRevisionVersion != nil {
		return nil, apperrors.WrapConflict("revisioned examination requires the reconfirm workflow")
	}

	version, err := s.revisions.AppendOfficialRevision(
		ctx,
		clinicID,
		current.ID,
		*actorID,
		examinationInitialConfirmReason,
	)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to append official examination revision")
	}
	after := *current
	after.Status = model.ExaminationStatusConfirmed
	after.CurrentRevisionVersion = cloneUint64(version)
	if err := s.logParentMutationTx(ctx, clinicID, actorID, action, operation, before, &after); err != nil {
		return nil, err
	}
	confirmed, err := s.revisions.ConfirmWithRevisionCAS(
		ctx,
		clinicID,
		current.ID,
		current.Status,
		version,
	)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to advance examination revision pointer")
	}
	return confirmed, nil
}

// GetOfficialByID reads only the append-only official revision projection and fails
// closed when a legacy confirmed parent has no official revision.
func (s *examinationService) GetOfficialByID(
	ctx context.Context,
	clinicID, examinationID uint64,
) (*ExaminationOfficialProjection, error) {
	if s.revisions == nil {
		return nil, apperrors.WrapInternalServerError("examination revision repository capability is required")
	}
	exam, err := s.revisions.FindOfficialByID(ctx, clinicID, examinationID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read official examination revision", "error", err)
		return nil, apperrors.Wrap(err, "failed to read official examination revision")
	}
	return exam, nil
}
