// Package inventory owns inventory and merchandise-item HTTP, use-case, and
// persistence behavior as one ADR-006 vertical slice.
package inventory

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Repository is the inventory persistence surface used by the domain service
// and type-safe application composition.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error)
	Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InventoryItem, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	DecreaseStock(ctx context.Context, clinicID, id uint64, quantity float64) error
	CountUsageByInventoryID(ctx context.Context, clinicID, inventoryID uint64) (int64, error)
	// BUG-381: 薬剤マスタ削除時に BUG-320 で自動作成された連携在庫をカスケード削除するため、
	// (clinic_id, name, category=medicine) で在庫を削除する。マッチなしは no-op。
	DeleteByNameAndMedicineCategory(ctx context.Context, clinicID uint64, name string) error
	// TASK-215: 薬剤名変更時に BUG-320 で自動作成された連携在庫の name を同期する。
	// (clinic_id, oldName, category=medicine) にマッチするレコードを newName に更新する。マッチなしは no-op。
	UpdateNameByMedicineCategory(ctx context.Context, clinicID uint64, oldName, newName string) error
}

type repository struct {
	db *gorm.DB
}

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

// isUniqueConstraintErr はPostgreSQLのユニーク制約違反（23505）を判定する
// （親 repository パッケージの同名ヘルパーの最小限の複製。repohelpers 未収載のため import cycle 回避のためローカル定義。BE8-4 batch18）。
func isUniqueConstraintErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// paginate は 1-origin の page と limit を Offset/Limit に変換する共有 Scope。
// （親 repository パッケージの同名ヘルパーの最小限の複製。repohelpers 未収載のため import cycle 回避のためローカル定義。BE8-4 batch18）。
func paginate(page, limit int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64, category, status *string, page, limit int) ([]model.InventoryItem, int64, error) {
	items := make([]model.InventoryItem, 0)
	var total int64

	q := persistence.DBOrTx(ctx, r.db).Model(&model.InventoryItem{}).Scopes(persistence.ClinicScope(clinicID))
	if category != nil {
		q = q.Where("category = ?", *category)
	}
	if status != nil {
		q = q.Where(inventoryStatusFilterClause(*status))
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "inventory_item", "")
	}
	if err := q.Scopes(paginate(page, limit)).Order("name ASC").Find(&items).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "inventory_item", "")
	}
	return items, total, nil
}

// inventoryStatusFilterClause は SD-4 決裁A（q&a.html SD-4: status を保存せず読み取り時に
// 導出する）に伴い、status クエリパラメータを quantity/min_stock_level に対する固定 SQL 述語へ
// 変換する。model.DeriveInventoryStatus と同一の判定ロジックを SQL で表現したもの — 保存された
// status 列はもはや信頼できる情報源ではないため、一覧フィルタも参照してはならない
// （参照すると、一覧で表示される導出済みステータスとフィルタ結果が食い違う）。
// 未知の値は 1=0（該当なし）とし、従来の厳密一致フィルタと同じ「ヒットなし」挙動を保つ。
// 呼び出し元の *status はこの固定候補群からの選択にのみ使い、SQL へ直接埋め込まない。
func inventoryStatusFilterClause(status string) string {
	switch model.InventoryStatus(status) {
	case model.InventoryStatusOutOfStock:
		return "quantity <= 0"
	case model.InventoryStatusLow:
		return "quantity > 0 AND min_stock_level > 0 AND quantity <= min_stock_level"
	case model.InventoryStatusSufficient:
		return "quantity > 0 AND (min_stock_level <= 0 OR quantity > min_stock_level)"
	default:
		return "1 = 0"
	}
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error) {
	return persistence.FindByIDScoped[model.InventoryItem](ctx, persistence.DBOrTx(ctx, r.db), "inventory_item", clinicID, id)
}

func (r *repository) Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error {
	item.ClinicID = clinicID
	err := persistence.DBOrTx(ctx, r.db).Create(item).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("inventory_item", item.Name)
		}
		return apperrors.FromGORM(err, "inventory_item", "")
	}
	return nil
}

