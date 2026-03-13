// Package repository provides data access implementations for ServiceType entity.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ServiceType ----

type ServiceTypeRepository interface {
	FindAll(ctx context.Context) ([]model.ServiceType, error)
	FindByID(ctx context.Context, id uint64) (*model.ServiceType, error)
	Create(ctx context.Context, serviceType *model.ServiceType) error
	Update(ctx context.Context, serviceType *model.ServiceType) error
	Delete(ctx context.Context, id uint64) error
}

type serviceTypeRepository struct{ db *gorm.DB }

func NewServiceTypeRepository(db *gorm.DB) ServiceTypeRepository {
	return &serviceTypeRepository{db: db}
}

func (r *serviceTypeRepository) FindAll(ctx context.Context) ([]model.ServiceType, error) {
	var serviceTypes []model.ServiceType
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&serviceTypes).Error; err != nil {
		return nil, apperrors.Wrap(err, "find service types")
	}
	return serviceTypes, nil
}

func (r *serviceTypeRepository) FindByID(ctx context.Context, id uint64) (*model.ServiceType, error) {
	var serviceType model.ServiceType
	if err := r.db.WithContext(ctx).First(&serviceType, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("service_type", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find service type by id")
	}
	return &serviceType, nil
}

func (r *serviceTypeRepository) Create(ctx context.Context, serviceType *model.ServiceType) error {
	if err := r.db.WithContext(ctx).Create(serviceType).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("service_type", serviceType.Name)
		}
		return apperrors.Wrap(err, "create service type")
	}
	return nil
}

func (r *serviceTypeRepository) Update(ctx context.Context, serviceType *model.ServiceType) error {
	if err := r.db.WithContext(ctx).Save(serviceType).Error; err != nil {
		return apperrors.Wrap(err, "update service type")
	}
	return nil
}

func (r *serviceTypeRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.ServiceType{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete service type")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("service_type", fmt.Sprintf("%d", id))
	}
	return nil
}
