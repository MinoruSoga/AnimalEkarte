// Package repository provides data access implementations for Insurance entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Insurance ----

type InsuranceRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Insurance, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error)
	Create(ctx context.Context, insurance *model.Insurance) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Insurance, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByInsuranceID(ctx context.Context, clinicID, insuranceID uint64) (int64, error)
}

type insuranceRepository struct{ db *gorm.DB }

func NewInsuranceRepository(db *gorm.DB) InsuranceRepository { return &insuranceRepository{db: db} }

func (r *insuranceRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Insurance, error) {
	insurances := make([]model.Insurance, 0)
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&insurances).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "insurance", "")
	}
	return insurances, nil
}

func (r *insuranceRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error) {
	var insurance model.Insurance
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&insurance).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "insurance", fmt.Sprintf("%d", id))
	}
	return &insurance, nil
}

func (r *insuranceRepository) Create(ctx context.Context, insurance *model.Insurance) error {
	err := r.db.WithContext(ctx).Create(insurance).Error
	if err != nil {
		return apperrors.FromGORM(err, "insurance", "")
	}
	return nil
}

func (r *insuranceRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Insurance, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Insurance{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "insurance", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("insurance", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *insuranceRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.Insurance{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "insurance", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("insurance", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountUsageByInsuranceID は指定保険を参照しているペット数を返す（BUG-110）
// pets テーブルは直接 clinic_id を持つためテナント分離を直接適用する
func (r *insuranceRepository) CountUsageByInsuranceID(ctx context.Context, clinicID, insuranceID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Pet{}).
		Scopes(clinicScope(clinicID)).
		Where("insurance_id = ? AND deleted_at IS NULL", insuranceID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "pet", "")
	}
	return count, nil
}

func (r *insuranceRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.Insurance{}, "insurance", clinicID, ids)
}
