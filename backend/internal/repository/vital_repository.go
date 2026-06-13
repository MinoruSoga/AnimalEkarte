package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// VitalRepository はバイタル記録のデータアクセスインターフェース
type VitalRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.VitalRecord, error)
	Create(ctx context.Context, vital *model.VitalRecord) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type vitalRepository struct {
	db *gorm.DB
}

// NewVitalRepository はVitalRepositoryを初期化して返す
func NewVitalRepository(db *gorm.DB) VitalRepository {
	return &vitalRepository{db: db}
}

func (r *vitalRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error) {
	vitals := make([]model.VitalRecord, 0)
	if err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("vital_records.medical_record_id = ? AND vital_records.deleted_at IS NULL", medicalRecordID).
		Order("vital_records.recorded_at ASC").
		Find(&vitals).Error; err != nil {
		return nil, apperrors.FromGORM(err, "vital", "")
	}
	return vitals, nil
}

func (r *vitalRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.VitalRecord, error) {
	var vital model.VitalRecord
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("vital_records.id = ? AND vital_records.deleted_at IS NULL", id).
		First(&vital).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vital", fmt.Sprintf("%d", id))
	}
	return &vital, nil
}

func (r *vitalRepository) Create(ctx context.Context, vital *model.VitalRecord) error {
	if err := r.db.WithContext(ctx).Create(vital).Error; err != nil {
		return apperrors.FromGORM(err, "vital", "")
	}
	return nil
}

func (r *vitalRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.VitalRecord{}).
		Scopes(clinicScope(clinicID)).
		Where("vital_records.id = ? AND vital_records.deleted_at IS NULL", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "vital", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vital", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *vitalRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("vital_records.id = ?", id).
		Delete(&model.VitalRecord{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "vital", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vital", fmt.Sprintf("%d", id))
	}
	return nil
}
