package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockRefundRepo struct {
	createFn          func(ctx context.Context, refund *model.BillingRefund) error
	sumFn             func(ctx context.Context, clinicID, billingID uint64) (int64, error)
	sumByMethodFn     func(ctx context.Context, clinicID, billingID uint64, method model.PaymentMethod) (int64, error)
	findByBillingIDFn func(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error)
}

// noopAuditTxLogger は billingAuditTxLogger の noop 実装（テスト用）
type noopAuditTxLogger struct{}

func (n *noopAuditTxLogger) LogEntryTx(_ context.Context, _ *AuditEntry) error { return nil }

func (m *mockRefundRepo) Create(ctx context.Context, refund *model.BillingRefund) error {
	return m.createFn(ctx, refund)
}

func (m *mockRefundRepo) FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingRefund, error) {
	if m.findByBillingIDFn != nil {
		return m.findByBillingIDFn(ctx, clinicID, billingID)
	}
	return nil, nil
}

func (m *mockRefundRepo) SumByBillingID(ctx context.Context, clinicID, billingID uint64) (int64, error) {
	if m.sumFn != nil {
		return m.sumFn(ctx, clinicID, billingID)
	}
	return 0, nil
}

func (m *mockRefundRepo) SumByBillingIDAndPaymentMethod(ctx context.Context, clinicID, billingID uint64, method model.PaymentMethod) (int64, error) {
	if m.sumByMethodFn != nil {
		return m.sumByMethodFn(ctx, clinicID, billingID, method)
	}
	return 0, nil
}

func completedBillingForRefund() *model.Billing {
	return &model.Billing{ID: 1, ClinicID: 1, Status: model.BillingStatusCompleted, TotalAmount: 10000}
}

func billingWithSplits(total int64, splits ...model.PaymentSplit) *model.Billing {
	return &model.Billing{
		ID: 1, ClinicID: 1, Status: model.BillingStatusCompleted, TotalAmount: total, PaymentSplits: splits,
	}
}

func split(method model.PaymentMethod, amount int64) model.PaymentSplit {
	return model.PaymentSplit{ID: 1, ClinicID: 1, BillingID: 1, Method: method, Amount: amount}
}

// #60: 支払方法(ENUM)を指定した返金で PaymentMethod が保存されること。
func TestRefundService_Create_RecordsPaymentMethod(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return billingWithSplits(10000, split(model.PaymentMethodCreditCard, 5000)), nil
		},
	}
	var saved *model.BillingRefund
	refundRepo := &mockRefundRepo{
		createFn: func(_ context.Context, r *model.BillingRefund) error {
			saved = r
			r.ID = 100
			return nil
		},
	}
	mockTx := &mockTransactor{withTxErr: nil}
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

	method := model.PaymentMethodCreditCard
	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{
		StaffID:       ptrUint64(1),
		Amount:        3000,
		Reason:        "返金テスト",
		PaymentMethod: &method,
	})

	assert.NoError(t, err)
	if assert.NotNil(t, saved) {
		assert.Equal(t, &method, saved.PaymentMethod)
		assert.Equal(t, int64(3000), saved.Amount)
	}
}

// #60: 支払方法未指定（単一支払い・後方互換）では PaymentMethod が nil のまま、全体チェックのみ。
func TestRefundService_Create_NilPaymentMethodWhenUnspecified(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return completedBillingForRefund(), nil
		},
	}
	var saved *model.BillingRefund
	refundRepo := &mockRefundRepo{
		createFn: func(_ context.Context, r *model.BillingRefund) error {
			saved = r
			r.ID = 101
			return nil
		},
	}
	mockTx := &mockTransactor{withTxErr: nil}
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{
		StaffID: ptrUint64(1),
		Amount:  2000,
		Reason:  "",
	})

	assert.NoError(t, err)
	if assert.NotNil(t, saved) {
		assert.Nil(t, saved.PaymentMethod)
	}
}

