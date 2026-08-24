package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockReservationRepository は appointment_service_test.go で定義済み
// mockTransactor は trimming_service_test.go で定義済み

// ---- ReservationAdmin モック ----

type mockReservationAdminRepository struct {
	findByMonthFn               func(ctx context.Context, clinicID uint64, year int, month time.Month) ([]model.Reservation, error)
	findByDayFn                 func(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error)
	createFn                    func(ctx context.Context, r *model.Reservation) error
	softDeleteFn                func(ctx context.Context, clinicID, id uint64) error
	findByCustomerIDFn          func(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error)
	cancelByIDFn                func(ctx context.Context, clinicID, customerID, id uint64) error
	findByIDForNotify           func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	findTimeRangesByDateRangeFn func(ctx context.Context, clinicID uint64, from, to time.Time) ([]model.Reservation, error)
}

func (m *mockReservationAdminRepository) FindAllByMonth(ctx context.Context, clinicID uint64, year int, month time.Month) ([]model.Reservation, error) {
	return m.findByMonthFn(ctx, clinicID, year, month)
}
func (m *mockReservationAdminRepository) FindAllByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error) {
	return m.findByDayFn(ctx, clinicID, date)
}
func (m *mockReservationAdminRepository) FindTimeRangesByDateRange(ctx context.Context, clinicID uint64, from, to time.Time) ([]model.Reservation, error) {
	if m.findTimeRangesByDateRangeFn != nil {
		return m.findTimeRangesByDateRangeFn(ctx, clinicID, from, to)
	}
	return nil, nil
}
func (m *mockReservationAdminRepository) Create(ctx context.Context, r *model.Reservation) error {
	if m.createFn != nil {
		return m.createFn(ctx, r)
	}
	return nil
}
func (m *mockReservationAdminRepository) SoftDelete(ctx context.Context, clinicID, id uint64) error {
	return m.softDeleteFn(ctx, clinicID, id)
}
func (m *mockReservationAdminRepository) FindAllByCustomerID(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error) {
	if m.findByCustomerIDFn != nil {
		return m.findByCustomerIDFn(ctx, clinicID, customerID)
	}
	return nil, nil
}
func (m *mockReservationAdminRepository) CancelByID(ctx context.Context, clinicID, customerID, id uint64) error {
	if m.cancelByIDFn != nil {
		return m.cancelByIDFn(ctx, clinicID, customerID, id)
	}
	return nil
}
func (m *mockReservationAdminRepository) FindByIDForNotify(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if m.findByIDForNotify != nil {
		return m.findByIDForNotify(ctx, clinicID, id)
	}
	return nil, nil
}

// ---- Tests ----

func TestReservationAdminService_ListByMonth(t *testing.T) {
	items := []model.Reservation{{ID: 1}}

	tests := []struct {
		name      string
		yearMonth string
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "lists by month successfully",
			yearMonth: "2026-04",
			wantErr:   false,
		},
		{
			name:      "returns error for invalid format",
			yearMonth: "not-a-date",
			wantErr:   true,
		},
		{
			name:      "propagates repository error",
			yearMonth: "2026-04",
			repoErr:   errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationAdminRepository{
				findByMonthFn: func(_ context.Context, _ uint64, _ int, _ time.Month) ([]model.Reservation, error) {
					return items, tt.repoErr
				},
			}
			svc := NewReservationAdminServiceWithAvailabilityAndType(repo, &mockReservationRepository{}, nil, &mockTransactor{}, nil, nil)
			result, err := svc.ListByMonth(context.Background(), 1, tt.yearMonth)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, items, result)
			}
		})
	}
}

func TestReservationAdminService_ListByDay(t *testing.T) {
	items := []model.Reservation{{ID: 2}}
	day := time.Date(2026, 4, 19, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{
			name:    "lists by day successfully",
			wantErr: false,
		},
		{
			name:    "propagates repository error",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationAdminRepository{
				findByDayFn: func(_ context.Context, _ uint64, _ time.Time) ([]model.Reservation, error) {
					return items, tt.repoErr
				},
			}
			svc := NewReservationAdminServiceWithAvailabilityAndType(repo, &mockReservationRepository{}, nil, &mockTransactor{}, nil, nil)
			result, err := svc.ListByDay(context.Background(), 1, day)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, items, result)
			}
		})
	}
}

