package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockRefundRepo struct {
	createFn      func(ctx context.Context, refund *model.BillingRefund) error
	sumFn         func(ctx context.Context, clinicID, billingID uint64) (int64, error)
	sumByMethodFn func(ctx context.Context, clinicID, billingID uint64, method model.PaymentMethod) (int64, error)
}

// noopAuditService is a minimal audit service for testing
type noopAuditService struct{}

func (n *noopAuditService) Log(ctx context.Context, log *model.AuditLog) error       { return nil }
func (n *noopAuditService) LogEntry(ctx context.Context, input *AuditLogInput) error { return nil }
func (n *noopAuditService) LogAuthLogin(ctx context.Context, clinicID, staffID *uint64, action, ip, ua string) error {
	return nil
}
func (n *noopAuditService) LogLstepOperation(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64) error {
	return nil
}
func (n *noopAuditService) LogLstepOperationWithMetadata(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64, metadata any) error {
	return nil
}
func (n *noopAuditService) LogMedicalRecordChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, recordID uint64, oldValue, newValue map[string]any) error {
	return nil
}
func (n *noopAuditService) LogVitalChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, vitalID, medicalRecordID uint64, oldValue, newValue map[string]any) error {
	return nil
}
func (n *noopAuditService) LogAddendumCreate(ctx context.Context, clinicID uint64, actorID *uint64, addendumID, medicalRecordID uint64, addendum *model.MedicalRecordAddendum) error {
	return nil
}
func (n *noopAuditService) LogClinicSwitch(ctx context.Context, actorID *uint64, fromClinicID, toClinicID uint64, ipAddress, userAgent string) error {
	return nil
}

func (m *mockRefundRepo) Create(ctx context.Context, refund *model.BillingRefund) error {
	return m.createFn(ctx, refund)
}

func (m *mockRefundRepo) FindByBillingID(_ context.Context, _, _ uint64) ([]model.BillingRefund, error) {
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
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditService{}, mockTx)

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
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditService{}, mockTx)

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
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditService{}, mockTx)

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
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditService{}, mockTx)

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
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditService{}, mockTx)

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
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditService{}, mockTx)

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
	svc := NewRefundService(refundRepo, accountRepo, &noopAuditService{}, mockTx)

	method := model.PaymentMethodCreditCard
	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 1500, PaymentMethod: &method})
	assert.NoError(t, err) // 1000 + 1500 <= 3000
	if assert.NotNil(t, saved) {
		assert.Equal(t, &method, saved.PaymentMethod)
	}
}
