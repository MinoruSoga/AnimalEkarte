package trimming

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type trimmingTxContextKey struct{}

type trimmingTxMarker struct{}

func (trimmingTxMarker) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(context.WithValue(ctx, trimmingTxContextKey{}, true))
}

func newTrimmingClinicIsolationTestService(
	reserv *mockTrimmingReservationRepository,
	detail *mockTrimmingDetailRepository,
) Service {
	return withTrimmingTestActor(NewServiceWithAudit(
		reserv,
		&mockTrimmingReservationTypeRepository{},
		nil,
		&mockTrimmingUnavailableTimeRepository{},
		detail,
		nil,
		nil,
		trimmingTxMarker{},
		noopTrimmingAuditTxLogger{},
	))
}

func assertTrimmingTxContext(ctx context.Context, t *testing.T) {
	t.Helper()
	assert.Equal(t, true, ctx.Value(trimmingTxContextKey{}), "Pet ownership checks and writes must share the transaction context")
}

func TestService_Create_RejectsPetFromAnotherClinic(t *testing.T) {
	petID := uint64(202)
	reserv := &mockTrimmingReservationRepository{
		findPetOwnerFn: func(ctx context.Context, clinicID, actualPetID uint64) (uint64, error) {
			assertTrimmingTxContext(ctx, t)
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, petID, actualPetID)
			return 0, apperrors.WrapNotFound("pet", "202")
		},
		createFn: func(_ context.Context, _ *model.Reservation) error {
			t.Fatal("appointment must not be created for a pet from another clinic")
			return nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			t.Fatal("trimming detail must not be created for a pet from another clinic")
			return nil
		},
	}
	svc := newTrimmingClinicIsolationTestService(reserv, detail)

	got, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		ReservationTypeID: 9,
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(time.Hour),
		PetID:             &petID,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, got)
}

func TestService_CreateExistingDetail_RejectsPetFromAnotherClinic(t *testing.T) {
	appointmentID := uint64(77)
	petID := uint64(202)
	category := model.ReservationTypeCategoryTrimming
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:                id,
				ClinicID:          clinicID,
				ReservationTypeID: 9,
				ReservationType:   &model.ReservationType{ID: 9, Category: category},
			}, nil
		},
		findPetOwnerFn: func(ctx context.Context, clinicID, actualPetID uint64) (uint64, error) {
			assertTrimmingTxContext(ctx, t)
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, petID, actualPetID)
			return 0, apperrors.WrapNotFound("pet", "202")
		},
		updateFieldsFn: func(_ context.Context, _ uint64, _ uint64, _ map[string]any) (*model.Reservation, error) {
			t.Fatal("appointment must not be updated for a pet from another clinic")
			return nil, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			return nil, apperrors.WrapNotFound("appointment_trimming_detail", "missing")
		},
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			t.Fatal("trimming detail must not be created for a pet from another clinic")
			return nil
		},
	}
	svc := newTrimmingClinicIsolationTestService(reserv, detail)

	got, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		AppointmentID: &appointmentID,
		PetID:         &petID,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, got)
}

func TestService_Update_RejectsPetFromAnotherClinic(t *testing.T) {
	appointmentID := uint64(77)
	petID := uint64(202)
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID}, nil
		},
		findPetOwnerFn: func(ctx context.Context, clinicID, actualPetID uint64) (uint64, error) {
			assertTrimmingTxContext(ctx, t)
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, petID, actualPetID)
			return 0, apperrors.WrapNotFound("pet", "202")
		},
		updateFieldsFn: func(_ context.Context, _ uint64, _ uint64, _ map[string]any) (*model.Reservation, error) {
			t.Fatal("appointment must not be updated for a pet from another clinic")
			return nil, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			t.Fatal("trimming detail must not be loaded after Pet ownership validation fails")
			return nil, nil
		},
	}
	svc := newTrimmingClinicIsolationTestService(reserv, detail)

	got, err := svc.Update(context.Background(), 1, appointmentID, &UpdateTrimmingInput{PetID: &petID})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, got)
}