func TestReservationAdminService_Create(t *testing.T) {
	now := time.Now()
	start := now
	end := now.Add(30 * time.Minute)

	validInput := &CreateReservationAdminInput{
		StartTime:         start,
		EndTime:           end,
		ReservationTypeID: 1,
	}

	tests := []struct {
		name          string
		input         *CreateReservationAdminInput
		txErr         error
		doctorCount   int64
		conflictCount int64
		createErr     error
		wantErr       bool
	}{
		{
			name:        "creates appointment successfully",
			input:       validInput,
			doctorCount: 1,
			wantErr:     false,
		},
		{
			name: "returns error when end time before start time",
			input: &CreateReservationAdminInput{
				StartTime: end,
				EndTime:   start,
			},
			wantErr: true,
		},
		{
			name:        "returns error when no doctors on duty",
			input:       validInput,
			doctorCount: 0,
			wantErr:     true,
		},
		{
			name:          "returns conflict when slot is full",
			input:         validInput,
			doctorCount:   1,
			conflictCount: 1,
			wantErr:       true,
		},
		{
			name:        "propagates create error",
			input:       validInput,
			doctorCount: 1,
			createErr:   errors.New("db error"),
			wantErr:     true,
		},
		{
			name:    "propagates tx error",
			input:   validInput,
			txErr:   errors.New("tx error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resRepo := &mockReservationRepository{
				countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
					return tt.doctorCount, nil
				},
				countConflictsFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
					return tt.conflictCount, nil
				},
				createFn: func(_ context.Context, _ *model.Reservation) error {
					return tt.createErr
				},
			}
			repo := &mockReservationAdminRepository{}
			tx := &mockTransactor{withTxErr: tt.txErr}
			svc := NewReservationAdminServiceWithClinicHolidays(repo, resRepo, nil, tx, nil, nil, nil, openDayHolidayFinder())
			result, err := svc.Create(context.Background(), 1, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestReservationAdminService_Create_RejectsDeceasedPet(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	deceasedAt := start.Add(-24 * time.Hour)
	petID := uint64(5)
	ownerID := uint64(2)
	createCalled := false
	resRepo := &mockReservationRepository{
		findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
			return ownerID, nil
		},
		findPetByIDInClinicFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, OwnerID: ownerID, DeceasedAt: &deceasedAt, Status: model.PetStatusDeceased}, nil
		},
		createFn: func(_ context.Context, _ *model.Reservation) error {
			createCalled = true
			return nil
		},
	}
	svc := NewReservationAdminServiceWithClinicHolidays(
		&mockReservationAdminRepository{}, resRepo, nil, &mockTransactor{}, nil, nil, nil, openDayHolidayFinder(),
	)
	result, err := svc.Create(context.Background(), 1, &CreateReservationAdminInput{
		StartTime:         start,
		EndTime:           start.Add(time.Hour),
		OwnerID:           &ownerID,
		PetID:             &petID,
		ReservationTypeID: 1,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "expected InvalidInput, got: %v", err)
	assert.Nil(t, result)
	assert.False(t, createCalled)
}

func TestReservationAdminService_Create_RejectsFullReservationTypeCapacity(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	maxConcurrent := 2
	createCalled := false
	resRepo := &mockReservationRepository{
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
			return 3, nil
		},
		countConflictsFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
			return 0, nil
		},
		countByTypeAndStartTimeFn: func(_ context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(9), reservationTypeID)
			assert.Equal(t, start, startTime)
			assert.Nil(t, excludeID)
			return 2, nil
		},
		createFn: func(_ context.Context, _ *model.Reservation) error {
			createCalled = true
			return nil
		},
	}
	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{ID: id, ClinicID: clinicID, MaxConcurrent: &maxConcurrent}, nil
		},
	}
	svc := NewReservationAdminServiceWithClinicHolidays(
		&mockReservationAdminRepository{},
		resRepo,
		typeRepo,
		&mockTransactor{},
		nil,
		nil,
		nil,
		openDayHolidayFinder(),
	)

	result, err := svc.Create(context.Background(), 1, &CreateReservationAdminInput{
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 9,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "expected conflict but got: %v", err)
	assert.Nil(t, result)
	assert.False(t, createCalled)
}

