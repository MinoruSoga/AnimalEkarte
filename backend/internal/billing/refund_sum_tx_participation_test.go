package billing

// sum_tx_participation_test.go — BE-refactor.md R1-1 (D2) の DB-backed 原子性証明
//
// 背景: refund_service.Create は WithTx 内で SumByBillingID / SumByBillingIDAndPaymentMethod を
// 呼び「返金可能残額チェック（トランザクション内で再計算）」する。しかし実装は r.db.WithContext(ctx)
// を直接参照しており dbOrTx を経由しないため、ambient tx に参加しない別セッションで実行される。
//
// 標準の READ COMMITTED 分離レベルでは、別セッションは他 tx の未コミット行を見えない。
// このテストは Create（既に dbOrTx で正しく ambient tx に参加している）で ambient tx 内に
// 未コミットの refund 行を作り、同一 tx 内で Sum* を呼んでその未コミット行が合計に含まれることを検証する。
// これが #211 refund_tx_atomicity_test.go 同様、tx 参加の直接証明になる。
//
// temp-revert RED の手順:
//   - refund_repository.go の SumByBillingID / SumByBillingIDAndPaymentMethod を
//     dbOrTx(ctx, r.db) → r.db.WithContext(ctx) に戻す
//   - 本ファイルのテストを実行 → 未コミット行を合計に含められず sum が 0 のまま RED
//   - 元に戻す → GREEN

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestRefundRepository_SumByBillingID_SeesUncommittedInsertWithinAmbientTx(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForRefund(t, db, clinicA)
	repo := NewRefundRepository(db)

	err := withTx(ctx, db, func(txCtx context.Context) error {
		refund := &model.BillingRefund{
			ClinicID:   clinicA,
			BillingID:  billing.ID,
			Amount:     4000,
			Reason:     "tx参加確認",
			RefundedAt: time.Now(),
		}
		// Create は既に dbOrTx で ambient tx に参加している（正本）。この INSERT は未コミット。
		if err := repo.Create(txCtx, refund); err != nil {
			return err
		}

		sum, err := repo.SumByBillingID(txCtx, clinicA, billing.ID)
		if err != nil {
			return err
		}
		assert.EqualValues(t, 4000, sum,
			"SumByBillingID は同一 ambient tx 内の未コミット INSERT を合計に含められるべき（tx 参加の証明）")
		return nil
	})
	require.NoError(t, err)

	// tx 成功でコミットされた後も、通常の読み取りで確認できる（回帰確認）。
	committedSum, sumErr := NewRefundRepository(db).SumByBillingID(ctx, clinicA, billing.ID)
	require.NoError(t, sumErr)
	assert.EqualValues(t, 4000, committedSum)
}

func TestRefundRepository_SumByBillingIDAndPaymentMethod_SeesUncommittedInsertWithinAmbientTx(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForRefund(t, db, clinicA)
	repo := NewRefundRepository(db)

	err := withTx(ctx, db, func(txCtx context.Context) error {
		method := model.PaymentMethodCreditCard
		refund := &model.BillingRefund{
			ClinicID:      clinicA,
			BillingID:     billing.ID,
			Amount:        2500,
			Reason:        "tx参加確認（支払方法別）",
			PaymentMethod: &method,
			RefundedAt:    time.Now(),
		}
		if err := repo.Create(txCtx, refund); err != nil {
			return err
		}

		sum, err := repo.SumByBillingIDAndPaymentMethod(txCtx, clinicA, billing.ID, method)
		if err != nil {
			return err
		}
		assert.EqualValues(t, 2500, sum,
			"SumByBillingIDAndPaymentMethod は同一 ambient tx 内の未コミット INSERT を合計に含められるべき（tx 参加の証明）")
		return nil
	})
	require.NoError(t, err)
}
