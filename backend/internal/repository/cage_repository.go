// Package repository provides data access implementations for Cage entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Cage ----

type CageRepository interface {
	FindAll(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	Create(ctx context.Context, cage *model.Cage) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Cage, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type cageRepository struct{ db *gorm.DB }

func NewCageRepository(db *gorm.DB) CageRepository { return &cageRepository{db: db} }

func (r *cageRepository) FindAll(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error) {
	cages := make([]model.Cage, 0)
	q := r.db.WithContext(ctx).Model(&model.Cage{}).Scopes(clinicScope(clinicID))
	if cageType != nil {
		q = q.Where("cage_type = ?", *cageType)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&cages).Error; err != nil {
		return nil, apperrors.FromGORM(err, "cage", "")
	}
	return cages, nil
}

func (r *cageRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	var cage model.Cage
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&cage).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "cage", fmt.Sprintf("%d", id))
	}
	return &cage, nil
}

func (r *cageRepository) Create(ctx context.Context, cage *model.Cage) error {
	if err := r.db.WithContext(ctx).Create(cage).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "cage", "")
	}
	return nil
}

func (r *cageRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Cage, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Cage{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "cage", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("cage", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *cageRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.Cage{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "cage", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("cage", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *cageRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.Cage{}, "cage", clinicID, ids)
}
