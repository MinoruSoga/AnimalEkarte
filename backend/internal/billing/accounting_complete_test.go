package billing

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// ---- pure helpers ----

func TestValidateIdempotencyKeyUUID(t *testing.T) {
	t.Parallel()
	assert.Error(t, ValidateIdempotencyKeyUUID(""))
	assert.Error(t, ValidateIdempotencyKeyUUID("not-a-uuid"))
	assert.NoError(t, ValidateIdempotencyKeyUUID(uuid.NewString()))
}

func TestComputeCompleteAccountingDigest_Deterministic(t *testing.T) {
	t.Parallel()
	ownerID := uint64(1)
	petID := uint64(2)
	base := &CompleteAccountingInput{
		OwnerID:       &ownerID,
		PetID:         &petID,
		ScheduledDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Memo:          "memo",
		Items: []CompleteAccountingItemInput{
			{Name: "診察料", UnitPrice: 1000, Quantity: 1, Category: "other", Source: "manual", TaxType: "excluded", TaxRate: 0.1},
		},
		PaymentSplits: []PaymentSplitInput{
			{Method: model.PaymentMethodCash, Amount: 1100, ReceivedAmount: 2000, ChangeAmount: 900},
		},
	}
	d1, err := ComputeCompleteAccountingDigest(base)
	require.NoError(t, err)
	d2, err := ComputeCompleteAccountingDigest(base)
	require.NoError(t, err)
	assert.Equal(t, d1, d2)

	changed := *base
	changed.Memo = "other"
	d3, err := ComputeCompleteAccountingDigest(&changed)
	require.NoError(t, err)
	assert.NotEqual(t, d1, d3)
}

// ---- mock collaborator doubles ----

type mockCompleteItemWriter struct {
	calls    int
	failAt   int // 1-based; 0 = never fail
	failErr  error
	createFn func(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error)
}

func (m *mockCompleteItemWriter) CreateItemForComplete(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
	m.calls++
	if m.failAt > 0 && m.calls == m.failAt {
		if m.failErr != nil {
			return nil, m.failErr
		}
		return nil, errors.New("injected item failure")
	}
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &model.BillingItem{
		ID:        uint64(m.calls),
		BillingID: input.BillingID,
		Name:      input.Name,
		UnitPrice: input.UnitPrice,
		Quantity:  input.Quantity,
		TaxType:   model.TaxTypeExcluded,
		TaxRate:   0.1,
	}, nil
}

type mockCompleteTotalsWriter struct {
	subtotal, taxTotal, total int64
	err                       error
}

func (m *mockCompleteTotalsWriter) RecalculateTotalsForComplete(_ context.Context, _, _ uint64) (int64, int64, int64, error) {
	if m.err != nil {
		return 0, 0, 0, m.err
	}
	if m.total == 0 && m.subtotal == 0 {
		return 1000, 100, 1100, nil
	}
	return m.subtotal, m.taxTotal, m.total, nil
}

func matchingReservationRepo() *mockReservationRepository {
	return &mockReservationRepository{
		// Complete 入力の pet/owner 相関を通過させる（ownerID=10, petID=20 を既定）。
		findPetOwnerInClinicFn: func(_ context.Context, _, petID uint64) (uint64, error) {
			if petID == 20 {
				return 10, nil
			}
			return petID, nil
		},
	}
}

func newCompleteTestService(
	repo *mockAccountingRepository,
	audit *mockAuditService,
	itemWriter completeItemWriter,
	totals completeTotalsWriter,
	opts ...accountingServiceOption,
) AccountingService {
	base := []accountingServiceOption{
		WithCompleteItemWriter(itemWriter),
		WithCompleteTotalsWriter(totals),
	}
	base = append(base, opts...)
	return NewAccountingService(
		repo, nil, nil, matchingReservationRepo(), nil,
		&mockTransactor{}, audit, seededPayMethodMock(),
		base...,
	)
}

