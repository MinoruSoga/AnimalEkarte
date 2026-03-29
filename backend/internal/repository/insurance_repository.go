// Package repository provides data access implementations for Insurance entity.
package repository

import (
	"context"
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
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Insurance, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type insuranceRepository struct{ db *gorm.DB }

func NewInsuranceRepository(db *gorm.DB) InsuranceRepository { return &insuranceRepository{db: db} }

func (r *insuranceRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Insurance, error) {
	insurances := make([]model.Insurance, 0)
	err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).Order("sort_order ASC, name ASC").Find(&insurances).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "insurance", "")
	}
	return insurances, nil
}

func (r *insuranceRepository) FindByID(ctx context.Context, id uint64) (*model.Insurance, error) {
	var insurance model.Insurance
	err := r.db.WithContext(ctx).First(&insurance, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "insurance", fmt.Sprintf("%d", id))
	}
	return &insurance, nil
}

func (r *insuranceRepository) Create(ctx context.Context, insurance *model.Insurance) error {
	err := r.db.WithContext(ctx).Create(insurance).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("insurance", insurance.Name)
		}
		return apperrors.FromGORM(err, "insurance", "")
	}
	return nil
}

func (r *insuranceRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Insurance, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Insurance{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "insurance", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("insurance", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, id)
}

func (r *insuranceRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Insurance{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "insurance", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("insurance", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *insuranceRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.Insurance{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "insurance", fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("insurance id %d not found in this clinic", id))
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("reorder insurance: %w", err)
	}
	return nil
}
