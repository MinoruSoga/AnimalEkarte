package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

var errHospitalizationAuditWrite = errors.New("hospitalization audit write failed")

type hospitalizationAuditRecorder struct {
	entries []*AuditEntry
	err     error
}

func (r *hospitalizationAuditRecorder) LogEntryTx(_ context.Context, entry *AuditEntry) error {
	r.entries = append(r.entries, entry)
	return r.err
}

// NOTE (EMR-499 U1): createDischargeBilling is unreachable: create_accounting=true is refused with
// stay_discharge_billing_unavailable (422) before any transaction starts. The tests that exercised billing
// creation, reuse and its audit payload were replaced by refusal tests in this file and in
// hospitalization_discharge_guard_test.go; EMR-505/506/508 (unit 4) re-add them against the rebuilt billing seam.
// updateFn is retained only so call sites stay unchanged; the movement side no longer updates a row.
func admittedHospitalizationRepo(updateFn func(context.Context, uint64, uint64, UpdateHospitalizationInput) (*model.Hospitalization, error)) *mockHospitalizationRepository {
	return &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{
				ID:       id,
				ClinicID: clinicID,
				OwnerID:  2,
				PetID:    5,
			}, nil
		},
	}
}

// create_accounting=true is refused with stay_discharge_billing_unavailable (422) before the transaction starts
// (hospitalization_discharge.go), whatever the actor. Nothing is locked, read, billed or audited.
func TestHospitalizationService_DischargeWithBilling_CreateAccountingRefusedBeforeAnyWork(t *testing.T) {
	for _, tc := range []struct {
		name    string
		actorID *uint64
	}{
		{name: "actor present", actorID: uint64PtrHosp(42)},
		{name: "actor missing", actorID: nil},
		{name: "actor zero", actorID: uint64PtrHosp(0)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			txCalls := 0
			hospRepo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					t.Fatal("hospitalization must not be read or locked when billing is unavailable")
					return nil, nil
				},
			}
			carePlanRepo := &mockCarePlanItemRepository{
				listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
					t.Fatal("care plan items must not be fetched when billing is unavailable")
					return nil, nil
				},
			}
			accountingRepo := &mockAccountingRepository{
				createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
					t.Fatal("billing must not be created when billing is unavailable")
					return nil
				},
			}
			billingItemRepo := &mockBillingItemRepository{
				createFn: func(_ context.Context, _ *model.BillingItem) error {
					t.Fatal("billing items must not be created when billing is unavailable")
					return nil
				},
				updateBillingTotals: func(_ context.Context, _, _ uint64, _, _, _ int64) error {
					t.Fatal("billing totals must not be written when billing is unavailable")
					return nil
				},
			}
			audit := &hospitalizationAuditRecorder{}
			transactor := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				txCalls++
				return fn(ctx)
			}}
			deps := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, billingItemRepo)
			svc := NewHospitalizationServiceWithAudit(
				hospRepo, deps.reservation, nil, nil, carePlanRepo, accountingRepo, billingItemRepo, transactor, audit,
				WithPetLifeStateReader(stubPetLifeReader{}),
			)

			result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
				DischargeDate:    time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
				CreateAccounting: true,
				ActorID:          tc.actorID,
			})

			require.Nil(t, result)
			requireCoded422(t, err, codeStayDischargeBillingUnavailable)
			assert.Zero(t, txCalls, "no transaction is started")
			assert.Empty(t, audit.entries, "no audit entry is written")
		})
	}
}

func TestHospitalizationService_DischargeWithBilling_MissingAuditDependencyFailsBeforeStatusWrite(t *testing.T) {
	actorID := uint64(42)
	statusWritten := false
	hospRepo := admittedHospitalizationRepo(func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
		statusWritten = true
		return &model.Hospitalization{ID: 10}, nil
	})
	deps := newDischargeTestDeps(hospRepo, nil, nil, nil)
	svc := NewHospitalizationService(hospRepo, deps.reservation, nil, nil, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
		DischargeDate:    time.Now(),
		CreateAccounting: true,
		ActorID:          &actorID,
	})

	// create_accounting=true is refused before the transaction starts, so nothing is written.
	requireCoded422(t, err, codeStayDischargeBillingUnavailable)
	assert.Nil(t, result)
	assert.False(t, statusWritten)
}

func TestHospitalizationService_DischargeWithBilling_MissingOrZeroActorFailsBeforeStatusWrite(t *testing.T) {
	for _, tc := range []struct {
		name    string
		actorID *uint64
	}{
		{name: "missing", actorID: nil},
		{name: "zero", actorID: uint64PtrHosp(0)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			statusWritten := false
			hospRepo := admittedHospitalizationRepo(func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
				statusWritten = true
				return &model.Hospitalization{ID: 10}, nil
			})
			deps := newDischargeTestDeps(hospRepo, nil, nil, nil)
			svc := NewHospitalizationServiceWithAudit(
				hospRepo, deps.reservation, nil, nil, nil, nil, nil, &mockTransactor{}, &hospitalizationAuditRecorder{},
				WithPetLifeStateReader(stubPetLifeReader{}),
			)

			result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
				DischargeDate:    time.Now(),
				CreateAccounting: true,
				ActorID:          tc.actorID,
			})

			// create_accounting=true is refused before the actor is validated, so the code is the same for every actor.
			requireCoded422(t, err, codeStayDischargeBillingUnavailable)
			assert.Nil(t, result)
			assert.False(t, statusWritten)
		})
	}
}