func TestReservationAdminService_Create_RejectsExcludedStaff(t *testing.T) {
	now := time.Now()
	doctorID := uint64(10)
	resRepo := &mockReservationRepository{
		createFn: func(_ context.Context, _ *model.Reservation) error {
			t.Fatal("reservation must not be created when staff cannot handle reservation type")
			return nil
		},
	}
	staffRepo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, doctorID, id)
			return &model.Staff{ID: id, IsActive: true}, nil
		},
		supportsReservationTypeFn: func(_ context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, doctorID, staffID)
			assert.Equal(t, uint64(5), reservationTypeID)
			return false, nil
		},
	}
	svc := NewReservationAdminServiceWithClinicHolidays(&mockReservationAdminRepository{}, resRepo, nil, &mockTransactor{}, staffRepo, nil, nil, openDayHolidayFinder())

	result, err := svc.Create(context.Background(), 1, &CreateReservationAdminInput{
		StartTime:         now,
		EndTime:           now.Add(30 * time.Minute),
		ReservationTypeID: 5,
		DoctorID:          &doctorID,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "expected ErrInvalidInput but got: %v", err)
}

func TestReservationAdminService_Create_RejectsUnavailableTime(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 30, 0, 0, config.JST)
	end := start.Add(30 * time.Minute)
	specificDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	resRepo := &mockReservationRepository{
		createFn: func(_ context.Context, _ *model.Reservation) error {
			t.Fatal("reservation must not be created during unavailable time")
			return nil
		},
	}
	unavailableRepo := &mockUnavailableTimeRepository{
		findAllFn: func(_ context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(5), reservationTypeID)
			return []model.ReservationTypeUnavailableTime{
				{
					ReservationTypeID: 5,
					UnavailableType:   model.UnavailableTypeSpecific,
					SpecificDate:      &specificDate,
					StartTime:         "10:00",
					EndTime:           "11:00",
				},
			}, nil
		},
	}
	svc := NewReservationAdminServiceWithClinicHolidays(&mockReservationAdminRepository{}, resRepo, nil, &mockTransactor{}, nil, unavailableRepo, nil, openDayHolidayFinder())

	result, err := svc.Create(context.Background(), 1, &CreateReservationAdminInput{
		StartTime:         start,
		EndTime:           end,
		ReservationTypeID: 5,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "expected ErrInvalidInput but got: %v", err)
}

