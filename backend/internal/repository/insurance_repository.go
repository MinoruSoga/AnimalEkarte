// Package repository provides data access implementations for Insurance entity.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Insurance ----

type InsuranceRepository interface {
	FindAll(ctx context.Context) ([]model.Insurance, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Insurance, error)
	Create(ctx context.Context, insurance *model.Insurance) error
	Update(ctx context.Context, insurance *model.Insurance) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type insuranceRepository struct{ db *gorm.DB }

func NewInsuranceRepository(db *gorm.DB) InsuranceRepository { return &insuranceRepository{db: db} }

func (r *insuranceRepository) FindAll(ctx context.Context) ([]model.Insurance, error) {
	var insurances []model.Insurance
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&insurances).Error; err != nil {
		return nil, apperrors.Wrap(err, "find insurances")
	}
	return insurances, nil
}

func (r *insuranceRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Insurance, error) {
	var insurance model.Insurance
	if err := r.db.WithContext(ctx).First(&insurance, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("insurance", id.String())
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
	if err := r.db.WithContext(ctx).Save(insurance).Error; err != nil {
		return apperrors.Wrap(err, "update insurance")
	}
	return nil
}

func (r *insuranceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Insurance{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete insurance")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("insurance", id.String())
	}
	return nil
}
