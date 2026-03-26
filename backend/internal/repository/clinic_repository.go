package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ClinicRepository interface {
	FindAll(ctx context.Context) ([]model.Clinic, error)
	FindByID(ctx context.Context, id uint64) (*model.Clinic, error)
	GetCompany(ctx context.Context) (*model.Company, error)
	Create(ctx context.Context, clinic *model.Clinic) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
}

type clinicRepository struct {
	db *gorm.DB
}

func NewClinicRepository(db *gorm.DB) ClinicRepository {
	return &clinicRepository{db: db}
}

func (r *clinicRepository) FindAll(ctx context.Context) ([]model.Clinic, error) {
	clinics := make([]model.Clinic, 0)
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&clinics).Error; err != nil {
		return nil, apperrors.Wrap(err, "find clinics")
	}
	return clinics, nil
}

func (r *clinicRepository) FindByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	var clinic model.Clinic
	if err := r.db.WithContext(ctx).First(&clinic, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("clinic", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find clinic by id")
	}
	return &clinic, nil
}

func (r *clinicRepository) GetCompany(ctx context.Context) (*model.Company, error) {
	var company model.Company
	if err := r.db.WithContext(ctx).First(&company).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("company", "singleton")
		}
		return nil, apperrors.Wrap(err, "get company")
	}
	return &company, nil
}

func (r *clinicRepository) Create(ctx context.Context, clinic *model.Clinic) error {
	if err := r.db.WithContext(ctx).Create(clinic).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("clinic", clinic.Name)
		}
		return apperrors.Wrap(err, "create clinic")
	}
	return nil
}

func (r *clinicRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).Model(&model.Clinic{}).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update clinic")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinic", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *clinicRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Clinic{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete clinic")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinic", fmt.Sprintf("%d", id))
	}
	return nil
}
