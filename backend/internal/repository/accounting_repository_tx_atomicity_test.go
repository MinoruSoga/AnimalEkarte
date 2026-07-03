package repository

// accounting_repository_tx_atomicity_test.go — BE-refactor.md R1-1 (D2) の DB-backed 原子性証明
//
// 背景: accounting_repository の LockAndFindByID / SavePayment / SavePaymentSplits は
// refund_service.Create と accounting_service_correction.CorrectCreditPayment の両方から
// txCtx 付きで呼ばれ、呼び出し元は Transactor.WithTx の ambient tx に参加することを前提にしている
// （LockAndFindByID の docstring は明示的に「TOCTOU を防止する」と主張している）。
// しかし実装は r.db.WithContext(ctx) を直接参照しており dbOrTx を経由しないため、
// 実際には ambient tx に参加しない別セッションで実行される。
//
// SavePaymentSplits はさらに深刻: r.db.WithContext(ctx).Transaction(...) で独立した
// 新規トランザクションを開始する。ambient tx の後続処理が失敗して rollback しても、
// SavePaymentSplits の書込は独立して既にコミット済みのため巻き戻らない
// （= 部分コミット。返金・クレジット訂正の金額整合が壊れうる）。
//
// temp-revert RED の手順:
//   - accounting_repository.go の対象メソッドを dbOrTx(ctx, r.db) → r.db.WithContext(ctx) に戻す
//     （SavePaymentSplits は dbOrTx(ctx, r.db).Transaction(...) → r.db.WithContext(ctx).Transaction(...)）
//   - 本ファイルのテストを実行 → Rollback 系が RED（書込が残ってしまう）、
//     BlocksConcurrentAmbientTx が RED（後続呼び出しがブロックされず即完了する）
//   - 元に戻す → GREEN
//
// 注意: LockAndFindByID の検証は自己参照的な自己可視性テストではなく、独立した2 ambient tx 間の
// 排他制御テストにしている。理由は下記 LockAndFindByID セクションのコメントを参照。

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// makeBillingForAccountingTx は payments/payment_splits 系テスト用の最小 Billing を作成する。
// payment_splits FK を満たすため clinic_id を明示する。ownerID/petID は nil（FK 制約違反なし）。
func makeBillingForAccountingTx(t *testing.T, db *gorm.DB, clinicID uint64) *model.Billing {
	t.Helper()
	b := &model.Billing{
		ClinicID:      clinicID,
		TotalAmount:   8000,
		Status:        model.BillingStatusCompleted,
		ScheduledDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(b).Error)
	return b
}

// ─── Update（Cancel の tx 化により新たに txCtx 付きで呼ばれるようになった経路） ──────────

// TestAccountingRepository_Update_RollsBackWhenAmbientTxFails は、Cancel の R1-2 移行で
// Update が初めて ambient tx 内から呼ばれるようになったため、その tx 参加を検証する。
func TestAccountingRepository_Update_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := NewTransactor(db)

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := repo.Update(txCtx, clinicA, billing.ID, map[string]any{"status": model.BillingStatusCancelled}); err != nil {
			return err
		}
		return errSentinelAccountingTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelAccountingTx)

	var reloaded model.Billing
	require.NoError(t, db.WithContext(ctx).First(&reloaded, billing.ID).Error)
	assert.Equal(t, model.BillingStatusCompleted, reloaded.Status,
		"ambient tx 失敗時、Update による status 変更はロールバックされる")
}

func TestAccountingRepository_Update_CommitsWithinAmbientTx(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := NewTransactor(db)

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		_, err := repo.Update(txCtx, clinicA, billing.ID, map[string]any{"status": model.BillingStatusCancelled})
		return err
	}))

	var reloaded model.Billing
	require.NoError(t, db.WithContext(ctx).First(&reloaded, billing.ID).Error)
	assert.Equal(t, model.BillingStatusCancelled, reloaded.Status, "commit 後は status 変更が永続化される")
}

// ─── SavePayment ─────────────────────────────────────────────────────────────

func TestAccountingRepository_SavePayment_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := NewTransactor(db)

	payment := &model.Payment{
		BillingID:     billing.ID,
		TotalAmount:   8000,
		BillingAmount: 8000,
		Method:        model.PaymentMethodCash,
	}
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.SavePayment(txCtx, payment); err != nil {
			return err
		}
		return errSentinelAccountingTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelAccountingTx)

	var count int64
	db.Model(&model.Payment{}).Where("billing_id = ?", billing.ID).Count(&count)
	assert.EqualValues(t, 0, count, "ambient tx 失敗時、SavePayment の新規作成はロールバックされる（fail-closed）")
}

func TestAccountingRepository_SavePayment_CommitsWithinAmbientTx(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := NewTransactor(db)

	payment := &model.Payment{
		BillingID:     billing.ID,
		TotalAmount:   8000,
		BillingAmount: 8000,
		Method:        model.PaymentMethodCash,
	}
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		return repo.SavePayment(txCtx, payment)
	}))

	var count int64
	db.Model(&model.Payment{}).Where("billing_id = ? AND billing_amount = ?", billing.ID, int64(8000)).Count(&count)
	assert.EqualValues(t, 1, count, "commit 後は payments に行が永続化される")
}

// ─── SavePaymentSplits（部分コミットの実証: r.db.Transaction が独立 tx を開始する誤り） ──

