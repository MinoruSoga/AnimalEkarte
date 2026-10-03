package billing

// accounting_complete_takeover_test.go — EMR-253: discharge-created waiting billing の
// in-place settlement（takeover）テスト。
//   - billing_id 明示: waiting 行を FOR UPDATE で掴み、同一 ID で確定（新規 INSERT しない）
//   - hospitalization 占有者: waiting→暗黙 takeover / cancelled→解放後 INSERT / completed→409
//   - 明細は request items が権威のため DeleteItemsForComplete で replace する
//   - digest は billing_id を含むため、同一 key で別対象を指す retry は 409 になる

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// takeoverFixture は takeover 対象の既存 waiting billing（退院作成を想定: hosp=99/owner=10/pet=20）。
func takeoverWaitingBilling() *model.Billing {
	hospID := uint64(99)
	ownerID := uint64(10)
	petID := uint64(20)
	return &model.Billing{
		ID:                7,
		ClinicID:          1,
		HospitalizationID: &hospID,
		OwnerID:           &ownerID,
		PetID:             &petID,
		Status:            model.BillingStatusWaiting,
		ScheduledDate:     time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC),
		Subtotal:          2000,
		TaxTotal:          200,
		TotalAmount:       2200,
	}
}

// takeoverInput は billing_id 明示の complete input（waiting 行の親参照と一致させる）。
func takeoverInput(key string, billingID uint64) *CompleteAccountingInput {
	input := validCompleteInput(key)
	input.BillingID = &billingID
	hospID := uint64(99)
	input.HospitalizationID = &hospID
	return input
}

// takeoverHospRepo は hosp=99 (owner=10/pet=20) を返す fixture。
func takeoverHospRepo() *mockHospitalizationRepository {
	return &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: id, ClinicID: clinicID, OwnerID: 10, PetID: 20}, nil
		},
	}
}

func newTakeoverTestService(
	repo *mockAccountingRepository,
	hospRepo billingHospitalizationFinder,
	itemWriter completeItemWriter,
	totals completeTotalsWriter,
) AccountingService {
	return NewAccountingService(
		repo, nil, hospRepo, matchingReservationRepo(), nil,
		&mockTransactor{}, &mockAuditService{}, seededPayMethodMock(),
		WithCompleteItemWriter(itemWriter),
		WithCompleteTotalsWriter(totals),
	)
}

// takeoverReloadRow は確定後 reload が返す completed 行。
func takeoverReloadRow(id uint64) *model.Billing {
	hospID := uint64(99)
	ownerID := uint64(10)
	petID := uint64(20)
	return &model.Billing{
		ID: id, ClinicID: 1, Status: model.BillingStatusCompleted,
		HospitalizationID: &hospID, OwnerID: &ownerID, PetID: &petID,
		Subtotal: 1000, TaxTotal: 100, TotalAmount: 1100,
	}
}

func TestAccountingService_Complete_TakeoverWaitingBillingInPlace(t *testing.T) {
	key := uuid.NewString()
	bound := takeoverWaitingBilling()
	var headerUpdate, finalizeUpdate *AccountingUpdate
	var updateCalls int
	var savedPayment *model.Payment
	var savedSplits []model.PaymentSplit
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, id uint64) (*model.Billing, error) {
			require.Equal(t, uint64(7), id)
			return bound, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("takeover must not INSERT a new billing row")
			return nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, cmd AccountingUpdate) (*model.Billing, error) {
			require.Equal(t, uint64(7), id, "takeover must update the bound row in place")
			updateCalls++
			copied := cmd
			if headerUpdate == nil {
				headerUpdate = &copied
			} else {
				finalizeUpdate = &copied
			}
			return takeoverReloadRow(id), nil
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
			return takeoverReloadRow(id), nil
		},
	}
	items := &mockCompleteItemWriter{}
	totals := &mockCompleteTotalsWriter{}
	svc := newTakeoverTestService(repo, takeoverHospRepo(), items, totals)

	result, err := svc.Complete(context.Background(), takeoverInput(key, 7))

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Created, "first takeover is a fresh completion (201)")
	assert.Equal(t, uint64(7), result.Accounting.ID, "same row settled in place")
	assert.Equal(t, model.BillingStatusCompleted, result.Accounting.Status)

	// 明細 replace: 既存明細を一括削除してから request items を作成する。
	assert.Equal(t, 1, items.deleteCalls, "existing waiting items must be bulk-deleted")
	assert.Equal(t, 2, items.calls, "request items must be re-created on the bound row")

	// header takeover update + finalize update の 2 回。
	require.Equal(t, 2, updateCalls)
	require.NotNil(t, headerUpdate)
	require.NotNil(t, headerUpdate.CompletionRequestID, "takeover must stamp the idempotency key on the existing row")
	assert.Equal(t, key, *headerUpdate.CompletionRequestID)
	require.NotNil(t, headerUpdate.CompletionRequestHash)
	assert.Equal(t, mustDigest(takeoverInput(key, 7)), *headerUpdate.CompletionRequestHash)
	require.NotNil(t, headerUpdate.ScheduledDate)
	assert.True(t, headerUpdate.ScheduledDate.Equal(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)))
	assert.Nil(t, headerUpdate.Status, "header phase must not finalize status")
	require.NotNil(t, finalizeUpdate)
	require.NotNil(t, finalizeUpdate.Status)
	assert.Equal(t, model.BillingStatusCompleted, *finalizeUpdate.Status)
	require.NotNil(t, finalizeUpdate.CompletedAt)
	require.NotNil(t, finalizeUpdate.TotalAmount)
	assert.Equal(t, int64(1100), *finalizeUpdate.TotalAmount)

	require.NotNil(t, savedPayment)
	assert.Equal(t, uint64(7), savedPayment.BillingID, "payment must attach to the taken-over row")
	require.Len(t, savedSplits, 1)
	assert.Equal(t, uint64(7), savedSplits[0].BillingID)
}

