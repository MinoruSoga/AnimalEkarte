package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestLiffCreate_RejectsPetTransferredAfterIdentityResolution(t *testing.T) {
	inTx, wrote := false, false
	repo := &mockReservationRepository{
		countOnDutyDoctorsFn:       func(context.Context, uint64, time.Time) (int64, error) { return 1, nil },
		countConflictsFn:           func(context.Context, uint64, time.Time, time.Time, *uint64) (int64, error) { return 0, nil },
		findPetOwnerInClinicFn:     func(context.Context, uint64, uint64) (uint64, error) { assert.True(t, inTx); return 8, nil },
		absorbTentativeOnBookingFn: func(_ context.Context, b *model.Reservation) (*model.Reservation, error) { wrote = true; return b, nil },
	}
	tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		inTx = true
		defer func() { inTx = false }()
		return fn(ctx)
	}}
	types := mockReservationTypeFinder{findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
		return &model.ReservationType{ID: id, ClinicID: clinicID, IsActive: true, ReservationVisible: true, Category: model.ReservationTypeCategoryGeneral}, nil
	}}
	validators := NewReservationValidators(tx, repo, types, nil, nil, nil, nil, openDayHolidayFinder())
	got, _, err := validators.ValidateAndCreate(context.Background(), &CreateReservationInput{ClinicID: 3, CustomerID: 1, ReservationTypeID: 9, OwnerID: uint64Ptr(7), PetID: uint64Ptr(22), Date: dateInDays(3), StartTime: "1000", EndTime: "1030", Settings: newSettingForValidation()})
	require.ErrorIs(t, err, apperrors.ErrNotFound)
	assert.Nil(t, got)
	assert.False(t, wrote, "stale pre-resolved owner/pet must never reach insert or absorb")
}

func TestLiffCreate_VerifiesLockedCustomerOwnerLink(t *testing.T) {
	for _, tt := range []struct {
		name              string
		resolved, current *uint64
		wantErr           bool
	}{
		{name: "relinked after resolution", resolved: uint64Ptr(7), current: uint64Ptr(8), wantErr: true},
		{name: "unlinked after resolution", resolved: uint64Ptr(7), wantErr: true},
		{name: "unchanged linked owner", resolved: uint64Ptr(7), current: uint64Ptr(7)},
		{name: "unresolved customer keeps post-create attach path", current: uint64Ptr(8)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			inTx, wrote := false, false
			repo := &mockReservationRepository{
				assertLineCustomerReservationOwnerFn: func(_ context.Context, clinicID, customerID uint64, resolved *uint64) error {
					assert.True(t, inTx)
					assert.Equal(t, uint64(3), clinicID)
					assert.Equal(t, uint64(1), customerID)
					return validateLineReservationCustomerOwner(resolved, tt.current)
				},
				findPetOwnerInClinicFn:     func(context.Context, uint64, uint64) (uint64, error) { return 7, nil },
				countOnDutyDoctorsFn:       func(context.Context, uint64, time.Time) (int64, error) { return 1, nil },
				countConflictsFn:           func(context.Context, uint64, time.Time, time.Time, *uint64) (int64, error) { return 0, nil },
				absorbTentativeOnBookingFn: func(_ context.Context, b *model.Reservation) (*model.Reservation, error) { wrote = true; return b, nil },
			}
			tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				inTx = true
				defer func() { inTx = false }()
				return fn(ctx)
			}}
			types := mockReservationTypeFinder{findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
				return &model.ReservationType{ID: id, ClinicID: clinicID, IsActive: true, ReservationVisible: true, Category: model.ReservationTypeCategoryGeneral}, nil
			}}
			validators := NewReservationValidators(tx, repo, types, nil, nil, nil, nil, openDayHolidayFinder())
			input := &CreateReservationInput{ClinicID: 3, CustomerID: 1, ReservationTypeID: 9, OwnerID: tt.resolved, Date: dateInDays(3), StartTime: "1000", EndTime: "1030", Settings: newSettingForValidation()}
			if tt.resolved != nil {
				input.PetID = uint64Ptr(22)
			}
			got, _, err := validators.ValidateAndCreate(context.Background(), input)
			if tt.wantErr {
				require.ErrorIs(t, err, apperrors.ErrConflict)
				assert.Nil(t, got)
				assert.False(t, wrote)
			} else {
				require.NoError(t, err)
				assert.True(t, wrote)
			}
		})
	}
}

func TestLiffCustomerOwnerVerification_RequiresAmbientTransaction(t *testing.T) {
	repo := &reservationRepository{}
	err := repo.assertLineCustomerReservationOwner(context.Background(), 3, 1, uint64Ptr(7))
	require.Error(t, err)
	assert.ErrorContains(t, err, "requires an ambient transaction")
}
