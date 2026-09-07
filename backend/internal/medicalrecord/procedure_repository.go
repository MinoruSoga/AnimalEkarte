// Package procedure owns procedures master data access.
package medicalrecord

// Moved from internal/repository/procedure (BE9-2D ⑥ Batch A roll-up・BE8-4 subpackage)。
// generic Repository/New は entity-specific 名へ改名（①⑤先例）— 外部は facade alias 経由で不変。

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Repository is the data access interface for procedures.
type ProcedureRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Procedure, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error)
	Create(ctx context.Context, procedure *model.Procedure) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateProcedureInput) (*model.Procedure, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByProcedureID(ctx context.Context, clinicID, procedureID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type procedureRepositoryImpl struct{ db *gorm.DB }

// New constructs a Repository.
func NewProcedureRepository(db *gorm.DB) ProcedureRepository {
	return &procedureRepositoryImpl{db: db}
}

func (r *procedureRepositoryImpl) FindAll(ctx context.Context, clinicID uint64) ([]model.Procedure, error) {
	procedures := make([]model.Procedure, 0)
	if err := persistence.DBOrTx(ctx, r.db).Scopes(persistence.ClinicScope(clinicID)).Order("sort_order ASC, name ASC").Limit(persistence.MaxMasterListRows).Find(&procedures).Error; err != nil {
		return nil, apperrors.FromGORM(err, "procedure", "")
	}
	return procedures, nil
}

func (r *procedureRepositoryImpl) FindByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error) {
	// MRC-07: ambient-tx participation for delete usage re-check under WithTx.
	return persistence.FindByIDScoped[model.Procedure](ctx, persistence.DBOrTx(ctx, r.db), "procedure", clinicID, id)
}

func (r *procedureRepositoryImpl) Create(ctx context.Context, procedure *model.Procedure) error {
	db := persistence.DBOrTx(ctx, r.db)
	wantActive := procedure.IsActive
	if err := db.Create(procedure).Error; err != nil {
		return apperrors.FromGORM(err, "procedure", "")
	}
	if !wantActive {
		if err := db.Model(procedure).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "procedure", fmt.Sprintf("%d", procedure.ID))
		}
		procedure.IsActive = false
	}
	return nil
}

func (r *procedureRepositoryImpl) Update(ctx context.Context, clinicID, id uint64, cmd UpdateProcedureInput) (*model.Procedure, error) {
	if err := r.update(ctx, clinicID, id, buildProcedureUpdate(&cmd)); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *procedureRepositoryImpl) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.Procedure{}, "procedure", clinicID, id, fields)
}

func (r *procedureRepositoryImpl) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM procedures children
			WHERE children.parent_id = procedures.id
			  AND children.clinic_id = ?
			  AND children.deleted_at IS NULL
		)`, clinicID).
		Where(`NOT EXISTS (
			SELECT 1 FROM treatments
			JOIN medical_records ON medical_records.id = treatments.medical_record_id
			  AND medical_records.clinic_id = ?
			  AND medical_records.deleted_at IS NULL
			WHERE treatments.procedure_id = procedures.id
			  AND treatments.deleted_at IS NULL
		)`, clinicID).
		Where(`NOT EXISTS (
			SELECT 1 FROM care_plan_items
			JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id
			  AND hospitalizations.clinic_id = ?
			  AND hospitalizations.deleted_at IS NULL
			WHERE care_plan_items.procedure_id = procedures.id
		)`, clinicID).
		Delete(&model.Procedure{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "procedure", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeProcedureDeleteMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *procedureRepositoryImpl) normalizeProcedureDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	childCount, err := r.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return apperrors.WrapConflict("この処置は子処置が存在するため削除できません")
	}
	return apperrors.WrapConflict("この診療項目は診療記録で使用中のため削除できません")
}

// CountUsageByProcedureID は treatments と care_plan_items で参照されている件数の合計を返す（BUG-107）
// treatments/care_plan_items は直接 clinic_id を持たないため JOIN でテナント分離する
func (r *procedureRepositoryImpl) CountUsageByProcedureID(ctx context.Context, clinicID, procedureID uint64) (int64, error) {
	var treatmentCount, carePlanCount int64
	db := persistence.DBOrTx(ctx, r.db)
	if err := db.
		Model(&model.Treatment{}).
		Scopes(persistence.MedicalRecordTenantScope("treatments", clinicID)).
		Where("treatments.procedure_id = ? AND treatments.deleted_at IS NULL", procedureID).
		Count(&treatmentCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "treatment", "")
	}
	if err := db.
		Model(&model.CarePlanItem{}).
		Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id AND hospitalizations.clinic_id = ? AND hospitalizations.deleted_at IS NULL", clinicID).
		Where("care_plan_items.procedure_id = ?", procedureID).
		Count(&carePlanCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return treatmentCount + carePlanCount, nil
}

func (r *procedureRepositoryImpl) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.Procedure{}, "procedure", clinicID, ids, "sort_order")
}

// CountChildrenByParentID は指定した処置の子処置数を返す (BUG-390)
func (r *procedureRepositoryImpl) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Procedure{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "procedure", fmt.Sprintf("%d", parentID))
	}
	return count, nil
}