func TestAccountingService_Complete_TakeoverRejectsCompletedTarget(t *testing.T) {
	key := uuid.NewString()
	completed := takeoverWaitingBilling()
	completed.Status = model.BillingStatusCompleted
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return completed, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("completed target must not create a new row")
			return nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ AccountingUpdate) (*model.Billing, error) {
			t.Fatal("completed target must not be updated")
			return nil, nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return completed, nil
		},
	}
	items := &mockCompleteItemWriter{}
	svc := newTakeoverTestService(repo, takeoverHospRepo(), items, &mockCompleteTotalsWriter{})

	result, err := svc.Complete(context.Background(), takeoverInput(key, 7))

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "got %v", err)
	assert.Nil(t, result)
	var alreadyCompleted *accountingAlreadyCompletedError
	require.True(t, errors.As(err, &alreadyCompleted))
	require.NotNil(t, alreadyCompleted.Existing)
	assert.Equal(t, uint64(7), alreadyCompleted.Existing.ID)
	assert.Zero(t, items.deleteCalls, "completed target must not touch items")
	assert.Zero(t, items.calls)
}

func TestAccountingService_Complete_TakeoverRejectsCancelledTarget(t *testing.T) {
	key := uuid.NewString()
	cancelled := takeoverWaitingBilling()
	cancelled.Status = model.BillingStatusCancelled
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return cancelled, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("cancelled target must not create a new row")
			return nil
		},
	}
	items := &mockCompleteItemWriter{}
	svc := newTakeoverTestService(repo, takeoverHospRepo(), items, &mockCompleteTotalsWriter{})

	result, err := svc.Complete(context.Background(), takeoverInput(key, 7))

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "got %v", err)
	assert.Nil(t, result)
	var alreadyCompleted *accountingAlreadyCompletedError
	assert.False(t, errors.As(err, &alreadyCompleted), "cancelled target is a plain 409, not ALREADY_COMPLETED")
	assert.Zero(t, items.deleteCalls)
}

func TestAccountingService_Complete_TakeoverTargetNotFound(t *testing.T) {
	key := uuid.NewString()
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, id uint64) (*model.Billing, error) {
			return nil, apperrors.WrapNotFound("billing", "missing")
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("missing target must not create a new row")
			return nil
		},
	}
	items := &mockCompleteItemWriter{}
	svc := newTakeoverTestService(repo, takeoverHospRepo(), items, &mockCompleteTotalsWriter{})

	result, err := svc.Complete(context.Background(), takeoverInput(key, 999))

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "got %v", err)
	assert.Nil(t, result)
	assert.Zero(t, items.deleteCalls)
}

func TestAccountingService_Complete_TakeoverRejectsMismatchedHospitalization(t *testing.T) {
	key := uuid.NewString()
	bound := takeoverWaitingBilling()
	otherHosp := uint64(55)
	bound.HospitalizationID = &otherHosp
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return bound, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("mismatched target must not create a new row")
			return nil
		},
	}
	items := &mockCompleteItemWriter{}
	svc := newTakeoverTestService(repo, takeoverHospRepo(), items, &mockCompleteTotalsWriter{})

	result, err := svc.Complete(context.Background(), takeoverInput(key, 7))

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "got %v", err)
	assert.Nil(t, result)
	assert.Zero(t, items.deleteCalls)
}

