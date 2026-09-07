package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// completedCardBilling は credit_card 単一支払いの確定済み会計を生成するテストヘルパー。
func completedCardBilling() *model.Billing {
	return &model.Billing{
		ID:          10,
		ClinicID:    1,
		Status:      model.BillingStatusCompleted,
		Subtotal:    9091,
		TaxTotal:    909,
		TotalAmount: 10000,
		Payments: []model.Payment{{
			BillingID:      10,
			Subtotal:       9091,
			TaxTotal:       909,
			TotalAmount:    10000,
			BillingAmount:  10000,
			ReceivedAmount: 0,
			ChangeAmount:   0,
			InsuranceName:  "アニコム",
			InsuranceRatio: 0.5,
			DiscountAmount: 0,
			Method:         model.PaymentMethodCreditCard,
		}},
		PaymentSplits: []model.PaymentSplit{{
			ID:             1,
			ClinicID:       1,
			BillingID:      10,
			Method:         model.PaymentMethodCreditCard,
			Amount:         10000,
			ReceivedAmount: 0,
			ChangeAmount:   0,
		}},
	}
}

// TestAccountingService_Cancel_NilAuditFailsClosed pins BIL-02 for Cancel (owned test file).
func TestAccountingService_Cancel_NilAuditFailsClosed(t *testing.T) {
	actorID := uint64(42)
	repo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return &model.Billing{ID: 10, ClinicID: 1, Status: model.BillingStatusWaiting}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ AccountingUpdate) (*model.Billing, error) {
			return &model.Billing{ID: 10, ClinicID: 1, Status: model.BillingStatusCancelled}, nil
		},
	}
	svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, nil /* auditTx */, &mockPaymentMethodMasterRepository{})
	err := svc.Cancel(context.Background(), 1, 10, &actorID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "billing audit dependency is required")
}

