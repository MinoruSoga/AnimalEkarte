package medicalrecord

// hospitalization_discharge_guard_test.go — EMR-253: 退院時会計の no-duplicate ガード。
//   - 非 cancelled 占有者（waiting/completed 等）がいる入院では Create を呼ばず既存 ID を再利用する
//   - cancelled 占有者は同一 tx で soft-delete してスロットを解放し、新規 waiting を作成する
//   - 再利用時は care-plan 明細の再取込・totals 上書きを行わない（既存行の明細を維持）

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func dischargeOccupant(id uint64, status model.BillingStatus) *model.Billing {
	hospID := uint64(10)
	ownerID := uint64(2)
	petID := uint64(5)
	return &model.Billing{
		ID:                id,
		ClinicID:          1,
		HospitalizationID: &hospID,
		OwnerID:           &ownerID,
		PetID:             &petID,
		Status:            status,
		Subtotal:          2000,
		TaxTotal:          200,
		TotalAmount:       2200,
	}
}

func TestHospitalizationService_DischargeWithBilling_ReusesExistingNonCancelledBilling(t *testing.T) {
	actorID := uint64(42)
	audit := &hospitalizationAuditRecorder{}
	hospRepo := admittedHospitalizationRepo(func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
		return &model.Hospitalization{ID: 10}, nil
	})
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			t.Fatal("reused billing must not re-fetch care plan items")
			return nil, nil
		},
	}
	accountingRepo := &mockAccountingRepository{
		findByHospitalizationIDFn: func(_ context.Context, clinicID, hospID uint64) (*model.Billing, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(10), hospID)
			return dischargeOccupant(55, model.BillingStatusWaiting), nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("existing non-cancelled billing must not be duplicated")
			return nil
		},
	}
	billingItemRepo := &mockBillingItemRepository{
		createFn: func(_ context.Context, _ *model.BillingItem) error {
			t.Fatal("reused billing must not append care-plan items")
			return nil
		},
		updateBillingTotals: func(_ context.Context, _, _ uint64, _, _, _ int64) error {
			t.Fatal("reused billing totals must not be rewritten")
			return nil
		},
	}
	deps := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, billingItemRepo)
	svc := NewHospitalizationServiceWithAudit(
		hospRepo, deps.reservation, nil, nil, carePlanRepo, accountingRepo, billingItemRepo, &mockTransactor{}, audit,
	)

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
		DischargeDate:    time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
		CreateAccounting: true,
		ActorID:          &actorID,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.AccountingID)
	assert.Equal(t, uint64(55), *result.AccountingID, "reused billing id must be returned")

	require.Len(t, audit.entries, 1)
	entry := audit.entries[0]
	assert.Equal(t, model.AuditActionHospitalizationDischargeWithBilling, entry.Action)
	assert.Equal(t, map[string]any{
		"billing_id":      uint64(55),
		"subtotal_amount": int64(2000),
		"tax_amount":      int64(200),
		"total_amount":    int64(2200),
		"reused":          true,
	}, entry.NewValue)
}

func TestHospitalizationService_DischargeWithBilling_ReleasesCancelledBillingThenCreates(t *testing.T) {
	actorID := uint64(42)
	audit := &hospitalizationAuditRecorder{}
	hospRepo := admittedHospitalizationRepo(func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
		return &model.Hospitalization{ID: 10}, nil
	})
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			return []model.CarePlanItem{{Name: "点滴", UnitPrice: 2000}}, nil
		},
	}
	var released []uint64
	var created *model.Billing
	accountingRepo := &mockAccountingRepository{
		findByHospitalizationIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return dischargeOccupant(5, model.BillingStatusCancelled), nil
		},
		softDeleteCancelledFn: func(_ context.Context, clinicID, id uint64) error {
			assert.Equal(t, uint64(1), clinicID)
			released = append(released, id)
			return nil
		},
		createFn: func(_ context.Context, _ uint64, b *model.Billing) error {
			b.ID = 57
			created = b
			return nil
		},
	}
	billingItemRepo := &mockBillingItemRepository{
		createFn: func(_ context.Context, _ *model.BillingItem) error { return nil },
		updateBillingTotals: func(_ context.Context, _, _ uint64, _, _, _ int64) error {
			return nil
		},
	}
	deps := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, billingItemRepo)
	svc := NewHospitalizationServiceWithAudit(
		hospRepo, deps.reservation, nil, nil, carePlanRepo, accountingRepo, billingItemRepo, &mockTransactor{}, audit,
	)

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
		DischargeDate:    time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
		CreateAccounting: true,
		ActorID:          &actorID,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, []uint64{5}, released, "cancelled occupant must be soft-deleted to free the slot")
	require.NotNil(t, created)
	assert.Equal(t, model.BillingStatusWaiting, created.Status)
	require.NotNil(t, result.AccountingID)
	assert.Equal(t, uint64(57), *result.AccountingID)

	require.Len(t, audit.entries, 1)
	newValue, ok := audit.entries[0].NewValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, uint64(57), newValue["billing_id"])
	_, hasReused := newValue["reused"]
	assert.False(t, hasReused, "fresh-create audit keeps the existing payload shape")
}

func TestHospitalizationService_DischargeWithBilling_ReusesCompletedBilling(t *testing.T) {
	actorID := uint64(42)
	hospRepo := admittedHospitalizationRepo(func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
		return &model.Hospitalization{ID: 10}, nil
	})
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			t.Fatal("reused billing must not re-fetch care plan items")
			return nil, nil
		},
	}
	accountingRepo := &mockAccountingRepository{
		findByHospitalizationIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return dischargeOccupant(60, model.BillingStatusCompleted), nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("completed billing must not be duplicated")
			return nil
		},
	}
	billingItemRepo := &mockBillingItemRepository{}
	deps := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, billingItemRepo)
	svc := NewHospitalizationServiceWithAudit(
		hospRepo, deps.reservation, nil, nil, carePlanRepo, accountingRepo, billingItemRepo, &mockTransactor{}, &hospitalizationAuditRecorder{},
	)

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
		DischargeDate:    time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
		CreateAccounting: true,
		ActorID:          &actorID,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.AccountingID)
	assert.Equal(t, uint64(60), *result.AccountingID)
}
