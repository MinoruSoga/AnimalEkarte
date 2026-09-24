package reservation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockOwnerNameOnlyRepo は FindByLineUserID capability を持たない最小 ownerRepo。
// liffOwnerRepo (FindByNameAndPhone のみ) を満たす。
type mockOwnerNameOnlyRepo struct{}

func (m *mockOwnerNameOnlyRepo) FindByNameAndPhone(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
	return nil, nil
}

type ownerSyncFixture struct {
	customer      *model.LineCustomer
	healed        *model.LineCustomer
	owner         *model.Owner
	ownerErr      error
	updateErr     error
	expectedCalls int
	assert        func(t *testing.T, got *model.LineCustomer)
}

func TestLiffService_findCustomerWithOwnerSync(t *testing.T) {
	ownerID := uint64(50)
	otherOwnerID := uint64(99)
	unlinked := &model.LineCustomer{ID: 100, ClinicID: 1, LineUserID: "Uabc", DisplayName: "LINE表示名"}
	linked := &model.LineCustomer{ID: 100, ClinicID: 1, LineUserID: "Uabc", DisplayName: "LINE表示名", OwnerID: &ownerID}
	healed := &model.LineCustomer{
		ID: 100, ClinicID: 1, LineUserID: "Uabc", DisplayName: "LINE表示名", OwnerID: &ownerID,
		Owner: &model.Owner{ID: ownerID, ClinicID: 1, Name: "山田太郎"},
	}
	owner := &model.Owner{ID: ownerID, ClinicID: 1, Name: "山田太郎"}

	cases := map[string]ownerSyncFixture{
		"heals NULL owner_id via verified line_user_id": {
			customer:      unlinked,
			healed:        healed,
			owner:         owner,
			expectedCalls: 1,
			assert: func(t *testing.T, got *model.LineCustomer) {
				require.NotNil(t, got.OwnerID)
				assert.Equal(t, ownerID, *got.OwnerID)
				require.NotNil(t, got.Owner)
				assert.Equal(t, ownerID, got.Owner.ID)
			},
		},
		"already-linked is a no-op without write": {
			customer:      linked,
			owner:         owner,
			expectedCalls: 0,
			assert: func(t *testing.T, got *model.LineCustomer) {
				assert.Same(t, linked, got)
			},
		},
		"conflicting owner_id is never overwritten": {
			customer:      &model.LineCustomer{ID: 100, ClinicID: 1, LineUserID: "Uabc", OwnerID: &otherOwnerID},
			owner:         owner,
			expectedCalls: 0,
			assert: func(t *testing.T, got *model.LineCustomer) {
				require.NotNil(t, got.OwnerID)
				assert.Equal(t, otherOwnerID, *got.OwnerID)
			},
		},
		"no matching owner stays unlinked": {
			customer:      unlinked,
			ownerErr:      apperrors.ErrNotFound,
			expectedCalls: 0,
			assert: func(t *testing.T, got *model.LineCustomer) {
				assert.Nil(t, got.OwnerID)
			},
		},
		"generic owner lookup error degrades to unlinked customer": {
			customer:      unlinked,
			ownerErr:      errors.New("db down"),
			expectedCalls: 0,
			assert: func(t *testing.T, got *model.LineCustomer) {
				assert.Nil(t, got.OwnerID)
			},
		},
		"UpdateOwnerLink failure returns unlinked customer": {
			customer:      unlinked,
			owner:         owner,
			updateErr:     errors.New("update failed"),
			expectedCalls: 1,
			assert: func(t *testing.T, got *model.LineCustomer) {
				assert.Nil(t, got.OwnerID)
			},
		},
	}

	for name, fixture := range cases {
		t.Run(name, func(t *testing.T) {
			var updateCalls []uint64
			findCalls := 0
			customerRepo := &mockLiffCustomerRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.LineCustomer, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(100), id)
					findCalls++
					if findCalls == 1 || fixture.healed == nil {
						return fixture.customer, nil
					}
					return fixture.healed, nil
				},
				updateOwnerLinkFn: func(_ context.Context, clinicID, id uint64, ownerID *uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(100), id)
					require.NotNil(t, ownerID)
					updateCalls = append(updateCalls, *ownerID)
					return fixture.updateErr
				},
			}
			ownerRepo := &mockLiffOwnerRepository{
				findByLineUserIDFn: func(_ context.Context, clinicID uint64, lineUserID string) (*model.Owner, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "Uabc", lineUserID)
					return fixture.owner, fixture.ownerErr
				},
			}
			svc := &liffService{customerRepo: customerRepo, ownerRepo: ownerRepo}

			got, err := svc.findCustomerWithOwnerSync(context.Background(), 1, 100)

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Len(t, updateCalls, fixture.expectedCalls)
			if fixture.expectedCalls > 0 {
				assert.Equal(t, ownerID, updateCalls[0])
			}
			fixture.assert(t, got)
		})
	}

	t.Run("customer without line_user_id skips owner lookup", func(t *testing.T) {
		lookupCalled := false
		customerRepo := &mockLiffCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				return &model.LineCustomer{ID: 100, ClinicID: 1, LineUserID: ""}, nil
			},
			updateOwnerLinkFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
				t.Error("UpdateOwnerLink must not be called")
				return nil
			},
		}
		ownerRepo := &mockLiffOwnerRepository{
			findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				lookupCalled = true
				return nil, nil
			},
		}
		svc := &liffService{customerRepo: customerRepo, ownerRepo: ownerRepo}

		got, err := svc.findCustomerWithOwnerSync(context.Background(), 1, 100)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.False(t, lookupCalled)
	})

	t.Run("owner repo without lookup capability degrades gracefully", func(t *testing.T) {
		customerRepo := &mockLiffCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				return unlinked, nil
			},
			updateOwnerLinkFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
				t.Error("UpdateOwnerLink must not be called")
				return nil
			},
		}
		svc := &liffService{customerRepo: customerRepo, ownerRepo: &mockOwnerNameOnlyRepo{}}

		got, err := svc.findCustomerWithOwnerSync(context.Background(), 1, 100)

		require.NoError(t, err)
		assert.Same(t, unlinked, got)
	})

	t.Run("cross-clinic match does not heal", func(t *testing.T) {
		customerRepo := &mockLiffCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				return unlinked, nil
			},
			updateOwnerLinkFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
				t.Error("UpdateOwnerLink must not be called")
				return nil
			},
		}
		ownerRepo := &mockLiffOwnerRepository{
			findByLineUserIDFn: func(_ context.Context, clinicID uint64, _ string) (*model.Owner, error) {
				// The lookup is clinic-scoped; a foreign-clinic link yields not-found.
				assert.Equal(t, uint64(1), clinicID)
				return nil, apperrors.ErrNotFound
			},
		}
		svc := &liffService{customerRepo: customerRepo, ownerRepo: ownerRepo}

		got, err := svc.findCustomerWithOwnerSync(context.Background(), 1, 100)

		require.NoError(t, err)
		assert.Same(t, unlinked, got)
	})

	t.Run("second call is idempotent once linked", func(t *testing.T) {
		var updateCalls int
		current := unlinked
		customerRepo := &mockLiffCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				return current, nil
			},
			updateOwnerLinkFn: func(_ context.Context, _, _ uint64, ownerID *uint64) error {
				updateCalls++
				current = healed
				return nil
			},
		}
		ownerRepo := &mockLiffOwnerRepository{
			findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return owner, nil
			},
		}
		svc := &liffService{customerRepo: customerRepo, ownerRepo: ownerRepo}

		first, err := svc.findCustomerWithOwnerSync(context.Background(), 1, 100)
		require.NoError(t, err)
		require.NotNil(t, first.OwnerID)

		second, err := svc.findCustomerWithOwnerSync(context.Background(), 1, 100)
		require.NoError(t, err)
		require.NotNil(t, second.OwnerID)
		assert.Equal(t, 1, updateCalls, "already-linked customer converges without a second write")
	})

	t.Run("initial FindByID error propagates", func(t *testing.T) {
		underlying := errors.New("db down")
		customerRepo := &mockLiffCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				return nil, underlying
			},
		}
		svc := &liffService{customerRepo: customerRepo, ownerRepo: &mockLiffOwnerRepository{}}

		got, err := svc.findCustomerWithOwnerSync(context.Background(), 1, 100)

		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, underlying))
	})
}