func TestService_Create_AcceptsPetFromSameClinicBeforeWrites(t *testing.T) {
	petID := uint64(101)
	events := make([]string, 0, 3)
	reserv := &mockTrimmingReservationRepository{
		findPetOwnerFn: func(ctx context.Context, _, _ uint64) (uint64, error) {
			assertTrimmingTxContext(ctx, t)
			events = append(events, "validate-pet")
			return 501, nil
		},
		createFn: func(ctx context.Context, appt *model.Reservation) error {
			assertTrimmingTxContext(ctx, t)
			events = append(events, "create-appointment")
			appt.ID = 88
			return nil
		},
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID, PetID: &petID}, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		createFn: func(ctx context.Context, _ *model.AppointmentTrimmingDetail) error {
			assertTrimmingTxContext(ctx, t)
			events = append(events, "create-detail")
			return nil
		},
		findByAppointmentIDFn: func(_ context.Context, clinicID, appointmentID uint64) (*model.AppointmentTrimmingDetail, error) {
			return &model.AppointmentTrimmingDetail{ClinicID: clinicID, AppointmentID: appointmentID}, nil
		},
	}
	svc := newTrimmingClinicIsolationTestService(reserv, detail)

	got, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		ReservationTypeID: 9,
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(time.Hour),
		PetID:             &petID,
	})

	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, []string{"validate-pet", "create-appointment", "create-detail"}, events)
}

func TestService_Update_AcceptsPetFromSameClinicAfterLock(t *testing.T) {
	appointmentID := uint64(77)
	ownerID := uint64(501)
	petID := uint64(101)
	events := make([]string, 0, 6)
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID}, nil
		},
		acquireBookingLockFn: func(ctx context.Context, clinicID uint64) error {
			assertTrimmingTxContext(ctx, t)
			assert.Equal(t, uint64(1), clinicID)
			events = append(events, "acquire-booking-lock")
			return nil
		},
		lockAndFindByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
			assertTrimmingTxContext(ctx, t)
			events = append(events, "lock-appointment")
			return &model.Reservation{ID: id, ClinicID: clinicID, OwnerID: &ownerID}, nil
		},
		findPetOwnerFn: func(ctx context.Context, _, _ uint64) (uint64, error) {
			assertTrimmingTxContext(ctx, t)
			events = append(events, "validate-pet")
			return ownerID, nil
		},
		updateFieldsFn: func(ctx context.Context, _ uint64, _ uint64, fields map[string]any) (*model.Reservation, error) {
			assertTrimmingTxContext(ctx, t)
			events = append(events, "update-appointment")
			assert.Equal(t, petID, fields["pet_id"])
			return &model.Reservation{ID: appointmentID, ClinicID: 1, OwnerID: &ownerID, PetID: &petID}, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(ctx context.Context, clinicID, id uint64) (*model.AppointmentTrimmingDetail, error) {
			if ctx.Value(trimmingTxContextKey{}) == true {
				events = append(events, "find-detail")
			}
			return &model.AppointmentTrimmingDetail{ClinicID: clinicID, AppointmentID: id}, nil
		},
		updateFn: func(ctx context.Context, _ *model.AppointmentTrimmingDetail) error {
			assertTrimmingTxContext(ctx, t)
			events = append(events, "update-detail")
			return nil
		},
	}
	svc := newTrimmingClinicIsolationTestService(reserv, detail)

	got, err := svc.Update(context.Background(), 1, appointmentID, &UpdateTrimmingInput{PetID: &petID})

	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, []string{
		"acquire-booking-lock",
		"lock-appointment",
		"validate-pet",
		"update-appointment",
		"find-detail",
		"update-detail",
		"find-detail",
	}, events)
}

func TestService_Update_RejectsPetBelongingToDifferentOwner(t *testing.T) {
	appointmentID := uint64(77)
	ownerID := uint64(501)
	petOwnerID := uint64(502)
	petID := uint64(101)
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID}, nil
		},
		lockAndFindByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID, OwnerID: &ownerID}, nil
		},
		findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
			return petOwnerID, nil
		},
		updateFieldsFn: func(_ context.Context, _ uint64, _ uint64, _ map[string]any) (*model.Reservation, error) {
			t.Fatal("appointment must not be updated when Pet does not belong to the reservation Owner")
			return nil, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			t.Fatal("trimming detail must not be loaded when Owner-Pet validation fails")
			return nil, nil
		},
	}
	svc := newTrimmingClinicIsolationTestService(reserv, detail)

	got, err := svc.Update(context.Background(), 1, appointmentID, &UpdateTrimmingInput{PetID: &petID})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, got)
}
