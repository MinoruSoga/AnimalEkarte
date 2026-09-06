package billing

// billing_item_write_clinic_isolation_test.go — BUG-417（BE9-2A監査検出・billing着手前提）
//
// テスト対象: billingItemRepository.Update / Delete の clinic_id 境界。
// 保護する不変条件: clinic A から clinic B の billing_items を更新/削除できない。
// 旧実装は Joins() を UPDATE/DELETE へ連結しており GORM が述語を落とすため実質 no-op だった。
// EXISTS subquery 述語を削除すると本テストは必ず失敗する（mutation で検証済みであること）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupBillingItemIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Billing{}, &model.BillingItem{}))
	db.Exec("TRUNCATE TABLE billing_items, billings CASCADE")
	return db
}

func makeBillingWithItem(t *testing.T, db *gorm.DB, clinicID uint64) (*model.Billing, *model.BillingItem) {
	t.Helper()
	b := &model.Billing{ClinicID: clinicID, Status: model.BillingStatusWaiting}
	require.NoError(t, db.WithContext(context.Background()).Create(b).Error)
	item := &model.BillingItem{BillingID: b.ID, Category: model.ItemCategoryExamination, Name: "検査A", Quantity: 1, UnitPrice: 1000}
	require.NoError(t, db.WithContext(context.Background()).Create(item).Error)
	return b, item
}

// TestBillingItemRepository_Update_ClinicIsolation は clinic A から clinic B の
// 明細を更新できない（NotFound・無変更）ことを検証する。
func TestBillingItemRepository_Update_ClinicIsolation(t *testing.T) {
	db := setupBillingItemIsolationTestDB(t)
	repo := NewBillingItemRepository(db)
	ctx := context.Background()

	_, itemB := makeBillingWithItem(t, db, 2)

	price := int64(9999)
	err := repo.Update(ctx, 1, itemB.ID, UpdateBillingItemInput{UnitPrice: &price})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	var reloaded model.BillingItem
	require.NoError(t, db.First(&reloaded, itemB.ID).Error)
	assert.Equal(t, int64(1000), reloaded.UnitPrice)
}

// TestBillingItemRepository_Update_SameClinic は同一クリニックの更新が成功する
// （是正後も正常系が壊れていない）ことを検証する。
func TestBillingItemRepository_Update_SameClinic(t *testing.T) {
	db := setupBillingItemIsolationTestDB(t)
	repo := NewBillingItemRepository(db)
	ctx := context.Background()

	_, item := makeBillingWithItem(t, db, 1)

	price := int64(2500)
	require.NoError(t, repo.Update(ctx, 1, item.ID, UpdateBillingItemInput{UnitPrice: &price}))
	var reloaded model.BillingItem
	require.NoError(t, db.First(&reloaded, item.ID).Error)
	assert.Equal(t, int64(2500), reloaded.UnitPrice)
}

// TestBillingItemRepository_Delete_ClinicIsolation は clinic A から clinic B の
// 明細を削除できない（NotFound・行残存）ことを検証する。
func TestBillingItemRepository_Delete_ClinicIsolation(t *testing.T) {
	db := setupBillingItemIsolationTestDB(t)
	repo := NewBillingItemRepository(db)
	ctx := context.Background()

	_, itemB := makeBillingWithItem(t, db, 2)

	err := repo.Delete(ctx, 1, itemB.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	var count int64
	require.NoError(t, db.Model(&model.BillingItem{}).
		Where("id = ? AND deleted_at IS NULL", itemB.ID).Count(&count).Error)
	assert.Equal(t, int64(1), count, "clinic B の明細が削除されてはならない")
}

// TestBillingItemRepository_Delete_SameClinic は同一クリニックの削除が成功することを検証する。
func TestBillingItemRepository_Delete_SameClinic(t *testing.T) {
	db := setupBillingItemIsolationTestDB(t)
	repo := NewBillingItemRepository(db)
	ctx := context.Background()

	_, item := makeBillingWithItem(t, db, 1)

	require.NoError(t, repo.Delete(ctx, 1, item.ID))
	var count int64
	require.NoError(t, db.Model(&model.BillingItem{}).
		Where("id = ? AND deleted_at IS NULL", item.ID).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}
