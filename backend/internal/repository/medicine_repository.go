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
	FindAll(ctx context.Context) ([]model.Medicine, error)
	FindByID(ctx context.Context, id uint64) (*model.Medicine, error)
	Create(ctx context.Context, medicine *model.Medicine) error
	Update(ctx context.Context, medicine *model.Medicine) error
	Delete(ctx context.Context, id uint64) error
}

type medicineRepository struct{ db *gorm.DB }

func NewMedicineRepository(db *gorm.DB) MedicineRepository { return &medicineRepository{db: db} }

func (r *medicineRepository) FindAll(ctx context.Context) ([]model.Medicine, error) {
	medicines := make([]model.Medicine, 0)
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&medicines).Error; err != nil {
		return nil, apperrors.Wrap(err, "find medicines")
	}
	return medicines, nil
}

func (r *medicineRepository) FindByID(ctx context.Context, id uint64) (*model.Medicine, error) {
	var medicine model.Medicine
	if err := r.db.WithContext(ctx).First(&medicine, "id = ?", id).Error; err != nil {
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

func (r *medicineRepository) Update(ctx context.Context, medicine *model.Medicine) error {
	result := r.db.WithContext(ctx).
		Model(&model.Medicine{}).
		Where("id = ? AND clinic_id = ?", medicine.ID, medicine.ClinicID).
		Updates(medicine)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update medicine")
	}
	if result.RowsAffected == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "update medicine")
	}
	return nil
}

func (r *medicineRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Medicine{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete medicine")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("medicine", fmt.Sprintf("%d", id))
	}
	return nil
}