// 既存バリデーション維持: 支払済み以外の請求は返金不可。
func TestRefundService_Create_RejectsNonCompletedBilling(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return &model.Billing{ID: 1, ClinicID: 1, Status: model.BillingStatusPending, TotalAmount: 10000}, nil
		},
	}
	refundRepo := &mockRefundRepo{
		createFn: func(_ context.Context, _ *model.BillingRefund) error { return nil },
	}
	mockTx := &mockTransactor{withTxErr: nil}
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 1000})
	assert.Error(t, err)
}

// #60 Phase 2: 支払方法別上限超過は 400（カード ¥3,000 受取に ¥3,500 カード返金）。
func TestRefundService_Create_PaymentMethodExceedsReceived(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return billingWithSplits(5000, split(model.PaymentMethodCreditCard, 3000), split(model.PaymentMethodCash, 2000)), nil
		},
	}
	refundRepo := &mockRefundRepo{
		createFn: func(_ context.Context, _ *model.BillingRefund) error { return nil },
	}
	mockTx := &mockTransactor{withTxErr: nil}
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

	method := model.PaymentMethodCreditCard
	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 3500, PaymentMethod: &method})
	assert.Error(t, err) // 全体(3500<=5000)は通るが方法別(3500>3000)で拒否
}

// #60 Phase 2: その支払方法での支払がない場合は 400。
func TestRefundService_Create_PaymentMethodNotUsed(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return billingWithSplits(5000, split(model.PaymentMethodCash, 5000)), nil
		},
	}
	refundRepo := &mockRefundRepo{
		createFn: func(_ context.Context, _ *model.BillingRefund) error { return nil },
	}
	mockTx := &mockTransactor{withTxErr: nil}
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

	method := model.PaymentMethodCreditCard // 現金のみの会計にカード返金
	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 1000, PaymentMethod: &method})
	assert.Error(t, err)
}

// #60 Phase 2: 同一支払方法の返金累積で上限超過は 400（カード ¥3,000 受取、既返金 ¥1,000、追加 ¥2,500）。
func TestRefundService_Create_PaymentMethodAccumulates(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return billingWithSplits(3000, split(model.PaymentMethodCreditCard, 3000)), nil
		},
	}
	refundRepo := &mockRefundRepo{
		createFn: func(_ context.Context, _ *model.BillingRefund) error { return nil },
		sumByMethodFn: func(_ context.Context, _, _ uint64, _ model.PaymentMethod) (int64, error) {
			return 1000, nil // カードで既に ¥1,000 返金済み
		},
	}
	mockTx := &mockTransactor{withTxErr: nil}
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

	method := model.PaymentMethodCreditCard
	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 2500, PaymentMethod: &method})
	assert.Error(t, err) // 1000 + 2500 > 3000
}

// #60 Phase 2: 上限内なら成功（カード ¥3,000 受取、既返金 ¥1,000、追加 ¥1,500）。
func TestRefundService_Create_PaymentMethodWithinLimit(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return billingWithSplits(5000, split(model.PaymentMethodCreditCard, 3000), split(model.PaymentMethodCash, 2000)), nil
		},
	}
	var saved *model.BillingRefund
	refundRepo := &mockRefundRepo{
		createFn: func(_ context.Context, r *model.BillingRefund) error {
			saved = r
			r.ID = 200
			return nil
		},
		sumByMethodFn: func(_ context.Context, _, _ uint64, _ model.PaymentMethod) (int64, error) {
			return 1000, nil
		},
	}
	mockTx := &mockTransactor{withTxErr: nil}
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

	method := model.PaymentMethodCreditCard
	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 1500, PaymentMethod: &method})
	assert.NoError(t, err) // 1000 + 1500 <= 3000
	if assert.NotNil(t, saved) {
		assert.Equal(t, &method, saved.PaymentMethod)
	}
}

// capturingAuditTxLogger は refund 監査テスト用の billingAuditTxLogger モック。LogEntryTx のみ capture する。
type capturingAuditTxLogger struct {
	lastInput *AuditEntry
	err       error
}