// TestLiffService_GetHealthCard_OwnerSync は BUG-LIFF-HEALTHCARD-OWNER-SYNC の
// 回帰: owners.line_user_id 連携済みで line_customers.owner_id が NULL の顧客でも
// 健康手帳が表示名フォールバックではなく実owner名+ペット+ワクチンを返す。
func TestLiffService_GetHealthCard_OwnerSync(t *testing.T) {
	ownerID := uint64(50)
	unlinked := &model.LineCustomer{ID: 100, ClinicID: 1, LineUserID: "Uabc", DisplayName: "LINE表示名"}
	healed := &model.LineCustomer{
		ID: 100, ClinicID: 1, LineUserID: "Uabc", DisplayName: "LINE表示名", OwnerID: &ownerID,
		Owner: &model.Owner{
			ID: ownerID, ClinicID: 1, Name: "山田太郎",
			Pets: []model.Pet{{ID: 10, OwnerID: ownerID, Name: "ポチ", AnimalSpeciesID: 1,
				AnimalSpecies: &model.AnimalSpecies{ID: 1, Name: "犬"}}},
		},
	}

	var updateCalls []uint64
	findCalls := 0
	customerRepo := &mockLiffCustomerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
			findCalls++
			if findCalls == 1 {
				return unlinked, nil
			}
			return healed, nil
		},
		updateOwnerLinkFn: func(_ context.Context, clinicID, id uint64, ownerIDPtr *uint64) error {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(100), id)
			require.NotNil(t, ownerIDPtr)
			updateCalls = append(updateCalls, *ownerIDPtr)
			return nil
		},
	}
	ownerRepo := &mockLiffOwnerRepository{
		findByLineUserIDFn: func(_ context.Context, clinicID uint64, lineUserID string) (*model.Owner, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, "Uabc", lineUserID)
			return healed.Owner, nil
		},
	}
	petID := uint64(10)
	vaccinationRepo := &mockVaccinationRepository{
		findByOwnerFn: func(_ context.Context, clinicID, requestedOwnerID uint64) ([]model.Vaccination, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, ownerID, requestedOwnerID)
			return []model.Vaccination{{PetID: &petID, Vaccine: &model.Vaccine{ID: 5, Name: "混合ワクチン"}}}, nil
		},
	}
	svc := newHealthCardTestServiceWithOwner(customerRepo, vaccinationRepo, ownerRepo)

	result, err := svc.GetHealthCard(context.Background(), 1, 100)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, []uint64{ownerID}, updateCalls, "heal write runs exactly once with the resolved owner id")
	assert.Equal(t, "山田太郎", result.OwnerName, "owner 名が返ること（表示名フォールバックではない）")
	require.Len(t, result.Pets, 1)
	assert.Equal(t, "ポチ", result.Pets[0].PetName)
	require.Len(t, result.Pets[0].Vaccines, 1)
	assert.Equal(t, "混合ワクチン", result.Pets[0].Vaccines[0].VaccineName)
}

