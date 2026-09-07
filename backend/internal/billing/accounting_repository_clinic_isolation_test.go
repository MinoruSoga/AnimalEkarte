package billing

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

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupAccountingIsolationTestDB は accounting clinic_id 隔離テスト用の DB を整備する。
// setupTestDB はコアテーブルのみ共有 AutoMigrate する。billing_items / payment_splits は
// ここで ensureAutoMigrated する（ENUM を毎 setup で DROP するわけではない）。
// billings / payments / billing_refunds / owners は setupTestDB が既に AutoMigrate 済み。
func setupAccountingIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.BillingItem{}, &model.PaymentSplit{}))
	return db
}

// makeBillingRet は指定 clinicID の最小 Billing を作成して返す。
// makeBilling は戻り値なしのため clinic_id 隔離テスト用に別途定義する。
// ownerID / petID は nil — FK 制約違反なし。
// makeBillingWith（billing_test_fixtures_test.go）への thin wrapper。
//
//nolint:unparam // clinicID は将来の隔離テスト拡張のため引数として保持する
func makeBillingRet(t *testing.T, db *gorm.DB, clinicID uint64) *model.Billing {
	t.Helper()
	return makeBillingWith(t, db, billingFixtureOpts{
		ClinicID:      clinicID,
		TotalAmount:   1000,
		Status:        model.BillingStatusWaiting,
		ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	})
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

// TestAccountingRepository_FindAll_ClinicIsolation は
// clinic A の会計一覧を clinic B の clinicID で取得できないことを検証する。
// clinicScope を accounting_repository.go から削除すると「別クリニックIDでは0件」が失敗する。
func TestAccountingRepository_FindAll_ClinicIsolation(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	makeBillingRet(t, db, clinicA)

	t.Run("別クリニックIDでは0件（clinic_id 隔離）", func(t *testing.T) {
		billings, total, err := repo.FindAll(ctx, clinicB, AccountingListFilters{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total, "clinic B から clinic A の会計一覧は0件でなければならない")
		assert.Empty(t, billings)
	})

	t.Run("同一クリニックIDでは取得できる", func(t *testing.T) {
		billings, total, err := repo.FindAll(ctx, clinicA, AccountingListFilters{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, billings, 1)
	})
}

// TestAccountingRepository_SavePaymentSplits_ClinicIsolation は
// clinic B のコンテキストで clinic A の billing_id を使った SavePaymentSplits が
// clinic A の payment_splits を削除しないことを検証する。
// DELETE に clinic_id フィルタがなければ「clinic A の payment_split が残存する」が失敗する。
func TestAccountingRepository_SavePaymentSplits_ClinicIsolation(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	// clinic A の billing に payment_split を直接挿入する
	billingA := makeBillingRet(t, db, clinicA)
	splitA := model.PaymentSplit{
		ClinicID:  clinicA,
		BillingID: billingA.ID,
		Method:    model.PaymentMethodCash,
		Amount:    1000,
	}
	require.NoError(t, db.WithContext(ctx).Create(&splitA).Error)

	// clinic B のコンテキストで billingA.ID を指定して SavePaymentSplits を呼ぶ
	// （サービス層のバグで cross-tenant billing_id が渡された場合のシミュレーション）
	crossSplits := []model.PaymentSplit{{
		ClinicID:  clinicB,
		BillingID: billingA.ID, // clinic A の billing_id を指定
		Method:    model.PaymentMethodCreditCard,
		Amount:    9999,
	}}
	// clinic_id フィルタがあれば DELETE は 0 件（エラーなし）
	_ = repo.SavePaymentSplits(ctx, crossSplits)

	t.Run("clinic A の payment_split は clinic B からの SavePaymentSplits で削除されない", func(t *testing.T) {
		var splits []model.PaymentSplit
		require.NoError(t, db.Where("billing_id = ? AND clinic_id = ?", billingA.ID, clinicA).Find(&splits).Error)
		assert.Len(t, splits, 1, "clinic B からの SavePaymentSplits で clinic A の payment_split が削除されてはならない")
		if len(splits) > 0 {
			assert.Equal(t, splitA.ID, splits[0].ID)
		}
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
		amt := int64(9999)
		_, err := repo.Update(ctx, clinicB, billingA.ID, AccountingUpdate{TotalAmount: &amt})
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
		amt := int64(2000)
		got, err := repo.Update(ctx, clinicA, billingA.ID, AccountingUpdate{TotalAmount: &amt})
		require.NoError(t, err)
		assert.Equal(t, int64(2000), got.TotalAmount)
	})
}
