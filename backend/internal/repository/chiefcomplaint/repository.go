// Package chiefcomplaint owns chief_complaint_types master data access.
package chiefcomplaint

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for chief complaint types.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error)
	CountUsageByChiefComplaintTypeID(ctx context.Context, clinicID, id uint64) (int64, error)
	Create(ctx context.Context, category *model.ChiefComplaintType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ChiefComplaintType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error) {
	categories := make([]model.ChiefComplaintType, 0)
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "chief_complaint_type", "")
	}
	return categories, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error) {
	return repohelpers.FindByIDScoped[model.ChiefComplaintType](ctx, r.db, "chief_complaint_type", clinicID, id)
}

func (r *repository) Create(ctx context.Context, category *model.ChiefComplaintType) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		return apperrors.FromGORM(err, "chief_complaint_type", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ChiefComplaintType, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, ids, "sort_order")
}

// CountUsageByChiefComplaintTypeID returns inquiry references.
// inquiries lack clinic_id; tenant isolation is via medical_records JOIN.
func (r *repository) CountUsageByChiefComplaintTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Inquiry{}).
		Scopes(repohelpers.MedicalRecordTenantScope("inquiries", clinicID)).
		Where("inquiries.chief_complaint_type_id = ? AND inquiries.deleted_at IS NULL", id).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "inquiry", "")
	}
	return count, nil
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, id)
}