// TestLiffService_GetProfile_OwnerSync は GetProfile でも同じ heal が働くことを固定する。
func TestLiffService_GetProfile_OwnerSync(t *testing.T) {
	ownerID := uint64(50)
	unlinked := &model.LineCustomer{ID: 100, ClinicID: 1, LineUserID: "Uabc"}
	healed := &model.LineCustomer{
		ID: 100, ClinicID: 1, LineUserID: "Uabc", OwnerID: &ownerID,
		Owner: &model.Owner{ID: ownerID, ClinicID: 1, Name: "山田太郎"},
	}
	var updateCalls int
	findCalls := 0
	svc := &liffService{
		customerRepo: &mockLiffCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				findCalls++
				if findCalls == 1 {
					return unlinked, nil
				}
				return healed, nil
			},
			updateOwnerLinkFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
				updateCalls++
				return nil
			},
		},
		ownerRepo: &mockLiffOwnerRepository{
			findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return healed.Owner, nil
			},
		},
	}

	got, err := svc.GetProfile(context.Background(), 1, 100)

	require.NoError(t, err)
	require.NotNil(t, got.OwnerID)
	assert.Equal(t, ownerID, *got.OwnerID)
	require.NotNil(t, got.Owner)
	assert.Equal(t, 1, updateCalls)
}

// TestLiffService_ReservationOwnerAttach_OwnerSync は予約の owner/pet 反映経路でも
// heal 済み顧客が使われることを固定する。
func TestLiffService_ReservationOwnerAttach_OwnerSync(t *testing.T) {
	ownerID := uint64(50)
	unlinked := &model.LineCustomer{ID: 100, ClinicID: 1, LineUserID: "Uabc"}
	healed := &model.LineCustomer{
		ID: 100, ClinicID: 1, LineUserID: "Uabc", OwnerID: &ownerID,
		Owner: &model.Owner{
			ID: ownerID, ClinicID: 1, Name: "山田太郎",
			Pets: []model.Pet{{ID: 10, OwnerID: ownerID, Name: "ポチ"}},
		},
	}
	findCalls := 0
	var updateCalls int
	var attachedFields map[string]any
	svc := &liffService{
		reservationRepo: &mockLiffReservationRepository{
			updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
				attachedFields = fields
				return &model.Reservation{ID: 7, ClinicID: 1, OwnerID: &ownerID, PetID: func() *uint64 { v := uint64(10); return &v }()}, nil
			},
		},
		customerRepo: &mockLiffCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				findCalls++
				if findCalls == 1 {
					return unlinked, nil
				}
				return healed, nil
			},
			updateOwnerLinkFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
				updateCalls++
				return nil
			},
		},
		ownerRepo: &mockLiffOwnerRepository{
			findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return healed.Owner, nil
			},
		},
	}
	appt := &model.Reservation{ID: 7, ClinicID: 1}

	svc.tryAttachReservationOwnerPet(context.Background(), 1, 100, appt, nil)

	assert.Equal(t, 1, updateCalls, "owner_id heal write runs once")
	require.NotNil(t, attachedFields, "healed customer reaches the reservation attach")
	assert.Equal(t, ownerID, attachedFields["owner_id"])
	assert.Equal(t, uint64(10), attachedFields["pet_id"])
}
