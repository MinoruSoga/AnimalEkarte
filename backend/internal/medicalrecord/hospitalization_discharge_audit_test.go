package medicalrecord

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

var errHospitalizationAuditWrite = errors.New("hospitalization audit write failed")

type hospitalizationAuditRecorder struct {
	entries []*AuditEntry
	err     error
}

func (r *hospitalizationAuditRecorder) LogEntryTx(_ context.Context, entry *AuditEntry) error {
	r.entries = append(r.entries, entry)
	return r.err
}

func admittedHospitalizationRepo(updateFn func(context.Context, uint64, uint64, map[string]any) (*model.Hospitalization, error)) *mockHospitalizationRepository {
	return &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{
				ID:       id,
				ClinicID: clinicID,
				OwnerID:  2,
				PetID:    5,
				Status:   model.HospitalizationStatusAdmitted,
			}, nil
		},
		updateIfNotDischargedFn: updateFn,
	}
}

func TestHospitalizationService_DischargeWithBilling_AuditsBillingTotals(t *testing.T) {
	actorID := uint64(42)
	order := make([]string, 0, 5)
	audit := &hospitalizationAuditRecorder{}
	hospRepo := admittedHospitalizationRepo(func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Hospitalization, error) {
		order = append(order, "status")
		return &model.Hospitalization{ID: 10}, nil
	})
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			return []model.CarePlanItem{
				{Name: "食事介助", UnitPrice: 1000},
				{Name: "点滴", UnitPrice: 2000},
			}, nil
		},
	}
	accountingRepo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, billing *model.Billing) error {
			order = append(order, "billing")
			billing.ID = 55
			return nil
		},
	}
	billingItemRepo := &mockBillingItemRepository{
		createFn: func(_ context.Context, _ *model.BillingItem) error {
			order = append(order, "item")
			return nil
		},
		updateBillingTotals: func(_ context.Context, _, _ uint64, _, _, _ int64) error {
			order = append(order, "totals")
			return nil
		},
	}
	auditWithOrder := AuditTxLoggerFunc(func(ctx context.Context, entry *AuditEntry) error {
		order = append(order, "audit")
		return audit.LogEntryTx(ctx, entry)
	})
	svc := NewHospitalizationServiceWithAudit(
		hospRepo, newDischargeTestDeps(hospRepo, nil, nil, nil).reservation, nil, nil,
		carePlanRepo, accountingRepo, billingItemRepo, &mockTransactor{}, auditWithOrder,
	)

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
		DischargeDate:    time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
		CreateAccounting: true,
		ActorID:          &actorID,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, []string{"status", "billing", "item", "item", "totals", "audit"}, order)
	require.Len(t, audit.entries, 1)
	entry := audit.entries[0]
	require.NotNil(t, entry.ClinicID)
	require.NotNil(t, entry.ActorID)
	require.NotNil(t, entry.ResourceID)
	assert.Equal(t, uint64(1), *entry.ClinicID)
	assert.Equal(t, actorID, *entry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, entry.ActorType)
	assert.Equal(t, model.AuditActionHospitalizationDischargeWithBilling, entry.Action)
	assert.Equal(t, model.AuditResourceHospitalization, entry.Resource)
	assert.Equal(t, uint64(10), *entry.ResourceID)
	assert.Nil(t, entry.OldValue)
	assert.Equal(t, map[string]any{
		"billing_id":      uint64(55),
		"subtotal_amount": int64(3000),
		"tax_amount":      int64(300),
		"total_amount":    int64(3300),
	}, entry.NewValue)
}

