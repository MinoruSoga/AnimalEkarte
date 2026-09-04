package identitylink

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *service) ListLinkedTreatmentHistory(
	ctx context.Context,
	actor ActorContext,
	seedClinicID, seedPetID uint64,
	includeLinked bool,
	page, limit int,
) ([]LinkedTreatmentHistoryItem, int64, error) {
	if err := requireActorClinics(actor); err != nil {
		return nil, 0, err
	}
	if !containsUint64(actor.VerifiedClinics, seedClinicID) {
		return nil, 0, apperrors.WrapForbidden("seed pet clinic outside actor scope")
	}

	var pairs []ClinicPetPair
	var err error
	if includeLinked {
		pairs, err = s.repo.ResolveLinkedPetPairs(ctx, seedClinicID, seedPetID, actor.VerifiedClinics)
	} else {
		// Default include_linked=false: seed pair only after existence check.
		if _, resolveErr := s.repo.ResolveLinkedPetPairs(ctx, seedClinicID, seedPetID, []uint64{seedClinicID}); resolveErr != nil {
			err = resolveErr
		} else {
			pairs = []ClinicPetPair{{ClinicID: seedClinicID, PetID: seedPetID}}
		}
	}
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to resolve linked pet pairs")
	}
	items, total, listErr := s.repo.ListLinkedTreatmentHistory(ctx, pairs, page, limit)
	if listErr != nil {
		return nil, 0, apperrors.Wrap(listErr, "failed to list linked treatment history")
	}
	return items, total, nil
}

func (s *service) writeAudit(
	ctx context.Context,
	actor ActorContext,
	action string,
	resourceID uint64,
	oldValue, newValue any,
) error {
	if s.audit == nil {
		return apperrors.WrapInternalServerError("identity link audit logger is required")
	}
	clinicID := actor.HomeClinicID
	if clinicID == 0 && len(actor.VerifiedClinics) > 0 {
		clinicID = actor.VerifiedClinics[0]
	}
	rid := resourceID
	staffID := actor.StaffID
	return s.audit.LogEntryTx(ctx, &audit.Entry{
		ClinicID:   &clinicID,
		ActorID:    &staffID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     action,
		Resource:   model.AuditResourceIdentityLink,
		ResourceID: &rid,
		// Non-PHI only: IDs and group metadata. Never names/phones.
		OldValue:  oldValue,
		NewValue:  newValue,
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	})
}
