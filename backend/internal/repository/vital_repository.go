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
	ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.VitalRecord, error)
	FindByID(ctx context.Context, id uint64) (*model.VitalRecord, error)
	Create(ctx context.Context, vital *model.VitalRecord) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
}

type vitalRepository struct {
	db *gorm.DB
}

// NewVitalRepository はVitalRepositoryを初期化して返す
func NewVitalRepository(db *gorm.DB) VitalRepository {
	return &vitalRepository{db: db}
}

func (r *vitalRepository) ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.VitalRecord, error) {
	vitals := make([]model.VitalRecord, 0)
	if err := r.db.WithContext(ctx).
		Where("medical_record_id = ?", medicalRecordID).
		Order("recorded_at ASC").
		Find(&vitals).Error; err != nil {
		return nil, apperrors.Wrap(err, "list vitals by medical_record_id")
	}
	return vitals, nil
}

func (r *vitalRepository) FindByID(ctx context.Context, id uint64) (*model.VitalRecord, error) {
	var vital model.VitalRecord
	err := r.db.WithContext(ctx).First(&vital, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vital", fmt.Sprintf("%d", id))
	}
	return &vital, nil
}

func (r *vitalRepository) Create(ctx context.Context, vital *model.VitalRecord) error {
	if err := r.db.WithContext(ctx).Create(vital).Error; err != nil {
		return apperrors.Wrap(err, "create vital")
	}
	return nil
}

func (r *vitalRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.VitalRecord{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update vital")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vital", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *vitalRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.VitalRecord{})
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete vital")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vital", fmt.Sprintf("%d", id))
	}
	return nil
}