func (c *capturingAuditTxLogger) LogEntryTx(_ context.Context, input *AuditEntry) error {
	c.lastInput = input
	if c.err != nil {
		return c.err
	}
	return nil
}

// TestRefundService_Create_AuditMinimalJSON は #122: NewValue が全Struct保存でなく
// 最小JSONフィールド（amount/reason/payment_method）のみを含むことを検証する。
func TestRefundService_Create_AuditMinimalJSON(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return completedBillingForRefund(), nil
		},
	}
	var saved *model.BillingRefund
	refundRepo := &mockRefundRepo{
		createFn: func(_ context.Context, r *model.BillingRefund) error {
			saved = r
			r.ID = 42
			return nil
		},
	}
	mockTx := &mockTransactor{withTxErr: nil}
	auditSvc := &capturingAuditTxLogger{}
	svc := NewRefundService(refundRepo, accountRepo, auditSvc, mockTx)

	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{
		StaffID: ptrUint64(1),
		Amount:  3000,
		Reason:  "返金テスト",
	})

	assert.NoError(t, err)
	require.NotNil(t, auditSvc.lastInput, "audit LogEntryTx should be called")
	assert.Equal(t, model.AuditActionBillingRefundCreate, auditSvc.lastInput.Action)
	assert.Equal(t, "billing_refund", auditSvc.lastInput.Resource)

	require.NotNil(t, auditSvc.lastInput.NewValue, "NewValue must not be nil")
	newVal, ok := auditSvc.lastInput.NewValue.(map[string]any)
	require.True(t, ok, "NewValue は map[string]any でなければならない（全Struct禁止）")
	assert.EqualValues(t, int64(3000), newVal["amount"])
	assert.Equal(t, "返金テスト", newVal["reason"])
	_, hasBillingID := newVal["billing_id"]
	assert.False(t, hasBillingID, "全Struct保存禁止: billing_id は NewValue に含まれてはいけない")
	_ = saved
}

// 金額が0以下の場合は即座にバリデーションエラー（トランザクションに入る前）。
func TestRefundService_Create_RejectsNonPositiveAmount(t *testing.T) {
	accountRepo := &mockAccountingRepository{}
	refundRepo := &mockRefundRepo{}
	mockTx := &mockTransactor{withTxErr: nil}
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 0})
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))

	_, err = svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: -100})
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

// LockAndFindByID（billing取得）が失敗した場合はラップされたエラーを返す。
func TestRefundService_Create_LockAndFindByIDError(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return nil, errors.New("db error")
		},
	}
	refundRepo := &mockRefundRepo{}
	mockTx := &mockTransactor{withTxErr: nil}
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 1000})
	assert.Error(t, err)
}

// SumByBillingID の集計エラーはラップされて伝播する。
func TestRefundService_Create_SumByBillingIDError(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return completedBillingForRefund(), nil
		},
	}
	refundRepo := &mockRefundRepo{
		sumFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, errors.New("sum error")
		},
	}
	mockTx := &mockTransactor{withTxErr: nil}
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 1000})
	assert.Error(t, err)
}

// 返金額が請求残高を超える場合は400エラー。
func TestRefundService_Create_ExceedsAvailableBalance(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return completedBillingForRefund(), nil // TotalAmount: 10000
		},
	}
	refundRepo := &mockRefundRepo{
		sumFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 9000, nil // already refunded 9000, available = 1000
		},
	}
	mockTx := &mockTransactor{withTxErr: nil}
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 1500})
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

// refund repo.Create の失敗はラップされて返る。
func TestRefundService_Create_RefundCreateRepoError(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return completedBillingForRefund(), nil
		},
	}
	refundRepo := &mockRefundRepo{
		createFn: func(_ context.Context, _ *model.BillingRefund) error {
			return errors.New("insert failed")
		},
	}
	mockTx := &mockTransactor{withTxErr: nil}
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 1000})
	assert.Error(t, err)
}

