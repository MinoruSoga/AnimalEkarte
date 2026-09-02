package billing

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
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// makeBillingForAccountingTx は payments/payment_splits 系テスト用の最小 Billing を作成する。
// payment_splits FK を満たすため clinic_id を明示する。ownerID/petID は nil（FK 制約違反なし）。
// makeBillingWith（billing_test_fixtures_test.go）への thin wrapper。
func makeBillingForAccountingTx(t *testing.T, db *gorm.DB, clinicID uint64) *model.Billing {
	t.Helper()
	return makeBillingWith(t, db, billingFixtureOpts{
		ClinicID:      clinicID,
		TotalAmount:   8000,
		Status:        model.BillingStatusCompleted,
		ScheduledDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
}

// ─── Update（Cancel の tx 化により新たに txCtx 付きで呼ばれるようになった経路） ──────────

// TestAccountingRepository_Update_RollsBackWhenAmbientTxFails は、Cancel の R1-2 移行で
// Update が初めて ambient tx 内から呼ばれるようになったため、その tx 参加を検証する。
func TestAccountingRepository_Update_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := testNewTransactor(db)

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
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := testNewTransactor(db)

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		_, err := repo.Update(txCtx, clinicA, billing.ID, map[string]any{"status": model.BillingStatusCancelled})
		return err
	}))

	var reloaded model.Billing
	require.NoError(t, db.WithContext(ctx).First(&reloaded, billing.ID).Error)
	assert.Equal(t, model.BillingStatusCancelled, reloaded.Status, "commit 後は status 変更が永続化される")
}

func TestAccountingRepository_FindByID_SeesAmbientTransactionState(t *testing.T) {
	tests := []struct {
		name string
		find func(context.Context, AccountingRepository, uint64, uint64) (*model.Billing, error)
	}{
		{
			name: "single clinic",
			find: func(ctx context.Context, repo AccountingRepository, clinicID, billingID uint64) (*model.Billing, error) {
				return repo.FindByID(ctx, clinicID, billingID)
			},
		},
		{
			name: "multiple clinics",
			find: func(ctx context.Context, repo AccountingRepository, clinicID, billingID uint64) (*model.Billing, error) {
				return repo.FindByIDForClinics(ctx, []uint64{clinicID}, billingID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testdb.SetupTestDB(t)
			ctx := context.Background()
			const clinicID = uint64(1)

			billing := makeBillingForAccountingTx(t, db, clinicID)
			repo := NewAccountingRepository(db)
			memo := "ambient transaction reload"
			payment := &model.Payment{
				BillingID:     billing.ID,
				TotalAmount:   8000,
				BillingAmount: 8000,
				Method:        model.PaymentMethodCash,
			}

			require.NoError(t, testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
				if _, err := repo.Update(txCtx, clinicID, billing.ID, map[string]any{"memo": memo}); err != nil {
					return err
				}
				if err := repo.SavePayment(txCtx, payment); err != nil {
					return err
				}
				got, err := tt.find(txCtx, repo, clinicID, billing.ID)
				if err != nil {
					return err
				}
				assert.Equal(t, memo, got.Memo, "commit前reloadは同じtxのbilling更新を読む")
				require.Len(t, got.Payments, 1, "commit前reloadは同じtxのpayment作成を読む")
				assert.Equal(t, int64(8000), got.Payments[0].BillingAmount)
				return nil
			}))
		})
	}
}