func validCompleteInput(key string) *CompleteAccountingInput {
	ownerID := uint64(10)
	petID := uint64(20)
	staffID := uint64(1)
	return &CompleteAccountingInput{
		ClinicID:       1,
		StaffID:        &staffID,
		IdempotencyKey: key,
		OwnerID:        &ownerID,
		PetID:          &petID,
		ScheduledDate:  time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Items: []CompleteAccountingItemInput{
			{Name: "診察料", UnitPrice: 1000, Quantity: 1, Category: "other", Source: "manual", TaxType: "excluded", TaxRate: 0.1},
			{Name: "薬", UnitPrice: 500, Quantity: 1, Category: "medicine", Source: "manual", TaxType: "excluded", TaxRate: 0.1},
		},
		PaymentSplits: []PaymentSplitInput{
			{Method: model.PaymentMethodCash, Amount: 1100, ReceivedAmount: 2000, ChangeAmount: 900},
		},
	}
}

func TestAccountingService_CompleteAccounting_OpenPeriodAtomicSuccess(t *testing.T) {
	key := uuid.NewString()
	var created *model.Billing
	var savedPayment *model.Payment
	var savedSplits []model.PaymentSplit
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		createFn: func(_ context.Context, clinicID uint64, b *model.Billing) error {
			b.ID = 42
			b.ClinicID = clinicID
			created = b
			return nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, fields map[string]any) (*model.Billing, error) {
			b := &model.Billing{ID: id, ClinicID: 1, Status: model.BillingStatusCompleted, TotalAmount: 1100, Subtotal: 1000, TaxTotal: 100}
			if v, ok := fields["total_amount"].(int64); ok {
				b.TotalAmount = v
			}
			return b, nil
		},
		savePaymentFn: func(_ context.Context, p *model.Payment) error {
			savedPayment = p
			return nil
		},
		savePaymentSplitsFn: func(_ context.Context, splits []model.PaymentSplit) error {
			savedSplits = splits
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Billing, error) {
			return &model.Billing{
				ID: id, ClinicID: 1, Status: model.BillingStatusCompleted,
				Subtotal: 1000, TaxTotal: 100, TotalAmount: 1100,
				CompletionRequestID:   &key,
				CompletionRequestHash: strPtr(mustDigest(validCompleteInput(key))),
			}, nil
		},
	}
	items := &mockCompleteItemWriter{}
	totals := &mockCompleteTotalsWriter{subtotal: 1000, taxTotal: 100, total: 1100}
	svc := newCompleteTestService(repo, &mockAuditService{}, items, totals)

	// Client total intentionally differs from server: server uses totalsWriter values.
	result, err := svc.Complete(context.Background(), validCompleteInput(key))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Created)
	assert.Equal(t, uint64(42), result.Accounting.ID)
	assert.Equal(t, 2, items.calls, "both items created in same command")
	require.NotNil(t, created)
	require.NotNil(t, created.CompletionRequestID)
	assert.Equal(t, key, *created.CompletionRequestID)
	require.NotNil(t, savedPayment)
	assert.Equal(t, int64(1100), savedPayment.TotalAmount, "server total, not client")
	assert.Len(t, savedSplits, 1)
}

func TestAccountingService_CompleteAccounting_PostCloseReasonMissing_NoWrites(t *testing.T) {
	key := uuid.NewString()
	createCalled := false
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			createCalled = true
			return nil
		},
	}
	// closeRepo re-evaluates as closed.
	closeRepo := &mockCashRegisterCloseRepository{
		hasCloseOnDateFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) {
			return true, nil
		},
	}
	items := &mockCompleteItemWriter{}
	svc := newCompleteTestService(repo, &mockAuditService{}, items, &mockCompleteTotalsWriter{},
		WithCashRegisterCloseRepository(closeRepo),
	)
	input := validCompleteInput(key)
	input.IsPostClose = false // handler may have missed; write-time re-eval catches it
	input.PostCloseReason = nil

	result, err := svc.Complete(context.Background(), input)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "got %v", err)
	assert.Nil(t, result)
	assert.False(t, createCalled, "header must not be created when post_close_reason missing")
	assert.Equal(t, 0, items.calls)
}

func TestAccountingService_CompleteAccounting_NthItemFailure_FullRollback(t *testing.T) {
	key := uuid.NewString()
	createCalled := false
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		createFn: func(_ context.Context, clinicID uint64, b *model.Billing) error {
			createCalled = true
			b.ID = 7
			b.ClinicID = clinicID
			return nil
		},
		savePaymentFn: func(_ context.Context, _ *model.Payment) error {
			t.Fatal("payment must not be saved after item failure")
			return nil
		},
	}
	// mockTransactor runs fn without real tx; simulate rollback by asserting no payment and Created=false.
	// Production uses real WithTx; here we verify error propagation and no post-item side effects.
	items := &mockCompleteItemWriter{failAt: 2, failErr: errors.New("N-th item reference invalid")}
	svc := newCompleteTestService(repo, &mockAuditService{}, items, &mockCompleteTotalsWriter{})

	result, err := svc.Complete(context.Background(), validCompleteInput(key))
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, createCalled, "header attempted inside tx before item failure")
	assert.Equal(t, 2, items.calls)
}

