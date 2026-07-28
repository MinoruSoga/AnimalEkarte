package billing

// repository_test.go — RefundRepository の実 DB 結合テスト（#212: internal/repository カバレッジ向上）。
//
// ローカル実測: FindByBillingID は 0% だった。Create/SumByBillingID/SumByBillingIDAndPaymentMethod は
// tx_atomicity_test.go / sum_tx_participation_test.go で正常系・tx参加は
// ある程度カバー済みのため、本ファイルは異常系・clinic_id 隔離・FK バリデーションを優先する。
//
// makeBillingForRefund は tx_atomicity_test.go に定義済みのものを再利用する。

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

// setupRefundRepoTestDB は billing_refunds の FK/Preload テストに必要な staffs と
// staff_clinic_assignments を追加で AutoMigrate する
// （setupSharedTestSchema の基本テーブル集合には含まれないため）。
// assignment の clinic 親行は makeStaffForRefundRepositoryTest が作成する。
func setupRefundRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Staff{},
		&model.StaffClinicAssignment{},
	))
	return db
}

// makeStaffForRefundRepositoryTest は billing_refunds.refunded_by の Preload テスト用に
// 最小構成の Staff を作成する。
func makeStaffForRefundRepositoryTest(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	s := &model.Staff{
		ClinicID:  clinicID,
		Name:      name,
		StaffType: model.StaffTypeDoctor,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	seedBillingClinicForFK(t, db, clinicID)
	require.NoError(t, db.WithContext(context.Background()).Create(
		&model.StaffClinicAssignment{StaffID: s.ID, ClinicID: clinicID},
	).Error)
	return s
}

func TestRefundRepository_FindByBillingID(t *testing.T) {
	db := setupRefundRepoTestDB(t)
	repo := NewRefundRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("正常系(RefundedByStaff Preload込)", func(t *testing.T) {
		billing := makeBillingForRefund(t, db, clinicA)
		staff := makeStaffForRefundRepositoryTest(t, db, clinicA, "返金担当スタッフ")
		refund := &model.BillingRefund{
			ClinicID:   clinicA,
			BillingID:  billing.ID,
			Amount:     3000,
			Reason:     "破損返金",
			RefundedBy: &staff.ID,
			RefundedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, refund))

		result, err := repo.FindByBillingID(ctx, clinicA, billing.ID)
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, refund.ID, result[0].ID)
		assert.EqualValues(t, 3000, result[0].Amount)
		require.NotNil(t, result[0].RefundedByStaff, "RefundedByStaff が Preload されている")
		assert.Equal(t, "返金担当スタッフ", result[0].RefundedByStaff.Name)
	})

	t.Run("別クリニックのbilling_idでは空配列", func(t *testing.T) {
		billing := makeBillingForRefund(t, db, clinicA)
		refund := &model.BillingRefund{
			ClinicID:   clinicA,
			BillingID:  billing.ID,
			Amount:     1500,
			Reason:     "クリニックB混入確認用",
			RefundedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, refund))

		// FindByBillingID は billing_refunds.clinic_id を直接 Scopes(clinicScope) で絞り込む実装のため、
		// billing_id が一致していても clinic_id が異なれば空配列（エラーではない）。
		result, err := repo.FindByBillingID(ctx, clinicB, billing.ID)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("該当なしで空配列", func(t *testing.T) {
		result, err := repo.FindByBillingID(ctx, clinicA, 999999999)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("複数返金は作成日時降順で返る", func(t *testing.T) {
		billing := makeBillingForRefund(t, db, clinicA)
		older := &model.BillingRefund{
			ClinicID:   clinicA,
			BillingID:  billing.ID,
			Amount:     1000,
			Reason:     "1件目",
			RefundedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, older))
		time.Sleep(2 * time.Millisecond) // created_at DESC 順を安定させるため作成時刻をずらす
		newer := &model.BillingRefund{
			ClinicID:   clinicA,
			BillingID:  billing.ID,
			Amount:     2000,
			Reason:     "2件目",
			RefundedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, newer))

		result, err := repo.FindByBillingID(ctx, clinicA, billing.ID)
		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, newer.ID, result[0].ID, "created_at DESC で新しい方が先頭")
		assert.Equal(t, older.ID, result[1].ID)
	})
}

