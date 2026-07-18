// Package procedure owns procedures master data access.
package procedure

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for procedures.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Procedure, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error)
	Create(ctx context.Context, procedure *model.Procedure) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Procedure, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByProcedureID(ctx context.Context, clinicID, procedureID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64) ([]model.Procedure, error) {
	procedures := make([]model.Procedure, 0)
	if err := r.db.WithContext(ctx).Scopes(repohelpers.ClinicScope(clinicID)).Order("sort_order ASC, name ASC").Limit(repohelpers.MaxMasterListRows).Find(&procedures).Error; err != nil {
		return nil, apperrors.FromGORM(err, "procedure", "")
	}
	return procedures, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error) {
	return repohelpers.FindByIDScoped[model.Procedure](ctx, r.db, "procedure", clinicID, id)
}

func (r *repository) Create(ctx context.Context, procedure *model.Procedure) error {
	if err := r.db.WithContext(ctx).Create(procedure).Error; err != nil {
		return apperrors.FromGORM(err, "procedure", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Procedure, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.Procedure{}, "procedure", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.Procedure{}, "procedure", clinicID, id)
}

// CountUsageByProcedureID は treatments と care_plan_items で参照されている件数の合計を返す（BUG-107）
// treatments/care_plan_items は直接 clinic_id を持たないため JOIN でテナント分離する
func (r *repository) CountUsageByProcedureID(ctx context.Context, clinicID, procedureID uint64) (int64, error) {
	var treatmentCount, carePlanCount int64
	if err := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Scopes(repohelpers.MedicalRecordTenantScope("treatments", clinicID)).
		Where("treatments.procedure_id = ? AND treatments.deleted_at IS NULL", procedureID).
		Count(&treatmentCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "treatment", "")
	}
	if err := r.db.WithContext(ctx).
		Model(&model.CarePlanItem{}).
		Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id AND hospitalizations.clinic_id = ? AND hospitalizations.deleted_at IS NULL", clinicID).
		Where("care_plan_items.procedure_id = ? AND care_plan_items.deleted_at IS NULL", procedureID).
		Count(&carePlanCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return treatmentCount + carePlanCount, nil
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.Procedure{}, "procedure", clinicID, ids, "sort_order")
}

// CountChildrenByParentID は指定した処置の子処置数を返す (BUG-390)
func (r *repository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Procedure{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "procedure", fmt.Sprintf("%d", parentID))
	}
	return count, nil
}
