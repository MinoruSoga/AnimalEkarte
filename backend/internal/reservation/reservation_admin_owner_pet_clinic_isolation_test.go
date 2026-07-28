package reservation

// reservation_admin_owner_pet_clinic_isolation_test.go — BE9-2C R③: 旧
// reservation_owner_pet_clinic_isolation_test.go の ReservationAdminService（R④未移行）節を
// R④で internal/reservation へ合流（分離時の予定通り）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestReservationAdminService_Create_RejectsCrossClinicOwnerPetLineCustomer(t *testing.T) {
	const clinicID = uint64(1)
	const ownedOwnerID = uint64(10)
	const foreignOwnerID = uint64(999)
	const ownedPetID = uint64(20)
	const foreignPetID = uint64(998)
	const ownedLineCustomerID = uint64(30)
	const foreignLineCustomerID = uint64(997)
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
			countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
				return 1, nil
			},
			assertOwnerInClinicFn: func(_ context.Context, gotClinicID, ownerID uint64) error {
				if gotClinicID != clinicID || ownerID != ownedOwnerID {
					return apperrors.WrapNotFound("owner", "foreign")
				}
				return nil
			},
			findPetOwnerInClinicFn: func(_ context.Context, gotClinicID, petID uint64) (uint64, error) {
				if gotClinicID != clinicID || petID != ownedPetID {
					return 0, apperrors.WrapNotFound("pet", "foreign")
				}
				return ownedOwnerID, nil
			},
			assertLineCustomerInClinicFn: func(_ context.Context, gotClinicID, id uint64) error {
				if gotClinicID != clinicID || id != ownedLineCustomerID {
					return apperrors.WrapNotFound("line_customer", "foreign")
				}
				return nil
			},
		}
	}

	base := func(ownerID, petID, lineCustomerID *uint64) *CreateReservationAdminInput {
		return &CreateReservationAdminInput{
			StartTime:         start,
			EndTime:           start.Add(30 * time.Minute),
			OwnerID:           ownerID,
			PetID:             petID,
			LineCustomerID:    lineCustomerID,
			ReservationTypeID: 50,
		}
	}

	newSvc := func(created *bool) ReservationAdminService {
		return NewReservationAdminServiceWithAvailabilityAndType(
			&mockReservationAdminRepository{}, linkAwareRepo(created), typeRepo, &mockTransactor{}, nil, nil,
		)
	}

	t.Run("rejects cross-clinic owner and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, base(ptrU64(foreignOwnerID), nil, nil))
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("rejects cross-clinic pet and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, base(nil, ptrU64(foreignPetID), nil))
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("rejects cross-clinic line customer and does not persist", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, base(nil, nil, ptrU64(foreignLineCustomerID)))
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("accepts same-clinic owner pet and line customer", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, base(ptrU64(ownedOwnerID), ptrU64(ownedPetID), ptrU64(ownedLineCustomerID)))
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, created)
	})
}

func TestReservationAdminService_Create_OwnerPetConsistencyAndNilContract(t *testing.T) {
	const clinicID = uint64(1)
	const ownedOwnerID = uint64(10)
	const otherOwnerID = uint64(11)
	const ownedPetID = uint64(20)
	const otherOwnerPetID = uint64(21)
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
			countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
				return 1, nil
			},
			assertOwnerInClinicFn: func(_ context.Context, gotClinicID, ownerID uint64) error {
				if gotClinicID != clinicID || (ownerID != ownedOwnerID && ownerID != otherOwnerID) {
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
				case otherOwnerPetID:
					return otherOwnerID, nil
				default:
					return 0, apperrors.WrapNotFound("pet", "foreign")
				}
			},
		}
	}
	newSvc := func(created *bool) ReservationAdminService {
		return NewReservationAdminServiceWithAvailabilityAndType(
			&mockReservationAdminRepository{}, linkAwareRepo(created), typeRepo, &mockTransactor{}, nil, nil,
		)
	}

	t.Run("rejects mismatched owner-pet", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, &CreateReservationAdminInput{
			StartTime: start, EndTime: start.Add(30 * time.Minute),
			OwnerID: ptrU64(ownedOwnerID), PetID: ptrU64(otherOwnerPetID), ReservationTypeID: 50,
		})
		assert.Error(t, err)
		assert.Nil(t, out)
		assert.False(t, created)
	})

	t.Run("accepts nil owner pet and line customer", func(t *testing.T) {
		created := false
		svc := newSvc(&created)
		out, err := svc.Create(context.Background(), clinicID, &CreateReservationAdminInput{
			StartTime: start, EndTime: start.Add(30 * time.Minute), ReservationTypeID: 50,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, created)
	})
}