func TestReservationAdminService_Create_RejectsClinicHoliday(t *testing.T) {
	start := time.Date(2026, 6, 2, 10, 0, 0, 0, config.JST)
	createCalled := false
	resRepo := &mockReservationRepository{
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
			return 1, nil
		},
		createFn: func(_ context.Context, _ *model.Reservation) error {
			createCalled = true
			return nil
		},
	}
	holidayFinder := &mockClinicHolidayFinder{
		findByDateFn: func(_ context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, "2026-06-02", date.In(config.JST).Format(time.DateOnly))
			return &model.ClinicHoliday{ID: 1, ClinicID: clinicID, Date: date, Reason: "臨時休診"}, nil
		},
	}
	svc := NewReservationAdminServiceWithMedicalRecord(
		&mockReservationAdminRepository{},
		resRepo,
		nil,
		&mockTransactor{},
		nil,
		nil,
		nil,
		nil,
		holidayFinder,
	)

	result, err := svc.Create(context.Background(), 1, &CreateReservationAdminInput{
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 5,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input but got: %v", err)
	assert.Contains(t, err.Error(), "休診日")
	assert.False(t, createCalled, "clinic holiday must not persist an admin reservation")
}

func TestReservationAdminService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		findErr   error
		deleteErr error
		wantErr   bool
	}{
		{
			name:    "deletes appointment successfully",
			wantErr: false,
		},
		{
			name:      "propagates soft delete error",
			deleteErr: errors.New("db error"),
			wantErr:   true,
		},
		{
			name:    "returns error when reservation appointment is not found",
			findErr: apperrors.WrapNotFound("reservation", "1"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resRepo := &mockReservationRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
					return &model.Reservation{ID: id}, tt.findErr
				},
			}
			repo := &mockReservationAdminRepository{
				softDeleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := NewReservationAdminServiceWithAvailabilityAndType(repo, resRepo, nil, &mockTransactor{}, nil, nil)
			err := svc.Delete(context.Background(), 1, 1)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReservationAdminService_Delete_CleansUpDraftMedicalRecord(t *testing.T) {
	const (
		clinicID      = uint64(3)
		reservationID = uint64(77)
	)

	t.Run("soft delete 成功後に同じ clinic と reservation で cleanup を呼ぶ", func(t *testing.T) {
		var cleanupCalls int
		medicalRecord := &mockMedicalRecordService{
			deleteDraftFromReservationFn: func(ctx context.Context, gotClinicID, gotReservationID uint64) {
				cleanupCalls++
				assert.NotNil(t, ctx)
				assert.Equal(t, clinicID, gotClinicID)
				assert.Equal(t, reservationID, gotReservationID)
			},
		}
		resRepo := &mockReservationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: id}, nil
			},
		}
		repo := &mockReservationAdminRepository{
			softDeleteFn: func(_ context.Context, _, _ uint64) error { return nil },
		}
		svc := NewReservationAdminServiceWithMedicalRecord(
			repo, resRepo, nil, &mockTransactor{}, nil, nil, nil, medicalRecord, nil,
		)

		err := svc.Delete(context.Background(), clinicID, reservationID)

		require.NoError(t, err)
		assert.Equal(t, 1, cleanupCalls)
	})

	t.Run("soft delete 失敗時は cleanup を呼ばない", func(t *testing.T) {
		var cleanupCalls int
		medicalRecord := &mockMedicalRecordService{
			deleteDraftFromReservationFn: func(_ context.Context, _, _ uint64) {
				cleanupCalls++
			},
		}
		resRepo := &mockReservationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: id}, nil
			},
		}
		repo := &mockReservationAdminRepository{
			softDeleteFn: func(_ context.Context, _, _ uint64) error {
				return errors.New("delete failed")
			},
		}
		svc := NewReservationAdminServiceWithMedicalRecord(
			repo, resRepo, nil, &mockTransactor{}, nil, nil, nil, medicalRecord, nil,
		)

		err := svc.Delete(context.Background(), clinicID, reservationID)

		require.Error(t, err)
		assert.Zero(t, cleanupCalls)
	})
}

// ---- NewReservationAdminServiceWithAvailabilityAndType ----

func TestNewReservationAdminServiceWithAvailabilityAndType(t *testing.T) {
	t.Run("wires optional available slot repository when provided", func(t *testing.T) {
		slotRepo := &mockAvailableSlotRepository{}

		svc := NewReservationAdminServiceWithAvailabilityAndType(
			&mockReservationAdminRepository{},
			&mockReservationRepository{},
			nil,
			&mockTransactor{},
			nil,
			nil,
			slotRepo,
		)

		impl, ok := svc.(*reservationAdminService)
		require.True(t, ok)
		assert.Same(t, slotRepo, impl.availableSlotRepo)
	})

	t.Run("leaves available slot repository nil when not provided", func(t *testing.T) {
		svc := NewReservationAdminServiceWithAvailabilityAndType(
			&mockReservationAdminRepository{},
			&mockReservationRepository{},
			nil,
			&mockTransactor{},
			nil,
			nil,
		)

		impl, ok := svc.(*reservationAdminService)
		require.True(t, ok)
		assert.Nil(t, impl.availableSlotRepo)
	})
}
