package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockRefundRepo struct {
	createFn func(ctx context.Context, refund *model.BillingRefund) error
	sumFn    func(ctx context.Context, clinicID, billingID uint64) (int64, error)
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

func completedBillingForRefund() *model.Billing {
	return &model.Billing{ID: 1, ClinicID: 1, Status: model.BillingStatusCompleted, TotalAmount: 10000}
}

// #60: 支払方法を指定した返金で PaymentMethod / PaymentMethodID が保存されること。
func TestRefundService_Create_RecordsPaymentMethod(t *testing.T) {
	accountRepo := &mockAccountingRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return completedBillingForRefund(), nil
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
	svc := NewRefundService(refundRepo, accountRepo)

	method := model.PaymentMethodCreditCard
	methodID := uint64(5)
	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{
		StaffID:         ptrUint64(1),
		Amount:          3000,
		Reason:          "返金テスト",
		PaymentMethod:   &method,
		PaymentMethodID: &methodID,
	})

	assert.NoError(t, err)
	if assert.NotNil(t, saved) {
		assert.Equal(t, &method, saved.PaymentMethod)
		assert.Equal(t, &methodID, saved.PaymentMethodID)
		assert.Equal(t, int64(3000), saved.Amount)
	}
}

// #60: 支払方法未指定（単一支払い・後方互換）では PaymentMethod / PaymentMethodID が nil のままであること。
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
	svc := NewRefundService(refundRepo, accountRepo)

	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{
		StaffID: ptrUint64(1),
		Amount:  2000,
		Reason:  "",
	})

	assert.NoError(t, err)
	if assert.NotNil(t, saved) {
		assert.Nil(t, saved.PaymentMethod)
		assert.Nil(t, saved.PaymentMethodID)
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
	svc := NewRefundService(refundRepo, accountRepo)

	_, err := svc.Create(context.Background(), 1, 1, CreateRefundInput{Amount: 1000})
	assert.Error(t, err)
}
