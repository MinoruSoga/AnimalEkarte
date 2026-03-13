// Package repository provides data access implementations for Vaccine entity.
package repository

import (
	"context"
	"errors"
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
	Update(ctx context.Context, vaccine *model.Vaccine) error
	Delete(ctx context.Context, id uint64) error
}

type vaccineRepository struct{ db *gorm.DB }

func NewVaccineRepository(db *gorm.DB) VaccineRepository { return &vaccineRepository{db: db} }

func (r *vaccineRepository) FindAll(ctx context.Context, species *string) ([]model.Vaccine, error) {
	var vaccines []model.Vaccine
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
	if err := r.db.WithContext(ctx).First(&vaccine, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("vaccine", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find vaccine by id")
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

func (r *vaccineRepository) Update(ctx context.Context, vaccine *model.Vaccine) error {
	if err := r.db.WithContext(ctx).Save(vaccine).Error; err != nil {
		return apperrors.Wrap(err, "update vaccine")
	}
	return nil
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
