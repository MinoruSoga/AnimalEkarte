// Package repository provides data access implementations for Medicine entity.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Medicine ----

type MedicineRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Medicine, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error)
	Create(ctx context.Context, medicine *model.Medicine) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type medicineRepository struct{ db *gorm.DB }

func NewMedicineRepository(db *gorm.DB) MedicineRepository { return &medicineRepository{db: db} }

func (r *medicineRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Medicine, error) {
	medicines := make([]model.Medicine, 0)
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("sort_order ASC, name ASC").
		Find(&medicines).Error; err != nil {
		return nil, apperrors.Wrap(err, "find medicines")
	}
	return medicines, nil
}

func (r *medicineRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error) {
	var medicine model.Medicine
	if err := r.db.WithContext(ctx).
		First(&medicine, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("medicine", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find medicine by id")
	}
	return &medicine, nil
}

func (r *medicineRepository) Create(ctx context.Context, medicine *model.Medicine) error {
	if err := r.db.WithContext(ctx).Create(medicine).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("medicine", medicine.Name)
		}
		return apperrors.Wrap(err, "create medicine")
	}
	return nil
}

func (r *medicineRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.Medicine{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update medicine")
	}
	if result.RowsAffected == 0 {
		var count int64
		r.db.WithContext(ctx).Model(&model.Medicine{}).
			Where("id = ? AND clinic_id = ?", id, clinicID).
			Count(&count)
		if count == 0 {
			return apperrors.WrapNotFound("medicine", fmt.Sprintf("%d", id))
		}
	}
	return nil
}

func (r *medicineRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Medicine{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete medicine")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("medicine", fmt.Sprintf("%d", id))
	}
	return nil
}