func TestRefundRepository_Create_InvalidBillingID(t *testing.T) {
	db := setupRefundRepoTestDB(t)
	repo := NewRefundRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	t.Run("存在しないbilling_idはFK違反エラー", func(t *testing.T) {
		refund := &model.BillingRefund{
			ClinicID:   clinicA,
			BillingID:  999999999,
			Amount:     1000,
			Reason:     "FK違反テスト",
			RefundedAt: time.Now(),
		}
		err := repo.Create(ctx, refund)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "FK違反は参照先不存在としてInvalidInputに変換される")
	})
}

func TestRefundRepository_SumByBillingID(t *testing.T) {
	db := setupRefundRepoTestDB(t)
	repo := NewRefundRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("複数返金の合計が正しい", func(t *testing.T) {
		billing := makeBillingForRefund(t, db, clinicA)
		require.NoError(t, repo.Create(ctx, &model.BillingRefund{
			ClinicID:   clinicA,
			BillingID:  billing.ID,
			Amount:     1000,
			Reason:     "1件目",
			RefundedAt: time.Now(),
		}))
		require.NoError(t, repo.Create(ctx, &model.BillingRefund{
			ClinicID:   clinicA,
			BillingID:  billing.ID,
			Amount:     2500,
			Reason:     "2件目",
			RefundedAt: time.Now(),
		}))

		total, err := repo.SumByBillingID(ctx, clinicA, billing.ID)
		require.NoError(t, err)
		assert.EqualValues(t, 3500, total)
	})

	t.Run("対象データなしで0", func(t *testing.T) {
		billing := makeBillingForRefund(t, db, clinicA)
		total, err := repo.SumByBillingID(ctx, clinicA, billing.ID)
		require.NoError(t, err)
		assert.EqualValues(t, 0, total)
	})

	t.Run("別クリニックの返金は合計に含まれない", func(t *testing.T) {
		billing := makeBillingForRefund(t, db, clinicA)
		require.NoError(t, repo.Create(ctx, &model.BillingRefund{
			ClinicID:   clinicA,
			BillingID:  billing.ID,
			Amount:     4000,
			Reason:     "クリニックA分",
			RefundedAt: time.Now(),
		}))

		total, err := repo.SumByBillingID(ctx, clinicB, billing.ID)
		require.NoError(t, err)
		assert.EqualValues(t, 0, total)
	})
}

func TestRefundRepository_SumByBillingIDAndPaymentMethod(t *testing.T) {
	db := setupRefundRepoTestDB(t)
	repo := NewRefundRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	t.Run("method絞り込みが正しく機能する", func(t *testing.T) {
		billing := makeBillingForRefund(t, db, clinicA)
		cash := model.PaymentMethodCash
		card := model.PaymentMethodCreditCard
		require.NoError(t, repo.Create(ctx, &model.BillingRefund{
			ClinicID:      clinicA,
			BillingID:     billing.ID,
			Amount:        1000,
			Reason:        "現金分1",
			PaymentMethod: &cash,
			RefundedAt:    time.Now(),
		}))
		require.NoError(t, repo.Create(ctx, &model.BillingRefund{
			ClinicID:      clinicA,
			BillingID:     billing.ID,
			Amount:        1500,
			Reason:        "現金分2",
			PaymentMethod: &cash,
			RefundedAt:    time.Now(),
		}))
		require.NoError(t, repo.Create(ctx, &model.BillingRefund{
			ClinicID:      clinicA,
			BillingID:     billing.ID,
			Amount:        5000,
			Reason:        "カード分",
			PaymentMethod: &card,
			RefundedAt:    time.Now(),
		}))

		cashTotal, err := repo.SumByBillingIDAndPaymentMethod(ctx, clinicA, billing.ID, cash)
		require.NoError(t, err)
		assert.EqualValues(t, 2500, cashTotal, "現金分のみ合計される")

		cardTotal, err := repo.SumByBillingIDAndPaymentMethod(ctx, clinicA, billing.ID, card)
		require.NoError(t, err)
		assert.EqualValues(t, 5000, cardTotal)
	})

	t.Run("対象データなしで0", func(t *testing.T) {
		billing := makeBillingForRefund(t, db, clinicA)
		total, err := repo.SumByBillingIDAndPaymentMethod(ctx, clinicA, billing.ID, model.PaymentMethodCash)
		require.NoError(t, err)
		assert.EqualValues(t, 0, total)
	})
}
