package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// PrescriptionRepository is the data access interface for prescriptions.
// Moved from internal/repository/prescription (BE8-4 batch7) — BE9-2D roll-up. Renamed from that
// subpackage's generic "Repository" to this entity-specific name only because medicalrecord
// holds multiple repository interfaces in one package; every external caller only ever saw
// this name via the internal/repository facade, so no call site changes.
type PrescriptionRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Prescription, error)
	FindActiveByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Prescription, error)
	Create(ctx context.Context, prescription *model.Prescription) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type prescriptionRepository struct {
	db *gorm.DB
}

// NewPrescriptionRepository constructs a PrescriptionRepository.
func NewPrescriptionRepository(db *gorm.DB) PrescriptionRepository {
	return &prescriptionRepository{db: db}
}

func prescriptionParentClinicScope(db *gorm.DB) *gorm.DB {
	return db.Where(`
		EXISTS (
			SELECT 1
			FROM owners
			WHERE owners.id = prescriptions.owner_id
			  AND owners.clinic_id = prescriptions.clinic_id
		)
		AND
		(prescriptions.pet_id IS NULL OR EXISTS (
			SELECT 1
			FROM pets
			WHERE pets.id = prescriptions.pet_id
			  AND pets.clinic_id = prescriptions.clinic_id
		))
		AND
		(prescriptions.medical_record_id IS NULL OR EXISTS (
			SELECT 1
			FROM medical_records
			WHERE medical_records.id = prescriptions.medical_record_id
			  AND medical_records.clinic_id = prescriptions.clinic_id
		))
	`)
}

func (r *prescriptionRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error) {
	prescriptions := make([]model.Prescription, 0)
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Scopes(prescriptionParentClinicScope).
		Where("prescriptions.medical_record_id = ?", medicalRecordID).
		Order("prescriptions.prescribed_at DESC").
		Find(&prescriptions).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "prescription", "")
	}
	return prescriptions, nil
}

// FindByID participates in an ambient transaction so Update can complete its
// response re-fetch before commit and roll back when that re-fetch fails (MRC-01).
func (r *prescriptionRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Prescription, error) {
	var prescription model.Prescription
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Scopes(prescriptionParentClinicScope).
		Where("prescriptions.id = ?", id).
		First(&prescription).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "prescription", fmt.Sprintf("%d", id))
	}
	return &prescription, nil
}

// FindActiveByOwner は補充推奨日計算用に飼い主の全アクティブ処方を返す（LSTEP-BE-009）。
func (r *prescriptionRepository) FindActiveByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Prescription, error) {
	prescriptions := make([]model.Prescription, 0)
	err := r.db.WithContext(ctx).
		Scopes(prescriptionParentClinicScope).
		Where(`
			prescriptions.clinic_id = ?
			AND prescriptions.deleted_at IS NULL
			AND EXISTS (
				SELECT 1
				FROM pets current_owner_pet
				JOIN owners current_owner
				  ON current_owner.id = current_owner_pet.owner_id
				 AND current_owner.clinic_id = current_owner_pet.clinic_id
				WHERE current_owner_pet.id = prescriptions.pet_id
				  AND current_owner_pet.clinic_id = prescriptions.clinic_id
				  AND current_owner.id = ?
			)
		`, clinicID, ownerID).
		Find(&prescriptions).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "prescription", "")
	}
	return prescriptions, nil
}

// Create は dbOrTx(ctx, r.db) で ambient tx に参加する（BE-refactor.md X-11）。
// LockByIDForUpdate の行ロック保持 tx 内から呼ばれた場合、別コネクションで INSERT すると
// prescriptions.medical_record_id の FK 制約チェックが同一行への FOR UPDATE ロックと
// デッドロックする（FK チェックは FOR KEY SHARE を要求し FOR UPDATE と競合するため）。
func (r *prescriptionRepository) Create(ctx context.Context, prescription *model.Prescription) error {
	err := persistence.DBOrTx(ctx, r.db).Create(prescription).Error
	if err != nil {
		return apperrors.FromGORM(err, "prescription", "")
	}
	return nil
}

// Update は dbOrTx(ctx, r.db) で ambient tx に参加する（Create と同じ理由、BE-refactor.md X-11）。
func (r *prescriptionRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Prescription{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "prescription", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("prescription", fmt.Sprintf("%d", id))
	}
	return nil
}

// Delete は dbOrTx(ctx, r.db) で ambient tx に参加する（Create/Update と同じ理由、BE-refactor.md H-8e）。
func (r *prescriptionRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Delete(&model.Prescription{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "prescription", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("prescription", fmt.Sprintf("%d", id))
	}
	return nil
}
