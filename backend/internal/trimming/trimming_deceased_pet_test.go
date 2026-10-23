package trimming

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// TASK-030: 死亡ペット検証が input.PetID 有無に依存せず finalPetID を見ることを経路別に固定する。

func trimmingApptWithPet(id, clinicID, petID uint64) *model.Reservation {
	start := time.Date(2026, time.August, 1, 10, 0, 0, 0, time.UTC)
	return &model.Reservation{
		ID:                id,
		ClinicID:          clinicID,
		ReservationTypeID: 9,
		ReservationType: &model.ReservationType{
			ID:       9,
			ClinicID: clinicID,
			Category: model.ReservationTypeCategoryTrimming,
			IsActive: true,
		},
		PetID:     &petID,
		StartTime: start,
		EndTime:   start.Add(time.Hour),
		Status:    model.ReservationStatusPending,
		Source:    model.ReservationSourceManual,
	}
}

func newDeceasedGuardTestService(
	reserv *mockTrimmingReservationRepository,
	detail *mockTrimmingDetailRepository,
	audit *trimmingAuditRecorder,
) Service {
	return withTrimmingTestActor(newTrimmingAuditTestService(reserv, detail, &mockTransactor{}, audit))
}

// ① detail 作成: input.PetID == nil かつ予約由来ペットが死亡 → 拒否・write/audit 0
func TestService_CreateExistingDetail_RejectsDeceasedPetWhenPetIDOmitted(t *testing.T) {
	appointmentID := uint64(77)
	petID := uint64(501)
	detailCreateCalls := 0
	appointmentUpdateCalls := 0
	audit := &trimmingAuditRecorder{}

	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return trimmingApptWithPet(id, clinicID, petID), nil
		},
		findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
			return 10, nil
		},
		petLifeStateFn: func(uint64) (sharedkernel.PetLifeState, error) {
			return sharedkernel.PetLifeStateDeceased, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			appointmentUpdateCalls++
			t.Fatal("appointment must not be updated for a deceased appointment pet when pet_id is omitted")
			return nil, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			return nil, apperrors.WrapNotFound("appointment_trimming_detail", "missing")
		},
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			detailCreateCalls++
			t.Fatal("trimming detail must not be created for a deceased appointment pet when pet_id is omitted")
			return nil
		},
	}
	svc := newDeceasedGuardTestService(reserv, detail, audit)

	got, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		AppointmentID: &appointmentID,
		// PetID intentionally omitted — finalPetID resolves to locked.PetID
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, got)
	assert.Zero(t, detailCreateCalls)
	assert.Zero(t, appointmentUpdateCalls)
	assert.Empty(t, audit.entries)
}

// ② Update: input.PetID == nil かつ予約由来ペットが死亡 → 拒否・write/audit 0
func TestService_Update_RejectsDeceasedPetWhenPetIDOmitted(t *testing.T) {
	appointmentID := uint64(77)
	petID := uint64(501)
	detailUpdateCalls := 0
	appointmentUpdateCalls := 0
	remarks := "weight recheck"
	audit := &trimmingAuditRecorder{}

	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return trimmingApptWithPet(id, clinicID, petID), nil
		},
		findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
			return 10, nil
		},
		petLifeStateFn: func(uint64) (sharedkernel.PetLifeState, error) {
			return sharedkernel.PetLifeStateDeceased, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			appointmentUpdateCalls++
			t.Fatal("appointment must not be updated for a deceased appointment pet when pet_id is omitted")
			return nil, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, clinicID, id uint64) (*model.AppointmentTrimmingDetail, error) {
			t.Fatal("trimming detail must not be loaded after deceased pet validation fails")
			return &model.AppointmentTrimmingDetail{ClinicID: clinicID, AppointmentID: id}, nil
		},
		updateFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			detailUpdateCalls++
			t.Fatal("trimming detail must not be updated for a deceased appointment pet when pet_id is omitted")
			return nil
		},
	}
	svc := newDeceasedGuardTestService(reserv, detail, audit)

	got, _, err := svc.Update(context.Background(), 1, appointmentID, &UpdateTrimmingInput{
		Remarks: &remarks,
		// PetID intentionally omitted — finalPetID resolves to locked.PetID
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, got)
	assert.Zero(t, detailUpdateCalls)
	assert.Zero(t, appointmentUpdateCalls)
	assert.Empty(t, audit.entries)
}