func TestAccountingRepository_SavePaymentSplits_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := NewTransactor(db)

	splits := []model.PaymentSplit{
		{ClinicID: clinicA, BillingID: billing.ID, Method: model.PaymentMethodCash, Amount: 8000, ReceivedAmount: 8000},
	}
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.SavePaymentSplits(txCtx, splits); err != nil {
			return err
		}
		return errSentinelAccountingTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelAccountingTx)

	var count int64
	db.Model(&model.PaymentSplit{}).Where("billing_id = ?", billing.ID).Count(&count)
	assert.EqualValues(t, 0, count,
		"ambient tx 失敗時、SavePaymentSplits の書込はロールバックされる。"+
			"バグ時は r.db.Transaction が独立 tx で即コミットするため count=1 になり FAIL する（部分コミットの実証）")
}

func TestAccountingRepository_SavePaymentSplits_CommitsWithinAmbientTx(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := NewTransactor(db)

	splits := []model.PaymentSplit{
		{ClinicID: clinicA, BillingID: billing.ID, Method: model.PaymentMethodCash, Amount: 8000, ReceivedAmount: 8000},
	}
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		return repo.SavePaymentSplits(txCtx, splits)
	}))

	var count int64
	db.Model(&model.PaymentSplit{}).Where("billing_id = ?", billing.ID).Count(&count)
	assert.EqualValues(t, 1, count, "commit 後は payment_splits に行が永続化される")
}

// ─── LockAndFindByID（FOR UPDATE が ambient tx をまたいで実際に排他するか） ──────────
//
// 設計メモ: 当初は「同一 tx 内で billings.status を未コミット更新した直後に
// LockAndFindByID で読めるか」という自己参照的な自己可視性テストを試みたが、これは
// バグのある実装（別セッション）に対して本物のデッドロックを引き起こした
// （ambient tx が保持する暗黙ロックと、別セッションの FOR UPDATE が相互待機し、
// go test のデフォルト 600s タイムアウトでプロセスごと強制終了された）。
// この経験自体が「FOR UPDATE ロックが ambient tx を保護していない」ことの動作実証でもある
// （本番では該当コネクションプールの枯渇として現れうる）。
//
// 安全かつ本質的な検証のため、自己参照ではなく「独立した2つの ambient tx が同一行を
// LockAndFindByID しようとしたとき、後続が先行の commit までブロックされるか」を
// goroutine + channel で検証する（真の排他制御の直接証明。デッドロックの起きようがない）。

// TestAccountingRepository_LockAndFindByID_BlocksConcurrentAmbientTx は、
// 先行 tx が LockAndFindByID で行ロックを保持している間、後続の独立した tx の
// LockAndFindByID が commit までブロックされることを検証する。
//
// バグ時（r.db.WithContext(ctx) 直参照）は FOR UPDATE が単一ステートメントの暗黙 tx で
// 即座に解放されるため、後続呼び出しはブロックされず即完了する（本テストは FAIL する）。
func TestAccountingRepository_LockAndFindByID_BlocksConcurrentAmbientTx(t *testing.T) {
	db := setupTestDB(t)
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := NewTransactor(db)

	lockAcquired := make(chan struct{})
	proceedToCommit := make(chan struct{})
	secondCallDone := make(chan struct{})

	// go-reviewer 指摘: t.Fatal で早期リターンした場合でも先行 tx の goroutine が
	// <-proceedToCommit で永遠にブロックされたまま残らないよう、テスト終了時に必ず解放する。
	// closeProceedToCommit は sync.Once で二重 close panic を防ぐ（通常経路でも明示的に呼ぶため）。
	var closeOnce sync.Once
	closeProceedToCommit := func() { closeOnce.Do(func() { close(proceedToCommit) }) }
	t.Cleanup(closeProceedToCommit)

	go func() {
		_ = tx.WithTx(context.Background(), func(txCtx context.Context) error {
			if _, err := repo.LockAndFindByID(txCtx, clinicA, billing.ID); err != nil {
				return err
			}
			close(lockAcquired)
			<-proceedToCommit // 明示的に指示されるまで tx を開いたまま保持し、ロックを維持する
			return nil
		})
	}()

	select {
	case <-lockAcquired:
	case <-time.After(5 * time.Second):
		t.Fatal("先行 tx が LockAndFindByID でロックを取得できなかった")
	}

	go func() {
		_ = tx.WithTx(context.Background(), func(txCtx context.Context) error {
			_, err := repo.LockAndFindByID(txCtx, clinicA, billing.ID)
			return err
		})
		close(secondCallDone)
	}()

	// 後続呼び出しが即座に完了しない（＝ブロックされている）ことを確認する。
	select {
	case <-secondCallDone:
		t.Fatal("後続の LockAndFindByID は先行 tx の commit までブロックされるべき" +
			"（FOR UPDATE が ambient tx を実際に保護している証明）。即完了した場合、" +
			"別セッションで実行されておりロックが機能していない（バグ時の挙動）")
	case <-time.After(300 * time.Millisecond):
		// 期待通りブロックされている。
	}

	closeProceedToCommit() // 先行 tx を commit させ、行ロックを解放する

	select {
	case <-secondCallDone:
		// OK: ロック解放後に後続呼び出しが完了した。
	case <-time.After(5 * time.Second):
		t.Fatal("先行 tx の commit 後も後続の LockAndFindByID が完了しなかった")
	}
}

var errSentinelAccountingTx = &accountingTxSentinelError{}

type accountingTxSentinelError struct{}

func (e *accountingTxSentinelError) Error() string {
	return "simulated post-write failure in ambient tx"
}
