// Package repository provides data access implementations for Insurance entity.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Insurance ----

type InsuranceRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Insurance, error)
	FindByID(ctx context.Context, id uint64) (*model.Insurance, error)
	Create(ctx context.Context, insurance *model.Insurance) error
	Update(ctx context.Context, insurance *model.Insurance) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type insuranceRepository struct{ db *gorm.DB }

func NewInsuranceRepository(db *gorm.DB) InsuranceRepository { return &insuranceRepository{db: db} }

func (r *insuranceRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Insurance, error) {
	insurances := make([]model.Insurance, 0)
	if err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).Order("sort_order ASC, name ASC").Find(&insurances).Error; err != nil {
		return nil, apperrors.Wrap(err, "find insurances")
	}
	return insurances, nil
}

func (r *insuranceRepository) FindByID(ctx context.Context, id uint64) (*model.Insurance, error) {
	var insurance model.Insurance
	if err := r.db.WithContext(ctx).First(&insurance, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("insurance", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find insurance by id")
	}
	return &insurance, nil
}

func (r *insuranceRepository) Create(ctx context.Context, insurance *model.Insurance) error {
	if err := r.db.WithContext(ctx).Create(insurance).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("insurance", insurance.Name)
		}
		return apperrors.Wrap(err, "create insurance")
	}
	return nil
}

func (r *insuranceRepository) Update(ctx context.Context, insurance *model.Insurance) error {
	result := r.db.WithContext(ctx).
		Model(&model.Insurance{}).
		Where("id = ? AND clinic_id = ?", insurance.ID, insurance.ClinicID).
		Updates(insurance)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update insurance")
	}
	if result.RowsAffected == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "update insurance")
	}
	return nil
}

func (r *insuranceRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Insurance{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete insurance")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("insurance", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *insuranceRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.Insurance{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, "reorder insurance")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("insurance id %d not found in this clinic", id))
			}
		}
		return nil
	})
}