// ③ 明示的な pet 差し替え先が死亡 → 拒否・write/audit 0
func TestService_Update_RejectsDeceasedPetReplacement(t *testing.T) {
	appointmentID := uint64(77)
	alivePetID := uint64(501)
	deceasedPetID := uint64(999)
	detailUpdateCalls := 0
	appointmentUpdateCalls := 0
	audit := &trimmingAuditRecorder{}

	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return trimmingApptWithPet(id, clinicID, alivePetID), nil
		},
		findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
			return 10, nil
		},
		petLifeStateFn: func(id uint64) (sharedkernel.PetLifeState, error) {
			if id == deceasedPetID {
				return sharedkernel.PetLifeStateDeceased, nil
			}
			return sharedkernel.PetLifeStateAlive, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			appointmentUpdateCalls++
			t.Fatal("appointment must not be updated when replacement pet is deceased")
			return nil, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			t.Fatal("trimming detail must not be loaded when replacement pet is deceased")
			return nil, nil
		},
		updateFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			detailUpdateCalls++
			t.Fatal("trimming detail must not be updated when replacement pet is deceased")
			return nil
		},
	}
	svc := newDeceasedGuardTestService(reserv, detail, audit)

	got, _, err := svc.Update(context.Background(), 1, appointmentID, &UpdateTrimmingInput{
		PetID: &deceasedPetID,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, got)
	assert.Zero(t, detailUpdateCalls)
	assert.Zero(t, appointmentUpdateCalls)
	assert.Empty(t, audit.entries)
}

// ④ 通常 create のペットが死亡 → 拒否・write/audit 0
func TestService_Create_RejectsDeceasedPet(t *testing.T) {
	petID := uint64(501)
	createCalls := 0
	detailCreateCalls := 0
	audit := &trimmingAuditRecorder{}

	reserv := &mockTrimmingReservationRepository{
		findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
			return 10, nil
		},
		petLifeStateFn: func(uint64) (sharedkernel.PetLifeState, error) {
			return sharedkernel.PetLifeStateDeceased, nil
		},
		createFn: func(_ context.Context, _ *model.Reservation) error {
			createCalls++
			t.Fatal("appointment must not be created for a deceased pet")
			return nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			detailCreateCalls++
			t.Fatal("trimming detail must not be created for a deceased pet")
			return nil
		},
	}
	svc := newDeceasedGuardTestService(reserv, detail, audit)

	got, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		ReservationTypeID: 9,
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(time.Hour),
		PetID:             &petID,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, got)
	assert.Zero(t, createCalls)
	assert.Zero(t, detailCreateCalls)
	assert.Empty(t, audit.entries)
}

// EMR-491 D6: 生死を確定できない（unresolved）/ reader 失敗の pet は生存扱いにせず、
// 業務拒否（InvalidInput）ではなくエラーとして write/audit 0 で拒否する。
func TestService_Create_RejectsUnresolvedOrFailedPetLifeState(t *testing.T) {
	petID := uint64(501)
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "unresolved", err: apperrors.Wrap(sharedkernel.ErrPetLifeStateUnresolved, "pet status is not recorded")},
		{name: "reader failure", err: errors.New("db down")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			createCalls := 0
			detailCreateCalls := 0
			audit := &trimmingAuditRecorder{}

			reserv := &mockTrimmingReservationRepository{
				findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
					return 10, nil
				},
				petLifeStateFn: func(uint64) (sharedkernel.PetLifeState, error) {
					return "", tc.err
				},
				createFn: func(_ context.Context, _ *model.Reservation) error {
					createCalls++
					t.Fatal("appointment must not be created when the pet life state is not confirmed alive")
					return nil
				},
			}
			detail := &mockTrimmingDetailRepository{
				createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
					detailCreateCalls++
					t.Fatal("trimming detail must not be created when the pet life state is not confirmed alive")
					return nil
				},
			}
			svc := newDeceasedGuardTestService(reserv, detail, audit)

			got, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
				ReservationTypeID: 9,
				StartTime:         time.Now(),
				EndTime:           time.Now().Add(time.Hour),
				PetID:             &petID,
			})

			require.Error(t, err)
			assert.ErrorIs(t, err, tc.err)
			assert.False(t, apperrors.IsInvalidInput(err), "unknown must not be reported as a deceased business rejection: %v", err)
			assert.Nil(t, got)
			assert.Zero(t, createCalls)
			assert.Zero(t, detailCreateCalls)
			assert.Empty(t, audit.entries)
		})
	}
}
