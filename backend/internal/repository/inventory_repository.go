package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type InventoryRepository interface {
	FindAll(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error)
	Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InventoryItem, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	DecreaseStock(ctx context.Context, id uint64, quantity float64) error
	CountUsageByInventoryID(ctx context.Context, clinicID, inventoryID uint64) (int64, error)
	// BUG-381: 薬剤マスタ削除時に BUG-320 で自動作成された連携在庫をカスケード削除するため、
	// (clinic_id, name, category=medicine) で在庫を削除する。マッチなしは no-op。
	DeleteByNameAndMedicineCategory(ctx context.Context, clinicID uint64, name string) error
	// TASK-215: 薬剤名変更時に BUG-320 で自動作成された連携在庫の name を同期する。
	// (clinic_id, oldName, category=medicine) にマッチするレコードを newName に更新する。マッチなしは no-op。
	UpdateNameByMedicineCategory(ctx context.Context, clinicID uint64, oldName, newName string) error
}

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) FindAll(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error) {
	items := make([]model.InventoryItem, 0)
	var total int64

	q := dbOrTx(ctx, r.db).Model(&model.InventoryItem{}).Scopes(clinicScope(clinicID))
	if category != nil {
		q = q.Where("category = ?", *category)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "inventory_item", "")
	}
	if err := q.Scopes(paginate(page, limit)).Order("name ASC").Find(&items).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "inventory_item", "")
	}
	return items, total, nil
}

func (r *inventoryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error) {
	return findByIDScoped[model.InventoryItem](ctx, dbOrTx(ctx, r.db), "inventory_item", clinicID, id)
}

func (r *inventoryRepository) Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error {
	item.ClinicID = clinicID
	err := dbOrTx(ctx, r.db).Create(item).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("inventory_item", item.Name)
		}
		return apperrors.FromGORM(err, "inventory_item", "")
	}
	return nil
}

func (r *inventoryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InventoryItem, error) {
	if err := updateScopedByID(ctx, dbOrTx(ctx, r.db), &model.InventoryItem{}, "inventory_item", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *inventoryRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return deleteScopedByID(ctx, dbOrTx(ctx, r.db), &model.InventoryItem{}, "inventory_item", clinicID, id)
}

func (r *inventoryRepository) DecreaseStock(ctx context.Context, id uint64, quantity float64) error {
	result := dbOrTx(ctx, r.db).
		Model(&model.InventoryItem{}).
		Where("id = ?", id).
		UpdateColumn("quantity", gorm.Expr("quantity - ?", int(quantity)))
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "inventory_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("inventory_item", fmt.Sprintf("%d", id))
	}
	return nil
}

// DeleteByNameAndMedicineCategory は BUG-320 で自動作成された在庫レコードを
// (clinic_id, name, category=medicine) で特定して削除する（BUG-381）。
// マッチなしは no-op（エラーなし）で返す。複数マッチは全件削除。
func (r *inventoryRepository) DeleteByNameAndMedicineCategory(ctx context.Context, clinicID uint64, name string) error {
	result := dbOrTx(ctx, r.db).
		Scopes(clinicScope(clinicID)).
		Where("name = ? AND category = ?", name, model.InventoryCategoryMedicine).
		Delete(&model.InventoryItem{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "inventory_item", fmt.Sprintf("name=%s", name))
	}
	return nil
}

// UpdateNameByMedicineCategory は TASK-215 で薬剤名変更時に連携在庫の name を同期する。
// (clinic_id, oldName, category=medicine) にマッチするレコードを newName に更新する（マッチなしは no-op）。
func (r *inventoryRepository) UpdateNameByMedicineCategory(ctx context.Context, clinicID uint64, oldName, newName string) error {
	result := dbOrTx(ctx, r.db).
		Model(&model.InventoryItem{}).
		Scopes(clinicScope(clinicID)).
		Where("name = ? AND category = ?", oldName, model.InventoryCategoryMedicine).
		Update("name", newName)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "inventory_item", fmt.Sprintf("name=%s", oldName))
	}
	return nil
}

// CountUsageByInventoryID は在庫アイテムを参照している治療明細・ワクチン・薬剤の件数を返す（BUG-195）
// clinic_id フィルタを JOIN またはスコープで適用しテナント分離を保証する（BUG-383）
func (r *inventoryRepository) CountUsageByInventoryID(ctx context.Context, clinicID, inventoryID uint64) (int64, error) {
	var treatmentCount, vaccineCount, medicineCount int64
	// treatments は clinic_id を直接持たないため medical_records を JOIN してテナント分離
	if err := dbOrTx(ctx, r.db).
		Model(&model.Treatment{}).
		Scopes(medicalRecordTenantScope("treatments", clinicID)).
		Where("treatments.inventory_id = ? AND treatments.deleted_at IS NULL", inventoryID).
		Count(&treatmentCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "inventory_item", fmt.Sprintf("%d", inventoryID))
	}
	// vaccines は clinic_id を直接持つ
	if err := dbOrTx(ctx, r.db).
		Model(&model.Vaccine{}).
		Scopes(clinicScope(clinicID)).
		Where("inventory_id = ? AND deleted_at IS NULL", inventoryID).
		Count(&vaccineCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "inventory_item", fmt.Sprintf("%d", inventoryID))
	}
	// medicines は clinic_id を直接持つ
	if err := dbOrTx(ctx, r.db).
		Model(&model.Medicine{}).
		Scopes(clinicScope(clinicID)).
		Where("inventory_id = ? AND deleted_at IS NULL", inventoryID).
		Count(&medicineCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "inventory_item", fmt.Sprintf("%d", inventoryID))
	}
	return treatmentCount + vaccineCount + medicineCount, nil
}
