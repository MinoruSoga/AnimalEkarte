package billing

// tx_atomicity_test.go — 返金 INSERT + 監査の tx 内原子性（DB-backed）
//
// billing_refunds の INSERT が caller の ambient transaction に参加し、
// INSERT 後に同一 tx 内の後続処理（= 監査書込を模倣）が失敗したら
// INSERT がロールバックして billing_refunds が空のままであることを実 DB で実証する。
//
// - この原子性は mock では検証不可能（security M2 の教訓）。repository が dbOrTx で
//   ambient tx に join することが正本である。
// - temp-revert RED: refund_repository.go の Create を
//   dbOrTx(ctx, r.db) → r.db.WithContext(ctx) に戻すと、INSERT が独立 tx で即 commit され、
//   ambient tx の rollback では巻き戻らない → RollsBackWhenAmbientTxFails が RED になる。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// makeBillingForRefund は billing_refunds の FK を満たす最小 Billing を作成して返す。
// ownerID / petID は nil — FK 制約違反なし。フラット package の makeBillingWith
// （billing_test_fixtures_test.go）と同型を import cycle 回避のため直接構築する
// （BE8-4: 移動時の型リネームはしない方針の対象外の最小限複製）。
func makeBillingForRefund(t *testing.T, db *gorm.DB, clinicID uint64) *model.Billing {
	t.Helper()
	b := &model.Billing{
		ClinicID:      clinicID,
		TotalAmount:   10000,
		Status:        model.BillingStatusCompleted,
		ScheduledDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(b).Error)
	return b
}

// withTx mirrors repository.Transactor.WithTx (repohelpers-based ambient tx) without
// importing the flat repository package, which would create an import cycle
// (repository imports this subpackage via its facade).
func withTx(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx))
	})
}

// TestRefundRepository_Create_RollsBackWhenAmbientTxFails は
// ambient tx 内で refund INSERT した直後に後続処理が失敗したとき
// billing_refunds テーブルに行が残らないことを検証する（fail-closed）。
//
// SC1: audit + refund が同一 tx にある → audit 失敗（sentinel error）で両者がロールバックされる。
// SC2 temp-revert RED の手順:
//   - refund_repository.go の Create を dbOrTx(ctx, r.db) → r.db.WithContext(ctx) に戻す
//   - 本テストを実行 → INSERT が独立 tx で即 commit → rollback 対象外 → count=1 で FAIL
//   - 元に戻す → GREEN
func TestRefundRepository_Create_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForRefund(t, db, clinicA)
	repo := NewRefundRepository(db)

	refund := &model.BillingRefund{
		ClinicID:   clinicA,
		BillingID:  billing.ID,
		Amount:     3000,
		Reason:     "原子性テスト",
		RefundedAt: time.Now(),
	}
	sentinel := errors.New("simulated post-create audit failure")
	txErr := withTx(ctx, db, func(txCtx context.Context) error {
		if err := repo.Create(txCtx, refund); err != nil {
			return err
		}
		return sentinel // fail-closed: INSERT 後に監査書込失敗を模倣 → tx 中断
	})
	require.Error(t, txErr, "ambient tx の後続失敗で WithTx はエラーを返す")
	require.ErrorIs(t, txErr, sentinel)

	var count int64
	db.Model(&model.BillingRefund{}).Where("billing_id = ?", billing.ID).Count(&count)
	assert.EqualValues(t, 0, count, "監査失敗で INSERT はロールバックされ billing_refunds は空のまま（fail-closed）")
}

// TestRefundRepository_Create_CommitsWithinAmbientTx は
// ambient tx 内で refund INSERT が成功し commit されると
// billing_refunds テーブルに行が永続化されることを検証する。
func TestRefundRepository_Create_CommitsWithinAmbientTx(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForRefund(t, db, clinicA)
	repo := NewRefundRepository(db)

	refund := &model.BillingRefund{
		ClinicID:   clinicA,
		BillingID:  billing.ID,
		Amount:     5000,
		Reason:     "コミットテスト",
		RefundedAt: time.Now(),
	}
	require.NoError(t, withTx(ctx, db, func(txCtx context.Context) error {
		return repo.Create(txCtx, refund)
	}))

	var count int64
	db.Model(&model.BillingRefund{}).Where("billing_id = ? AND amount = ?", billing.ID, int64(5000)).Count(&count)
	assert.EqualValues(t, 1, count, "commit 後は billing_refunds に行が永続化される")
}
