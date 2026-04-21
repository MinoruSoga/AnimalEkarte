package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type MedicalRecordRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	Create(ctx context.Context, record *model.MedicalRecord) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicalRecord, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error)
	CountEstimatesByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (int64, error)
}

type medicalRecordRepository struct {
	db *gorm.DB
}

func NewMedicalRecordRepository(db *gorm.DB) MedicalRecordRepository {
	return &medicalRecordRepository{db: db}
}

func (r *medicalRecordRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
	records := make([]model.MedicalRecord, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.MedicalRecord{}).Scopes(clinicScope(clinicID))
	if petID != nil {
		q = q.Where("pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Where("owner_id = ?", *ownerID)
	}
	if startDate != nil {
		q = q.Where("date >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("date <= ?", *endDate)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medical_record", "")
	}
	if err := q.Offset((page-1)*limit).Limit(limit).Order("date DESC, created_at DESC").
		Preload("Owner", "deleted_at IS NULL").Preload("Pet", "deleted_at IS NULL").Preload("Pet.AnimalSpecies", "deleted_at IS NULL").Preload("Doctor", "deleted_at IS NULL").Preload("EnteredByStaff", "deleted_at IS NULL").Preload("Inquiry", "deleted_at IS NULL").Preload("Billing", "deleted_at IS NULL").
		Find(&records).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medical_record", "")
	}
	return records, total, nil
}

func (r *medicalRecordRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	var record model.MedicalRecord
	err := r.db.WithContext(ctx).
		Preload("Treatments", "deleted_at IS NULL").
		Preload("Vitals").
		Preload("Doctor", "deleted_at IS NULL").
		Preload("EnteredByStaff", "deleted_at IS NULL").
		Preload("Owner", "deleted_at IS NULL").
		Preload("Pet", "deleted_at IS NULL").
		Preload("Pet.AnimalSpecies", "deleted_at IS NULL").
		Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("%d", id))
	}
	return &record, nil
}

func (r *medicalRecordRepository) Create(ctx context.Context, record *model.MedicalRecord) error {
	err := r.db.WithContext(ctx).Create(record).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("medical_record", record.RecordNo)
		}
		return apperrors.FromGORM(err, "medical_record", "")
	}
	return nil
}

func (r *medicalRecordRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicalRecord, error) {
	result := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "medical_record", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *medicalRecordRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.MedicalRecord{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "medical_record", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountByPetID は指定されたペットに関連するカルテ数を返す
func (r *medicalRecordRepository) CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Scopes(clinicScope(clinicID)).
		Where("pet_id = ?", petID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "medical_record", "")
	}
	return count, nil
}

// CountEstimatesByMedicalRecordID はカルテに紐付く見積書の件数を返す（BUG-201）
// estimates.medical_record_id は ON DELETE RESTRICT のため削除前チェックが必要。
func (r *medicalRecordRepository) CountEstimatesByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Estimate{}).
		Where("medical_record_id = ?", medicalRecordID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "estimate", "")
	}
	return count, nil
}
