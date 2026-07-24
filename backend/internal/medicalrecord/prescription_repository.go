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

func (r *prescriptionRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error) {
	prescriptions := make([]model.Prescription, 0)
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("medical_record_id = ?", medicalRecordID).
		Order("prescribed_at DESC").
		Find(&prescriptions).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "prescription", "")
	}
	return prescriptions, nil
}

func (r *prescriptionRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Prescription, error) {
	var prescription model.Prescription
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
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
		Where("clinic_id = ? AND owner_id = ? AND deleted_at IS NULL", clinicID, ownerID).
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
