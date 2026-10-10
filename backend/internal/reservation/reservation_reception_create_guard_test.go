package reservation

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestReceptionCreationGuardRejectsWrongRoute(t *testing.T) {
	body := `{"prevent_duplicate_reception":true,"reservation_route":"phone","status":"checked_in"}`
	var req createReservationRequest
	require.NoError(t, json.Unmarshal([]byte(body), &req))
	req.StartTime = time.Now()
	req.EndTime = req.StartTime.Add(15 * time.Minute)
	_, err := req.toServiceInput(1, 2)
	require.Error(t, err)
}

type receptionCreateFake struct {
	*mockReservationRepository
	guardedFn func(context.Context, *model.Reservation) (*model.Reservation, error)
}

func (f *receptionCreateFake) createGuardedReception(ctx context.Context, b *model.Reservation) (*model.Reservation, *uint64, error) {
	created, err := f.guardedFn(ctx, b)
	return created, f.previousAbsorbedOwnerID, err
}

func receptionCreateInput() *CreateManualReservationInput {
	route := "reception"
	owner, pet := uint64(10), uint64(1)
	now := time.Now()
	return &CreateManualReservationInput{ClinicID: 1, OwnerID: &owner, PetID: &pet, StartTime: now, EndTime: now.Add(time.Minute * 15), Status: model.ReservationStatusCheckedIn, ReservationRoute: &route, PreventDuplicateReception: true}
}

func TestReceptionCreationGuardInputAndDependency(t *testing.T) {
	for _, change := range []func(*CreateManualReservationInput){
		func(i *CreateManualReservationInput) { i.Status = model.ReservationStatusConfirmed },
		func(i *CreateManualReservationInput) { i.StartTime = i.StartTime.Add(48 * time.Hour) },
		func(i *CreateManualReservationInput) { i.ReservationRoute = nil },
	} {
		i := receptionCreateInput()
		change(i)
		require.Error(t, validateReceptionCreationGuard(i))
	}
	svc := &reservationService{repo: &mockReservationRepository{findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) { return 10, nil }}, tx: &mockTransactor{}}
	got, err := svc.Create(context.Background(), receptionCreateInput())
	require.Error(t, err)
	require.Nil(t, got)
	var app *apperrors.AppError
	require.True(t, errors.As(err, &app))
	require.Equal(t, "INTERNAL", app.Code)
	_, _, err = (&reservationRepository{}).createGuardedReception(context.Background(), nil)
	require.Error(t, err)
}

func TestReceptionCreationGuardDispatchAndConflict(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		repo := &receptionCreateFake{mockReservationRepository: &mockReservationRepository{findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) { return 10, nil }}, guardedFn: func(ctx context.Context, b *model.Reservation) (*model.Reservation, error) {
			require.Equal(t, true, ctx.Value(checkInTxKey{}))
			if conflict {
				return nil, apperrors.WrapConflict("retry")
			}
			copy := *b
			copy.ID = 42
			return &copy, nil
		}}
		svc := &reservationService{repo: repo, tx: &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(context.WithValue(ctx, checkInTxKey{}, true))
		}}}
		got, err := svc.Create(context.Background(), receptionCreateInput())
		if conflict {
			require.True(t, apperrors.IsConflict(err))
			require.Nil(t, got)
		} else {
			require.NoError(t, err)
			require.Equal(t, uint64(42), got.ID)
		}
	}
}

func TestReceptionCreationBatchLocksAscendingAndReturnsInputOrder(t *testing.T) {
	for _, guarded := range []bool{false, true} {
		var order []uint64
		create := func(_ context.Context, b *model.Reservation) (*model.Reservation, error) {
			order = append(order, *b.PetID)
			copy := *b
			return &copy, nil
		}
		repo := &receptionCreateFake{mockReservationRepository: &mockReservationRepository{findPetOwnerInClinicFn: func(_ context.Context, _, pet uint64) (uint64, error) { return pet * 10, nil }, promoteTentativeOnCheckinFn: create}, guardedFn: create}
		svc := &reservationService{repo: repo, tx: &mockTransactor{}}
		input := receptionCreateInput()
		input.PreventDuplicateReception = guarded
		pets := []ReservationBatchPet{{OwnerID: 20, PetID: 2}, {OwnerID: 10, PetID: 1}}
		got, err := svc.CreateBatch(context.Background(), input, pets)
		require.NoError(t, err)
		require.Equal(t, []uint64{1, 2}, order)
		require.Equal(t, uint64(2), *got[0].PetID)
		require.Equal(t, uint64(1), *got[1].PetID)
		require.Equal(t, uint64(2), pets[0].PetID)
	}
}

func TestReceptionCreationBatchReturnsNoPartialResultOnConflict(t *testing.T) {
	calls := 0
	repo := &receptionCreateFake{mockReservationRepository: &mockReservationRepository{findPetOwnerInClinicFn: func(_ context.Context, _, pet uint64) (uint64, error) { return pet * 10, nil }}, guardedFn: func(_ context.Context, b *model.Reservation) (*model.Reservation, error) {
		calls++
		if calls == 2 {
			return nil, apperrors.WrapConflict("retry")
		}
		return b, nil
	}}
	svc := &reservationService{repo: repo, tx: &mockTransactor{}}
	got, err := svc.CreateBatch(context.Background(), receptionCreateInput(), []ReservationBatchPet{{OwnerID: 20, PetID: 2}, {OwnerID: 10, PetID: 1}})
	require.True(t, apperrors.IsConflict(err))
	require.Nil(t, got)
}

func TestReceptionCreationBatchRequestPropagatesGuard(t *testing.T) {
	var req createReservationBatchRequest
	require.NoError(t, json.Unmarshal([]byte(`{"prevent_duplicate_reception":true,"reservation_route":"reception","status":"checked_in","pets":[{"owner_id":10,"pet_id":1},{"owner_id":20,"pet_id":2}]}`), &req))
	req.StartTime = time.Now()
	req.EndTime = req.StartTime.Add(time.Minute * 15)
	input, pets, err := req.toServiceInput(1, 2)
	require.NoError(t, err)
	require.True(t, input.PreventDuplicateReception)
	require.Len(t, pets, 2)
}

func TestReceptionCreationGuardRejectsNilSuccess(t *testing.T) {
	repo := &receptionCreateFake{mockReservationRepository: &mockReservationRepository{}, guardedFn: func(context.Context, *model.Reservation) (*model.Reservation, error) { return nil, nil }}
	got, _, err := createOrPromoteTentativeWithGuard(context.Background(), repo, &model.Reservation{}, true)
	require.Nil(t, got)
	var app *apperrors.AppError
	require.True(t, errors.As(err, &app))
	require.Equal(t, "INTERNAL", app.Code)
}