func TestAccountingService_CompleteAccounting_AuditFailure_FullRollback(t *testing.T) {
	key := uuid.NewString()
	paymentSaved := false
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		createFn: func(_ context.Context, clinicID uint64, b *model.Billing) error {
			b.ID = 9
			b.ClinicID = clinicID
			return nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.Billing, error) {
			return &model.Billing{ID: id, ClinicID: 1, Status: model.BillingStatusCompleted, TotalAmount: 1100}, nil
		},
		savePaymentFn: func(_ context.Context, _ *model.Payment) error {
			paymentSaved = true
			return nil
		},
		savePaymentSplitsFn: func(_ context.Context, _ []model.PaymentSplit) error { return nil },
	}
	audit := &mockAuditService{logEntryTxErr: errors.New("audit write failed")}
	closeRepo := &mockCashRegisterCloseRepository{
		hasCloseOnDateFn: func(_ context.Context, _ uint64, _ time.Time) (bool, error) {
			return true, nil
		},
		findByDateAndPeriodFn: func(_ context.Context, clinicID uint64, _ time.Time, period string) (*model.CashRegisterClose, error) {
			if period == "pm" {
				return &model.CashRegisterClose{ID: 1, ClinicID: clinicID}, nil
			}
			return nil, nil
		},
		createAdjustmentFn: func(_ context.Context, _ *model.CashRegisterCloseAdjustment) error {
			return nil
		},
	}
	reason := "締め後の追加会計"
	input := validCompleteInput(key)
	input.IsPostClose = true
	input.PostCloseReason = &reason

	svc := newCompleteTestService(repo, audit, &mockCompleteItemWriter{}, &mockCompleteTotalsWriter{subtotal: 1000, taxTotal: 100, total: 1100},
		WithCashRegisterCloseRepository(closeRepo),
	)
	result, err := svc.Complete(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, paymentSaved, "payment ran before audit; real WithTx would roll it back")
	assert.True(t, audit.logEntryTxCalled)
}

func TestAccountingService_CompleteAccounting_IdempotentReplaySameDigest(t *testing.T) {
	key := uuid.NewString()
	input := validCompleteInput(key)
	digest := mustDigest(input)
	existing := &model.Billing{
		ID: 55, ClinicID: 1, Status: model.BillingStatusCompleted,
		Subtotal: 1000, TaxTotal: 100, TotalAmount: 1100,
		CompletionRequestID:   &key,
		CompletionRequestHash: &digest,
	}
	createCalled := false
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, requestID string) (*model.Billing, error) {
			if requestID == key {
				return existing, nil
			}
			return nil, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			createCalled = true
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Billing, error) {
			assert.Equal(t, uint64(55), id)
			return existing, nil
		},
	}
	items := &mockCompleteItemWriter{}
	svc := newCompleteTestService(repo, &mockAuditService{}, items, &mockCompleteTotalsWriter{})

	result, err := svc.Complete(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Created, "same digest replay must be 200 path")
	assert.Equal(t, uint64(55), result.Accounting.ID)
	assert.False(t, createCalled)
	assert.Equal(t, 0, items.calls, "replay must not re-create items/payments")
}

func TestAccountingService_CompleteAccounting_IdempotentConflictDifferentDigest(t *testing.T) {
	key := uuid.NewString()
	storedHash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	existing := &model.Billing{
		ID: 56, ClinicID: 1, Status: model.BillingStatusCompleted,
		CompletionRequestID:   &key,
		CompletionRequestHash: &storedHash,
	}
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return existing, nil
		},
	}
	svc := newCompleteTestService(repo, &mockAuditService{}, &mockCompleteItemWriter{}, &mockCompleteTotalsWriter{})
	result, err := svc.Complete(context.Background(), validCompleteInput(key))
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "got %v", err)
	assert.Nil(t, result)
}

