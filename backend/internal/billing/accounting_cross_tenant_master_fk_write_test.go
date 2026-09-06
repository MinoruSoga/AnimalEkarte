package billing

// accounting_cross_tenant_master_fk_write_test.go — BE9-2C B④:
// service/cross_tenant_master_fk_write_test.go から accountingService 節を同名移動。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestAccountingService_Update_RejectsForeignPaymentMethodID(t *testing.T) {
	const clinicA = uint64(1)
	const clinicB = uint64(2)
	const clinicACashID = uint64(101)
	const clinicBCashID = uint64(201)
	billingAmount := int64(5000)
	cashKey := "cash"

	payMethodRepo := &mockPaymentMethodMasterRepository{
		findAllFn: func(_ context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
			switch clinicID {
			case clinicA:
				return []model.PaymentMethodMaster{{ID: clinicACashID, ClinicID: clinicA, SystemKey: &cashKey}}, nil
			case clinicB:
				return []model.PaymentMethodMaster{{ID: clinicBCashID, ClinicID: clinicB, SystemKey: &cashKey}}, nil
			default:
				return nil, nil
			}
		},
	}

	newSvc := func(saved *bool) AccountingService {
		repo := &mockAccountingRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Billing, error) {
				return &model.Billing{ID: id, ClinicID: clinicID, Status: model.BillingStatusCompleted}, nil
			},
			updateFieldsFn: func(_ context.Context, clinicID, id uint64, _ AccountingUpdate) (*model.Billing, error) {
				return &model.Billing{ID: id, ClinicID: clinicID}, nil
			},
			savePaymentSplitsFn: func(_ context.Context, _ []model.PaymentSplit) error {
				*saved = true
				return nil
			},
		}
		return NewAccountingService(repo, nil, nil, nil, nil, &mockTransactor{}, &mockAuditService{}, payMethodRepo)
	}

	t.Run("rejects clinic B's payment_method_id on clinic A's billing update and does not persist", func(t *testing.T) {
		saved := false
		svc := newSvc(&saved)
		foreign := clinicBCashID // clinic B の cash master id（実在するが clinic A のものではない）
		out, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID:            1,
			ClinicID:      clinicA,
			BillingAmount: &billingAmount,
			PaymentSplits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, PaymentMethodID: &foreign, Amount: 5000, ReceivedAmount: 5000},
			},
		})
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, out)
		assert.False(t, saved, "billing must NOT be persisted with another clinic's payment_method_id")
	})

	t.Run("accepts clinic A's own payment_method_id (no false-reject)", func(t *testing.T) {
		saved := false
		svc := newSvc(&saved)
		owned := clinicACashID
		out, err := svc.Update(context.Background(), &UpdateAccountingInput{
			ID:            1,
			ClinicID:      clinicA,
			BillingAmount: &billingAmount,
			PaymentSplits: []PaymentSplitInput{
				{Method: model.PaymentMethodCash, PaymentMethodID: &owned, Amount: 5000, ReceivedAmount: 5000},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.True(t, saved)
	})
}

// ── campaign (target-item FK): TargetItemIDs → campaign_target_items.merchandise_item_id ──
