package inventory

// repository_test.go — inventory Repository の統合テスト（内部カバレッジ向上）。
//
// 対象: FindAll / FindByID / Create / Update / Delete / DeleteIfUnused / DecreaseStock /
//       CountUsageByInventoryID / DeleteByNameAndMedicineCategory / UpdateNameByMedicineCategory
// 検証観点: 正常系、clinic_id 隔離、ソフトデリート除外、NotFound ラップ、使用数カウントの参照種別別合算。
//
// makeSpeciesAndPet/makeHistoryMedicalRecord はフラット package の同名ヘルパーの複製
// （BE8-4: import cycle を避けるための最小限の重複、移動時の型リネームはしない方針の対象外）。
// makeTestOwner はフラット側と同様 testdb.MakeTestOwner に直接委譲する。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupInventoryTestDB は inventory_items と CountUsageByInventoryID の JOIN 先
// treatments/vaccines/medicines/medical_records を整備する。
// Owner/AnimalSpecies/Pet は medical_records.pet_id の FK (fk_medical_records_pet) を満たすため
// AutoMigrate 対象に含める（makeHistoryMedicalRecord は実在の pet.ID を要求する）。
func setupInventoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Owner{}, &model.AnimalSpecies{}, &model.Pet{}, &model.MedicalRecord{},
		&model.InventoryItem{}, &model.Treatment{}, &model.Vaccine{}, &model.Medicine{},
	))
	db.Exec("TRUNCATE TABLE treatments CASCADE")
	db.Exec("TRUNCATE TABLE vaccines CASCADE")
	db.Exec("TRUNCATE TABLE medicines CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE inventory_items CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE owners CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func makeInventoryItem(t *testing.T, db *gorm.DB, clinicID uint64, name string, category model.InventoryCategory, status model.InventoryStatus, quantity int) *model.InventoryItem {
	t.Helper()
	item := &model.InventoryItem{ClinicID: clinicID, Name: name, Category: category, Status: status, Quantity: quantity}
	require.NoError(t, db.WithContext(context.Background()).Create(item).Error)
	return item
}

func makeSpeciesAndPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, petName string) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: species.ID,
		Name:            petName,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

func makeHistoryMedicalRecord(t *testing.T, db *gorm.DB, clinicID, petID uint64, recordNo string, date time.Time) *model.MedicalRecord {
	t.Helper()
	pet := petID
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     date,
		PetID:    &pet,
		Status:   model.MedicalRecordStatusFinalized,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

func TestInventoryRepository_FindAll_ClinicIsolationFiltersAndPagination(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	itemB := makeInventoryItem(t, db, clinicB, "医院Bの在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
	medA := makeInventoryItem(t, db, clinicA, "薬剤A", model.InventoryCategoryMedicine, model.InventoryStatusLow, 5)
	foodA := makeInventoryItem(t, db, clinicA, "フードA", model.InventoryCategoryFood, model.InventoryStatusSufficient, 20)

	t.Run("clinic_id 隔離: 他院の在庫は含まれない", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, 1, 10)
		require.NoError(t, err)
		assert.EqualValues(t, 2, total)
		require.Len(t, got, 2)
		for _, i := range got {
			assert.NotEqual(t, itemB.ID, i.ID)
		}
	})

	t.Run("category フィルタ", func(t *testing.T) {
		cat := string(model.InventoryCategoryMedicine)
		got, total, err := repo.FindAll(ctx, clinicA, &cat, nil, 1, 10)
		require.NoError(t, err)
		assert.EqualValues(t, 1, total)
		require.Len(t, got, 1)
		assert.Equal(t, medA.ID, got[0].ID)
	})

	t.Run("name ASC でページネーションされる", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, 1, 1)
		require.NoError(t, err)
		assert.EqualValues(t, 2, total, "total は limit に依存せず全件数")
		require.Len(t, got, 1)
		assert.Equal(t, foodA.ID, got[0].ID, "name ASC でフードAが先")

		got2, _, err := repo.FindAll(ctx, clinicA, nil, nil, 2, 1)
		require.NoError(t, err)
		require.Len(t, got2, 1)
		assert.Equal(t, medA.ID, got2[0].ID)
	})

	// status フィルタは他 subtest の total 件数アサーションを汚染しないよう最後に置く
	// （新規アイテムを clinicA に追加するため）。
	t.Run("status フィルタ", func(t *testing.T) {
		// SD-4 決裁A: status フィルタは保存列でなく quantity/min_stock_level の導出条件で絞り込む
		// （inventoryStatusFilterClause）。medA/foodA は min_stock_level が 0（未設定）のため、
		// 保存 status に関わらずどちらも sufficient と導出され low には該当しない。
		// low 導出条件を満たす専用アイテムを用意して検証する。
		lowItem := &model.InventoryItem{
			ClinicID: clinicA, Name: "残少アイテム", Category: model.InventoryCategoryMedicine,
			Status: model.InventoryStatusSufficient, Quantity: 2, MinStockLevel: 5,
		}
		require.NoError(t, db.WithContext(ctx).Create(lowItem).Error)

		status := string(model.InventoryStatusLow)
		got, total, err := repo.FindAll(ctx, clinicA, nil, &status, 1, 10)
		require.NoError(t, err)
		assert.EqualValues(t, 1, total)
		require.Len(t, got, 1)
		assert.Equal(t, lowItem.ID, got[0].ID)
	})
}

