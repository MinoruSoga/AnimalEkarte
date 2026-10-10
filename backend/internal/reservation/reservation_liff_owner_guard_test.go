package reservation

import (
	"context"
	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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
