package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

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

func NewPrescriptionRepository(db *gorm.DB) PrescriptionRepository {
	return &prescriptionRepository{db: db}
}

func (r *prescriptionRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error) {
	prescriptions := make([]model.Prescription, 0)
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
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
		Scopes(clinicScope(clinicID)).
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

func (r *prescriptionRepository) Create(ctx context.Context, prescription *model.Prescription) error {
	err := r.db.WithContext(ctx).Create(prescription).Error
	if err != nil {
		return apperrors.FromGORM(err, "prescription", "")
	}
	return nil
}

func (r *prescriptionRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.Prescription{}).
		Scopes(clinicScope(clinicID)).
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

func (r *prescriptionRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
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