func TestHospitalizationService_DischargeWithBilling_WithoutAccountingDoesNotAudit(t *testing.T) {
	audit := &hospitalizationAuditRecorder{}
	hospRepo := admittedHospitalizationRepo(func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
		return &model.Hospitalization{ID: 10}, nil
	})
	deps := newDischargeTestDeps(hospRepo, nil, nil, nil)
	svc := NewHospitalizationServiceWithAudit(
		hospRepo, deps.reservation, nil, nil, nil, nil, nil, &mockTransactor{}, audit,
		WithPetLifeStateReader(stubPetLifeReader{}),
	)

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
		DischargeDate:    time.Now(),
		CreateAccounting: false,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, audit.entries)
}

type AuditTxLoggerFunc func(context.Context, *AuditEntry) error

func (f AuditTxLoggerFunc) LogEntryTx(ctx context.Context, entry *AuditEntry) error {
	return f(ctx, entry)
}

type dischargePersistenceState struct {
	status         model.HospitalizationStatus
	billingCount   int
	billingItems   int
	subtotalAmount int64
	taxAmount      int64
	totalAmount    int64
}

type snapshotRollbackTransactor struct {
	state      *dischargePersistenceState
	rolledBack bool
}

func (t *snapshotRollbackTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	snapshot := *t.state
	err := fn(ctx)
	if err != nil {
		*t.state = snapshot
		t.rolledBack = true
	}
	return err
}

type statefulHospitalizationRepo struct {
	*mockHospitalizationRepository
	state *dischargePersistenceState
}

func (r *statefulHospitalizationRepo) FindByID(_ context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	return &model.Hospitalization{ID: id, ClinicID: clinicID, OwnerID: 2, PetID: 5}, nil
}

func (r *statefulHospitalizationRepo) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	return r.FindByID(ctx, clinicID, id)
}

func (r *statefulHospitalizationRepo) UpdateIfNotDischarged(_ context.Context, _, id uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
	r.state.status = model.HospitalizationStatusDischarged
	return &model.Hospitalization{ID: id}, nil
}

type statefulAccountingRepo struct {
	state *dischargePersistenceState
}

func (r *statefulAccountingRepo) Create(_ context.Context, _ uint64, billing *model.Billing) error {
	r.state.billingCount++
	billing.ID = 57
	return nil
}

// FindByHospitalizationID / SoftDeleteCancelled は EMR-253 の occupancy guard 用。
// 本 fixture は「占有者なし」を返す（rollback テストは新規作成経路のまま挙動不変）。
func (r *statefulAccountingRepo) FindByHospitalizationID(_ context.Context, _, _ uint64) (*model.Billing, error) {
	return nil, nil
}

func (r *statefulAccountingRepo) SoftDeleteCancelled(_ context.Context, _, _ uint64) error {
	return nil
}

type statefulBillingItemRepo struct {
	state *dischargePersistenceState
}

func (r *statefulBillingItemRepo) Create(_ context.Context, _ *model.BillingItem) error {
	r.state.billingItems++
	return nil
}

func (r *statefulBillingItemRepo) UpdateBillingTotals(_ context.Context, _, _ uint64, subtotal, taxAmount, totalAmount int64) error {
	r.state.subtotalAmount = subtotal
	r.state.taxAmount = taxAmount
	r.state.totalAmount = totalAmount
	return nil
}

// create_accounting=true is refused before the transaction starts (hospitalization_discharge.go), so the audit sink is
// never reached and no hospitalization, billing or billing item write happens; the transaction is never rolled back.
func TestHospitalizationService_DischargeWithBilling_CreateAccountingRefusedLeavesPersistenceUntouched(t *testing.T) {
	actorID := uint64(42)
	state := &dischargePersistenceState{status: model.HospitalizationStatusAdmitted}
	hospRepo := &statefulHospitalizationRepo{mockHospitalizationRepository: &mockHospitalizationRepository{}, state: state}
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			return []model.CarePlanItem{{Name: "点滴", UnitPrice: 2000}}, nil
		},
	}
	accountingRepo := &statefulAccountingRepo{state: state}
	billingItemRepo := &statefulBillingItemRepo{state: state}
	transactor := &snapshotRollbackTransactor{state: state}
	audit := &hospitalizationAuditRecorder{err: errHospitalizationAuditWrite}
	deps := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, billingItemRepo)
	svc := NewHospitalizationServiceWithAudit(
		hospRepo, deps.reservation, nil, nil, carePlanRepo, accountingRepo, billingItemRepo, transactor, audit,
		WithPetLifeStateReader(stubPetLifeReader{}),
	)

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
		DischargeDate:    time.Now(),
		CreateAccounting: true,
		ActorID:          &actorID,
	})

	requireCoded422(t, err, codeStayDischargeBillingUnavailable)
	assert.NotErrorIs(t, err, errHospitalizationAuditWrite, "the audit sink is never reached")
	assert.Nil(t, result)
	assert.Empty(t, audit.entries)
	assert.False(t, transactor.rolledBack, "no transaction is started, so nothing is rolled back")
	assert.Equal(t, model.HospitalizationStatusAdmitted, state.status)
	assert.Zero(t, state.billingCount)
	assert.Zero(t, state.billingItems)
	assert.Zero(t, state.subtotalAmount)
	assert.Zero(t, state.taxAmount)
	assert.Zero(t, state.totalAmount)
}
