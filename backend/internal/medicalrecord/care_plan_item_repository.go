package medicalrecord

// Moved from internal/repository (BE9-2D ⑤ Batch A). 旧 package-private helper は repohelpers 同等物へ
// 置換（同一述語/ambient-tx参加）。外部呼び出しは internal/repository の facade alias 経由で不変。

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// CarePlanItemRepository はケアプランアイテムのデータアクセスインターフェース
type CarePlanItemRepository interface {
	FindByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.CarePlanItem, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.CarePlanItem, error)
	Create(ctx context.Context, item *model.CarePlanItem) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type carePlanItemRepository struct {
	db *gorm.DB
}

// NewCarePlanItemRepository は CarePlanItemRepository を初期化して返す
func NewCarePlanItemRepository(db *gorm.DB) CarePlanItemRepository {
	return &carePlanItemRepository{db: db}
}

// FindByHospitalizationID は dbOrTx で ambient tx に参加する（BE9-2D ⑤: DischargeWithBilling が
// FOR UPDATE 保持中に読み billing_items へ変換する read＝旧 repos.Transaction の tx-bound clone と等価維持）。
func (r *carePlanItemRepository) FindByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.CarePlanItem, error) {
	items := make([]model.CarePlanItem, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id AND hospitalizations.deleted_at IS NULL").
		Where("hospitalizations.clinic_id = ? AND care_plan_items.hospitalization_id = ?", clinicID, hospitalizationID).
		Order("care_plan_items.sort_order ASC").
		Preload("Medicine", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Procedure", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return items, nil
}

// FindByID participates in an ambient transaction so Create/Update can complete
// response re-fetch before commit (MRA-02).
func (r *carePlanItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.CarePlanItem, error) {
	var item model.CarePlanItem
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id AND hospitalizations.deleted_at IS NULL").
		Where("hospitalizations.clinic_id = ? AND care_plan_items.id = ?", clinicID, id).
		Preload("Medicine", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Procedure", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		First(&item).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "care_plan_item", fmt.Sprintf("%d", id))
	}
	return &item, nil
}

func (r *carePlanItemRepository) Create(ctx context.Context, item *model.CarePlanItem) error {
	err := persistence.DBOrTx(ctx, r.db).Create(item).Error
	if err != nil {
		return apperrors.FromGORM(err, "care_plan_item", "")
	}
	return nil
}

func (r *carePlanItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	// NOTE: GORM does not propagate Joins() into the generated UPDATE statement's SQL
	// (it is a SELECT-only clause), so a WHERE referencing the joined table fails with
	// "missing FROM-clause entry". clinic_id isolation must be expressed as a subquery instead.
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.CarePlanItem{}).
		Where("care_plan_items.id = ? AND care_plan_items.hospitalization_id IN (SELECT id FROM hospitalizations WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "care_plan_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("care_plan_item", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *carePlanItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	// NOTE: see Update — Joins() does not propagate into DELETE's SQL either.
	// .Unscoped() is a no-op here (CarePlanItem has no DeletedAt field, i.e. it is a hard-delete
	// model already) but is kept to make that intent explicit for future maintainers.
	// Hard-delete durability for MRA-01 is provided by fail-closed audit of the pre-delete snapshot.
	result := persistence.DBOrTx(ctx, r.db).
		Where("care_plan_items.id = ? AND care_plan_items.hospitalization_id IN (SELECT id FROM hospitalizations WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID).
		Unscoped().
		Delete(&model.CarePlanItem{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "care_plan_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("care_plan_item", fmt.Sprintf("%d", id))
	}
	return nil
}