// 監査ログ書き込み失敗は fail-closed（返金自体も失敗として扱われる #211）。
func TestRefundService_Create_AuditLogFailureIsFailClosed(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return completedBillingForRefund(), nil
		},
	}
	refundRepo := &mockRefundRepo{
		createFn: func(_ context.Context, r *model.BillingRefund) error {
			r.ID = 500
			return nil
		},
	}
	mockTx := &mockTransactor{withTxErr: nil}
	auditSvc := &capturingAuditTxLogger{err: errors.New("audit write failed")}
	svc := NewRefundService(refundRepo, accountRepo, auditSvc, mockTx)

	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 1000, Reason: "test"})
	assert.Error(t, err)
}

// トランザクション自体が失敗（WithTx が伝播するエラー）した場合はラップされて返る。
func TestRefundService_Create_TransactorError(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return completedBillingForRefund(), nil
		},
	}
	refundRepo := &mockRefundRepo{}
	mockTx := &mockTransactor{withTxErr: errors.New("tx failed")}
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 1000})
	assert.Error(t, err)
}

// ---- ListByBillingID tests ----

func TestRefundService_ListByBillingID(t *testing.T) {
	t.Run("returns refunds when billing exists", func(t *testing.T) {
		accountRepo := &mockAccountingRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
				return completedBillingForRefund(), nil
			},
		}
		refundRepo := &mockRefundRepo{
			findByBillingIDFn: func(_ context.Context, _, _ uint64) ([]model.BillingRefund, error) {
				return []model.BillingRefund{{ID: 1, Amount: 1000}, {ID: 2, Amount: 2000}}, nil
			},
		}
		mockTx := &mockTransactor{}
		svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

		items, err := svc.ListByBillingID(context.Background(), 1, 1)
		assert.NoError(t, err)
		assert.Len(t, items, 2)
	})

	t.Run("returns wrapped error when billing lookup fails (multi-tenant protection)", func(t *testing.T) {
		accountRepo := &mockAccountingRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
				return nil, apperrors.WrapNotFound("billing", "999")
			},
		}
		refundRepo := &mockRefundRepo{}
		mockTx := &mockTransactor{}
		svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

		items, err := svc.ListByBillingID(context.Background(), 1, 999)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, items)
	})

	t.Run("returns wrapped error when repo.FindByBillingID fails", func(t *testing.T) {
		accountRepo := &mockAccountingRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
				return completedBillingForRefund(), nil
			},
		}
		refundRepo := &mockRefundRepo{
			findByBillingIDFn: func(_ context.Context, _, _ uint64) ([]model.BillingRefund, error) {
				return nil, errors.New("db error")
			},
		}
		mockTx := &mockTransactor{}
		svc := NewRefundService(refundRepo, accountRepo, &noopAuditTxLogger{}, mockTx)

		items, err := svc.ListByBillingID(context.Background(), 1, 1)
		assert.Error(t, err)
		assert.Nil(t, items)
	})
}

func TestRefundService_Create_LocksCloseBoundary(t *testing.T) {
	scheduled := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			b := completedBillingForRefund()
			b.ScheduledDate = scheduled
			return b, nil
		},
	}
	refundRepo := &mockRefundRepo{
		createFn: func(_ context.Context, r *model.BillingRefund) error {
			r.ID = 200
			return nil
		},
	}
	var locked time.Time
	closeRepo := &mockCashRegisterCloseRepository{
		lockCloseBoundaryFn: func(_ context.Context, clinicID uint64, date time.Time) error {
			assert.Equal(t, uint64(1), clinicID)
			locked = date
			return nil
		},
	}
	svc := NewRefundService(
		refundRepo, accountRepo, &noopAuditTxLogger{}, &mockTransactor{},
		WithRefundCloseRepository(closeRepo),
	)

	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{
		StaffID: ptrUint64(1),
		Amount:  1000,
		Reason:  "返金",
	})
	assert.NoError(t, err)
	assert.Equal(t, "2026-06-01", locked.Format(time.DateOnly))
}
