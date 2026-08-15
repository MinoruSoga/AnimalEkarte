package reservation

// reservation_owner_pet_clinic_isolation_test.go — AUD-001
// 通常予約 Create/Update と管理予約 Create の Owner/Pet/LineCustomer clinic 所有確認回帰。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func ptrU64(v uint64) *uint64 { return &v }

func TestReservationService_Create_RejectsCrossClinicOwnerPet(t *testing.T) {
	const clinicID = uint64(1)
	const ownedOwnerID = uint64(10)
	const foreignOwnerID = uint64(999)
	const ownedPetID = uint64(20)
	const foreignPetID = uint64(998)
	const otherOwnerSameClinicPetID = uint64(21)
	const otherOwnerID = uint64(11)
	start := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)

	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{ID: id, ClinicID: gotClinicID}, nil
		},
	}

	linkAwareRepo := func(created *bool) *mockReservationRepository {
		return &mockReservationRepository{
			createFn: func(_ context.Context, _ *model.Reservation) error {
				*created = true
				return nil
			},
			assertOwnerInClinicFn: func(_ context.Context, gotClinicID, ownerID uint64) error {
				if gotClinicID != clinicID || ownerID != ownedOwnerID {
					return apperrors.WrapNotFound("owner", "foreign")
				}
				return nil
			},
			findPetOwnerInClinicFn: func(_ context.Context, gotClinicID, petID uint64) (uint64, error) {
				if gotClinicID != clinicID {
					return 0, apperrors.WrapNotFound("pet", "foreign")
				}
				switch petID {
				case ownedPetID:
					return ownedOwnerID, nil
				case otherOwnerSameClinicPetID:
					return otherOwnerID, nil
				default:
					return 0, apperrors.WrapNotFound("pet", "foreign")
				}
			},
		}
	}

	base := func(ownerID, petID *uint64) *CreateManualReservationInput {
		return &CreateManualReservationInput{
			ClinicID:          clinicID,
			StartTime:         start,
			EndTime:           start.Add(30 * time.Minute),
			OwnerID:           ownerID,
			PetID:             petID,
			ReservationTypeID: 50,
			Status:            model.ReservationStatusPending,
			Source:            model.ReservationSourceManual,
		}
	}

	t.Run("rejects cross-clinic owner and does not persist", func(t *testing.T) {
		created := false
		svc := NewReservationServiceWithAvailabilityAndType(linkAwareRepo(&created), typeRepo, &mockTransactor{}, nil, nil)
		out, err := svc.Create(context.Background(), base(ptrU64(foreignOwnerID), nil))
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "want NotFound, got %v", err)
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("rejects cross-clinic pet and does not persist", func(t *testing.T) {
		created := false
		svc := NewReservationServiceWithAvailabilityAndType(linkAwareRepo(&created), typeRepo, &mockTransactor{}, nil, nil)
		out, err := svc.Create(context.Background(), base(nil, ptrU64(foreignPetID)))
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "want NotFound, got %v", err)
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("rejects pet belonging to different owner and does not persist", func(t *testing.T) {
		created := false
		svc := NewReservationServiceWithAvailabilityAndType(linkAwareRepo(&created), typeRepo, &mockTransactor{}, nil, nil)
		out, err := svc.Create(context.Background(), base(ptrU64(ownedOwnerID), ptrU64(otherOwnerSameClinicPetID)))
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("accepts same-clinic matching owner and pet", func(t *testing.T) {
		created := false
		svc := NewReservationServiceWithAvailabilityAndType(linkAwareRepo(&created), typeRepo, &mockTransactor{}, nil, nil)
		out, err := svc.Create(context.Background(), base(ptrU64(ownedOwnerID), ptrU64(ownedPetID)))
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, created)
	})

	t.Run("accepts nil owner and nil pet (existing contract)", func(t *testing.T) {
		created := false
		svc := NewReservationServiceWithAvailabilityAndType(linkAwareRepo(&created), typeRepo, &mockTransactor{}, nil, nil)
		out, err := svc.Create(context.Background(), base(nil, nil))
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, created)
	})

	t.Run("accepts owner only", func(t *testing.T) {
		created := false
		svc := NewReservationServiceWithAvailabilityAndType(linkAwareRepo(&created), typeRepo, &mockTransactor{}, nil, nil)
		out, err := svc.Create(context.Background(), base(ptrU64(ownedOwnerID), nil))
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, created)
	})

	t.Run("accepts pet only", func(t *testing.T) {
		created := false
		svc := NewReservationServiceWithAvailabilityAndType(linkAwareRepo(&created), typeRepo, &mockTransactor{}, nil, nil)
		out, err := svc.Create(context.Background(), base(nil, ptrU64(ownedPetID)))
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestReservationService_Update_RejectsCrossClinicOwnerPet(t *testing.T) {
	const clinicID = uint64(1)
	const reservationID = uint64(7)
	const ownedOwnerID = uint64(10)
	const foreignOwnerID = uint64(999)
	const ownedPetID = uint64(20)
	const foreignPetID = uint64(998)
	const otherOwnerSameClinicPetID = uint64(21)
	const otherOwnerID = uint64(11)
	start := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)

	current := &model.Reservation{
		ID:                reservationID,
		ClinicID:          clinicID,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		OwnerID:           ptrU64(ownedOwnerID),
		PetID:             ptrU64(ownedPetID),
		ReservationTypeID: 50,
		Status:            model.ReservationStatusPending,
	}

	linkAwareRepo := func(updated *bool, snapshot **model.Reservation) *mockReservationRepository {
		cur := *current // copy
		*snapshot = &cur
		return &mockReservationRepository{
			findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.Reservation, error) {
				c := **snapshot
				c.ID = id
				c.ClinicID = gotClinicID
				return &c, nil
			},
			lockAndFindByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.Reservation, error) {
				c := **snapshot
				c.ID = id
				c.ClinicID = gotClinicID
				return &c, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
				*updated = true
				c := **snapshot
				if v, ok := fields["owner_id"]; ok {
					if v == nil {
						c.OwnerID = nil
					} else {
						id := v.(uint64)
						c.OwnerID = &id
					}
				}
				if v, ok := fields["pet_id"]; ok {
					if v == nil {
						c.PetID = nil
					} else {
						id := v.(uint64)
						c.PetID = &id
					}
				}
				*snapshot = &c
				return &c, nil
			},
			assertOwnerInClinicFn: func(_ context.Context, gotClinicID, ownerID uint64) error {
				if gotClinicID != clinicID || ownerID != ownedOwnerID {
					return apperrors.WrapNotFound("owner", "foreign")
				}
				return nil
			},
			findPetOwnerInClinicFn: func(_ context.Context, gotClinicID, petID uint64) (uint64, error) {
				if gotClinicID != clinicID {
					return 0, apperrors.WrapNotFound("pet", "foreign")
				}
				switch petID {
				case ownedPetID:
					return ownedOwnerID, nil
				case otherOwnerSameClinicPetID:
					return otherOwnerID, nil
				default:
					return 0, apperrors.WrapNotFound("pet", "foreign")
				}
			},
		}
	}

	t.Run("rejects cross-clinic owner and leaves reservation unchanged", func(t *testing.T) {
		updated := false
		var snap *model.Reservation
		svc := NewReservationServiceWithAvailabilityAndType(linkAwareRepo(&updated, &snap), nil, &mockTransactor{}, nil, nil)
		out, err := svc.Update(context.Background(), clinicID, reservationID, &UpdateReservationInput{OwnerID: ptrU64(foreignOwnerID)})
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, out)
		assert.False(t, updated)
		assert.Equal(t, ownedOwnerID, *snap.OwnerID)
		assert.Equal(t, ownedPetID, *snap.PetID)
	})

	t.Run("rejects cross-clinic pet and leaves reservation unchanged", func(t *testing.T) {
		updated := false
		var snap *model.Reservation
		svc := NewReservationServiceWithAvailabilityAndType(linkAwareRepo(&updated, &snap), nil, &mockTransactor{}, nil, nil)
		out, err := svc.Update(context.Background(), clinicID, reservationID, &UpdateReservationInput{PetID: ptrU64(foreignPetID)})
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, out)
		assert.False(t, updated)
	})

	t.Run("rejects pet-only change that breaks final owner-pet consistency", func(t *testing.T) {
		updated := false
		var snap *model.Reservation
		svc := NewReservationServiceWithAvailabilityAndType(linkAwareRepo(&updated, &snap), nil, &mockTransactor{}, nil, nil)
		out, err := svc.Update(context.Background(), clinicID, reservationID, &UpdateReservationInput{PetID: ptrU64(otherOwnerSameClinicPetID)})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated)
	})

	t.Run("rejects owner-only change that breaks final owner-pet consistency", func(t *testing.T) {
		// current has pet owned by ownedOwnerID; changing owner to otherOwnerID without pet change must fail
		updated := false
		var snap *model.Reservation
		repo := linkAwareRepo(&updated, &snap)
		repo.assertOwnerInClinicFn = func(_ context.Context, gotClinicID, ownerID uint64) error {
			if gotClinicID != clinicID || (ownerID != ownedOwnerID && ownerID != otherOwnerID) {
				return apperrors.WrapNotFound("owner", "foreign")
			}
			return nil
		}
		svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)
		out, err := svc.Update(context.Background(), clinicID, reservationID, &UpdateReservationInput{OwnerID: ptrU64(otherOwnerID)})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, updated)
	})

	t.Run("accepts clearing owner and pet via zero", func(t *testing.T) {
		updated := false
		var snap *model.Reservation
		svc := NewReservationServiceWithAvailabilityAndType(linkAwareRepo(&updated, &snap), nil, &mockTransactor{}, nil, nil)
		zero := uint64(0)
		out, err := svc.Update(context.Background(), clinicID, reservationID, &UpdateReservationInput{OwnerID: &zero, PetID: &zero})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, updated)
		assert.Nil(t, snap.OwnerID)
		assert.Nil(t, snap.PetID)
	})

	t.Run("accepts owner-only update when pet is cleared", func(t *testing.T) {
		updated := false
		var snap *model.Reservation
		svc := NewReservationServiceWithAvailabilityAndType(linkAwareRepo(&updated, &snap), nil, &mockTransactor{}, nil, nil)
		zero := uint64(0)
		out, err := svc.Update(context.Background(), clinicID, reservationID, &UpdateReservationInput{OwnerID: ptrU64(ownedOwnerID), PetID: &zero})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, updated)
	})

	t.Run("accepts both owner and pet update when consistent", func(t *testing.T) {
		updated := false
		var snap *model.Reservation
		svc := NewReservationServiceWithAvailabilityAndType(linkAwareRepo(&updated, &snap), nil, &mockTransactor{}, nil, nil)
		out, err := svc.Update(context.Background(), clinicID, reservationID, &UpdateReservationInput{
			OwnerID: ptrU64(ownedOwnerID),
			PetID:   ptrU64(ownedPetID),
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, updated)
	})
}

func TestReservationService_Update_RejectsLineCheckedInWhenOwnerPetCleared(t *testing.T) {
	const clinicID = uint64(1)
	const reservationID = uint64(7)
	start := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	status := model.ReservationStatusCheckedIn
	zero := uint64(0)
	updated := false
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, gotClinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID: id, ClinicID: gotClinicID, StartTime: start, EndTime: start.Add(30 * time.Minute),
				OwnerID: ptrU64(10), PetID: ptrU64(20), LineCustomerID: ptrU64(30),
				Source: model.ReservationSourceLine, Status: model.ReservationStatusPending,
				ReservationTypeID: 50,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			updated = true
			return &model.Reservation{ID: reservationID}, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)
	out, err := svc.Update(context.Background(), clinicID, reservationID, &UpdateReservationInput{
		Status: &status, OwnerID: &zero, PetID: &zero,
	})
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, out)
	assert.False(t, updated)
}
