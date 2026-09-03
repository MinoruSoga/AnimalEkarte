// Package billing provides billing item vaccination persistence.
package billing

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func (r *billingItemRepository) ValidateVaccinationCreateReference(
	ctx context.Context,
	clinicID, billingID uint64,
	vaccinationID uint64,
) (*vaccinationBillingValues, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return nil, apperrors.WrapInternalServerError("vaccination billing validation requires an active transaction")
	}
	tx = tx.WithContext(ctx)

	var billingRef vaccinationBillingParentRef
	if err := tx.
		Table("billings").
		Select("medical_record_id", "owner_id", "pet_id", "status").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", billingID, clinicID).
		Take(&billingRef).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", billingID))
	}
	if billingRef.Status == model.BillingStatusCompleted ||
		billingRef.Status == model.BillingStatusCancelled {
		return nil, apperrors.WrapConflict("確定済みまたは取消済みの会計には予防接種を追加できません")
	}

	var vaccinationRef vaccinationEventRef
	// Read the event graph without a lock first so its medical-record parent can
	// be locked before the event. Vaccination writes use the same canonical
	// MedicalRecord -> Vaccination lock order.
	if err := tx.
		Table("vaccinations").
		Select("medical_record_id", "pet_id", "vaccine_id").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", vaccinationID, clinicID).
		Take(&vaccinationRef).Error; err != nil {
		return nil, apperrors.FromGORM(err, "vaccination", fmt.Sprintf("%d", vaccinationID))
	}
	if billingRef.OwnerID == nil ||
		billingRef.PetID == nil ||
		vaccinationRef.PetID == nil ||
		*billingRef.PetID != *vaccinationRef.PetID {
		return nil, invalidBillingItemReferenceCombination()
	}

	return r.quoteLockedVaccinationBilling(ctx, tx, clinicID, vaccinationID, billingRef, vaccinationRef)
}