// Update updates fields and reloads the row in one transaction so a reload
// failure cannot invert a committed write into a failure response (BUG-465).
func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InventoryItem, error) {
	var loaded *model.InventoryItem
	err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := persistence.UpdateScopedByID(txCtx, tx, &model.InventoryItem{}, "inventory_item", clinicID, id, fields); err != nil {
			return err
		}
		var err error
		loaded, err = r.FindByID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "reload inventory_item after update")
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update and reload inventory_item")
	}
	return loaded, nil
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DeleteScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.InventoryItem{}, "inventory_item", clinicID, id)
}

// DecreaseStock は clinic_id + id + active row を同じ UPDATE 述語で検証して quantity のみを減算する。
// status は SD-4 決裁A（q&a.html SD-4）により
// 保存値を信頼しないことになったため、ここでの再計算は行わない — 読み取り時に
// model.DeriveInventoryStatus で quantity/min_stock_level から都度導出する
// （internal/inventory/inventory_response.go の toInventoryResponse を参照）。
//
// BUG-466: quantity >= 減算量 を同一 UPDATE 述語に含め、負数在庫を作らない。
// RowsAffected==0 のとき FindByID で存在確認し、存在すれば Conflict（在庫不足）、
// なければ NotFound を返す。
func (r *repository) DecreaseStock(ctx context.Context, clinicID, id uint64, quantity float64) error {
	qty := int(quantity)
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.InventoryItem{}).
		Where("clinic_id = ? AND id = ? AND deleted_at IS NULL AND quantity >= ?", clinicID, id, qty).
		UpdateColumn("quantity", gorm.Expr("quantity - ?", qty))
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "inventory_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		if _, err := r.FindByID(ctx, clinicID, id); err != nil {
			return err
		}
		return apperrors.WrapConflict("在庫が不足しています")
	}
	return nil
}

// DeleteByNameAndMedicineCategory は BUG-320 で自動作成された在庫レコードを
// (clinic_id, name, category=medicine) で特定して削除する（BUG-381）。
// マッチなしは no-op（エラーなし）で返す。複数マッチは全件削除。
func (r *repository) DeleteByNameAndMedicineCategory(ctx context.Context, clinicID uint64, name string) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("name = ? AND category = ?", name, model.InventoryCategoryMedicine).
		Delete(&model.InventoryItem{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "inventory_item", fmt.Sprintf("name=%s", name))
	}
	return nil
}

// UpdateNameByMedicineCategory は TASK-215 で薬剤名変更時に連携在庫の name を同期する。
// (clinic_id, oldName, category=medicine) にマッチするレコードを newName に更新する（マッチなしは no-op）。
func (r *repository) UpdateNameByMedicineCategory(ctx context.Context, clinicID uint64, oldName, newName string) error {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.InventoryItem{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("name = ? AND category = ?", oldName, model.InventoryCategoryMedicine).
		Update("name", newName)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "inventory_item", fmt.Sprintf("name=%s", oldName))
	}
	return nil
}

// CountUsageByInventoryID は在庫アイテムを参照している治療明細・ワクチン・薬剤の件数を返す（BUG-195）
// clinic_id フィルタを JOIN またはスコープで適用しテナント分離を保証する（BUG-383）
func (r *repository) CountUsageByInventoryID(ctx context.Context, clinicID, inventoryID uint64) (int64, error) {
	var treatmentCount, vaccineCount, medicineCount int64
	// treatments は clinic_id を直接持たないため medical_records を JOIN してテナント分離
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Treatment{}).
		Scopes(persistence.MedicalRecordTenantScope("treatments", clinicID)).
		Where("treatments.inventory_id = ? AND treatments.deleted_at IS NULL", inventoryID).
		Count(&treatmentCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "inventory_item", fmt.Sprintf("%d", inventoryID))
	}
	// vaccines は clinic_id を直接持つ
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Vaccine{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("inventory_id = ? AND deleted_at IS NULL", inventoryID).
		Count(&vaccineCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "inventory_item", fmt.Sprintf("%d", inventoryID))
	}
	// medicines は clinic_id を直接持つ
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Medicine{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("inventory_id = ? AND deleted_at IS NULL", inventoryID).
		Count(&medicineCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "inventory_item", fmt.Sprintf("%d", inventoryID))
	}
	return treatmentCount + vaccineCount + medicineCount, nil
}
