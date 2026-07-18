// Package consultation owns consultations master data access.
package consultation

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for consultations.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Consultation, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error)
	Create(ctx context.Context, consultation *model.Consultation) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByConsultationID(ctx context.Context, clinicID, consultationID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64) ([]model.Consultation, error) {
	consultations := make([]model.Consultation, 0)
	err := r.db.WithContext(ctx).Scopes(repohelpers.ClinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&consultations).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "consultation", "")
	}
	return consultations, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	return repohelpers.FindByIDScoped[model.Consultation](ctx, r.db, "consultation", clinicID, id)
}

func (r *repository) Create(ctx context.Context, consultation *model.Consultation) error {
	err := r.db.WithContext(ctx).Create(consultation).Error
	if err != nil {
		return apperrors.FromGORM(err, "consultation", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.Consultation{}, "consultation", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.Consultation{}, "consultation", clinicID, id)
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.Consultation{}, "consultation", clinicID, ids, "sort_order")
}

// CountChildrenByParentID は指定した親 ID を持つ子診察項目の件数を返す。
// 親を削除する前に孤立子が残らないことを確認するために使用する。
func (r *repository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Consultation{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "consultation", fmt.Sprintf("%d", parentID))
	}
	return count, nil
}

// CountUsageByConsultationID は診察マスタを参照している treatments の件数を返す（BUG-107）
// treatments テーブルに直接 clinic_id がないため medical_records を JOIN してテナント分離する
func (r *repository) CountUsageByConsultationID(ctx context.Context, clinicID, consultationID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Scopes(repohelpers.MedicalRecordTenantScope("treatments", clinicID)).
		Where("treatments.consultation_id = ? AND treatments.deleted_at IS NULL", consultationID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "treatment", "")
	}
	return count, nil
}
