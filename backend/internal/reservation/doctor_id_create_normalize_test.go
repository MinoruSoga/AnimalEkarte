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

func TestNormalizeCreateDoctorID(t *testing.T) {
	t.Parallel()
	zero := uint64(0)
	valid := uint64(42)
	assert.Nil(t, normalizeCreateDoctorID(nil))
	assert.Nil(t, normalizeCreateDoctorID(&zero))
	got := normalizeCreateDoctorID(&valid)
	require.NotNil(t, got)
	assert.Equal(t, uint64(42), *got)
	// caller input must not be mutated
	assert.Equal(t, uint64(42), valid)
	assert.Equal(t, uint64(0), zero)
}

func TestReservationService_Create_NormalizesUnsetDoctorID(t *testing.T) {
	now := time.Now()
	zero := uint64(0)
	valid := uint64(7)

	cases := []struct {
		name       string
		doctorID   *uint64
		wantDoctor *uint64
		staffOK    bool
		wantErr    bool
	}{
		{name: "omit/nil -> NULL", doctorID: nil, wantDoctor: nil},
		{name: "zero -> NULL", doctorID: &zero, wantDoctor: nil},
		{name: "valid positive kept", doctorID: &valid, wantDoctor: &valid, staffOK: true},
		{name: "incapable positive fails before write", doctorID: &valid, staffOK: false, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var persisted *model.Reservation
			createCalled := false
			repo := &mockReservationRepository{
				createFn: func(_ context.Context, r *model.Reservation) error {
					createCalled = true
					cp := *r
					persisted = &cp
					r.ID = 99
					return nil
				},
			}
			staffRepo := &mockReservationStaffRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, valid, id)
					return &model.Staff{ID: id, IsActive: true}, nil
				},
				supportsReservationTypeFn: func(_ context.Context, _, _, _ uint64) (bool, error) {
					return tc.staffOK, nil
				},
			}
			svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, staffRepo, nil)
			inputDoctor := tc.doctorID
			var inputCopy *uint64
			if tc.doctorID != nil {
				v := *tc.doctorID
				inputCopy = &v
			}
			result, err := svc.Create(context.Background(), &CreateManualReservationInput{
				ClinicID:          1,
				StartTime:         now,
				EndTime:           now.Add(time.Hour),
				ReservationTypeID: 1,
				DoctorID:          inputCopy,
				Status:            model.ReservationStatusInConsultation,
			})
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err), "got: %v", err)
				assert.False(t, createCalled)
				assert.Nil(t, result)
				return
			}
			require.NoError(t, err)
			require.True(t, createCalled)
			require.NotNil(t, persisted)
			if tc.wantDoctor == nil {
				assert.Nil(t, persisted.DoctorID)
			} else {
				require.NotNil(t, persisted.DoctorID)
				assert.Equal(t, *tc.wantDoctor, *persisted.DoctorID)
			}
			// input not destructively mutated
			if inputDoctor == nil {
				assert.Nil(t, inputCopy)
			} else {
				require.NotNil(t, inputCopy)
				assert.Equal(t, *inputDoctor, *inputCopy)
			}
		})
	}
}

func TestReservationService_CreateBatch_NormalizesZeroDoctorID(t *testing.T) {
	start := time.Date(2027, 7, 1, 10, 0, 0, 0, time.UTC)
	zero := uint64(0)
	var persisted []model.Reservation
	repo := &mockReservationRepository{
		createFn: func(_ context.Context, r *model.Reservation) error {
			persisted = append(persisted, *r)
			return nil
		},
		findPetOwnerInClinicFn: func(_ context.Context, _ uint64, _ uint64) (uint64, error) { return 1, nil },
		findPetByIDInClinicFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, OwnerID: 1, Status: model.PetStatusAlive}, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)
	input := &CreateManualReservationInput{
		ClinicID: 1, StartTime: start, EndTime: start.Add(time.Hour), ReservationTypeID: 1,
		DoctorID: &zero, Status: model.ReservationStatusInConsultation, Source: model.ReservationSourceManual,
	}
	got, err := svc.CreateBatch(context.Background(), input, []ReservationBatchPet{{OwnerID: 1, PetID: 10}, {OwnerID: 1, PetID: 11}})
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Len(t, persisted, 2)
	for _, r := range persisted {
		assert.Nil(t, r.DoctorID)
	}
	assert.Equal(t, uint64(0), *input.DoctorID, "caller input must remain unchanged")
}

func TestReservationAdminService_Create_NormalizesZeroDoctorID(t *testing.T) {
	now := time.Now()
	zero := uint64(0)
	var persisted *model.Reservation
	resRepo := &mockReservationRepository{
		createFn: func(_ context.Context, r *model.Reservation) error {
			cp := *r
			persisted = &cp
			return nil
		},
	}
	svc := NewReservationAdminServiceWithClinicHolidays(
		&mockReservationAdminRepository{}, resRepo, nil, &mockTransactor{}, nil, nil, nil, openDayHolidayFinder(),
	)
	input := &CreateReservationAdminInput{
		StartTime: now, EndTime: now.Add(time.Hour), ReservationTypeID: 1, DoctorID: &zero,
	}
	result, err := svc.Create(context.Background(), 1, input)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, persisted)
	assert.Nil(t, persisted.DoctorID)
	assert.Equal(t, uint64(0), *input.DoctorID)
}

func TestResolveUpdateParams_DoctorOmitKeepsZeroClears(t *testing.T) {
	t.Parallel()
	currentDoctor := uint64(5)
	current := &model.Reservation{
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		DoctorID:  &currentDoctor,
	}
	_, _, kept := resolveUpdateParams(current, &UpdateReservationInput{})
	require.NotNil(t, kept)
	assert.Equal(t, currentDoctor, *kept)

	zero := uint64(0)
	_, _, cleared := resolveUpdateParams(current, &UpdateReservationInput{DoctorID: &zero})
	assert.Nil(t, cleared)

	next := uint64(9)
	_, _, replaced := resolveUpdateParams(current, &UpdateReservationInput{DoctorID: &next})
	require.NotNil(t, replaced)
	assert.Equal(t, next, *replaced)
}

func TestBuildReservationUpdate_ZeroDoctorClearsToNull(t *testing.T) {
	t.Parallel()
	zero := uint64(0)
	fields := buildReservationUpdate(&UpdateReservationInput{DoctorID: &zero})
	require.Contains(t, fields, "doctor_id")
	assert.Nil(t, fields["doctor_id"])

	omitFields := buildReservationUpdate(&UpdateReservationInput{})
	assert.NotContains(t, omitFields, "doctor_id")
}