func TestAccountingService_Update_PostCloseMissingAuditDependencyRollsBack(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	originalMemo := billing.Memo
	reason := "締め後訂正"
	updatedMemo := "監査なしでは保存しない"
	repo := NewAccountingRepository(db)
	// close repo は adjustment 成功を返すモック。auditTx=nil なら監査欠落で fail-closed になることを固定する。
	// （実テーブル cash_register_close_adjustments は migration 003 適用後に存在。本テストは audit 欠落に焦点）
	closeRepo := &mockCashRegisterCloseRepository{
		findByDateAndPeriodFn: func(_ context.Context, _ uint64, _ time.Time, period string) (*model.CashRegisterClose, error) {
			if period == "am" {
				return &model.CashRegisterClose{ID: 1, ClinicID: clinicA, Period: "am"}, nil
			}
			return nil, nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, nil, nil, testNewTransactor(db), nil, nil,
		WithCashRegisterCloseRepository(closeRepo))

	got, err := svc.Update(ctx, &UpdateAccountingInput{
		ID:              billing.ID,
		ClinicID:        clinicA,
		Memo:            &updatedMemo,
		IsPostClose:     true,
		PostCloseReason: &reason,
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "audit")
	var reloaded model.Billing
	require.NoError(t, db.First(&reloaded, billing.ID).Error)
	assert.Equal(t, originalMemo, reloaded.Memo, "監査依存欠落時は会計更新もrollbackする")
}

// ─── SavePayment ─────────────────────────────────────────────────────────────

func TestAccountingRepository_SavePayment_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := testNewTransactor(db)

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
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := testNewTransactor(db)

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

// TestAccountingRepository_SavePayment_PersistsClinicIDFromBilling は TASK-445:
// SavePayment が lockBillingClinic で得た billing.clinic_id を payments.clinic_id に
// 永続化することを raw 列読取で証明する（モデル json:"-" のため FE wire には出ない）。
func TestAccountingRepository_SavePayment_PersistsClinicIDFromBilling(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	require.Equal(t, clinicA, billing.ClinicID)

	repo := NewAccountingRepository(db)
	payment := &model.Payment{
		BillingID:     billing.ID,
		TotalAmount:   8000,
		BillingAmount: 8000,
		Method:        model.PaymentMethodCash,
		// ClinicID intentionally omitted: SavePayment must derive it from billing.
	}
	require.NoError(t, repo.SavePayment(ctx, payment))
	assert.Equal(t, clinicA, payment.ClinicID, "SavePayment は in-memory payment.ClinicID も billing 由来にセットする")

	var rawClinicID uint64
	require.NoError(t, db.WithContext(ctx).
		Raw("SELECT clinic_id FROM payments WHERE billing_id = ?", billing.ID).
		Scan(&rawClinicID).Error)
	assert.Equal(t, clinicA, rawClinicID, "payments.clinic_id は billing.clinic_id と一致して永続化される")

	// Update path also refreshes clinic_id (fields map includes clinic_id).
	paymentUpdate := &model.Payment{
		BillingID:     billing.ID,
		TotalAmount:   9000,
		BillingAmount: 9000,
		Method:        model.PaymentMethodCash,
	}
	require.NoError(t, repo.SavePayment(ctx, paymentUpdate))
	require.NoError(t, db.WithContext(ctx).
		Raw("SELECT clinic_id FROM payments WHERE billing_id = ?", billing.ID).
		Scan(&rawClinicID).Error)
	assert.Equal(t, clinicA, rawClinicID, "update 後も payments.clinic_id は billing.clinic_id のまま")
	assert.Equal(t, clinicA, paymentUpdate.ClinicID)
}

// ─── SavePaymentSplits（部分コミットの実証: r.db.Transaction が独立 tx を開始する誤り） ──

func TestAccountingRepository_SavePaymentSplits_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := testNewTransactor(db)

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
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := testNewTransactor(db)

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

// ─── reservation.CompleteForAccounting（BE-refactor.md X-12: billing 確定と appointment ────
// 完了化の部分コミット修正） ──────────────────────────────────────────────

// TestReservationRepository_CompleteForAccounting_RollsBackWhenAmbientTxFails は、
// billing の status 更新（Update）と appointment 完了化（CompleteForAccounting）を
// 同一 ambient tx 内で行った場合、後続失敗で両方がロールバックされることを検証する。
//
// バグ時（CompleteForAccounting が r.db.WithContext(ctx) 直参照で dbOrTx 非参加）は
// 別セッションで即コミットされるため、billing の Update はロールバックされても appointment の
// 完了化だけは残ってしまう（このテストでは逆に、Update 側が正しく tx 参加している前提のもと
// CompleteForAccounting 側が非参加だと reloadAppointmentStatus が Completed のまま
// FAIL する——旧 X-12 failure mode の反対方向だが、本質は同じ「一部だけ確定する部分コミット」）。
func TestReservationRepository_CompleteForAccounting_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := &model.Billing{
		ClinicID:      clinicA,
		TotalAmount:   5000,
		Status:        model.BillingStatusWaiting,
		ScheduledDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	require.NoError(t, db.WithContext(ctx).Create(billing).Error)

	owner := testdb.MakeTestOwner(t, db, clinicA, "X-12ロールバック飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "X-12ロールバックペット")
	appt := makeAccountingAppointment(t, db, clinicA, &owner.ID, &pet.ID, model.ReservationStatusPending,
		time.Date(2026, 7, 1, 3, 0, 0, 0, time.UTC))
	mr := makeMedicalRecordForAppointment(t, db, clinicA, appt.ID, "MR-X12-rollback")

	repo := NewAccountingRepository(db)
	appointmentRepo := reservation.NewReservationRepository(db)
	tx := testNewTransactor(db)

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := repo.Update(txCtx, clinicA, billing.ID, map[string]any{"status": model.BillingStatusCompleted}); err != nil {
			return err
		}
		if _, err := appointmentRepo.CompleteForAccounting(txCtx, clinicA, &mr.ID, nil, nil, time.Time{}); err != nil {
			return err
		}
		return errSentinelAccountingTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelAccountingTx)

	var reloadedBilling model.Billing
	require.NoError(t, db.WithContext(ctx).First(&reloadedBilling, billing.ID).Error)
	assert.Equal(t, model.BillingStatusWaiting, reloadedBilling.Status,
		"ambient tx 失敗時、billing の status 更新はロールバックされる")

	assert.Equal(t, model.ReservationStatusPending, reloadAppointmentStatus(t, db, appt.ID),
		"ambient tx 失敗時、CompleteForAccounting による appointment 完了化もロールバックされる"+
			"（X-12 旧 failure mode = billing 確定済み・appointment 完了化のみ失敗の部分コミットが再現しないことの証明）。"+
			"バグ時（r.db.WithContext(ctx) 直参照）は独立セッションで即コミットするため Completed のままとなり FAIL する")
}

func TestReservationRepository_CompleteForAccounting_CommitsWithinAmbientTx(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := &model.Billing{
		ClinicID:      clinicA,
		TotalAmount:   5000,
		Status:        model.BillingStatusWaiting,
		ScheduledDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	require.NoError(t, db.WithContext(ctx).Create(billing).Error)

	owner := testdb.MakeTestOwner(t, db, clinicA, "X-12コミット飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "X-12コミットペット")
	appt := makeAccountingAppointment(t, db, clinicA, &owner.ID, &pet.ID, model.ReservationStatusPending,
		time.Date(2026, 7, 1, 3, 0, 0, 0, time.UTC))
	mr := makeMedicalRecordForAppointment(t, db, clinicA, appt.ID, "MR-X12-commit")

	repo := NewAccountingRepository(db)
	appointmentRepo := reservation.NewReservationRepository(db)
	tx := testNewTransactor(db)

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := repo.Update(txCtx, clinicA, billing.ID, map[string]any{"status": model.BillingStatusCompleted}); err != nil {
			return err
		}
		_, err := appointmentRepo.CompleteForAccounting(txCtx, clinicA, &mr.ID, nil, nil, time.Time{})
		return err
	}))

	var reloadedBilling model.Billing
	require.NoError(t, db.WithContext(ctx).First(&reloadedBilling, billing.ID).Error)
	assert.Equal(t, model.BillingStatusCompleted, reloadedBilling.Status, "commit 後は billing status が永続化される")
	assert.Equal(t, model.ReservationStatusCompleted, reloadAppointmentStatus(t, db, appt.ID), "commit 後は appointment も完了化される")
}

// TestAccountingRepository_Create_RollsBackWhenAmbientTxFails は、accounting_service_core.Create
// の X-12 修正（repo.Create + CompleteForAccounting を単一 tx で括る）の repo 側前提を検証する。
// バグ時（Create が r.db.WithContext(ctx) 直参照で dbOrTx 非参加）は billing の INSERT が独立
// セッションで即コミットされるため、後続失敗時も billing 行が残ってしまう（旧 X-12 Create failure mode:
// medical_record_id が NULL の手動会計・トリミング会計はリトライで二重 billing を作りうる）。
func TestAccountingRepository_Create_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	repo := NewAccountingRepository(db)
	tx := testNewTransactor(db)

	billing := &model.Billing{
		TotalAmount:   3000,
		Status:        model.BillingStatusCompleted,
		ScheduledDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Create(txCtx, clinicA, billing); err != nil {
			return err
		}
		return errSentinelAccountingTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelAccountingTx)

	var count int64
	db.Model(&model.Billing{}).Where("clinic_id = ? AND total_amount = ?", clinicA, int64(3000)).Count(&count)
	assert.EqualValues(t, 0, count,
		"ambient tx 失敗時、Create による billing 新規作成はロールバックされる（X-12 の二重billingリスク根絶）")
}

func TestAccountingRepository_Create_CommitsWithinAmbientTx(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	repo := NewAccountingRepository(db)
	tx := testNewTransactor(db)

	billing := &model.Billing{
		TotalAmount:   3000,
		Status:        model.BillingStatusCompleted,
		ScheduledDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		return repo.Create(txCtx, clinicA, billing)
	}))

	var count int64
	db.Model(&model.Billing{}).Where("clinic_id = ? AND total_amount = ?", clinicA, int64(3000)).Count(&count)
	assert.EqualValues(t, 1, count, "commit 後は billing が永続化される")
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
	db := testdb.SetupTestDB(t)
	const clinicA = uint64(1)

	billing := makeBillingForAccountingTx(t, db, clinicA)
	repo := NewAccountingRepository(db)
	tx := testNewTransactor(db)

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
