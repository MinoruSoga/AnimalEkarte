// Package repository provides data access implementations for Cage entity.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Cage ----

type CageRepository interface {
	FindAll(ctx context.Context, cageType *string) ([]model.Cage, error)
	FindByID(ctx context.Context, id uint64) (*model.Cage, error)
	Create(ctx context.Context, cage *model.Cage) error
	Update(ctx context.Context, cage *model.Cage) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type cageRepository struct{ db *gorm.DB }

func NewCageRepository(db *gorm.DB) CageRepository { return &cageRepository{db: db} }

func (r *cageRepository) FindAll(ctx context.Context, cageType *string) ([]model.Cage, error) {
	cages := make([]model.Cage, 0)
	q := r.db.WithContext(ctx).Model(&model.Cage{})
	if cageType != nil {
		q = q.Where("cage_type = ?", *cageType)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&cages).Error; err != nil {
		return nil, apperrors.Wrap(err, "find cages")
	}
	return cages, nil
}

func (r *cageRepository) FindByID(ctx context.Context, id uint64) (*model.Cage, error) {
	var cage model.Cage
	if err := r.db.WithContext(ctx).First(&cage, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("cage", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find cage by id")
	}
	return &cage, nil
}

func (r *cageRepository) Create(ctx context.Context, cage *model.Cage) error {
	if err := r.db.WithContext(ctx).Create(cage).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("cage", cage.Name)
		}
		return apperrors.Wrap(err, "create cage")
	}
	return nil
}

// buildCageUpdateFields は updateCageRequest から非ゼロ値/非nilフィールドのみを map に変換する。
// GORM の zero-value スキップ問題を回避し、PATCH セマンティクスを実現する。
func buildCageUpdateFields(cage *model.Cage) map[string]any {
	fields := make(map[string]any)
	if cage.Name != "" {
		fields["name"] = cage.Name
	}
	if cage.CageType != "" {
		fields["cage_type"] = cage.CageType
	}
	if cage.CageSize != "" {
		fields["cage_size"] = cage.CageSize
	}
	if cage.Price != nil {
		fields["price"] = cage.Price
	}
	fields["is_active"] = cage.IsActive
	if cage.Description != "" {
		fields["description"] = cage.Description
	}
	if cage.SortOrder != 0 {
		fields["sort_order"] = cage.SortOrder
	}
	return fields
}

func (r *cageRepository) Update(ctx context.Context, cage *model.Cage) error {
	result := r.db.WithContext(ctx).
		Model(&model.Cage{}).
		Where("id = ? AND clinic_id = ?", cage.ID, cage.ClinicID).
		Updates(buildCageUpdateFields(cage))
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update cage")
	}
	if result.RowsAffected == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "update cage")
	}
	return nil
}

func (r *cageRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Cage{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete cage")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("cage", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *cageRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.Cage{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, "reorder cage")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("cage id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("reorder cage: %w", err)
	}
	return nil
}
