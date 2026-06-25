package repository

// accounting_repository_clinic_isolation_test.go — #196 clinic_id テナント隔離回帰テスト
//
// テスト対象: AccountingRepository の clinic_id 境界
// 保護する不変条件: clinic A のスコープで clinic B の会計を
//   読み取れない、FOR UPDATE ロック取得できない、更新できない。
//
// このテストは clinicScope を accounting_repository.go から削除すると必ず失敗するよう設計されている。
//
// メソッド選定根拠:
//   FindByID       — clinicScope が回帰すると任意テナントの Billing が読み取り可能になる
//   LockAndFindByID— clinicScope + FOR UPDATE ロック越境は会計確定 TOCTOU 防止を破壊する最重大リスク
//   Update         — clinicScope 回帰で RowsAffected=0 チェックが機能せず任意テナントを上書き可能

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupAccountingIsolationTestDB は accounting clinic_id 隔離テスト用の DB を整備する。
// setupTestDB が item_category/item_source/payment_method ENUM を DROP TYPE ... CASCADE で削除するため
// billing_items / payment_splits テーブルが失われる。AutoMigrate で再追加する。
// billings / payments / billing_refunds / owners は setupTestDB が既に AutoMigrate 済み。
func setupAccountingIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.BillingItem{}, &model.PaymentSplit{}))
	return db
}

// makeBillingRet は指定 clinicID の最小 Billing を作成して返す。
// makeBilling は戻り値なしのため clinic_id 隔離テスト用に別途定義する。
// ownerID / petID は nil — FK 制約違反なし。
func makeBillingRet(t *testing.T, db *gorm.DB, clinicID uint64) *model.Billing {
	t.Helper()
	b := &model.Billing{
		ClinicID:      clinicID,
		TotalAmount:   1000,
		Status:        model.BillingStatusWaiting,
		ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(b).Error)
	return b
}

// TestAccountingRepository_FindByID_ClinicIsolation は
// clinic A の会計を clinic B の clinicID で取得できないことを検証する。
// clinicScope を accounting_repository.go から削除すると「別クリニックIDでは取得できない」が失敗する。
func TestAccountingRepository_FindByID_ClinicIsolation(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	billingA := makeBillingRet(t, db, clinicA)

	t.Run("同一クリニックIDでは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, billingA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, billingA.ID, got.ID)
		assert.Equal(t, clinicA, got.ClinicID)
	})

	t.Run("別クリニックIDでは取得できない（clinic_id 隔離）", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, billingA.ID)
		assert.Error(t, err, "clinic B から clinic A の会計を取得できてはならない")
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

// TestAccountingRepository_LockAndFindByID_ClinicIsolation は
// clinic A の会計を clinic B の clinicID で FOR UPDATE ロック取得できないことを検証する。
// clinicScope を削除すると「別クリニックIDでは取得できない」が失敗する。
// LockAndFindByID は会計確定の TOCTOU 防止に使われるため、テナント越境ロックは
// 支払確定フローの安全性を破壊する最重大リスク。
func TestAccountingRepository_LockAndFindByID_ClinicIsolation(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	billingA := makeBillingRet(t, db, clinicA)

	t.Run("同一クリニックIDではロック取得できる", func(t *testing.T) {
		got, err := repo.LockAndFindByID(ctx, clinicA, billingA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, billingA.ID, got.ID)
		assert.Equal(t, clinicA, got.ClinicID)
	})

	t.Run("別クリニックIDではロック取得できない（clinic_id 隔離）", func(t *testing.T) {
		got, err := repo.LockAndFindByID(ctx, clinicB, billingA.ID)
		assert.Error(t, err, "clinic B から clinic A の会計をロック取得できてはならない")
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

// TestAccountingRepository_Update_ClinicIsolation は
// 別クリニックIDからの Update が NotFound を返し、行が変更されないことを検証する。
// clinicScope を削除すると「行が変更されていない」が失敗する。
func TestAccountingRepository_Update_ClinicIsolation(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	billingA := makeBillingRet(t, db, clinicA)

	t.Run("別クリニックIDからの Update は NotFound を返す", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, billingA.ID, map[string]any{"total_amount": int64(9999)})
		require.Error(t, err, "clinic B から clinic A の会計を更新できてはならない")
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("行が変更されていない（データ改ざん防止）", func(t *testing.T) {
		// preload を使わず直接 DB 読み取りで total_amount が変化していないことを確認する。
		var b model.Billing
		require.NoError(t, db.Where("id = ? AND clinic_id = ?", billingA.ID, clinicA).First(&b).Error)
		assert.Equal(t, int64(1000), b.TotalAmount, "別クリニックからの Update で total_amount が変わってはならない")
	})

	t.Run("正しいクリニックIDからの Update は成功する", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, billingA.ID, map[string]any{"total_amount": int64(2000)})
		require.NoError(t, err)
		assert.Equal(t, int64(2000), got.TotalAmount)
	})
}