// 暗黙 takeover: billing_id 未指定でも hospitalization スロットの waiting 占有者を確定対象にする。
func TestAccountingService_Complete_TakeoverBindsHospitalizationOccupant(t *testing.T) {
	key := uuid.NewString()
	bound := takeoverWaitingBilling()
	var createCalls int
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		findByHospitalizationIDFn: func(_ context.Context, _, hospID uint64) (*model.Billing, error) {
			require.Equal(t, uint64(99), hospID)
			return bound, nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, id uint64) (*model.Billing, error) {
			require.Equal(t, bound.ID, id)
			return bound, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			createCalls++
			return nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ AccountingUpdate) (*model.Billing, error) {
			return takeoverReloadRow(id), nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Billing, error) {
			return takeoverReloadRow(id), nil
		},
	}
	items := &mockCompleteItemWriter{}
	totals := &mockCompleteTotalsWriter{}
	svc := newTakeoverTestService(repo, takeoverHospRepo(), items, totals)

	input := validCompleteInput(key)
	hospID := uint64(99)
	input.HospitalizationID = &hospID

	result, err := svc.Complete(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint64(7), result.Accounting.ID, "waiting occupant must be settled in place")
	assert.Zero(t, createCalls, "implicit takeover must not INSERT")
	assert.Equal(t, 1, items.deleteCalls)
	assert.Equal(t, 2, items.calls)
}

// cancelled 占有者: 部分 UNIQUE スロットを解放してから新規 INSERT する。
func TestAccountingService_Complete_ReleasesCancelledOccupantThenCreates(t *testing.T) {
	key := uuid.NewString()
	cancelled := takeoverWaitingBilling()
	cancelled.ID = 5
	cancelled.Status = model.BillingStatusCancelled
	var softDeleted []uint64
	var createdID uint64
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		findByHospitalizationIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return cancelled, nil
		},
		softDeleteCancelledFn: func(_ context.Context, _, id uint64) error {
			softDeleted = append(softDeleted, id)
			return nil
		},
		createFn: func(_ context.Context, _ uint64, b *model.Billing) error {
			b.ID = 42
			createdID = b.ID
			return nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ AccountingUpdate) (*model.Billing, error) {
			return &model.Billing{ID: id, ClinicID: 1, Status: model.BillingStatusCompleted}, nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Billing, error) {
			return &model.Billing{ID: id, ClinicID: 1, Status: model.BillingStatusCompleted}, nil
		},
	}
	items := &mockCompleteItemWriter{}
	totals := &mockCompleteTotalsWriter{}
	svc := newTakeoverTestService(repo, takeoverHospRepo(), items, totals)

	input := validCompleteInput(key)
	hospID := uint64(99)
	input.HospitalizationID = &hospID

	result, err := svc.Complete(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, []uint64{5}, softDeleted, "cancelled occupant must be released before insert")
	assert.Equal(t, uint64(42), createdID, "a fresh billing must be created after the release")
	assert.Equal(t, uint64(42), result.Accounting.ID)
	assert.Zero(t, items.deleteCalls, "insert path never bulk-deletes items")
	assert.Equal(t, 2, items.calls)
}

// waiting 占有者が検索→ロック間に completed へ遷移した場合は 409（ロック後再判定）。
func TestAccountingService_Complete_TakeoverOccupantRaceCompleted(t *testing.T) {
	key := uuid.NewString()
	waiting := takeoverWaitingBilling()
	completed := takeoverWaitingBilling()
	completed.Status = model.BillingStatusCompleted
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		findByHospitalizationIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return waiting, nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return completed, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("locked-completed occupant must not create a new row")
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return completed, nil
		},
	}
	items := &mockCompleteItemWriter{}
	svc := newTakeoverTestService(repo, takeoverHospRepo(), items, &mockCompleteTotalsWriter{})

	input := validCompleteInput(key)
	hospID := uint64(99)
	input.HospitalizationID = &hospID

	result, err := svc.Complete(context.Background(), input)

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Nil(t, result)
	var alreadyCompleted *accountingAlreadyCompletedError
	require.True(t, errors.As(err, &alreadyCompleted))
	assert.Zero(t, items.deleteCalls)
}

// 明細一括削除の失敗は takeover 全体を abort させる（新規 item 作成・支払いに進まない）。
func TestAccountingService_Complete_TakeoverAbortsWhenItemDeleteFails(t *testing.T) {
	key := uuid.NewString()
	bound := takeoverWaitingBilling()
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return bound, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("takeover must not INSERT")
			return nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ AccountingUpdate) (*model.Billing, error) {
			return takeoverReloadRow(id), nil
		},
	}
	items := &mockCompleteItemWriter{
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("injected delete failure")
		},
	}
	svc := newTakeoverTestService(repo, takeoverHospRepo(), items, &mockCompleteTotalsWriter{})

	result, err := svc.Complete(context.Background(), takeoverInput(key, 7))

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 1, items.deleteCalls)
	assert.Zero(t, items.calls, "item creation must not run after delete failure")
}

