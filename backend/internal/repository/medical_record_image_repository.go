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
	FindByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.MedicalRecordImage, error)
	Create(ctx context.Context, image *model.MedicalRecordImage) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.MedicalRecordImage, error)
}

type medicalRecordImageRepository struct {
	db *gorm.DB
}

// NewMedicalRecordImageRepository は MedicalRecordImageRepository を初期化して返す
func NewMedicalRecordImageRepository(db *gorm.DB) MedicalRecordImageRepository {
	return &medicalRecordImageRepository{db: db}
}

func (r *medicalRecordImageRepository) FindByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
	images := make([]model.MedicalRecordImage, 0)
	if err := r.db.WithContext(ctx).
		Where("medical_record_id = ?", medicalRecordID).
		Preload("Staff", "deleted_at IS NULL").
		Order("sort_order ASC, created_at ASC").
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

func (r *medicalRecordImageRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.MedicalRecordImage{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "medical_record_image", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("medical_record_image", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *medicalRecordImageRepository) FindByID(ctx context.Context, id uint64) (*model.MedicalRecordImage, error) {
	var image model.MedicalRecordImage
	err := r.db.WithContext(ctx).
		Preload("Staff", "deleted_at IS NULL").
		First(&image, id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record_image", fmt.Sprintf("%d", id))
	}
	return &image, nil
}
