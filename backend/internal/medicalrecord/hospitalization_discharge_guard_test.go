package medicalrecord

// hospitalization_discharge_guard_test.go — EMR-253: 退院時会計の no-duplicate ガード。
// EMR-499 U1: create_accounting=true は会計処理に入る前に stay_discharge_billing_unavailable (422) で拒否されるため、
// 既存会計の再利用・cancelled 解放・新規作成の各経路は到達不能になった。ここでは、拒否された退院が既存会計
// （waiting / cancelled / completed）を参照も解放も複製もしないことだけを固定する。
// ガード本体の挙動テストは EMR-505/506/508 が再設計後の会計 seam に対して復元する。

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

// A refused create_accounting=true leaves an existing billing alone: nothing is looked up, released
// (cancelled), reused (waiting / completed) or created, and no audit entry is written.
func TestHospitalizationService_DischargeWithBilling_RefusedWithoutTouchingExistingBilling(t *testing.T) {
	actorID := uint64(42)
	for _, tc := range []struct {
		name   string
		status model.BillingStatus
	}{
		{name: "waiting billing is not reused", status: model.BillingStatusWaiting},
		{name: "cancelled billing is not released", status: model.BillingStatusCancelled},
		{name: "completed billing is not reused", status: model.BillingStatusCompleted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			audit := &hospitalizationAuditRecorder{}
			hospRepo := admittedHospitalizationRepo(nil)
			lookups, releases, creates := 0, 0, 0
			accountingRepo := &mockAccountingRepository{
				findByHospitalizationIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
					lookups++
					return dischargeOccupant(5, tc.status), nil
				},
				softDeleteCancelledFn: func(_ context.Context, _, _ uint64) error {
					releases++
					return nil
				},
				createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
					creates++
					return nil
				},
			}
			billingItemRepo := &mockBillingItemRepository{}
			deps := newDischargeTestDeps(hospRepo, nil, accountingRepo, billingItemRepo)
			svc := NewHospitalizationServiceWithAudit(
				hospRepo, deps.reservation, nil, nil, nil, accountingRepo, billingItemRepo, &mockTransactor{}, audit,
				WithPetLifeStateReader(stubPetLifeReader{}),
			)

			result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
				DischargeDate:    time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
				CreateAccounting: true,
				ActorID:          &actorID,
			})

			require.Nil(t, result)
			requireCoded422(t, err, codeStayDischargeBillingUnavailable)
			assert.Zero(t, lookups, "an existing billing is not looked up")
			assert.Zero(t, releases, "a cancelled billing is not released")
			assert.Zero(t, creates, "no billing is created")
			assert.Empty(t, audit.entries, "no audit entry is written")
		})
	}
}