func TestInventoryRepository_FindByID(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	item := makeInventoryItem(t, db, clinicA, "在庫A", model.InventoryCategoryConsumable, model.InventoryStatusSufficient, 3)

	t.Run("同一クリニックで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, "在庫A", got.Name)
	})

	t.Run("別クリニックからは NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, item.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestInventoryRepository_Create(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	item := &model.InventoryItem{Name: "新規在庫", Category: model.InventoryCategoryOther}
	require.NoError(t, repo.Create(ctx, clinicA, item))
	assert.NotZero(t, item.ID)
	assert.Equal(t, clinicA, item.ClinicID, "Create は clinicID を強制的にセットする")

	got, err := repo.FindByID(ctx, clinicA, item.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規在庫", got.Name)
}

func TestInventoryRepository_Update(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	item := makeInventoryItem(t, db, clinicA, "旧在庫名", model.InventoryCategoryOther, model.InventoryStatusSufficient, 1)

	t.Run("同一クリニックで更新できる", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, item.ID, map[string]any{"name": "新在庫名"})
		require.NoError(t, err)
		assert.Equal(t, "新在庫名", got.Name)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, item.ID, map[string]any{"name": "乗っ取り"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の更新は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"name": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestInventoryRepository_Delete(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	item := makeInventoryItem(t, db, clinicA, "削除対象", model.InventoryCategoryOther, model.InventoryStatusSufficient, 1)

	t.Run("別クリニックからの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, item.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("同一クリニックで削除でき、ソフトデリートされ FindAll から除外される", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, item.ID))

		_, err := repo.FindByID(ctx, clinicA, item.ID)
		assert.True(t, apperrors.IsNotFound(err))

		var raw model.InventoryItem
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", item.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "deleted_at が設定されているべき")
	})
}

func TestInventoryRepository_DecreaseStock(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	item := makeInventoryItem(t, db, clinicA, "在庫減算対象", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
	foreignItem := makeInventoryItem(t, db, clinicB, "他院の在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)

	t.Run("在庫数が減算される", func(t *testing.T) {
		require.NoError(t, repo.DecreaseStock(ctx, clinicA, item.ID, 3))

		got, err := repo.FindByID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, 7, got.Quantity)
	})

	t.Run("別クリニックの在庫は NotFound となり減算されない", func(t *testing.T) {
		err := repo.DecreaseStock(ctx, clinicA, foreignItem.ID, 3)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, findErr := repo.FindByID(ctx, clinicB, foreignItem.ID)
		require.NoError(t, findErr)
		assert.Equal(t, 10, got.Quantity)
	})

	t.Run("ソフトデリート済みの在庫は NotFound となり減算されない", func(t *testing.T) {
		deletedItem := makeInventoryItem(t, db, clinicA, "削除済み在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
		require.NoError(t, repo.Delete(ctx, clinicA, deletedItem.ID))

		err := repo.DecreaseStock(ctx, clinicA, deletedItem.ID, 3)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var raw model.InventoryItem
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", deletedItem.ID).First(&raw).Error)
		assert.Equal(t, 10, raw.Quantity)
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		err := repo.DecreaseStock(ctx, clinicA, 999999, 1)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	// BUG-466: 在庫不足時は減算せず Conflict。quantity は変化しない。
	t.Run("在庫不足は Conflict となり数量は変わらない", func(t *testing.T) {
		shortItem := makeInventoryItem(t, db, clinicA, "不足在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 2)

		err := repo.DecreaseStock(ctx, clinicA, shortItem.ID, 5)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected conflict for insufficient stock, got: %v", err)

		got, findErr := repo.FindByID(ctx, clinicA, shortItem.ID)
		require.NoError(t, findErr)
		assert.Equal(t, 2, got.Quantity)
	})

	// BUG-466: ちょうどゼロになる減算は成功する。
	t.Run("ちょうどゼロになる減算は成功する", func(t *testing.T) {
		exactItem := makeInventoryItem(t, db, clinicA, "ゼロ丁度在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 4)

		require.NoError(t, repo.DecreaseStock(ctx, clinicA, exactItem.ID, 4))

		got, err := repo.FindByID(ctx, clinicA, exactItem.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, got.Quantity)
	})
}

func TestInventoryRepository_DecreaseStock_AmbientTxRollback(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	item := makeInventoryItem(t, db, clinicA, "ロールバック対象在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
	forcedErr := errors.New("force rollback after stock decrease")

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := repo.DecreaseStock(txCtx, clinicA, item.ID, 3); err != nil {
			return err
		}
		return forcedErr
	})
	require.ErrorIs(t, err, forcedErr)

	got, err := repo.FindByID(ctx, clinicA, item.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, got.Quantity, "ambient transaction rollback must restore the stock quantity")
}

// TestInventoryRepository_Update_AmbientTxRollback は BUG-465 の回帰:
// Update の書き込みと再取得が同一 tx に参加し、呼び出し元の ambient tx が
// rollback されると更新も取り消されることを検証する。
func TestInventoryRepository_Update_AmbientTxRollback(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	item := makeInventoryItem(t, db, clinicA, "更新前在庫名", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
	forcedErr := errors.New("force rollback after inventory update")

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		got, err := repo.Update(txCtx, clinicA, item.ID, map[string]any{"name": "更新後在庫名"})
		if err != nil {
			return err
		}
		assert.Equal(t, "更新後在庫名", got.Name, "同一 tx 内では更新後の値を再取得できる")
		return forcedErr
	})
	require.ErrorIs(t, err, forcedErr)

	got, err := repo.FindByID(ctx, clinicA, item.ID)
	require.NoError(t, err)
	assert.Equal(t, "更新前在庫名", got.Name, "ambient transaction rollback must restore the inventory name")
}

// TestInventoryRepository_Update_ReloadFailureRollsBackUpdate は BUG-465:
// Update 成功後の再取得が失敗した場合、コミット済み更新を失敗応答へ反転させないよう
// 同一 tx 内で rollback することを検証する。
func TestInventoryRepository_Update_ReloadFailureRollsBackUpdate(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	item := makeInventoryItem(t, db, clinicA, "更新前", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 5)
	const callbackName = "inventory:update_reload_failure"
	reloadErr := errors.New("forced inventory reload failure")
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register(callbackName, func(query *gorm.DB) {
		if query.Statement != nil && query.Statement.Table == "inventory_items" {
			query.AddError(reloadErr)
		}
	}))
	callbackRegistered := true
	t.Cleanup(func() {
		if callbackRegistered {
			require.NoError(t, db.Callback().Query().Remove(callbackName))
		}
	})

	got, err := repo.Update(ctx, clinicA, item.ID, map[string]any{"name": "更新後"})

	assert.Nil(t, got)
	require.Error(t, err)
	assert.ErrorIs(t, err, reloadErr)

	require.NoError(t, db.Callback().Query().Remove(callbackName))
	callbackRegistered = false

	var persisted model.InventoryItem
	require.NoError(t, db.WithContext(ctx).First(&persisted, item.ID).Error)
	assert.Equal(t, "更新前", persisted.Name, "reload failure must roll back the committed-looking update")
}

// TestInventoryRepository_FindAll_StatusFilterUsesQuantityDerivedPredicate は SD-4 決裁A
// （q&a.html SD-4: status を保存せず読み取り時に導出する）の回帰:
// status クエリフィルタが保存された status 列（もはや信頼できない）ではなく、
// quantity/min_stock_level から導出した predicate で絞り込むことを検証する。
// 保存 status とわざと矛盾させたレコードで検証し、保存値が無視されていることを証明する。
func TestInventoryRepository_FindAll_StatusFilterUsesQuantityDerivedPredicate(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	// 保存 status はすべて意図的に実態と矛盾させる（sufficient のまま放置された旧データを模す）。
	lowItem := &model.InventoryItem{
		ClinicID: clinicA, Name: "実態は残少", Category: model.InventoryCategoryMedicine,
		Status: model.InventoryStatusSufficient, Quantity: 3, MinStockLevel: 10,
	}
	outOfStockItem := &model.InventoryItem{
		ClinicID: clinicA, Name: "実態は在庫切れ", Category: model.InventoryCategoryMedicine,
		Status: model.InventoryStatusSufficient, Quantity: 0, MinStockLevel: 10,
	}
	sufficientItem := &model.InventoryItem{
		ClinicID: clinicA, Name: "実態は十分", Category: model.InventoryCategoryMedicine,
		Status: model.InventoryStatusLow, Quantity: 20, MinStockLevel: 10,
	}
	unsetThresholdItem := &model.InventoryItem{
		ClinicID: clinicA, Name: "閾値未設定なら十分", Category: model.InventoryCategoryMedicine,
		Status: model.InventoryStatusLow, Quantity: 1, MinStockLevel: 0,
	}
	for _, item := range []*model.InventoryItem{lowItem, outOfStockItem, sufficientItem, unsetThresholdItem} {
		require.NoError(t, db.WithContext(ctx).Create(item).Error)
	}

	t.Run("status=low は保存値でなく quantity<=min_stock_level(>0) で絞り込む", func(t *testing.T) {
		status := string(model.InventoryStatusLow)
		got, total, err := repo.FindAll(ctx, clinicA, nil, &status, 1, 20)
		require.NoError(t, err)
		assert.EqualValues(t, 1, total)
		assert.Equal(t, lowItem.ID, got[0].ID)
	})

	t.Run("status=out_of_stock は quantity<=0 で絞り込む", func(t *testing.T) {
		status := string(model.InventoryStatusOutOfStock)
		got, total, err := repo.FindAll(ctx, clinicA, nil, &status, 1, 20)
		require.NoError(t, err)
		assert.EqualValues(t, 1, total)
		assert.Equal(t, outOfStockItem.ID, got[0].ID)
	})

	t.Run("status=sufficient は quantity>0 かつ (閾値未設定 or quantity>閾値) で絞り込む", func(t *testing.T) {
		status := string(model.InventoryStatusSufficient)
		got, total, err := repo.FindAll(ctx, clinicA, nil, &status, 1, 20)
		require.NoError(t, err)
		assert.EqualValues(t, 2, total)
		gotIDs := []uint64{got[0].ID, got[1].ID}
		assert.ElementsMatch(t, []uint64{sufficientItem.ID, unsetThresholdItem.ID}, gotIDs)
	})

	t.Run("未知の status 値は該当なし", func(t *testing.T) {
		status := "unknown"
		got, total, err := repo.FindAll(ctx, clinicA, nil, &status, 1, 20)
		require.NoError(t, err)
		assert.EqualValues(t, 0, total)
		assert.Empty(t, got)
	})
}

func TestInventoryRepository_CountUsageByInventoryID(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	item := makeInventoryItem(t, db, clinicA, "使用中の在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 100)

	t.Run("参照が無ければ 0", func(t *testing.T) {
		count, err := repo.CountUsageByInventoryID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	owner := testdb.MakeTestOwner(t, db, clinicA, "在庫使用飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "在庫使用犬")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-INV-001", time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	inventoryID := item.ID
	treatment := &model.Treatment{MedicalRecordID: mr.ID, ItemType: model.TreatmentItemTypeMedicine, InventoryID: &inventoryID, Content: "薬剤投与"}
	require.NoError(t, db.WithContext(ctx).Create(treatment).Error)

	vaccine := &model.Vaccine{ClinicID: clinicA, Name: "混合ワクチン", InventoryID: &inventoryID}
	require.NoError(t, db.WithContext(ctx).Create(vaccine).Error)

	medicine := &model.Medicine{ClinicID: clinicA, Name: "抗生物質", InventoryID: &inventoryID}
	require.NoError(t, db.WithContext(ctx).Create(medicine).Error)

	t.Run("treatment/vaccine/medicine の参照が合算される", func(t *testing.T) {
		count, err := repo.CountUsageByInventoryID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})

	t.Run("別クリニックIDでは 0（クロステナント越境なし）", func(t *testing.T) {
		count, err := repo.CountUsageByInventoryID(ctx, clinicB, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリートされた treatment は除外される（P2）", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Delete(treatment).Error)
		count, err := repo.CountUsageByInventoryID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count, "treatment 分が除外され vaccine+medicine の2件のみ")
	})

	t.Run("ソフトデリートされた vaccine/medicine は除外される（P2）", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Delete(vaccine).Error)
		require.NoError(t, db.WithContext(ctx).Delete(medicine).Error)
		count, err := repo.CountUsageByInventoryID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestInventoryRepository_DeleteIfUnused(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	makePet := func(t *testing.T, clinicID, ownerID uint64, petName string) *model.Pet {
		t.Helper()
		pet := &model.Pet{
			ClinicID: clinicID, OwnerID: ownerID, AnimalSpeciesID: species.ID, Name: petName,
		}
		require.NoError(t, db.WithContext(ctx).Create(pet).Error)
		return pet
	}

	assertStillExists := func(t *testing.T, clinicID, id uint64) {
		t.Helper()
		got, err := repo.FindByID(ctx, clinicID, id)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
	}

	t.Run("参照が無ければ削除できる", func(t *testing.T) {
		item := makeInventoryItem(t, db, clinicA, "未使用在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)

		require.NoError(t, repo.DeleteIfUnused(ctx, clinicA, item.ID))

		_, err := repo.FindByID(ctx, clinicA, item.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("同一クリニックの treatment 参照は Conflict で行は残る", func(t *testing.T) {
		item := makeInventoryItem(t, db, clinicA, "治療参照在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
		owner := testdb.MakeTestOwner(t, db, clinicA, "治療参照飼主")
		pet := makePet(t, clinicA, owner.ID, "治療参照犬")
		mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-DEL-T", time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
		inventoryID := item.ID
		require.NoError(t, db.WithContext(ctx).Create(&model.Treatment{
			MedicalRecordID: mr.ID, ItemType: model.TreatmentItemTypeMedicine, InventoryID: &inventoryID, Content: "薬剤投与",
		}).Error)

		err := repo.DeleteIfUnused(ctx, clinicA, item.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Contains(t, err.Error(), "この項目は使用中のため削除できません")
		assertStillExists(t, clinicA, item.ID)
	})

	t.Run("同一クリニックの vaccine 参照は Conflict で行は残る", func(t *testing.T) {
		item := makeInventoryItem(t, db, clinicA, "ワクチン参照在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
		inventoryID := item.ID
		require.NoError(t, db.WithContext(ctx).Create(&model.Vaccine{
			ClinicID: clinicA, Name: "混合ワクチン", InventoryID: &inventoryID,
		}).Error)

		err := repo.DeleteIfUnused(ctx, clinicA, item.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Contains(t, err.Error(), "この項目は使用中のため削除できません")
		assertStillExists(t, clinicA, item.ID)
	})

	t.Run("同一クリニックの medicine 参照は Conflict で行は残る", func(t *testing.T) {
		item := makeInventoryItem(t, db, clinicA, "薬剤参照在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
		inventoryID := item.ID
		require.NoError(t, db.WithContext(ctx).Create(&model.Medicine{
			ClinicID: clinicA, Name: "抗生物質", InventoryID: &inventoryID,
		}).Error)

		err := repo.DeleteIfUnused(ctx, clinicA, item.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Contains(t, err.Error(), "この項目は使用中のため削除できません")
		assertStillExists(t, clinicA, item.ID)
	})

	t.Run("他院の vaccine 参照は自院削除を妨げない", func(t *testing.T) {
		item := makeInventoryItem(t, db, clinicA, "他院ワクチン参照", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
		inventoryID := item.ID
		require.NoError(t, db.WithContext(ctx).Create(&model.Vaccine{
			ClinicID: clinicB, Name: "他院ワクチン", InventoryID: &inventoryID,
		}).Error)

		require.NoError(t, repo.DeleteIfUnused(ctx, clinicA, item.ID))
		_, err := repo.FindByID(ctx, clinicA, item.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("他院の medicine 参照は自院削除を妨げない", func(t *testing.T) {
		item := makeInventoryItem(t, db, clinicA, "他院薬剤参照", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
		inventoryID := item.ID
		require.NoError(t, db.WithContext(ctx).Create(&model.Medicine{
			ClinicID: clinicB, Name: "他院薬剤", InventoryID: &inventoryID,
		}).Error)

		require.NoError(t, repo.DeleteIfUnused(ctx, clinicA, item.ID))
		_, err := repo.FindByID(ctx, clinicA, item.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("他院カルテの treatment 参照は自院削除を妨げない", func(t *testing.T) {
		item := makeInventoryItem(t, db, clinicA, "他院治療参照", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
		ownerB := testdb.MakeTestOwner(t, db, clinicB, "他院治療飼主")
		petB := makePet(t, clinicB, ownerB.ID, "他院治療犬")
		mrB := makeHistoryMedicalRecord(t, db, clinicB, petB.ID, "MR-DEL-B", time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC))
		inventoryID := item.ID
		require.NoError(t, db.WithContext(ctx).Create(&model.Treatment{
			MedicalRecordID: mrB.ID, ItemType: model.TreatmentItemTypeMedicine, InventoryID: &inventoryID, Content: "他院投与",
		}).Error)

		require.NoError(t, repo.DeleteIfUnused(ctx, clinicA, item.ID))
		_, err := repo.FindByID(ctx, clinicA, item.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("ソフトデリートされた参照は削除を妨げない", func(t *testing.T) {
		item := makeInventoryItem(t, db, clinicA, "削除済み参照在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
		inventoryID := item.ID
		vaccine := &model.Vaccine{ClinicID: clinicA, Name: "削除済みワクチン", InventoryID: &inventoryID}
		require.NoError(t, db.WithContext(ctx).Create(vaccine).Error)
		require.NoError(t, db.WithContext(ctx).Delete(vaccine).Error)

		require.NoError(t, repo.DeleteIfUnused(ctx, clinicA, item.ID))
		_, err := repo.FindByID(ctx, clinicA, item.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("他院からの削除は NotFound（対象は削除されない）", func(t *testing.T) {
		item := makeInventoryItem(t, db, clinicA, "越境削除対象", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)

		err := repo.DeleteIfUnused(ctx, clinicB, item.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assertStillExists(t, clinicA, item.ID)
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		err := repo.DeleteIfUnused(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("CountUsage==0 直後の参照挿入は Conflict（TOCTOU）", func(t *testing.T) {
		item := makeInventoryItem(t, db, clinicA, "TOCTOU在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)

		count, err := repo.CountUsageByInventoryID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), count)

		inventoryID := item.ID
		require.NoError(t, db.WithContext(ctx).Create(&model.Vaccine{
			ClinicID: clinicA, Name: "レース挿入ワクチン", InventoryID: &inventoryID,
		}).Error)

		err = repo.DeleteIfUnused(ctx, clinicA, item.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Contains(t, err.Error(), "この項目は使用中のため削除できません")
		assertStillExists(t, clinicA, item.ID)
	})
}

func TestInventoryRepository_DeleteIfUnused_AmbientTxRollback(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	item := makeInventoryItem(t, db, clinicA, "ロールバック削除在庫", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
	forcedErr := errors.New("force rollback after unused delete")

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := repo.DeleteIfUnused(txCtx, clinicA, item.ID); err != nil {
			return err
		}
		return forcedErr
	})
	require.ErrorIs(t, err, forcedErr)

	got, err := repo.FindByID(ctx, clinicA, item.ID)
	require.NoError(t, err)
	assert.Equal(t, item.ID, got.ID, "ambient transaction rollback must restore the inventory row")
}

func TestInventoryRepository_DeleteByNameAndMedicineCategory(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("マッチなしは no-op（エラーなし）", func(t *testing.T) {
		require.NoError(t, repo.DeleteByNameAndMedicineCategory(ctx, clinicA, "存在しない薬剤"))
	})

	medItem := makeInventoryItem(t, db, clinicA, "アモキシリン", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
	consumableSameName := makeInventoryItem(t, db, clinicA, "アモキシリン", model.InventoryCategoryConsumable, model.InventoryStatusSufficient, 10)
	otherClinicItem := makeInventoryItem(t, db, clinicB, "アモキシリン", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)

	require.NoError(t, repo.DeleteByNameAndMedicineCategory(ctx, clinicA, "アモキシリン"))

	t.Run("category=medicine の同名在庫のみ削除される", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, medItem.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("category が異なる在庫は削除されない", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, consumableSameName.ID)
		require.NoError(t, err)
		assert.Equal(t, "アモキシリン", got.Name)
	})

	t.Run("別クリニックの同名在庫は削除されない", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, otherClinicItem.ID)
		require.NoError(t, err)
		assert.Equal(t, "アモキシリン", got.Name)
	})
}

func TestInventoryRepository_UpdateNameByMedicineCategory(t *testing.T) {
	db := setupInventoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("マッチなしは no-op（エラーなし）", func(t *testing.T) {
		require.NoError(t, repo.UpdateNameByMedicineCategory(ctx, clinicA, "存在しない薬剤", "新名称"))
	})

	medItem := makeInventoryItem(t, db, clinicA, "旧薬剤名", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)
	consumableSameName := makeInventoryItem(t, db, clinicA, "旧薬剤名", model.InventoryCategoryConsumable, model.InventoryStatusSufficient, 10)
	otherClinicItem := makeInventoryItem(t, db, clinicB, "旧薬剤名", model.InventoryCategoryMedicine, model.InventoryStatusSufficient, 10)

	require.NoError(t, repo.UpdateNameByMedicineCategory(ctx, clinicA, "旧薬剤名", "新薬剤名"))

	t.Run("category=medicine の同名在庫のみ更新される", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, medItem.ID)
		require.NoError(t, err)
		assert.Equal(t, "新薬剤名", got.Name)
	})

	t.Run("category が異なる在庫は更新されない", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, consumableSameName.ID)
		require.NoError(t, err)
		assert.Equal(t, "旧薬剤名", got.Name)
	})

	t.Run("別クリニックの同名在庫は更新されない", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, otherClinicItem.ID)
		require.NoError(t, err)
		assert.Equal(t, "旧薬剤名", got.Name)
	})
}