func TestAccountingService_CompleteAccounting_InvalidKey(t *testing.T) {
	svc := newCompleteTestService(&mockAccountingRepository{}, &mockAuditService{}, &mockCompleteItemWriter{}, &mockCompleteTotalsWriter{})
	input := validCompleteInput("bad-key")
	result, err := svc.Complete(context.Background(), input)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, result)
}

func TestAccountingService_CompleteAccounting_BlockingUnbilled(t *testing.T) {
	key := uuid.NewString()
	createCalled := false
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			createCalled = true
			return nil
		},
	}
	guard := &mockBillingItemService{
		assertNoBlockingUnbilledFn: func(_ context.Context, _, _ uint64) error {
			return apperrors.WrapConflict("未請求候補に請求不能な予防接種が含まれるため会計を確定できません")
		},
	}
	svc := newCompleteTestService(repo, &mockAuditService{}, &mockCompleteItemWriter{}, &mockCompleteTotalsWriter{},
		WithUnbilledWriteGuard(guard),
	)
	result, err := svc.Complete(context.Background(), validCompleteInput(key))
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Nil(t, result)
	assert.False(t, createCalled)
}

// TestAccountingService_CompleteAccounting_DBAtomicRollback は実 DB で N 番目 item 失敗時に
// billings / billing_items / payments がすべて不変であることを検証する。
func TestAccountingService_CompleteAccounting_DBAtomicRollback(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedBillingClinicForFK(t, db, clinicID)

	// Payment method masters for cash
	cashKey := "cash"
	pm := &model.PaymentMethodMaster{ClinicID: clinicID, Name: "現金", SystemKey: &cashKey, IsActive: true}
	require.NoError(t, db.Create(pm).Error)

	owner := &model.Owner{ClinicID: clinicID, Name: "テスト飼主"}
	require.NoError(t, db.Create(owner).Error)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}))
	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.Create(species).Error)
	pet := &model.Pet{ClinicID: clinicID, OwnerID: owner.ID, AnimalSpeciesID: species.ID, Name: "テストペット"}
	require.NoError(t, db.Create(pet).Error)

	// Wire real repos + failing item writer after first success
	accRepo := NewAccountingRepository(db)
	itemRepo := NewBillingItemRepository(db)
	tx := testNewTransactor(db)
	itemSvc := NewBillingItemServiceWithCampaign(
		itemRepo, accRepo, nil, tx, nil, nil, nil, nil,
	)

	// Wrap itemSvc to fail on 2nd item
	failWriter := &countingFailItemWriter{inner: itemSvc, failAt: 2}

	payMethodRepo := NewPaymentMethodMasterRepository(db)
	// DB 経路でも owner/pet 相関検証を mock で通す（reservation 具象は別 domain）
	resRepo := &mockReservationRepository{
		findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
			return owner.ID, nil
		},
	}
	svc := NewAccountingService(
		accRepo, nil, nil, resRepo, nil,
		tx, &mockAuditService{}, payMethodRepo,
		WithCompleteItemWriter(failWriter),
		WithCompleteTotalsWriter(itemSvc),
	)

	key := uuid.NewString()
	ownerID, petID := owner.ID, pet.ID
	input := &CompleteAccountingInput{
		ClinicID:       clinicID,
		IdempotencyKey: key,
		OwnerID:        &ownerID,
		PetID:          &petID,
		ScheduledDate:  time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
		Items: []CompleteAccountingItemInput{
			{Name: "item1", UnitPrice: 1000, Quantity: 1, Category: "examination", Source: "manual", TaxType: "excluded", TaxRate: 0.1},
			{Name: "item2", UnitPrice: 500, Quantity: 1, Category: "examination", Source: "manual", TaxType: "excluded", TaxRate: 0.1},
		},
		PaymentSplits: []PaymentSplitInput{
			{Method: model.PaymentMethodCash, Amount: 1650, ReceivedAmount: 2000, ChangeAmount: 350},
		},
	}

	var billingsBefore, itemsBefore, paymentsBefore int64
	db.Model(&model.Billing{}).Count(&billingsBefore)
	db.Model(&model.BillingItem{}).Count(&itemsBefore)
	db.Model(&model.Payment{}).Count(&paymentsBefore)

	result, err := svc.Complete(ctx, input)
	require.Error(t, err)
	assert.Nil(t, result)

	var billingsAfter, itemsAfter, paymentsAfter int64
	db.Model(&model.Billing{}).Count(&billingsAfter)
	db.Model(&model.BillingItem{}).Count(&itemsAfter)
	db.Model(&model.Payment{}).Count(&paymentsAfter)
	assert.Equal(t, billingsBefore, billingsAfter, "header must roll back")
	assert.Equal(t, itemsBefore, itemsAfter, "items must roll back")
	assert.Equal(t, paymentsBefore, paymentsAfter, "payments must roll back")
}

