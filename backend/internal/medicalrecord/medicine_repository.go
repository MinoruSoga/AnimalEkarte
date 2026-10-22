// Package medicalrecord provides medicine persistence.
package medicalrecord

// Moved from internal/repository (BE9-2D ⑥ Batch A). 旧 package-private helper は repohelpers
// 同等物へ置換（同一述語/ambient-tx参加）。外部は internal/repository の facade alias 経由で不変。

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ---- Medicine ----

type MedicineRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
	CountUsageByMedicineID(ctx context.Context, clinicID, medicineID uint64) (int64, error)
	Create(ctx context.Context, medicine *model.Medicine) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateMedicineInput) (*model.Medicine, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type medicineRepository struct{ db *gorm.DB }

func NewMedicineRepository(db *gorm.DB) MedicineRepository { return &medicineRepository{db: db} }

// medicineReadColumns は EMR-447 A.3 の read 証拠。is_bill_discount_excluded / version を
// 列挙することで列未接続（欠落）が query error になり、DB NULL との区別がつく。
var medicineReadColumns = []string{
	"id", "clinic_id", "name", "price", "is_active", "description",
	"parent_id", "dosage_form", "medicine_unit", "inventory_id",
	"default_quantity", "tax_type", "tax_rate", "sort_order", "is_non_insurance",
	"calculation_type", "strength", "frequency_per_day", "default_duration_days",
	"is_bill_discount_excluded", "version",
	"created_at", "updated_at", "deleted_at",
}

func (r *medicineRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error) {
	medicines := make([]model.Medicine, 0)
	var total int64

	buildBase := func() *gorm.DB {
		return persistence.DBOrTx(ctx, r.db).Model(&model.Medicine{}).Scopes(persistence.ClinicScope(clinicID))
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medicine", "")
	}
	if err := buildBase().
		Select(medicineReadColumns).
		Scopes(paginate(page, limit)).
		Order("sort_order ASC, name ASC").
		Find(&medicines).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medicine", "")
	}
	return medicines, total, nil
}

func (r *medicineRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error) {
	var medicine model.Medicine
	if err := persistence.DBOrTx(ctx, r.db).
		Select(medicineReadColumns).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&medicine).Error; err != nil {
		return nil, apperrors.FromGORM(err, "medicine", fmt.Sprintf("%d", id))
	}
	return &medicine, nil
}

// CountUsageByMedicineID は treatments と care_plan_items で参照されている件数の合計を返す（BUG-108）
// clinic_id フィルタを JOIN で適用しテナント分離を保証する（BUG-377）
func (r *medicineRepository) CountUsageByMedicineID(ctx context.Context, clinicID, medicineID uint64) (int64, error) {
	var treatmentCount, carePlanCount int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Treatment{}).
		Scopes(persistence.MedicalRecordTenantScope("treatments", clinicID)).
		Where("treatments.medicine_id = ? AND treatments.deleted_at IS NULL", medicineID).
		Count(&treatmentCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "treatment", "")
	}
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.CarePlanItem{}).
		Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id AND hospitalizations.clinic_id = ? AND hospitalizations.deleted_at IS NULL", clinicID).
		Where("care_plan_items.medicine_id = ?", medicineID).
		Count(&carePlanCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return treatmentCount + carePlanCount, nil
}

func (r *medicineRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Medicine{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "medicine", "")
	}
	return count, nil
}

func (r *medicineRepository) Create(ctx context.Context, medicine *model.Medicine) error {
	db := persistence.DBOrTx(ctx, r.db)
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	// Use same DBOrTx handle so compensation stays inside ambient multi-write tx.
	wantActive := medicine.IsActive
	if err := db.Create(medicine).Error; err != nil {
		return apperrors.FromGORM(err, "medicine", "")
	}
	if !wantActive {
		if err := db.Model(medicine).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "medicine", fmt.Sprintf("%d", medicine.ID))
		}
		// この UPDATE は内容変更なので entity_version_guard が DB の version を加算する。
		// 応答・監査が古い版を返さないよう、同じ tx で再読取した行に揃える（EMR-447 A.4）。
		stored, err := r.FindByID(ctx, medicine.ClinicID, medicine.ID)
		if err != nil {
			return err
		}
		*medicine = *stored
	}
	return nil
}

// Update は ambient tx に参加し、cmd.Version が非 nil なら期待版を WHERE で照合する
// （EMR-447 A.4 楽観ロック・任意送信互換）。0 行は「存在しない（別医院含む）」と
// 「版不一致」に区別して返す。返却値は同じ tx 内で再読取し実保存版を返す。
func (r *medicineRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateMedicineInput) (*model.Medicine, error) {
	if cmd.Version != nil && *cmd.Version <= 0 {
		return nil, apperrors.WrapInvalidInput(errMsgMedicineVersionInvalid)
	}
	fields := buildMedicineUpdate(&cmd)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	q := persistence.DBOrTx(ctx, r.db).
		Model(&model.Medicine{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id)
	if cmd.Version != nil {
		q = q.Where("version = ?", *cmd.Version)
	}
	result := q.Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "medicine", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		if _, err := r.FindByID(ctx, clinicID, id); err != nil {
			return nil, err
		}
		return nil, apperrors.WrapConflict(errMsgMedicineVersionConflict)
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *medicineRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.Medicine{}, "medicine", clinicID, ids, "sort_order")
}

func (r *medicineRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM medicines children
			WHERE children.parent_id = medicines.id
			  AND children.clinic_id = ?
			  AND children.deleted_at IS NULL
		)`, clinicID).
		Where(`NOT EXISTS (
			SELECT 1 FROM treatments
			JOIN medical_records ON medical_records.id = treatments.medical_record_id
			  AND medical_records.clinic_id = ?
			  AND medical_records.deleted_at IS NULL
			WHERE treatments.medicine_id = medicines.id
			  AND treatments.deleted_at IS NULL
		)`, clinicID).
		Where(`NOT EXISTS (
			SELECT 1 FROM care_plan_items
			JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id
			  AND hospitalizations.clinic_id = ?
			  AND hospitalizations.deleted_at IS NULL
			WHERE care_plan_items.medicine_id = medicines.id
		)`, clinicID).
		Delete(&model.Medicine{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "medicine", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeMedicineDeleteMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *medicineRepository) normalizeMedicineDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	childCount, err := r.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return apperrors.WrapConflict(fmt.Sprintf("このカテゴリには%d件の薬剤が含まれています。先に薬剤を削除してください", childCount))
	}
	return apperrors.WrapConflict("この薬剤はカルテで使用中のため削除できません")
}