// 冪等 replay は takeover 解決より先に short-circuit する（lock/item 書込を走らない）。
func TestAccountingService_Complete_TakeoverIdempotentReplayShortCircuits(t *testing.T) {
	key := uuid.NewString()
	input := takeoverInput(key, 7)
	existing := takeoverWaitingBilling()
	existing.Status = model.BillingStatusCompleted
	existing.CompletionRequestID = &key
	hash := mustDigest(input)
	existing.CompletionRequestHash = &hash
	var lockCalls int
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return existing, nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			lockCalls++
			return existing, nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return existing, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("replay must not create")
			return nil
		},
	}
	items := &mockCompleteItemWriter{}
	svc := newTakeoverTestService(repo, takeoverHospRepo(), items, &mockCompleteTotalsWriter{})

	result, err := svc.Complete(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Created, "same-key same-digest replay returns the stored accounting")
	assert.Equal(t, uint64(7), result.Accounting.ID)
	assert.Zero(t, lockCalls, "replay must not lock the target row")
	assert.Zero(t, items.deleteCalls)
	assert.Zero(t, items.calls)
}

// stubUnbilledGuard は unbilledWriteGuard のテスト用 stub（EMR-253 takeover の
// revision 免除と blocking warning 検証の切り分けを pin する）。
type stubUnbilledGuard struct {
	assertNoBlockingFn  func(ctx context.Context, clinicID, petID uint64) error
	assertForCompleteFn func(ctx context.Context, clinicID, petID uint64, expectedRevision string) error
}

func (s *stubUnbilledGuard) AssertNoBlockingUnbilled(ctx context.Context, clinicID, petID uint64) error {
	if s.assertNoBlockingFn == nil {
		return nil
	}
	return s.assertNoBlockingFn(ctx, clinicID, petID)
}

func (s *stubUnbilledGuard) AssertUnbilledForComplete(ctx context.Context, clinicID, petID uint64, expectedRevision string) error {
	if s.assertForCompleteFn == nil {
		return nil
	}
	return s.assertForCompleteFn(ctx, clinicID, petID, expectedRevision)
}

// billing_id 明示の takeover では expected_unbilled_revision を必須としない。
// FE の waiting 確定画面は未請求集約を表示していないため「表示した版」が存在しない。
// blocking unbilled warning の検証は残る（AssertNoBlockingUnbilled 経路）。
func TestAccountingService_Complete_TakeoverAllowsEmptyUnbilledRevision(t *testing.T) {
	key := uuid.NewString()
	bound := takeoverWaitingBilling()
	var blockChecked bool
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Billing, error) {
			return bound, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			t.Fatal("takeover must not INSERT")
			return nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ AccountingUpdate) (*model.Billing, error) {
			return takeoverReloadRow(id), nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Billing, error) {
			return takeoverReloadRow(id), nil
		},
	}
	items := &mockCompleteItemWriter{}
	svc := NewAccountingService(
		repo, nil, takeoverHospRepo(), matchingReservationRepo(), nil,
		&mockTransactor{}, &mockAuditService{}, seededPayMethodMock(),
		WithCompleteItemWriter(items),
		WithCompleteTotalsWriter(&mockCompleteTotalsWriter{}),
		WithUnbilledWriteGuard(&stubUnbilledGuard{
			assertNoBlockingFn: func(_ context.Context, _ uint64, _ uint64) error {
				blockChecked = true
				return nil
			},
			assertForCompleteFn: func(_ context.Context, _ uint64, _ uint64, _ string) error {
				t.Fatal("revision unset takeover must not require AssertUnbilledForComplete")
				return nil
			},
		}),
	)

	input := takeoverInput(key, 7)
	input.ExpectedUnbilledRevision = ""
	result, err := svc.Complete(context.Background(), input)

	require.NoError(t, err, "billing_id takeover must not require expected_unbilled_revision")
	require.NotNil(t, result)
	assert.True(t, blockChecked, "blocking unbilled warnings must still be checked")
}

func TestComputeCompleteAccountingDigest_IncludesBillingID(t *testing.T) {
	key := uuid.NewString()
	without := validCompleteInput(key)
	d1, err := ComputeCompleteAccountingDigest(without)
	require.NoError(t, err)

	seven := uint64(7)
	eight := uint64(8)
	withSeven := validCompleteInput(key)
	withSeven.BillingID = &seven
	d2, err := ComputeCompleteAccountingDigest(withSeven)
	require.NoError(t, err)
	withEight := validCompleteInput(key)
	withEight.BillingID = &eight
	d3, err := ComputeCompleteAccountingDigest(withEight)
	require.NoError(t, err)

	assert.NotEqual(t, d1, d2, "billing_id presence must change the digest")
	assert.NotEqual(t, d2, d3, "different takeover targets must produce different digests")
}
