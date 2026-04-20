// Package repository provides data access implementations for Occupation entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Occupation ----

type OccupationRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Occupation, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error)
	Create(ctx context.Context, occupation *model.Occupation) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByOccupationID(ctx context.Context, clinicID, occupationID uint64) (int64, error)
}

type occupationRepository struct{ db *gorm.DB }

func NewOccupationRepository(db *gorm.DB) OccupationRepository {
	return &occupationRepository{db: db}
}

func (r *occupationRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Occupation, error) {
	occupations := make([]model.Occupation, 0)
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&occupations).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "occupation", "")
	}
	return occupations, nil
}

func (r *occupationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error) {
	var occupation model.Occupation
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&occupation).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "occupation", fmt.Sprintf("%d", id))
	}
	return &occupation, nil
}

func (r *occupationRepository) Create(ctx context.Context, occupation *model.Occupation) error {
	err := r.db.WithContext(ctx).Create(occupation).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "occupation", "")
	}
	return nil
}

func (r *occupationRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Occupation{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "occupation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("occupation", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *occupationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.Occupation{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "occupation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("occupation", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountUsageByOccupationID は指定役職を参照しているスタッフ数を返す（BUG-112）
// staffs テーブルに直接 clinic_id がないため staff_clinic_assignments を JOIN してテナント分離する
func (r *occupationRepository) CountUsageByOccupationID(ctx context.Context, clinicID, occupationID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Staff{}).
		Joins("JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL", clinicID).
		Where("staffs.occupation_id = ? AND staffs.deleted_at IS NULL", occupationID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "staff", "")
	}
	return count, nil
}

func (r *occupationRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.Occupation{}, "occupation", clinicID, ids)
}
