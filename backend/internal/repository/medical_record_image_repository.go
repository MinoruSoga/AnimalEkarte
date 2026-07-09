package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// MedicalRecordImageRepository は診療画像のデータアクセス層
type MedicalRecordImageRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error)
	Create(ctx context.Context, image *model.MedicalRecordImage) error
	Delete(ctx context.Context, clinicID, id uint64) error
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecordImage, error)
}

type medicalRecordImageRepository struct {
	db *gorm.DB
}

// NewMedicalRecordImageRepository は MedicalRecordImageRepository を初期化して返す
func NewMedicalRecordImageRepository(db *gorm.DB) MedicalRecordImageRepository {
	return &medicalRecordImageRepository{db: db}
}

func (r *medicalRecordImageRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
	images := make([]model.MedicalRecordImage, 0)
	if err := r.db.WithContext(ctx).
		Joins("JOIN medical_records ON medical_records.id = medical_record_images.medical_record_id AND medical_records.deleted_at IS NULL").
		Where("medical_records.clinic_id = ? AND medical_record_images.medical_record_id = ?", clinicID, medicalRecordID).
		Preload("Staff", "deleted_at IS NULL").
		Order("medical_record_images.sort_order ASC, medical_record_images.created_at ASC").
		Find(&images).Error; err != nil {
		return nil, apperrors.FromGORM(err, "medical_record_image", "")
	}
	return images, nil
}

func (r *medicalRecordImageRepository) Create(ctx context.Context, image *model.MedicalRecordImage) error {
	if err := r.db.WithContext(ctx).Create(image).Error; err != nil {
		return apperrors.FromGORM(err, "medical_record_image", "")
	}
	return nil
}

func (r *medicalRecordImageRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("medical_record_images.id = ? AND medical_record_images.medical_record_id IN "+
			"(SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID).
		Delete(&model.MedicalRecordImage{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "medical_record_image", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("medical_record_image", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *medicalRecordImageRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecordImage, error) {
	var image model.MedicalRecordImage
	err := r.db.WithContext(ctx).
		Scopes(medicalRecordTenantScope("medical_record_images", clinicID)).
		Where("medical_record_images.id = ?", id).
		Preload("Staff", "deleted_at IS NULL").
		First(&image).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record_image", fmt.Sprintf("%d", id))
	}
	return &image, nil
}
