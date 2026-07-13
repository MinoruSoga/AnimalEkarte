// Package repository provides data access implementations for Vaccine entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Vaccine ----

type VaccineRepository interface {
	FindAll(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error)
	Create(ctx context.Context, vaccine *model.Vaccine) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccine, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByVaccineID(ctx context.Context, clinicID, vaccineID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type vaccineRepository struct{ db *gorm.DB }

func NewVaccineRepository(db *gorm.DB) VaccineRepository { return &vaccineRepository{db: db} }

func (r *vaccineRepository) FindAll(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error) {
	vaccines := make([]model.Vaccine, 0)
	q := r.db.WithContext(ctx).Model(&model.Vaccine{}).Scopes(clinicScope(clinicID))
	if species != nil {
		q = q.Where("species = ?", *species)
	}
	if err := q.Order("sort_order ASC, name ASC").Limit(maxMasterListRows).Find(&vaccines).Error; err != nil {
		return nil, apperrors.FromGORM(err, "vaccine", "")
	}
	return vaccines, nil
}

func (r *vaccineRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error) {
	return findByIDScoped[model.Vaccine](ctx, r.db, "vaccine", clinicID, id)
}

func (r *vaccineRepository) Create(ctx context.Context, vaccine *model.Vaccine) error {
	if err := r.db.WithContext(ctx).Create(vaccine).Error; err != nil {
		return apperrors.FromGORM(err, "vaccine", "")
	}
	return nil
}

func (r *vaccineRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccine, error) {
	if err := updateScopedByID(ctx, r.db, &model.Vaccine{}, "vaccine", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *vaccineRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return deleteScopedByID(ctx, r.db, &model.Vaccine{}, "vaccine", clinicID, id)
}

func (r *vaccineRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.Vaccine{}, "vaccine", clinicID, ids, "sort_order")
}

// CountUsageByVaccineID はワクチンマスタを参照している vaccinations の件数を返す（BUG-107）
// vaccinations テーブルは直接 clinic_id を持つためテナント分離を直接適用する
func (r *vaccineRepository) CountUsageByVaccineID(ctx context.Context, clinicID, vaccineID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Vaccination{}).
		Scopes(clinicScope(clinicID)).
		Where("vaccine_id = ? AND deleted_at IS NULL", vaccineID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "vaccination_record", "")
	}
	return count, nil
}

// CountChildrenByParentID は指定したワクチンの子ワクチン数を返す (BUG-390)
func (r *vaccineRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Vaccine{}).
		Scopes(clinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "vaccine", fmt.Sprintf("%d", parentID))
	}
	return count, nil
}
