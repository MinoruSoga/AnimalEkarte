package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TreatmentSortUpdate は並び順一括更新に使う軽量DTO
type TreatmentSortUpdate struct {
	ID        uint64
	SortOrder int
}

// TreatmentRepository は治療項目の永続化インターフェース
type TreatmentRepository interface {
	ListByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Treatment, error)
	Create(ctx context.Context, treatment *model.Treatment) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	BulkUpdateSortOrder(ctx context.Context, updates []TreatmentSortUpdate) error
}

type treatmentRepository struct {
	db *gorm.DB
}

// NewTreatmentRepository はTreatmentRepositoryを初期化して返す
func NewTreatmentRepository(db *gorm.DB) TreatmentRepository {
	return &treatmentRepository{db: db}
}

func (r *treatmentRepository) ListByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error) {
	treatments := make([]model.Treatment, 0)
	if err := r.db.WithContext(ctx).
		Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id").
		Where("medical_records.clinic_id = ? AND treatments.medical_record_id = ? AND treatments.deleted_at IS NULL", clinicID, medicalRecordID).
		Preload("Consultation").
		Preload("Procedure").
		Preload("Medicine").
		Order("treatments.sort_order ASC").
		Find(&treatments).Error; err != nil {
		return nil, apperrors.FromGORM(err, "treatment", "")
	}
	return treatments, nil
}

func (r *treatmentRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Treatment, error) {
	var treatment model.Treatment
	err := r.db.WithContext(ctx).
		Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id").
		Where("medical_records.clinic_id = ? AND treatments.id = ? AND treatments.deleted_at IS NULL", clinicID, id).
		First(&treatment).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "treatment", fmt.Sprintf("%d", id))
	}
	return &treatment, nil
}

func (r *treatmentRepository) Create(ctx context.Context, treatment *model.Treatment) error {
	if err := r.db.WithContext(ctx).Create(treatment).Error; err != nil {
		return apperrors.FromGORM(err, "treatment", "")
	}
	return nil
}

func (r *treatmentRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id").
		Where("medical_records.clinic_id = ? AND treatments.id = ? AND treatments.deleted_at IS NULL", clinicID, id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("treatment", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *treatmentRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id").
		Where("medical_records.clinic_id = ? AND treatments.id = ? AND treatments.deleted_at IS NULL", clinicID, id).
		Delete(&model.Treatment{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("treatment", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *treatmentRepository) BulkUpdateSortOrder(ctx context.Context, updates []TreatmentSortUpdate) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, u := range updates {
			result := tx.Model(&model.Treatment{}).
				Where("id = ? AND deleted_at IS NULL", u.ID).
				Update("sort_order", u.SortOrder)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, fmt.Sprintf("bulk update sort_order for treatment %d", u.ID))
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "bulk update treatment sort order")
	}
	return nil
}