func TestHospitalizationService_DischargeWithBilling_MissingAuditDependencyFailsBeforeStatusWrite(t *testing.T) {
	actorID := uint64(42)
	statusWritten := false
	hospRepo := admittedHospitalizationRepo(func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Hospitalization, error) {
		statusWritten = true
		return &model.Hospitalization{ID: 10}, nil
	})
	deps := newDischargeTestDeps(hospRepo, nil, nil, nil)
	svc := NewHospitalizationService(hospRepo, deps.reservation, nil, nil, nil, nil, nil, &mockTransactor{})

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
		DischargeDate:    time.Now(),
		CreateAccounting: true,
		ActorID:          &actorID,
	})

	assert.Error(t, err)
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
			hospRepo := admittedHospitalizationRepo(func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Hospitalization, error) {
				statusWritten = true
				return &model.Hospitalization{ID: 10}, nil
			})
			deps := newDischargeTestDeps(hospRepo, nil, nil, nil)
			svc := NewHospitalizationServiceWithAudit(
				hospRepo, deps.reservation, nil, nil, nil, nil, nil, &mockTransactor{}, &hospitalizationAuditRecorder{},
			)

			result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
				DischargeDate:    time.Now(),
				CreateAccounting: true,
				ActorID:          tc.actorID,
			})

			assert.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
			assert.Nil(t, result)
			assert.False(t, statusWritten)
		})
	}
}

func TestHospitalizationService_DischargeWithBilling_WithoutAccountingDoesNotAudit(t *testing.T) {
	audit := &hospitalizationAuditRecorder{}
	hospRepo := admittedHospitalizationRepo(func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Hospitalization, error) {
		return &model.Hospitalization{ID: 10}, nil
	})
	deps := newDischargeTestDeps(hospRepo, nil, nil, nil)
	svc := NewHospitalizationServiceWithAudit(
		hospRepo, deps.reservation, nil, nil, nil, nil, nil, &mockTransactor{}, audit,
	)

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
		DischargeDate:    time.Now(),
		CreateAccounting: false,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, audit.entries)
}

func TestHospitalizationService_DischargeWithBilling_ZeroCarePlanCreatesZeroBillingAudit(t *testing.T) {
	actorID := uint64(42)
	audit := &hospitalizationAuditRecorder{}
	hospRepo := admittedHospitalizationRepo(func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Hospitalization, error) {
		return &model.Hospitalization{ID: 10}, nil
	})
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			return nil, nil
		},
	}
	billingCreated := false
	accountingRepo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, billing *model.Billing) error {
			billingCreated = true
			billing.ID = 56
			return nil
		},
	}
	deps := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, &mockBillingItemRepository{})
	svc := NewHospitalizationServiceWithAudit(
		hospRepo, deps.reservation, nil, nil, carePlanRepo, accountingRepo, deps.billingItem, &mockTransactor{}, audit,
	)

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
		DischargeDate:    time.Now(),
		CreateAccounting: true,
		ActorID:          &actorID,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, billingCreated)
	require.NotNil(t, result.AccountingID)
	assert.Equal(t, uint64(56), *result.AccountingID)
	require.Len(t, audit.entries, 1)
	assert.Equal(t, map[string]any{
		"billing_id":      uint64(56),
		"subtotal_amount": int64(0),
		"tax_amount":      int64(0),
		"total_amount":    int64(0),
	}, audit.entries[0].NewValue)
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
	return &model.Hospitalization{ID: id, ClinicID: clinicID, OwnerID: 2, PetID: 5, Status: r.state.status}, nil
}

func (r *statefulHospitalizationRepo) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	return r.FindByID(ctx, clinicID, id)
}

func (r *statefulHospitalizationRepo) UpdateIfNotDischarged(_ context.Context, _, id uint64, _ map[string]any) (*model.Hospitalization, error) {
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

func TestHospitalizationService_DischargeWithBilling_AuditFailureRollsBackAllWrites(t *testing.T) {
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
	)

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
		DischargeDate:    time.Now(),
		CreateAccounting: true,
		ActorID:          &actorID,
	})

	assert.ErrorIs(t, err, errHospitalizationAuditWrite)
	assert.Nil(t, result)
	assert.True(t, transactor.rolledBack)
	assert.Equal(t, model.HospitalizationStatusAdmitted, state.status)
	assert.Zero(t, state.billingCount)
	assert.Zero(t, state.billingItems)
	assert.Zero(t, state.subtotalAmount)
	assert.Zero(t, state.taxAmount)
	assert.Zero(t, state.totalAmount)
}
