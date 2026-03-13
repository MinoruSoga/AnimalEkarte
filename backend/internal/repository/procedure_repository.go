// Package repository provides data access implementations for Procedure entity.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Procedure ----

type ProcedureRepository interface {
	FindAll(ctx context.Context) ([]model.Procedure, error)
	FindByID(ctx context.Context, id uint64) (*model.Procedure, error)
	Create(ctx context.Context, procedure *model.Procedure) error
	Update(ctx context.Context, procedure *model.Procedure) error
	Delete(ctx context.Context, id uint64) error
}

type procedureRepository struct{ db *gorm.DB }

func NewProcedureRepository(db *gorm.DB) ProcedureRepository { return &procedureRepository{db: db} }

func (r *procedureRepository) FindAll(ctx context.Context) ([]model.Procedure, error) {
	var procedures []model.Procedure
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&procedures).Error; err != nil {
		return nil, apperrors.Wrap(err, "find procedures")
	}
	return procedures, nil
}

func (r *procedureRepository) FindByID(ctx context.Context, id uint64) (*model.Procedure, error) {
	var procedure model.Procedure
	if err := r.db.WithContext(ctx).First(&procedure, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("procedure", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find procedure by id")
	}
	return &procedure, nil
}

func (r *procedureRepository) Create(ctx context.Context, procedure *model.Procedure) error {
	if err := r.db.WithContext(ctx).Create(procedure).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("procedure", procedure.Name)
		}
		return apperrors.Wrap(err, "create procedure")
	}
	return nil
}

func (r *procedureRepository) Update(ctx context.Context, procedure *model.Procedure) error {
	result := r.db.WithContext(ctx).
		Model(&model.Procedure{}).
		Where("id = ? AND clinic_id = ?", procedure.ID, procedure.ClinicID).
		Updates(procedure)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update procedure")
	}
	if result.RowsAffected == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "update procedure")
	}
	return nil
}

func (r *procedureRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Procedure{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete procedure")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("procedure", fmt.Sprintf("%d", id))
	}
	return nil
}