func TestAccountingService_CorrectCreditPayment(t *testing.T) {
	staffID := uint64(7)

	t.Run("正常: クレジット金額を訂正し billing_amount を再計算・監査記録する", func(t *testing.T) {
		billing := completedCardBilling()
		var savedPayment *model.Payment
		savedSplits := &[]model.PaymentSplit{}
		repo := &mockAccountingRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) { return billing, nil },
			savePaymentFn: func(_ context.Context, p *model.Payment) error {
				cp := *p
				savedPayment = &cp
				return nil
			},
			savePaymentSplitsFn: func(_ context.Context, splits []model.PaymentSplit) error {
				cp := make([]model.PaymentSplit, len(splits))
				copy(cp, splits)
				*savedSplits = cp
				return nil
			},
		}
		audit := &mockAuditService{}
		svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, audit, &mockPaymentMethodMasterRepository{})

		_, err := svc.CorrectCreditPayment(context.Background(), &CorrectCreditPaymentInput{
			ClinicID:  1,
			BillingID: 10,
			StaffID:   &staffID,
			Method:    model.PaymentMethodCreditCard,
			Amount:    12000,
			Reason:    "クレジット端末の打ち間違い",
			Memo:      "本人確認済み",
		})
		assert.NoError(t, err)

		// 支払いヘッダ: billing_amount は訂正値に再計算、医療費(subtotal/tax/total)と保険は不変
		assert.NotNil(t, savedPayment)
		assert.Equal(t, int64(12000), savedPayment.BillingAmount, "billing_amount は訂正後 split 合計に再計算される")
		assert.Equal(t, int64(9091), savedPayment.Subtotal, "subtotal は不変")
		assert.Equal(t, int64(909), savedPayment.TaxTotal, "tax_total は不変")
		assert.Equal(t, int64(10000), savedPayment.TotalAmount, "total_amount(医療費) は不変")
		assert.Equal(t, "アニコム", savedPayment.InsuranceName, "保険名は不変")

		// payment_splits: カード split のみ訂正
		assert.Len(t, *savedSplits, 1)
		assert.Equal(t, model.PaymentMethodCreditCard, (*savedSplits)[0].Method)
		assert.Equal(t, int64(12000), (*savedSplits)[0].Amount)
		assert.Equal(t, int64(0), (*savedSplits)[0].ReceivedAmount)
		assert.Equal(t, int64(0), (*savedSplits)[0].ChangeAmount)

		// 監査: before/after + reason
		assert.True(t, audit.logEntryTxCalled, "訂正時は監査ログ必須")
		assert.Equal(t, model.AuditActionBillingCreditCorrection, audit.logEntryTxInput.Action)
		assert.Equal(t, "billing", audit.logEntryTxInput.Resource)
		assert.Equal(t, &staffID, audit.logEntryTxInput.ActorID)
		old, ok := audit.logEntryTxInput.OldValue.(map[string]any)
		assert.True(t, ok, "OldValue は map")
		assert.Equal(t, int64(10000), old["amount"])
		nv, ok := audit.logEntryTxInput.NewValue.(map[string]any)
		assert.True(t, ok, "NewValue は map")
		assert.Equal(t, int64(12000), nv["amount"])
		meta, ok := audit.logEntryTxInput.Metadata.(map[string]any)
		assert.True(t, ok, "Metadata は map")
		assert.Equal(t, "クレジット端末の打ち間違い", meta["reason"])
		assert.Equal(t, "本人確認済み", meta["memo"])

		// M-1: billing_amount の before/after と差額を監査に記録する（総額 10000 → 12000）
		assert.Equal(t, int64(10000), old["billing_amount"], "OldValue に訂正前 billing_amount")
		assert.Equal(t, int64(12000), nv["billing_amount"], "NewValue に訂正後 billing_amount")
		assert.Equal(t, int64(2000), meta["billing_amount_delta"], "Metadata に billing_amount 差額(delta)")
		// 締め前（通常）訂正は post_close を記録しない
		_, hasPostClose := meta["post_close"]
		assert.False(t, hasPostClose, "締め前の訂正では post_close を記録しない")
	})

	t.Run("M-2: 締め済み期間の訂正は post_close を監査に記録する（拒否はしない）", func(t *testing.T) {
		billing := completedCardBilling()
		billing.ScheduledDate = time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
		repo := &mockAccountingRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) { return billing, nil },
		}
		audit := &mockAuditService{}
		var capturedAdj *model.CashRegisterCloseAdjustment
		closeRepo := postCloseCloseRepoForTests(billing.ScheduledDate)
		closeRepo.createAdjustmentFn = func(_ context.Context, adj *model.CashRegisterCloseAdjustment) error {
			capturedAdj = adj
			return nil
		}
		svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, audit, &mockPaymentMethodMasterRepository{},
			WithCashRegisterCloseRepository(closeRepo))

		_, err := svc.CorrectCreditPayment(context.Background(), &CorrectCreditPaymentInput{
			ClinicID: 1, BillingID: 10, StaffID: &staffID,
			Method: model.PaymentMethodCreditCard, Amount: 12000, Reason: "締め後のカード端末訂正",
			IsPostClose: true,
		})
		assert.NoError(t, err, "締め済み期間でも訂正は成功する（M-2 は拒否しない・可視化のみ）")

		assert.True(t, audit.logEntryTxCalled, "締め済み訂正も監査ログを記録する")
		meta, ok := audit.logEntryTxInput.Metadata.(map[string]any)
		assert.True(t, ok, "Metadata は map")
		assert.Equal(t, true, meta["post_close"], "締め済み期間の訂正は post_close=true を記録")
		assert.Equal(t, "2026-06-30", meta["post_close_date"], "対象締めの識別子として予定日を記録")
		// W-013 HIGH-2: adjustment 台帳にも追記する
		if assert.NotNil(t, capturedAdj) {
			assert.Equal(t, int64(2000), capturedAdj.AccountingDelta)
			assert.Equal(t, "締め後のカード端末訂正", capturedAdj.Reason)
		}
	})

	// BIL-02: auditTx 欠落は成功扱いしない（logPostCloseEdit と同型 fail-closed）。
	t.Run("監査dependency欠落時は訂正を拒否する（fail-closed）", func(t *testing.T) {
		billing := completedCardBilling()
		saveCalled := false
		repo := &mockAccountingRepository{
			findByIDFn:          func(_ context.Context, _, _ uint64) (*model.Billing, error) { return billing, nil },
			savePaymentFn:       func(_ context.Context, _ *model.Payment) error { saveCalled = true; return nil },
			savePaymentSplitsFn: func(_ context.Context, _ []model.PaymentSplit) error { return nil },
		}
		svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, nil /* auditTx */, &mockPaymentMethodMasterRepository{})

		_, err := svc.CorrectCreditPayment(context.Background(), &CorrectCreditPaymentInput{
			ClinicID: 1, BillingID: 10, StaffID: &staffID,
			Method: model.PaymentMethodCreditCard, Amount: 12000, Reason: "監査依存欠落",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "billing audit dependency is required")
		// save may run before audit in the same tx; fail-closed means the method still errors.
		_ = saveCalled
	})

	// BE-refactor.md R1-2 手順3（go-reviewer HIGH-2 指摘の是正）: money-critical なクレジット訂正でも
	// 監査書込（LogEntryTx）が失敗すると tx 全体がロールバックし訂正ごと無効になる（fail-closed）ことを固定する。
	// 他の6移行サイト（Cancel/Update/medicine/dose-param/treatment）と同型の失敗注入回帰テスト。
	// これが無いと将来 logCreditCorrection を fire-and-forget に戻す退行を検出できない。
	t.Run("監査失敗時は訂正全体をロールバックする（fail-closed）", func(t *testing.T) {
		billing := completedCardBilling()
		saveCalled := false
		repo := &mockAccountingRepository{
			findByIDFn:          func(_ context.Context, _, _ uint64) (*model.Billing, error) { return billing, nil },
			savePaymentFn:       func(_ context.Context, _ *model.Payment) error { saveCalled = true; return nil },
			savePaymentSplitsFn: func(_ context.Context, _ []model.PaymentSplit) error { return nil },
		}
		audit := &mockAuditService{logEntryTxErr: errors.New("audit write failed")}
		svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, audit, &mockPaymentMethodMasterRepository{})

		_, err := svc.CorrectCreditPayment(context.Background(), &CorrectCreditPaymentInput{
			ClinicID: 1, BillingID: 10, StaffID: &staffID,
			Method: model.PaymentMethodCreditCard, Amount: 12000, Reason: "監査失敗の注入",
		})

		assert.Error(t, err, "監査失敗はクレジット訂正全体を失敗させる（fail-closed）")
		assert.True(t, audit.logEntryTxCalled, "監査書込は試行される")
		assert.True(t, saveCalled, "監査は payment 保存の後に呼ばれる（同一 tx 内で rollback 対象）")
	})

	denied := []struct {
		name   string
		input  *CorrectCreditPaymentInput
		setup  func() *model.Billing
		errSub string
	}{
		{
			name:  "拒否: 理由が空",
			setup: completedCardBilling,
			input: &CorrectCreditPaymentInput{ClinicID: 1, BillingID: 10, StaffID: &staffID, Method: model.PaymentMethodCreditCard, Amount: 12000, Reason: "  "},
		},
		{
			name:  "拒否: 現金(カード以外)は対象外",
			setup: completedCardBilling,
			input: &CorrectCreditPaymentInput{ClinicID: 1, BillingID: 10, StaffID: &staffID, Method: model.PaymentMethodCash, Amount: 12000, Reason: "理由"},
		},
		{
			name:  "拒否: 金額が0以下",
			setup: completedCardBilling,
			input: &CorrectCreditPaymentInput{ClinicID: 1, BillingID: 10, StaffID: &staffID, Method: model.PaymentMethodCreditCard, Amount: 0, Reason: "理由"},
		},
		{
			name:  "拒否: 金額が負",
			setup: completedCardBilling,
			input: &CorrectCreditPaymentInput{ClinicID: 1, BillingID: 10, StaffID: &staffID, Method: model.PaymentMethodCreditCard, Amount: -100, Reason: "理由"},
		},
		{
			name: "拒否: 確定済み(completed)以外は訂正不可",
			setup: func() *model.Billing {
				b := completedCardBilling()
				b.Status = model.BillingStatusWaiting
				return b
			},
			input: &CorrectCreditPaymentInput{ClinicID: 1, BillingID: 10, StaffID: &staffID, Method: model.PaymentMethodCreditCard, Amount: 12000, Reason: "理由"},
		},
		{
			name: "拒否: 対象のカード支払い内訳が存在しない",
			setup: func() *model.Billing {
				b := completedCardBilling()
				b.PaymentSplits = []model.PaymentSplit{{ID: 1, ClinicID: 1, BillingID: 10, Method: model.PaymentMethodCash, Amount: 10000, ReceivedAmount: 10000, ChangeAmount: 0}}
				return b
			},
			input: &CorrectCreditPaymentInput{ClinicID: 1, BillingID: 10, StaffID: &staffID, Method: model.PaymentMethodCreditCard, Amount: 12000, Reason: "理由"},
		},
	}
	for _, tc := range denied {
		t.Run(tc.name, func(t *testing.T) {
			billing := tc.setup()
			saveCalled := false
			repo := &mockAccountingRepository{
				findByIDFn:    func(_ context.Context, _, _ uint64) (*model.Billing, error) { return billing, nil },
				savePaymentFn: func(_ context.Context, _ *model.Payment) error { saveCalled = true; return nil },
			}
			audit := &mockAuditService{}
			svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, audit, &mockPaymentMethodMasterRepository{})

			_, err := svc.CorrectCreditPayment(context.Background(), tc.input)
			assert.Error(t, err, "不正入力/状態は決定的にエラーを返す")
			assert.False(t, saveCalled, "拒否時は支払いを保存しない")
			assert.False(t, audit.logEntryTxCalled, "拒否時は監査を記録しない")
		})
	}

	t.Run("回帰(#188): 現金 split のお釣り上書きを保持したままカードを訂正できる", func(t *testing.T) {
		// cash: amount=5000 received=5000 だが change=300(直接上書き、派生値0と不一致)、card: amount=5000
		billing := &model.Billing{
			ID: 10, ClinicID: 1, Status: model.BillingStatusCompleted,
			Subtotal: 9091, TaxTotal: 909, TotalAmount: 10000,
			Payments: []model.Payment{{BillingID: 10, BillingAmount: 10000, ReceivedAmount: 5000, ChangeAmount: 300, Method: model.PaymentMethodCash}},
			PaymentSplits: []model.PaymentSplit{
				{ID: 1, ClinicID: 1, BillingID: 10, Method: model.PaymentMethodCash, Amount: 5000, ReceivedAmount: 5000, ChangeAmount: 300},
				{ID: 2, ClinicID: 1, BillingID: 10, Method: model.PaymentMethodCreditCard, Amount: 5000, ReceivedAmount: 0, ChangeAmount: 0},
			},
		}
		savedSplits := &[]model.PaymentSplit{}
		repo := &mockAccountingRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) { return billing, nil },
			savePaymentSplitsFn: func(_ context.Context, splits []model.PaymentSplit) error {
				cp := make([]model.PaymentSplit, len(splits))
				copy(cp, splits)
				*savedSplits = cp
				return nil
			},
		}
		audit := &mockAuditService{}
		svc := NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, audit, &mockPaymentMethodMasterRepository{})

		_, err := svc.CorrectCreditPayment(context.Background(), &CorrectCreditPaymentInput{
			ClinicID: 1, BillingID: 10, StaffID: &staffID,
			Method: model.PaymentMethodCreditCard, Amount: 6000, Reason: "カード訂正",
		})
		assert.NoError(t, err, "現金お釣り上書きが存在してもカード訂正は成功する(#188非破壊)")
		assert.Len(t, *savedSplits, 2)
		for _, s := range *savedSplits {
			switch s.Method {
			case model.PaymentMethodCash:
				assert.Equal(t, int64(5000), s.Amount, "現金 split は不変")
				assert.Equal(t, int64(300), s.ChangeAmount, "現金お釣り上書きは保持")
			case model.PaymentMethodCreditCard:
				assert.Equal(t, int64(6000), s.Amount, "カードは訂正値")
			}
		}
	})
}
