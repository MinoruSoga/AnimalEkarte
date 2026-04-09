package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// RecordImageRepository は診療画像のデータアクセス層
type RecordImageRepository interface {
	ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.RecordImage, error)
	Create(ctx context.Context, image *model.RecordImage) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*model.RecordImage, error)
}

type recordImageRepository struct {
	db *gorm.DB
}

// NewRecordImageRepository は RecordImageRepository を初期化して返す
func NewRecordImageRepository(db *gorm.DB) RecordImageRepository {
	return &recordImageRepository{db: db}
}

func (r *recordImageRepository) ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.RecordImage, error) {
	images := make([]model.RecordImage, 0)
	if err := r.db.WithContext(ctx).
		Where("medical_record_id = ?", medicalRecordID).
		Preload("Staff").
		Order("sort_order ASC, created_at ASC").
		Find(&images).Error; err != nil {
		return nil, apperrors.FromGORM(err, "record_image", "")
	}
	return images, nil
}

func (r *recordImageRepository) Create(ctx context.Context, image *model.RecordImage) error {
	if err := r.db.WithContext(ctx).Create(image).Error; err != nil {
		return apperrors.FromGORM(err, "record_image", "")
	}
	return nil
}

func (r *recordImageRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.RecordImage{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "record_image", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("record_image", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *recordImageRepository) FindByID(ctx context.Context, id uint64) (*model.RecordImage, error) {
	var image model.RecordImage
	err := r.db.WithContext(ctx).
		Preload("Staff").
		First(&image, id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "record_image", fmt.Sprintf("%d", id))
	}
	return &image, nil
}
