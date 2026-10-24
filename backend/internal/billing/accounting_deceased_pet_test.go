package billing

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
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// BUG-001: 死亡ペットへの会計 Create/Complete を BE で拒否する。

func deceasedAtFixture() time.Time {
	return time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
}

func deceasedPetReservationRepo(ownerID, petID uint64) *mockReservationRepository {
	deceased := deceasedAtFixture()
	return &mockReservationRepository{
		findPetOwnerInClinicFn: func(_ context.Context, _, id uint64) (uint64, error) {
			if id == petID {
				return ownerID, nil
			}
			return id, nil
		},
		findPetByIDInClinicFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
			return &model.Pet{
				ID:         id,
				Status:     model.PetStatusDeceased,
				DeceasedAt: &deceased,
			}, nil
		},
	}
}

func alivePetReservationRepo(ownerID, petID uint64) *mockReservationRepository {
	return &mockReservationRepository{
		findPetOwnerInClinicFn: func(_ context.Context, _, id uint64) (uint64, error) {
			if id == petID {
				return ownerID, nil
			}
			return id, nil
		},
		findPetByIDInClinicFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, Status: model.PetStatusAlive, DeceasedAt: nil}, nil
		},
	}
}

func TestAccountingService_Create_RejectsDeceasedPet(t *testing.T) {
	t.Parallel()
	ownerID := uint64(10)
	petID := uint64(20)
	createCalled := false
	repo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			createCalled = true
			return nil
		},
	}
	svc := NewAccountingService(
		repo, nil, nil, deceasedPetReservationRepo(ownerID, petID), nil,
		&mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{},
		WithPetLifeStateReader(stubPetLifeReader{state: sharedkernel.PetLifeStateDeceased}),
	)

	_, err := svc.Create(context.Background(), &CreateAccountingInput{
		ClinicID:      1,
		OwnerID:       &ownerID,
		PetID:         &petID,
		ScheduledDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "want invalid input, got %v", err)
	assert.Contains(t, err.Error(), accountingDeceasedPetMessage)
	assert.False(t, createCalled, "deceased pet must not reach repository Create")
}

func TestAccountingService_Create_AllowsLivingPet(t *testing.T) {
	t.Parallel()
	ownerID := uint64(10)
	petID := uint64(20)
	createCalled := false
	repo := &mockAccountingRepository{
		createFn: func(_ context.Context, clinicID uint64, b *model.Billing) error {
			createCalled = true
			b.ID = 99
			b.ClinicID = clinicID
			return nil
		},
	}
	svc := NewAccountingService(
		repo, nil, nil, alivePetReservationRepo(ownerID, petID), nil,
		&mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{},
		WithPetLifeStateReader(alivePetLifeReader()),
	)

	out, err := svc.Create(context.Background(), &CreateAccountingInput{
		ClinicID:      1,
		OwnerID:       &ownerID,
		PetID:         &petID,
		ScheduledDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.True(t, createCalled)
	assert.Equal(t, uint64(99), out.ID)
}

func TestAccountingService_Complete_RejectsDeceasedPet_NoWrites(t *testing.T) {
	t.Parallel()
	key := uuid.NewString()
	createCalled := false
	itemCalls := 0
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			createCalled = true
			return nil
		},
	}
	items := &mockCompleteItemWriter{
		createFn: func(_ context.Context, _ *CreateBillingItemInput) (*model.BillingItem, error) {
			itemCalls++
			return &model.BillingItem{ID: 1}, nil
		},
	}
	totals := &mockCompleteTotalsWriter{subtotal: 1000, taxTotal: 100, total: 1100}
	ownerID := uint64(10)
	petID := uint64(20)
	svc := NewAccountingService(
		repo, nil, nil, deceasedPetReservationRepo(ownerID, petID), nil,
		&mockTransactor{}, &mockAuditService{}, seededPayMethodMock(),
		WithCompleteItemWriter(items),
		WithCompleteTotalsWriter(totals),
	)

	input := validCompleteInput(key)
	result, err := svc.Complete(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "want invalid input, got %v", err)
	assert.Contains(t, err.Error(), accountingDeceasedPetMessage)
	assert.False(t, createCalled, "deceased pet complete must not create billing")
	assert.Equal(t, 0, itemCalls, "deceased pet complete must not write items")
}

