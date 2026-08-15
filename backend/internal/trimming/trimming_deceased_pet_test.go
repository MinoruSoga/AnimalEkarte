package trimming

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TASK-030: 死亡ペット検証が input.PetID 有無に依存せず finalPetID を見ることを経路別に固定する。

func deceasedPetAt() time.Time {
	return time.Now().Add(-24 * time.Hour)
}

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
) TrimmingService {
	return withTrimmingTestActor(newTrimmingAuditTestService(reserv, detail, &mockTransactor{}, audit))
}

// ① detail 作成: input.PetID == nil かつ予約由来ペットが死亡 → 拒否・write/audit 0
func TestTrimmingService_CreateExistingDetail_RejectsDeceasedPetWhenPetIDOmitted(t *testing.T) {
	appointmentID := uint64(77)
	petID := uint64(501)
	deceasedAt := deceasedPetAt()
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
		findPetByIDFn: func(id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, DeceasedAt: &deceasedAt}, nil
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
func TestTrimmingService_Update_RejectsDeceasedPetWhenPetIDOmitted(t *testing.T) {
	appointmentID := uint64(77)
	petID := uint64(501)
	deceasedAt := deceasedPetAt()
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
		findPetByIDFn: func(id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, DeceasedAt: &deceasedAt}, nil
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

	got, err := svc.Update(context.Background(), 1, appointmentID, &UpdateTrimmingInput{
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
func TestTrimmingService_Update_RejectsDeceasedPetReplacement(t *testing.T) {
	appointmentID := uint64(77)
	alivePetID := uint64(501)
	deceasedPetID := uint64(999)
	deceasedAt := deceasedPetAt()
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
		findPetByIDFn: func(id uint64) (*model.Pet, error) {
			if id == deceasedPetID {
				return &model.Pet{ID: id, DeceasedAt: &deceasedAt}, nil
			}
			return &model.Pet{ID: id, Status: model.PetStatusAlive}, nil
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

	got, err := svc.Update(context.Background(), 1, appointmentID, &UpdateTrimmingInput{
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
func TestTrimmingService_Create_RejectsDeceasedPet(t *testing.T) {
	petID := uint64(501)
	deceasedAt := deceasedPetAt()
	createCalls := 0
	detailCreateCalls := 0
	audit := &trimmingAuditRecorder{}

	reserv := &mockTrimmingReservationRepository{
		findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
			return 10, nil
		},
		findPetByIDFn: func(id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, DeceasedAt: &deceasedAt}, nil
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
