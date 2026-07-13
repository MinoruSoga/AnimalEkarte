// Package repository provides data access implementations for ChiefComplaintType entity.
package repository

import (
	"context"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ChiefComplaintType ----

type ChiefComplaintTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error)
	CountUsageByChiefComplaintTypeID(ctx context.Context, clinicID, id uint64) (int64, error)
	Create(ctx context.Context, category *model.ChiefComplaintType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ChiefComplaintType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type chiefComplaintTypeRepository struct{ db *gorm.DB }

func NewChiefComplaintTypeRepository(db *gorm.DB) ChiefComplaintTypeRepository {
	return &chiefComplaintTypeRepository{db: db}
}

func (r *chiefComplaintTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error) {
	categories := make([]model.ChiefComplaintType, 0)
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "chief_complaint_type", "")
	}
	return categories, nil
}

func (r *chiefComplaintTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error) {
	return findByIDScoped[model.ChiefComplaintType](ctx, r.db, "chief_complaint_type", clinicID, id)
}

func (r *chiefComplaintTypeRepository) Create(ctx context.Context, category *model.ChiefComplaintType) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		return apperrors.FromGORM(err, "chief_complaint_type", "")
	}
	return nil
}

func (r *chiefComplaintTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ChiefComplaintType, error) {
	if err := updateScopedByID(ctx, r.db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *chiefComplaintTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, ids, "sort_order")
}

// CountUsageByChiefComplaintTypeID は主訴区分を参照している inquiries の件数を返す。
// inquiries テーブルは clinic_id を持たないため medical_records を JOIN してテナント分離する。
func (r *chiefComplaintTypeRepository) CountUsageByChiefComplaintTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Inquiry{}).
		Scopes(medicalRecordTenantScope("inquiries", clinicID)).
		Where("inquiries.chief_complaint_type_id = ? AND inquiries.deleted_at IS NULL", id).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "inquiry", "")
	}
	return count, nil
}

func (r *chiefComplaintTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return deleteScopedByID(ctx, r.db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, id)
}