func TestAccountingService_Complete_AllowsLivingPet(t *testing.T) {
	t.Parallel()
	key := uuid.NewString()
	repo := &mockAccountingRepository{
		findByCompletionRequestIDFn: func(_ context.Context, _ uint64, _ string) (*model.Billing, error) {
			return nil, nil
		},
		createFn: func(_ context.Context, clinicID uint64, b *model.Billing) error {
			b.ID = 42
			b.ClinicID = clinicID
			return nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ AccountingUpdate) (*model.Billing, error) {
			return &model.Billing{
				ID: id, ClinicID: 1, Status: model.BillingStatusCompleted,
				Subtotal: 1000, TaxTotal: 100, TotalAmount: 1100,
			}, nil
		},
		savePaymentFn:       func(_ context.Context, _ *model.Payment) error { return nil },
		savePaymentSplitsFn: func(_ context.Context, _ []model.PaymentSplit) error { return nil },
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Billing, error) {
			return &model.Billing{
				ID: id, ClinicID: 1, Status: model.BillingStatusCompleted,
				Subtotal: 1000, TaxTotal: 100, TotalAmount: 1100,
				CompletionRequestID:   &key,
				CompletionRequestHash: strPtr(mustDigest(validCompleteInput(key))),
			}, nil
		},
	}
	// newCompleteTestService uses matchingReservationRepo (alive by default).
	svc := newCompleteTestService(repo, &mockAuditService{}, &mockCompleteItemWriter{}, &mockCompleteTotalsWriter{subtotal: 1000, taxTotal: 100, total: 1100})

	result, err := svc.Complete(context.Background(), validCompleteInput(key))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Created)
	assert.Equal(t, uint64(42), result.Accounting.ID)
}

// BUG-001 / EMR-491 D6: 死亡 reader は ambient tx を要求するため、会計 Create は WithTx 内で生死を判定する。
func TestAccountingService_Create_PetLifeCheckRunsInsideTransaction(t *testing.T) {
	t.Parallel()
	ownerID := uint64(10)
	petID := uint64(20)
	createCalled := false
	repo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			createCalled = true
			return nil
		},
	}
	tx := &mockTransactor{
		withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(context.WithValue(ctx, petLifeTxMarkerKey{}, true))
		},
	}
	reader := alivePetLifeReader()
	reader.requireTxMarker = true
	svc := NewAccountingService(
		repo, nil, nil, alivePetReservationRepo(ownerID, petID), nil,
		tx, &mockAuditService{}, &mockPaymentMethodMasterRepository{},
		WithPetLifeStateReader(reader),
	)

	out, err := svc.Create(context.Background(), &CreateAccountingInput{
		ClinicID:      1,
		OwnerID:       &ownerID,
		PetID:         &petID,
		ScheduledDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err, "pet life read must run inside the WithTx scope")
	require.NotNil(t, out)
	assert.True(t, createCalled)
}

// BUG-001 / EMR-491 D6: reader 未配線・読取失敗・生死未解決・状態なしは、いずれも生存扱いにせず Create を拒否する。
func TestAccountingService_Create_PetLifeCheckFailsClosed(t *testing.T) {
	ownerID := uint64(10)
	petID := uint64(20)
	cases := []struct {
		name string
		opts []accountingServiceOption
	}{
		{name: "reader 未配線", opts: nil},
		{name: "reader 読取失敗", opts: []accountingServiceOption{WithPetLifeStateReader(stubPetLifeReader{err: errors.New("db down")})}},
		{name: "生死未解決", opts: []accountingServiceOption{WithPetLifeStateReader(stubPetLifeReader{err: sharedkernel.ErrPetLifeStateUnresolved})}},
		{name: "状態なし", opts: []accountingServiceOption{WithPetLifeStateReader(stubPetLifeReader{})}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			createCalled := false
			repo := &mockAccountingRepository{
				createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
					createCalled = true
					return nil
				},
			}
			svc := NewAccountingService(
				repo, nil, nil, alivePetReservationRepo(ownerID, petID), nil,
				&mockTransactor{}, &mockAuditService{}, &mockPaymentMethodMasterRepository{},
				tc.opts...,
			)

			out, err := svc.Create(context.Background(), &CreateAccountingInput{
				ClinicID:      1,
				OwnerID:       &ownerID,
				PetID:         &petID,
				ScheduledDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			})
			require.Error(t, err)
			assert.Nil(t, out)
			assert.False(t, createCalled, "unresolved pet life state must not reach repository Create")
		})
	}
}

// petLifeTxMarkerKey は WithTx 内で ctx に付ける目印（reader が tx 内で呼ばれたかの判定用）。
type petLifeTxMarkerKey struct{}

// stubPetLifeReader は sharedkernel.PetLifeStateReader の明示的な test double。
// requireTxMarker が true のとき、WithTx 内で付与した目印の無い ctx での呼び出しは本番 reader 同様に内部エラーにする。
type stubPetLifeReader struct {
	state           sharedkernel.PetLifeState
	err             error
	requireTxMarker bool
}

func (s stubPetLifeReader) ReadPetLifeState(ctx context.Context, _, _ uint64) (sharedkernel.PetLifeState, error) {
	if s.requireTxMarker && ctx.Value(petLifeTxMarkerKey{}) == nil {
		return "", apperrors.WrapInternalServerError("pet life state read requires an ambient transaction")
	}
	return s.state, s.err
}

// alivePetLifeReader は常に alive を返す明示的な test double。
func alivePetLifeReader() stubPetLifeReader {
	return stubPetLifeReader{state: sharedkernel.PetLifeStateAlive}
}
