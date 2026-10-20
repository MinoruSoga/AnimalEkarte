// Package procedure owns procedures master data access.
package medicalrecord

// Moved from internal/repository/procedure (BE9-2D ⑥ Batch A roll-up・BE8-4 subpackage)。
// generic Repository/New は entity-specific 名へ改名（①⑤先例）— 外部は facade alias 経由で不変。

import (
	"context"
	"fmt"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
	db := persistence.DBOrTx(ctx, r.db)
	return persistence.FindByIDScoped[model.Procedure](ctx, db, "procedure", clinicID, id)
}

// LockWriteRows locks target and requested parent before validation/mutation.
func (r *procedureRepositoryImpl) LockWriteRows(ctx context.Context, clinicID uint64, ids []uint64) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("master parent lock requires an ambient transaction")
	}
	// Sort again at the repository boundary; callers cannot invert the lock order.
	ordered := append([]uint64(nil), ids...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	for i, id := range ordered {
		if id == 0 {
			return apperrors.WrapInvalidInput("master write row id must be positive")
		}
		if i > 0 && id == ordered[i-1] {
			continue
		}
		db := persistence.DBOrTx(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"})
		if _, err := persistence.FindByIDScoped[model.Procedure](ctx, db, "procedure", clinicID, id); err != nil {
			return err
		}
	}
	return nil
}

// LockByIDForUpdate locks only the mutation target for an exact audit before-image.
func (r *procedureRepositoryImpl) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Procedure, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("master insurance lock requires an ambient transaction")
	}
	db := persistence.DBOrTx(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"})
	return persistence.FindByIDScoped[model.Procedure](ctx, db, "procedure", clinicID, id)
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

// Update compares the editing-start version; the storage trigger owns increments.
func (r *procedureRepositoryImpl) Update(ctx context.Context, clinicID, id uint64, cmd UpdateProcedureInput) (*model.Procedure, error) {
	if cmd.Version < 1 {
		return nil, apperrors.WrapInvalidInput("positive editing-start version is required")
	}
	fields := buildProcedureUpdate(&cmd)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	result := persistence.DBOrTx(ctx, r.db).Model(&model.Procedure{}).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ? AND version = ?", id, cmd.Version).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "procedure", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		if _, err := r.FindByID(ctx, clinicID, id); err != nil {
			return nil, err
		}
		return nil, apperrors.WrapConflict("編集後に変更されています。再読み込みしてから保存してください")
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *procedureRepositoryImpl) Delete(ctx context.Context, clinicID, id uint64) error {
	if persistence.TxFromContext(ctx) == nil {
		return persistence.NewTransactor(r.db).WithTx(ctx, func(txCtx context.Context) error { return r.Delete(txCtx, clinicID, id) })
	}
	// Acquire the target before starting the guarded mutation statement. A wait
	// must finish before its READ COMMITTED snapshot checks newly committed children.
	if err := r.LockWriteRows(ctx, clinicID, []uint64{id}); err != nil {
		return err
	}
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
	return apperrors.WrapConflict("この診療項目はカルテで使用中のため削除できません")
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