// TestAccountingService_CompleteAccounting_DBSuccess_ServerTotals は server 再計算 total で commit されることを検証する。
func TestAccountingService_CompleteAccounting_DBSuccess_ServerTotals(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedBillingClinicForFK(t, db, clinicID)

	cashKey := "cash"
	require.NoError(t, db.Create(&model.PaymentMethodMaster{ClinicID: clinicID, Name: "現金", SystemKey: &cashKey, IsActive: true}).Error)
	owner := &model.Owner{ClinicID: clinicID, Name: "飼主S"}
	require.NoError(t, db.Create(owner).Error)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}))
	species := &model.AnimalSpecies{Name: "猫"}
	require.NoError(t, db.Create(species).Error)
	pet := &model.Pet{ClinicID: clinicID, OwnerID: owner.ID, AnimalSpeciesID: species.ID, Name: "ペットS"}
	require.NoError(t, db.Create(pet).Error)

	accRepo := NewAccountingRepository(db)
	itemRepo := NewBillingItemRepository(db)
	tx := testNewTransactor(db)
	itemSvc := NewBillingItemServiceWithCampaign(itemRepo, accRepo, nil, tx, nil, nil, nil, nil)
	payMethodRepo := NewPaymentMethodMasterRepository(db)
	resRepo := &mockReservationRepository{
		findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
			return owner.ID, nil
		},
	}
	svc := NewAccountingService(
		accRepo, nil, nil, resRepo, nil,
		tx, &mockAuditService{}, payMethodRepo,
		WithCompleteItemWriter(itemSvc),
		WithCompleteTotalsWriter(itemSvc),
	)

	key := uuid.NewString()
	ownerID, petID := owner.ID, pet.ID
	// Payment amount must match server total: 1000*1 + 100 tax = 1100
	input := &CompleteAccountingInput{
		ClinicID:       clinicID,
		IdempotencyKey: key,
		OwnerID:        &ownerID,
		PetID:          &petID,
		ScheduledDate:  time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC),
		Items: []CompleteAccountingItemInput{
			{Name: "診察", UnitPrice: 1000, Quantity: 1, Category: "examination", Source: "manual", TaxType: "excluded", TaxRate: 0.1},
		},
		PaymentSplits: []PaymentSplitInput{
			{Method: model.PaymentMethodCash, Amount: 1100, ReceivedAmount: 2000, ChangeAmount: 900},
		},
	}

	result, err := svc.Complete(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Created)
	assert.Equal(t, model.BillingStatusCompleted, result.Accounting.Status)
	assert.EqualValues(t, 1000, result.Accounting.Subtotal)
	assert.EqualValues(t, 100, result.Accounting.TaxTotal)
	assert.EqualValues(t, 1100, result.Accounting.TotalAmount)

	var itemCount int64
	db.Model(&model.BillingItem{}).Where("billing_id = ?", result.Accounting.ID).Count(&itemCount)
	assert.EqualValues(t, 1, itemCount)

	// Same digest replay -> not created again
	result2, err := svc.Complete(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, result2)
	assert.False(t, result2.Created)
	assert.Equal(t, result.Accounting.ID, result2.Accounting.ID)

	var paymentCount int64
	db.Model(&model.Payment{}).Where("billing_id = ?", result.Accounting.ID).Count(&paymentCount)
	assert.EqualValues(t, 1, paymentCount, "replay must not duplicate payments")
}

type countingFailItemWriter struct {
	inner  completeItemWriter
	calls  int
	failAt int
}

func (w *countingFailItemWriter) CreateItemForComplete(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
	w.calls++
	if w.failAt > 0 && w.calls == w.failAt {
		return nil, fmt.Errorf("injected failure at item %d", w.calls)
	}
	return w.inner.CreateItemForComplete(ctx, input)
}

func mustDigest(input *CompleteAccountingInput) string {
	d, err := ComputeCompleteAccountingDigest(input)
	if err != nil {
		panic(err)
	}
	return d
}

func strPtr(s string) *string { return &s }
