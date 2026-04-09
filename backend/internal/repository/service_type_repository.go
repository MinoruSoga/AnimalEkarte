// Package repository provides data access implementations for ServiceType entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ServiceType ----

type ServiceTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ServiceType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ServiceType, error)
	Create(ctx context.Context, serviceType *model.ServiceType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type serviceTypeRepository struct{ db *gorm.DB }

func NewServiceTypeRepository(db *gorm.DB) ServiceTypeRepository {
	return &serviceTypeRepository{db: db}
}

func (r *serviceTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ServiceType, error) {
	serviceTypes := make([]model.ServiceType, 0)
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("sort_order ASC, name ASC").
		Find(&serviceTypes).Error; err != nil {
		return nil, apperrors.FromGORM(err, "service_type", "")
	}
	return serviceTypes, nil
}

func (r *serviceTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ServiceType, error) {
	var st model.ServiceType
	err := r.db.WithContext(ctx).First(&st, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "service_type", fmt.Sprintf("%d", id))
	}
	return &st, nil
}

func (r *serviceTypeRepository) Create(ctx context.Context, serviceType *model.ServiceType) error {
	if err := r.db.WithContext(ctx).Create(serviceType).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "service_type", "")
	}
	return nil
}

func (r *serviceTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ServiceType{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "service_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := r.db.WithContext(ctx).Model(&model.ServiceType{}).
			Where("id = ? AND clinic_id = ?", id, clinicID).
			Count(&count).Error; err != nil {
			return apperrors.FromGORM(err, "service_type", fmt.Sprintf("%d", id))
		}
		if count == 0 {
			return apperrors.WrapNotFound("service_type", fmt.Sprintf("%d", id))
		}
	}
	return nil
}

func (r *serviceTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.ServiceType{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		// BUG-030: ON DELETE RESTRICT の FK 制約違反は 409 Conflict に変換する
		if isFKConstraintErr(result.Error) {
			return apperrors.WrapConflict("このサービス種別は予約に使用されているため削除できません")
		}
		return apperrors.FromGORM(result.Error, "service_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("service_type", fmt.Sprintf("%d", id))
	}
	return nil
}

// Reorder はトランザクション内で予約区分の sort_order を ids の順序で更新する
func (r *serviceTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.ServiceType{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "service_type", fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("service_type id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
