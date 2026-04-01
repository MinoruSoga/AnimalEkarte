// Package repository provides data access implementations for Vaccine entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Vaccine ----

type VaccineRepository interface {
	FindAll(ctx context.Context, species *string) ([]model.Vaccine, error)
	FindByID(ctx context.Context, id uint64) (*model.Vaccine, error)
	Create(ctx context.Context, vaccine *model.Vaccine) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccine, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByVaccineID(ctx context.Context, vaccineID uint64) (int64, error)
}

type vaccineRepository struct{ db *gorm.DB }

func NewVaccineRepository(db *gorm.DB) VaccineRepository { return &vaccineRepository{db: db} }

func (r *vaccineRepository) FindAll(ctx context.Context, species *string) ([]model.Vaccine, error) {
	vaccines := make([]model.Vaccine, 0)
	q := r.db.WithContext(ctx).Model(&model.Vaccine{})
	if species != nil {
		q = q.Where("species = ?", *species)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&vaccines).Error; err != nil {
		return nil, apperrors.Wrap(err, "find vaccines")
	}
	return vaccines, nil
}

func (r *vaccineRepository) FindByID(ctx context.Context, id uint64) (*model.Vaccine, error) {
	var vaccine model.Vaccine
	err := r.db.WithContext(ctx).First(&vaccine, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "vaccine", fmt.Sprintf("%d", id))
	}
	return &vaccine, nil
}

func (r *vaccineRepository) Create(ctx context.Context, vaccine *model.Vaccine) error {
	if err := r.db.WithContext(ctx).Create(vaccine).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("vaccine", vaccine.Name)
		}
		return apperrors.Wrap(err, "create vaccine")
	}
	return nil
}

func (r *vaccineRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccine, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Vaccine{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.Wrap(result.Error, "update vaccine")
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("vaccine", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, id)
}

func (r *vaccineRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Vaccine{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete vaccine")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vaccine", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *vaccineRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.Vaccine{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, "reorder vaccine")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("vaccine id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "reorder vaccines")
	}
	return nil
}

// CountUsageByVaccineID はワクチンマスタを参照している vaccination_records の件数を返す（BUG-107）
func (r *vaccineRepository) CountUsageByVaccineID(ctx context.Context, vaccineID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Vaccination{}).
		Where("vaccine_id = ?", vaccineID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "vaccination_record", "")
	}
	return count, nil
}
