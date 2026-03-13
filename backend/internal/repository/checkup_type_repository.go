// Package repository provides data access implementations for CheckupType entity.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- CheckupType ----

type CheckupTypeRepository interface {
	FindAll(ctx context.Context) ([]model.CheckupType, error)
	FindByID(ctx context.Context, id uint64) (*model.CheckupType, error)
	Create(ctx context.Context, checkupType *model.CheckupType) error
	Update(ctx context.Context, checkupType *model.CheckupType) error
	Delete(ctx context.Context, id uint64) error
}

type checkupTypeRepository struct{ db *gorm.DB }

func NewCheckupTypeRepository(db *gorm.DB) CheckupTypeRepository {
	return &checkupTypeRepository{db: db}
}

func (r *checkupTypeRepository) FindAll(ctx context.Context) ([]model.CheckupType, error) {
	checkupTypes := make([]model.CheckupType, 0)
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&checkupTypes).Error; err != nil {
		return nil, apperrors.Wrap(err, "find checkup types")
	}
	return checkupTypes, nil
}

func (r *checkupTypeRepository) FindByID(ctx context.Context, id uint64) (*model.CheckupType, error) {
	var checkupType model.CheckupType
	if err := r.db.WithContext(ctx).First(&checkupType, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("checkup_type", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find checkup type by id")
	}
	return &checkupType, nil
}

func (r *checkupTypeRepository) Create(ctx context.Context, checkupType *model.CheckupType) error {
	if err := r.db.WithContext(ctx).Create(checkupType).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("checkup_type", checkupType.Name)
		}
		return apperrors.Wrap(err, "create checkup type")
	}
	return nil
}

func (r *checkupTypeRepository) Update(ctx context.Context, checkupType *model.CheckupType) error {
	result := r.db.WithContext(ctx).
		Model(&model.CheckupType{}).
		Where("id = ? AND clinic_id = ?", checkupType.ID, checkupType.ClinicID).
		Updates(checkupType)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update checkup type")
	}
	if result.RowsAffected == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "update checkup type")
	}
	return nil
}

func (r *checkupTypeRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.CheckupType{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete checkup type")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("checkup_type", fmt.Sprintf("%d", id))
	}
	return nil
}
