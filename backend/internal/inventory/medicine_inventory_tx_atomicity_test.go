package inventory_test

// medicine_inventory_tx_atomicity_test.go — BE-refactor.md Appendix-A X-6
//
// 背景: medicine_service.Create/Update/Delete は BUG-429/R1-2 のコメントで
// s.transactor.WithTx(...) による原子性を謳うが、medicine_repository.go / inventory_repository.go は
// dbOrTx を一切使用せず常に r.db.WithContext(ctx) を直接参照していた（billing_item_repository と
// 同型の部分コミットバグ）。薬剤行の作成は独立 tx で即コミットされるため、直後の連携在庫作成や
// per_weight 監査だけが失敗しても ambient tx の rollback で薬剤行は戻らず孤児レコードが残っていた。
//
// temp-revert RED の手順:
//   - medicine_repository.go / inventory_repository.go の Create を dbOrTx(ctx, r.db) → r.db.WithContext(ctx) に戻す
//   - 本ファイルの RollsBackWhenAmbientTxFails 系を実行 → RED（孤児行が残ってしまう）
//   - 元に戻す → GREEN

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/inventory"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupMedicineInventoryTxTestDB は BUG-429 の Create フロー（medicine + inventory_items）を
// 単一トランザクションで検証するために両テーブルを整備する。
func setupMedicineInventoryTxTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Medicine{}, &model.InventoryItem{}))
	db.Exec("TRUNCATE TABLE medicines CASCADE")
	db.Exec("TRUNCATE TABLE inventory_items CASCADE")
	return db
}

var errSentinelMedicineInventoryTx = errors.New("simulated post-write failure in ambient tx")

// TestMedicineInventoryRepository_Create_RollsBackWhenAmbientTxFails は BUG-429 の実フロー
// （medicineRepo.Create → inventoryRepo.Create）を模し、ambient tx 内での後続失敗で
// 薬剤行・連携在庫行の両方がロールバックされる（孤児が残らない）ことを検証する。
func TestMedicineInventoryRepository_Create_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupMedicineInventoryTxTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	medicineRepo := medicalrecord.NewMedicineRepository(db)
	inventoryRepo := inventory.New(db)
	tx := persistence.NewTransactor(db)

	medicine := &model.Medicine{ClinicID: clinicA, Name: "原子性テスト薬剤"}

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := medicineRepo.Create(txCtx, medicine); err != nil {
			return err
		}
		inventoryItem := &model.InventoryItem{
			ClinicID: clinicA,
			Name:     medicine.Name,
			Category: model.InventoryCategoryMedicine,
		}
		if err := inventoryRepo.Create(txCtx, clinicA, inventoryItem); err != nil {
			return err
		}
		// per_weight 監査など後続ステップの失敗を模す。
		return errSentinelMedicineInventoryTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelMedicineInventoryTx)

	var medicineCount int64
	db.Model(&model.Medicine{}).Where("clinic_id = ? AND name = ?", clinicA, "原子性テスト薬剤").Count(&medicineCount)
	assert.EqualValues(t, 0, medicineCount, "ambient tx 失敗時、薬剤行の作成はロールバックされる（孤児禁止）")

	var inventoryCount int64
	db.Model(&model.InventoryItem{}).Where("clinic_id = ? AND name = ?", clinicA, "原子性テスト薬剤").Count(&inventoryCount)
	assert.EqualValues(t, 0, inventoryCount, "ambient tx 失敗時、連携在庫行の作成もロールバックされる（孤児禁止）")
}

// TestMedicineInventoryRepository_Create_CommitsWithinAmbientTx は正常系で
// 薬剤行・連携在庫行の両方が commit 後に永続化されることを検証する。
func TestMedicineInventoryRepository_Create_CommitsWithinAmbientTx(t *testing.T) {
	db := setupMedicineInventoryTxTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	medicineRepo := medicalrecord.NewMedicineRepository(db)
	inventoryRepo := inventory.New(db)
	tx := persistence.NewTransactor(db)

	medicine := &model.Medicine{ClinicID: clinicA, Name: "コミットテスト薬剤"}

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := medicineRepo.Create(txCtx, medicine); err != nil {
			return err
		}
		inventoryItem := &model.InventoryItem{
			ClinicID: clinicA,
			Name:     medicine.Name,
			Category: model.InventoryCategoryMedicine,
		}
		return inventoryRepo.Create(txCtx, clinicA, inventoryItem)
	}))

	var medicineCount int64
	db.Model(&model.Medicine{}).Where("clinic_id = ? AND name = ?", clinicA, "コミットテスト薬剤").Count(&medicineCount)
	assert.EqualValues(t, 1, medicineCount, "commit 後は薬剤行が永続化される")

	var inventoryCount int64
	db.Model(&model.InventoryItem{}).Where("clinic_id = ? AND name = ?", clinicA, "コミットテスト薬剤").Count(&inventoryCount)
	assert.EqualValues(t, 1, inventoryCount, "commit 後は連携在庫行が永続化される")
}

// TestMedicineRepository_Update_ReadsOwnWriteWithinAmbientTx は medicine_repository.Update が
// 内部で呼ぶ FindByID が ambient tx に参加し、同一 tx 内でコミット前の更新後の値を読めることを検証する
// （dbOrTx 非対応時は別コネクションを読むため FindByID が更新前の値を返す可能性がある）。
func TestMedicineRepository_Update_ReadsOwnWriteWithinAmbientTx(t *testing.T) {
	db := setupMedicineInventoryTxTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	medicineRepo := medicalrecord.NewMedicineRepository(db)
	tx := persistence.NewTransactor(db)

	medicine := &model.Medicine{ClinicID: clinicA, Name: "更新前"}
	require.NoError(t, medicineRepo.Create(ctx, medicine))

	var updated *model.Medicine
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		name := "更新後"
		updated, err = medicineRepo.Update(txCtx, clinicA, medicine.ID, medicalrecord.UpdateMedicineInput{Name: &name})
		return err
	}))

	require.NotNil(t, updated)
	assert.Equal(t, "更新後", updated.Name, "Update 内部の FindByID は同一 tx 内の更新結果を読めるべき")
}
