package billing

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type vaccinationBillingParentRef struct {
	MedicalRecordID *uint64
	OwnerID         *uint64
	PetID           *uint64
	Status          model.BillingStatus
}

type vaccinationEventRef struct {
	MedicalRecordID *uint64
	PetID           *uint64
	VaccineID       uint64
}

func (r *billingItemRepository) quoteLockedVaccinationBilling(
	ctx context.Context,
	tx *gorm.DB,
	clinicID, vaccinationID uint64,
	billingRef vaccinationBillingParentRef,
	vaccinationRef vaccinationEventRef,
) (*vaccinationBillingValues, error) {
	_ = ctx
	var ownerID uint64
	if err := tx.
		Table("owners").
		Select("id").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", *billingRef.OwnerID, clinicID).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Take(&ownerID).Error; err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", *billingRef.OwnerID))
	}

	var petID uint64
	if err := tx.
		Table("pets").
		Select("id").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", *vaccinationRef.PetID, clinicID).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Take(&petID).Error; err != nil {
		return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", *vaccinationRef.PetID))
	}

	validateMedicalRecord := func(id uint64) error {
		var medicalRecordRef struct {
			PetID *uint64
		}
		if err := tx.
			Table("medical_records").
			Select("pet_id").
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", id, clinicID).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Take(&medicalRecordRef).Error; err != nil {
			return apperrors.FromGORM(err, "medical_record", fmt.Sprintf("%d", id))
		}
		if medicalRecordRef.PetID == nil ||
			*medicalRecordRef.PetID != *billingRef.PetID {
			return invalidBillingItemReferenceCombination()
		}
		return nil
	}
	medicalRecordIDs := make([]uint64, 0, 2)
	if billingRef.MedicalRecordID != nil {
		medicalRecordIDs = append(medicalRecordIDs, *billingRef.MedicalRecordID)
	}
	if vaccinationRef.MedicalRecordID != nil &&
		(billingRef.MedicalRecordID == nil ||
			*vaccinationRef.MedicalRecordID != *billingRef.MedicalRecordID) {
		medicalRecordIDs = append(medicalRecordIDs, *vaccinationRef.MedicalRecordID)
	}
	sort.Slice(medicalRecordIDs, func(i, j int) bool {
		return medicalRecordIDs[i] < medicalRecordIDs[j]
	})
	for _, medicalRecordID := range medicalRecordIDs {
		if err := validateMedicalRecord(medicalRecordID); err != nil {
			return nil, err
		}
	}
	if vaccinationRef.MedicalRecordID == nil {
		return nil, invalidBillingItemReferenceCombination()
	}
	if billingRef.MedicalRecordID != nil &&
		*billingRef.MedicalRecordID != *vaccinationRef.MedicalRecordID {
		return nil, invalidBillingItemReferenceCombination()
	}
	var confirmationStatus string
	if err := tx.
		Table("billing_confirmations").
		Select("status").
		Where("medical_record_id = ?", *vaccinationRef.MedicalRecordID).
		Take(&confirmationStatus).Error; err != nil || confirmationStatus != string(model.ConfirmationStatusConfirmed) {
		return nil, apperrors.WrapConflict("会計確認前の予防接種は請求できません")
	}

	var lockedVaccinationRef vaccinationEventRef
	if err := tx.
		Table("vaccinations").
		Select("medical_record_id", "pet_id", "vaccine_id").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", vaccinationID, clinicID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Take(&lockedVaccinationRef).Error; err != nil {
		return nil, apperrors.FromGORM(err, "vaccination", fmt.Sprintf("%d", vaccinationID))
	}
	if !sameOptionalBillingReference(vaccinationRef.MedicalRecordID, lockedVaccinationRef.MedicalRecordID) ||
		!sameOptionalBillingReference(vaccinationRef.PetID, lockedVaccinationRef.PetID) ||
		vaccinationRef.VaccineID != lockedVaccinationRef.VaccineID {
		return nil, apperrors.WrapConflict("予防接種情報が更新されたため再試行してください")
	}

	var vaccineRef struct {
		Name  string
		Price *int64
	}
	if err := tx.
		Table("vaccines").
		Select("name", "price").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", lockedVaccinationRef.VaccineID, clinicID).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Take(&vaccineRef).Error; err != nil {
		return nil, apperrors.FromGORM(err, "vaccine", fmt.Sprintf("%d", lockedVaccinationRef.VaccineID))
	}
	if strings.TrimSpace(vaccineRef.Name) == "" || vaccineRef.Price == nil || *vaccineRef.Price < 0 {
		return nil, apperrors.WrapInternalServerError("vaccination vaccine master is not billable")
	}

	var existingCount int64
	if err := tx.
		Table("billing_items AS bi").
		Where(
			"bi.vaccination_id = ? AND bi.clinic_id = ? AND bi.deleted_at IS NULL",
			vaccinationID,
			clinicID,
		).
		Count(&existingCount).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("vaccination:%d", vaccinationID))
	}
	if existingCount > 0 {
		return nil, apperrors.WrapConflict("この予防接種は既に会計明細へ取り込まれています")
	}

	return &vaccinationBillingValues{
		Name:      vaccineRef.Name,
		UnitPrice: *vaccineRef.Price,
	}, nil
}
