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

// mockReservationRepository は ReservationRepository のテスト用モック実装
type mockReservationRepository struct {
	findAllFn                          func(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error)
	findByIDFn                         func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	lockAndFindByIDFn                  func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	createFn                           func(ctx context.Context, reservation *model.Reservation) error
	updateFieldsFn                     func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error)
	deleteFn                           func(ctx context.Context, clinicID, id uint64) error
	countMedicalRecordsByReservationID func(ctx context.Context, reservationID uint64) (int64, error)
	hasDoctorConflictFn                func(ctx context.Context, clinicID, doctorID uint64, start, end time.Time, excludeID *uint64) (bool, error)
	countOnDutyDoctorsFn               func(ctx context.Context, clinicID uint64, date time.Time) (int64, error)
	countConflictsFn                   func(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error)
	countByTypeAndStartTimeFn          func(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error)
	assertOwnerInClinicFn              func(ctx context.Context, clinicID, ownerID uint64) error
	findPetOwnerInClinicFn             func(ctx context.Context, clinicID, petID uint64) (uint64, error)
	findPetByIDInClinicFn              func(ctx context.Context, clinicID, petID uint64) (*model.Pet, error)
	assertLineCustomerInClinicFn       func(ctx context.Context, clinicID, lineCustomerID uint64) error
	acquireBookingLockFn               func(ctx context.Context, clinicID uint64) error
}

func (m *mockReservationRepository) FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error) {
	return m.findAllFn(ctx, clinicIDs, page, limit, date, startDate, endDate, status, source, petID, ownerID)
}

func (m *mockReservationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockReservationRepository) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Reservation, error) {
	return nil, nil
}

func (m *mockReservationRepository) Create(ctx context.Context, reservation *model.Reservation) error {
	return m.createFn(ctx, reservation)
}

func (m *mockReservationRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockReservationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockReservationRepository) ExistsByReservationTypeID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationRepository) ExistsByStaffID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationRepository) CountMedicalRecordsByReservationID(ctx context.Context, _, reservationID uint64) (int64, error) {
	if m.countMedicalRecordsByReservationID != nil {
		return m.countMedicalRecordsByReservationID(ctx, reservationID)
	}
	return 0, nil
}

func (m *mockReservationRepository) AcquireBookingLock(ctx context.Context, clinicID uint64) error {
	if m.acquireBookingLockFn != nil {
		return m.acquireBookingLockFn(ctx, clinicID)
	}
	return nil
}

func (m *mockReservationRepository) LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if m.lockAndFindByIDFn != nil {
		return m.lockAndFindByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockReservationRepository) HasDoctorConflict(ctx context.Context, clinicID, doctorID uint64, start, end time.Time, excludeID *uint64) (bool, error) {
	if m.hasDoctorConflictFn != nil {
		return m.hasDoctorConflictFn(ctx, clinicID, doctorID, start, end, excludeID)
	}
	return false, nil
}

func (m *mockReservationRepository) CountOnDutyDoctors(ctx context.Context, clinicID uint64, date time.Time) (int64, error) {
	if m.countOnDutyDoctorsFn != nil {
		return m.countOnDutyDoctorsFn(ctx, clinicID, date)
	}
	return 1, nil
}

func (m *mockReservationRepository) CountConflicts(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error) {
	if m.countConflictsFn != nil {
		return m.countConflictsFn(ctx, clinicID, start, end, excludeID)
	}
	return 0, nil
}

func (m *mockReservationRepository) CountByTypeAndStartTime(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error) {
	if m.countByTypeAndStartTimeFn != nil {
		return m.countByTypeAndStartTimeFn(ctx, clinicID, reservationTypeID, startTime, excludeID)
	}
	return 0, nil
}

func (m *mockReservationRepository) CountByCustomerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockReservationRepository) CountByDateAndSource(_ context.Context, _ uint64, _ time.Time, _ model.ReservationSource) (int64, error) {
	return 0, nil
}

func (m *mockReservationRepository) FindAllByCategory(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}

func (m *mockReservationRepository) AssertOwnerInClinic(ctx context.Context, clinicID, ownerID uint64) error {
	if m.assertOwnerInClinicFn != nil {
		return m.assertOwnerInClinicFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockReservationRepository) FindPetOwnerInClinic(ctx context.Context, clinicID, petID uint64) (uint64, error) {
	if m.findPetOwnerInClinicFn != nil {
		return m.findPetOwnerInClinicFn(ctx, clinicID, petID)
	}
	return 0, nil
}

func (m *mockReservationRepository) FindPetByIDInClinic(ctx context.Context, clinicID, petID uint64) (*model.Pet, error) {
	if m.findPetByIDInClinicFn != nil {
		return m.findPetByIDInClinicFn(ctx, clinicID, petID)
	}
	return &model.Pet{ID: petID, Status: model.PetStatusAlive}, nil
}

func (m *mockReservationRepository) AssertLineCustomerInClinic(ctx context.Context, clinicID, lineCustomerID uint64) error {
	if m.assertLineCustomerInClinicFn != nil {
		return m.assertLineCustomerInClinicFn(ctx, clinicID, lineCustomerID)
	}
	return nil
}

func (m *mockReservationRepository) FindNoShowCandidates(_ context.Context, _ uint64) ([]model.Reservation, error) {
	return nil, nil
}

func (m *mockReservationRepository) CompleteForAccounting(_ context.Context, _ uint64, _, _, _ *uint64, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockReservationRepository) AssertMedicalRecordDoctorInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockReservationRepository) BackfillForMedicalRecord(_ context.Context, _, _ uint64, _, _, _ *uint64) error {
	return nil
}

func (m *mockReservationRepository) MarkNoShow(_ context.Context, _, _ uint64) (NoShowTransition, error) {
	return NoShowTransition{}, nil
}

func (m *mockReservationRepository) CreateForTrimming(_ context.Context, clinicID uint64, input CreateTrimmingReservationInput) (*model.Reservation, error) {
	return &model.Reservation{ClinicID: clinicID, ReservationTypeID: input.ReservationTypeID}, nil
}

func (m *mockReservationRepository) UpdateForTrimming(_ context.Context, clinicID, id uint64, _ UpdateTrimmingReservationInput) (*model.Reservation, error) {
	return &model.Reservation{ID: id, ClinicID: clinicID}, nil
}

func (m *mockReservationRepository) DeleteForTrimming(_ context.Context, _, _ uint64) error {
	return nil
}

func ptrTime(t time.Time) *time.Time { return &t }

func TestReservationService_List(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name             string
		clinicID         uint64
		page             int
		limit            int
		date             *time.Time
		status           *string
		petID            *uint64
		ownerID          *uint64
		repoReservations []model.Reservation
		repoTotal        int64
		repoErr          error
		wantLen          int
		wantTotal        int64
		wantErr          bool
	}{
		{
			name:     "returns all reservations for clinic",
			clinicID: 1,
			page:     1,
			limit:    20,
			date:     nil,
			status:   nil,
			petID:    nil,
			ownerID:  nil,
			repoReservations: []model.Reservation{
				{ID: 1, ClinicID: 1, StartTime: now, EndTime: now.Add(time.Hour), Status: model.ReservationStatusPending},
				{ID: 2, ClinicID: 1, StartTime: now, EndTime: now.Add(time.Hour), Status: model.ReservationStatusConfirmed},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:     "filters by date",
			clinicID: 1,
			page:     1,
			limit:    20,
			date:     ptrTime(now),
			status:   nil,
			petID:    nil,
			ownerID:  nil,
			repoReservations: []model.Reservation{
				{ID: 1, ClinicID: 1, StartTime: now, EndTime: now.Add(time.Hour), Status: model.ReservationStatusPending},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by status",
			clinicID: 1,
			page:     1,
			limit:    20,
			date:     nil,
			status:   ptrString("confirmed"),
			petID:    nil,
			ownerID:  nil,
			repoReservations: []model.Reservation{
				{ID: 2, ClinicID: 1, StartTime: now, EndTime: now.Add(time.Hour), Status: model.ReservationStatusConfirmed},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by pet_id",
			clinicID: 1,
			page:     1,
			limit:    20,
			date:     nil,
			status:   nil,
			petID:    ptrUint64(10),
			ownerID:  nil,
			repoReservations: []model.Reservation{
				{ID: 1, ClinicID: 1, StartTime: now, EndTime: now.Add(time.Hour), Status: model.ReservationStatusPending},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by owner_id",
			clinicID: 1,
			page:     1,
			limit:    20,
			date:     nil,
			status:   nil,
			petID:    nil,
			ownerID:  ptrUint64(5),
			repoReservations: []model.Reservation{
				{ID: 2, ClinicID: 1, StartTime: now, EndTime: now.Add(time.Hour), Status: model.ReservationStatusConfirmed},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:             "returns empty list when no reservations exist",
			clinicID:         1,
			page:             1,
			limit:            20,
			date:             nil,
			status:           nil,
			petID:            nil,
			ownerID:          nil,
			repoReservations: []model.Reservation{},
			repoTotal:        0,
			repoErr:          nil,
			wantLen:          0,
			wantTotal:        0,
			wantErr:          false,
		},
		{
			name:             "propagates repository error",
			clinicID:         1,
			page:             1,
			limit:            20,
			date:             nil,
			status:           nil,
			petID:            nil,
			ownerID:          nil,
			repoReservations: nil,
			repoTotal:        0,
			repoErr:          errors.New("db connection error"),
			wantLen:          0,
			wantTotal:        0,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationRepository{
				findAllFn: func(_ context.Context, _ []uint64, _, _ int, _, _, _ *time.Time, _ *string, _ *string, _ *uint64, _ *uint64) ([]model.Reservation, int64, error) {
					return tt.repoReservations, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)

			reservations, total, err := svc.List(context.Background(), []uint64{tt.clinicID}, tt.page, tt.limit, tt.date, nil, nil, tt.status, nil, tt.petID, tt.ownerID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, reservations, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestReservationService_GetByID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name            string
		clinicID        uint64
		id              uint64
		repoReservation *model.Reservation
		repoErr         error
		wantReservation *model.Reservation
		wantErr         error
	}{
		{
			name:            "returns reservation when found",
			clinicID:        1,
			id:              10,
			repoReservation: &model.Reservation{ID: 10, ClinicID: 1, StartTime: now, EndTime: now.Add(time.Hour), Status: model.ReservationStatusPending},
			repoErr:         nil,
			wantReservation: &model.Reservation{ID: 10, ClinicID: 1, StartTime: now, EndTime: now.Add(time.Hour), Status: model.ReservationStatusPending},
			wantErr:         nil,
		},
		{
			name:            "returns not found error when reservation does not exist",
			clinicID:        1,
			id:              999,
			repoReservation: nil,
			repoErr:         apperrors.WrapNotFound("reservation", "999"),
			wantReservation: nil,
			wantErr:         apperrors.ErrNotFound,
		},
		{
			name:            "returns error on repository failure",
			clinicID:        1,
			id:              10,
			repoReservation: nil,
			repoErr:         errors.New("db error"),
			wantReservation: nil,
			wantErr:         errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					return tt.repoReservation, tt.repoErr
				},
			}
			svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)

			reservation, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, apperrors.ErrNotFound) {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantReservation, reservation)
			}
		})
	}
}

func TestReservationService_GetByID_NotFound(t *testing.T) {
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
			return nil, apperrors.WrapNotFound("reservation", "999")
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)

	reservation, err := svc.GetByID(context.Background(), 1, 999)

	assert.Nil(t, reservation)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestReservationService_Create(t *testing.T) {
	now := time.Now()

	// バリデーション（end_time チェック）はトランザクション前に実行されるためモックテスト可能。
	// 競合チェック（SELECT FOR UPDATE + トランザクション）は統合テストで担保する。
	tests := []struct {
		name             string
		input            *CreateManualReservationInput
		wantErr          bool
		wantInvalidInput bool
	}{
		{
			// BUG-034: end_time == start_time は 400 Bad Request
			name: "returns invalid input when end_time equals start_time",
			input: &CreateManualReservationInput{
				ClinicID:          1,
				StartTime:         now,
				EndTime:           now,
				ReservationTypeID: 1,
			},
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			// BUG-034: end_time < start_time は 400 Bad Request
			name: "returns invalid input when end_time is before start_time",
			input: &CreateManualReservationInput{
				ClinicID:          1,
				StartTime:         now,
				EndTime:           now.Add(-time.Minute),
				ReservationTypeID: 1,
			},
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			name: "returns invalid input when reservation_route is invalid",
			input: &CreateManualReservationInput{
				ClinicID:          1,
				StartTime:         now,
				EndTime:           now.Add(time.Hour),
				ReservationTypeID: 1,
				ReservationRoute:  ptrString("fax"),
			},
			wantErr:          true,
			wantInvalidInput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationRepository{}
			svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)

			_, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantInvalidInput {
					assert.True(t, apperrors.IsInvalidInput(err), "expected ErrInvalidInput but got: %v", err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReservationService_Create_RejectsFullReservationTypeCapacity(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	maxConcurrent := 2
	createCalled := false
	repo := &mockReservationRepository{
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
	svc := NewReservationServiceWithClinicHolidays(repo, typeRepo, &mockTransactor{}, nil, nil, nil, &mockLineReservationSettingFinder{}, &mockClinicHolidayFinder{})

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 9,
		Status:            model.ReservationStatusPending,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "expected conflict but got: %v", err)
	assert.Nil(t, result)
	assert.False(t, createCalled)
}

func TestReservationService_Create_RejectsIncapableStaff(t *testing.T) {
	now := time.Now()
	doctorID := uint64(10)
	transactionActive := false
	transactionCalls := 0
	capabilityChecked := false
	repo := &mockReservationRepository{
		createFn: func(_ context.Context, _ *model.Reservation) error {
			t.Fatal("reservation must not be created when staff cannot handle reservation type")
			return nil
		},
	}
	staffRepo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
			assert.True(t, transactionActive, "staff must be validated inside the write transaction")
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, doctorID, id)
			return &model.Staff{ID: id, IsActive: true}, nil
		},
		supportsReservationTypeFn: func(_ context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
			assert.True(t, transactionActive, "staff capability must be validated inside the write transaction")
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, doctorID, staffID)
			assert.Equal(t, uint64(5), reservationTypeID)
			capabilityChecked = true
			return false, nil
		},
	}
	tx := &mockTransactor{
		withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			transactionCalls++
			transactionActive = true
			defer func() {
				transactionActive = false
			}()
			return fn(ctx)
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, tx, staffRepo, nil)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         now,
		EndTime:           now.Add(30 * time.Minute),
		ReservationTypeID: 5,
		DoctorID:          &doctorID,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "expected ErrInvalidInput but got: %v", err)
	assert.Equal(t, 1, transactionCalls)
	assert.True(t, capabilityChecked)
}

func TestReservationService_Create_FailsClosedWithoutTransactor(t *testing.T) {
	doctorID := uint64(10)
	tests := []struct {
		name     string
		doctorID *uint64
	}{
		{
			name:     "doctor assigned",
			doctorID: &doctorID,
		},
		{
			name:     "doctor unassigned",
			doctorID: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createCalled := false
			repo := &mockReservationRepository{
				createFn: func(_ context.Context, _ *model.Reservation) error {
					createCalled = true
					return nil
				},
			}
			svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)
			now := time.Now()

			result, err := svc.Create(context.Background(), &CreateManualReservationInput{
				ClinicID:          1,
				StartTime:         now,
				EndTime:           now.Add(30 * time.Minute),
				ReservationTypeID: 5,
				DoctorID:          tt.doctorID,
			})

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.False(t, createCalled)
			var appErr *apperrors.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, "INTERNAL", appErr.Code)
		})
	}
}

func TestReservationService_Create_RejectsUnavailableTime(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 30, 0, 0, config.JST)
	end := start.Add(30 * time.Minute)
	specificDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	repo := &mockReservationRepository{
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
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, unavailableRepo)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           end,
		ReservationTypeID: 5,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "expected ErrInvalidInput but got: %v", err)
}

type mockLineReservationSettingFinder struct {
	findByClinicIDFn func(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
}

func (m *mockLineReservationSettingFinder) FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return &model.LineReservationSetting{
		ClosedWeekdays:        []byte("[]"),
		ClosedDates:           []byte("[]"),
		NationalHolidayClosed: false,
	}, nil
}

func closedDayCreateRepo(t *testing.T) *mockReservationRepository {
	t.Helper()
	return &mockReservationRepository{
		createFn: func(_ context.Context, appt *model.Reservation) error {
			appt.ID = 10
			return nil
		},
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
			return 1, nil
		},
		countConflictsFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
			return 0, nil
		},
	}
}

func TestReservationService_Create_RejectsClosedWeekdayWhenConstraintsApply(t *testing.T) {
	// 2026-06-01 is Monday.
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	createCalled := false
	repo := closedDayCreateRepo(t)
	repo.createFn = func(_ context.Context, _ *model.Reservation) error {
		createCalled = true
		return nil
	}
	finder := &mockLineReservationSettingFinder{
		findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
			assert.Equal(t, uint64(1), clinicID)
			return &model.LineReservationSetting{
				ClosedWeekdays:        []byte("[1]"),
				ClosedDates:           []byte("[]"),
				NationalHolidayClosed: false,
			}, nil
		},
	}
	svc := NewReservationServiceWithLineSettings(repo, nil, &mockTransactor{}, nil, nil, nil, finder)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 5,
		Status:            model.ReservationStatusPending,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input but got: %v", err)
	assert.Contains(t, err.Error(), "休業曜日")
	assert.False(t, createCalled)
}

func TestReservationService_Create_RejectsClosedDateWhenConstraintsApply(t *testing.T) {
	start := time.Date(2026, 6, 2, 10, 0, 0, 0, config.JST)
	createCalled := false
	repo := closedDayCreateRepo(t)
	repo.createFn = func(_ context.Context, _ *model.Reservation) error {
		createCalled = true
		return nil
	}
	finder := &mockLineReservationSettingFinder{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			return &model.LineReservationSetting{
				ClosedWeekdays:        []byte("[]"),
				ClosedDates:           []byte(`["2026-06-02"]`),
				NationalHolidayClosed: false,
			}, nil
		},
	}
	svc := NewReservationServiceWithLineSettings(repo, nil, &mockTransactor{}, nil, nil, nil, finder)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 5,
		Status:            model.ReservationStatusConfirmed,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input but got: %v", err)
	assert.Contains(t, err.Error(), "休業日")
	assert.False(t, createCalled)
}

func TestReservationService_Create_RejectsNationalHolidayWhenConstraintsApply(t *testing.T) {
	start := time.Date(2026, 1, 1, 10, 0, 0, 0, config.JST)
	createCalled := false
	repo := closedDayCreateRepo(t)
	repo.createFn = func(_ context.Context, _ *model.Reservation) error {
		createCalled = true
		return nil
	}
	finder := &mockLineReservationSettingFinder{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			return &model.LineReservationSetting{
				ClosedWeekdays:        []byte("[]"),
				ClosedDates:           []byte("[]"),
				NationalHolidayClosed: true,
			}, nil
		},
	}
	svc := NewReservationServiceWithLineSettings(repo, nil, &mockTransactor{}, nil, nil, nil, finder)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 5,
		Status:            model.ReservationStatusPending,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input but got: %v", err)
	assert.Contains(t, err.Error(), "祝日")
	assert.False(t, createCalled)
}

func TestReservationService_Create_AllowsClosedDayForWalkInRoutes(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	finder := &mockLineReservationSettingFinder{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			t.Fatal("walk-in routes must not load closed-day settings")
			return nil, nil
		},
	}
	routes := []string{"reception", "exam_room", "record_shortcut"}
	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			created := false
			repo := closedDayCreateRepo(t)
			repo.createFn = func(_ context.Context, appt *model.Reservation) error {
				created = true
				appt.ID = 10
				return nil
			}
			svc := NewReservationServiceWithLineSettings(repo, nil, &mockTransactor{}, nil, nil, nil, finder)
			routeCopy := route

			result, err := svc.Create(context.Background(), &CreateManualReservationInput{
				ClinicID:          1,
				StartTime:         start,
				EndTime:           start.Add(30 * time.Minute),
				ReservationTypeID: 5,
				Status:            model.ReservationStatusPending,
				ReservationRoute:  &routeCopy,
			})

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, created)
		})
	}
}

func TestReservationService_Create_AllowsClosedDayForCheckedInStatus(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	created := false
	repo := closedDayCreateRepo(t)
	repo.createFn = func(_ context.Context, appt *model.Reservation) error {
		created = true
		appt.ID = 10
		return nil
	}
	finder := &mockLineReservationSettingFinder{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			t.Fatal("checked_in+ create must not load closed-day settings")
			return nil, nil
		},
	}
	svc := NewReservationServiceWithLineSettings(repo, nil, &mockTransactor{}, nil, nil, nil, finder)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 5,
		Status:            model.ReservationStatusCheckedIn,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, created)
}

func TestReservationService_Create_FailsClosedWhenSettingsCannotBeLoaded(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	createCalled := false
	repo := closedDayCreateRepo(t)
	repo.createFn = func(_ context.Context, _ *model.Reservation) error {
		createCalled = true
		return nil
	}
	finder := &mockLineReservationSettingFinder{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			return nil, errors.New("settings db error")
		},
	}
	svc := NewReservationServiceWithLineSettings(repo, nil, &mockTransactor{}, nil, nil, nil, finder)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 5,
		Status:            model.ReservationStatusPending,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.False(t, createCalled)
	assert.False(t, apperrors.IsInvalidInput(err), "load failure must not look like a closed-day validation error")
}

type mockClinicHolidayFinder struct {
	findByDateFn func(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error)
}

func (m *mockClinicHolidayFinder) FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error) {
	if m.findByDateFn != nil {
		return m.findByDateFn(ctx, clinicID, date)
	}
	return nil, apperrors.WrapNotFound("clinic_holiday", date.Format(time.DateOnly))
}

func TestReservationService_Create_RejectsClinicHolidayWhenConstraintsApply(t *testing.T) {
	start := time.Date(2026, 6, 2, 10, 0, 0, 0, config.JST)
	createCalled := false
	repo := closedDayCreateRepo(t)
	repo.createFn = func(_ context.Context, _ *model.Reservation) error {
		createCalled = true
		return nil
	}
	finder := &mockLineReservationSettingFinder{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			return &model.LineReservationSetting{
				ClosedWeekdays:        []byte("[]"),
				ClosedDates:           []byte("[]"),
				NationalHolidayClosed: false,
			}, nil
		},
	}
	holidayFinder := &mockClinicHolidayFinder{
		findByDateFn: func(_ context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, "2026-06-02", date.In(config.JST).Format(time.DateOnly))
			return &model.ClinicHoliday{ID: 1, ClinicID: clinicID, Date: date, Reason: "臨時休診"}, nil
		},
	}
	svc := NewReservationServiceWithClinicHolidays(repo, nil, &mockTransactor{}, nil, nil, nil, finder, holidayFinder)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 5,
		Status:            model.ReservationStatusConfirmed,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input but got: %v", err)
	assert.Contains(t, err.Error(), "休診日")
	assert.False(t, createCalled)
}

func TestReservationService_Create_AllowsWhenClinicHolidayNotFound(t *testing.T) {
	start := time.Date(2026, 6, 2, 10, 0, 0, 0, config.JST)
	createCalled := false
	holidayCalled := false
	repo := closedDayCreateRepo(t)
	repo.createFn = func(_ context.Context, appt *model.Reservation) error {
		createCalled = true
		appt.ID = 10
		return nil
	}
	finder := &mockLineReservationSettingFinder{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			return &model.LineReservationSetting{
				ClosedWeekdays:        []byte("[]"),
				ClosedDates:           []byte("[]"),
				NationalHolidayClosed: false,
			}, nil
		},
	}
	holidayFinder := &mockClinicHolidayFinder{
		findByDateFn: func(_ context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error) {
			holidayCalled = true
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, "2026-06-02", date.In(config.JST).Format(time.DateOnly))
			return nil, apperrors.WrapNotFound("clinic_holiday", date.Format(time.DateOnly))
		},
	}
	svc := NewReservationServiceWithClinicHolidays(repo, nil, &mockTransactor{}, nil, nil, nil, finder, holidayFinder)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 5,
		Status:            model.ReservationStatusPending,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, holidayCalled)
	assert.True(t, createCalled)
}

func TestReservationService_Create_WalkInRoutesDoNotCallClinicHolidayFinder(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	finder := &mockLineReservationSettingFinder{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			t.Fatal("walk-in routes must not load closed-day settings")
			return nil, nil
		},
	}
	holidayFinder := &mockClinicHolidayFinder{
		findByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClinicHoliday, error) {
			t.Fatal("walk-in routes must not load clinic_holidays")
			return nil, nil
		},
	}
	routes := []string{"reception", "exam_room", "record_shortcut"}
	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			created := false
			repo := closedDayCreateRepo(t)
			repo.createFn = func(_ context.Context, appt *model.Reservation) error {
				created = true
				appt.ID = 10
				return nil
			}
			svc := NewReservationServiceWithClinicHolidays(repo, nil, &mockTransactor{}, nil, nil, nil, finder, holidayFinder)
			routeCopy := route

			result, err := svc.Create(context.Background(), &CreateManualReservationInput{
				ClinicID:          1,
				StartTime:         start,
				EndTime:           start.Add(30 * time.Minute),
				ReservationTypeID: 5,
				Status:            model.ReservationStatusPending,
				ReservationRoute:  &routeCopy,
			})

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, created)
		})
	}
}

func TestReservationService_Create_FailsClosedWhenClinicHolidayFinderErrors(t *testing.T) {
	start := time.Date(2026, 6, 2, 10, 0, 0, 0, config.JST)
	createCalled := false
	repo := closedDayCreateRepo(t)
	repo.createFn = func(_ context.Context, _ *model.Reservation) error {
		createCalled = true
		return nil
	}
	finder := &mockLineReservationSettingFinder{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			return &model.LineReservationSetting{
				ClosedWeekdays:        []byte("[]"),
				ClosedDates:           []byte("[]"),
				NationalHolidayClosed: false,
			}, nil
		},
	}
	holidayFinder := &mockClinicHolidayFinder{
		findByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.ClinicHoliday, error) {
			return nil, errors.New("clinic_holidays db error")
		},
	}
	svc := NewReservationServiceWithClinicHolidays(repo, nil, &mockTransactor{}, nil, nil, nil, finder, holidayFinder)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 5,
		Status:            model.ReservationStatusPending,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.False(t, createCalled)
	assert.False(t, apperrors.IsInvalidInput(err), "finder failure must not look like a closed-day validation error")
}

func TestReservationService_Create_FailsClosedWhenClinicHolidayFinderMissing(t *testing.T) {
	start := time.Date(2026, 6, 2, 10, 0, 0, 0, config.JST)
	createCalled := false
	repo := closedDayCreateRepo(t)
	repo.createFn = func(_ context.Context, _ *model.Reservation) error {
		createCalled = true
		return nil
	}
	finder := &mockLineReservationSettingFinder{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
			return &model.LineReservationSetting{
				ClosedWeekdays:        []byte("[]"),
				ClosedDates:           []byte("[]"),
				NationalHolidayClosed: false,
			}, nil
		},
	}
	svc := NewReservationServiceWithLineSettings(repo, nil, &mockTransactor{}, nil, nil, nil, finder)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 5,
		Status:            model.ReservationStatusPending,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.False(t, createCalled)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
}

func TestReservationService_Create_FailsClosedWhenSettingFinderMissing(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	createCalled := false
	repo := closedDayCreateRepo(t)
	repo.createFn = func(_ context.Context, _ *model.Reservation) error {
		createCalled = true
		return nil
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: 5,
		Status:            model.ReservationStatusPending,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.False(t, createCalled)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
}

func TestReservationService_Create_SkipsBookingConstraintsForInConsultation(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 30, 0, 0, config.JST)
	end := start.Add(30 * time.Minute)
	specificDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	created := false
	repo := &mockReservationRepository{
		createFn: func(_ context.Context, appt *model.Reservation) error {
			created = true
			appt.ID = 10
			return nil
		},
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
			t.Fatal("in_consultation appointment must not run booking capacity checks")
			return 0, nil
		},
	}
	unavailableRepo := &mockUnavailableTimeRepository{
		findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeUnavailableTime, error) {
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
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, unavailableRepo)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         start,
		EndTime:           end,
		ReservationTypeID: 5,
		Status:            model.ReservationStatusInConsultation,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, created)
}

func TestReservationService_Update(t *testing.T) {
	_ = time.Now() // StartTime/EndTime/DoctorID を含むケースは統合テストで検証（s.db 必須）
	statusConfirmed := model.ReservationStatusConfirmed
	tests := []struct {
		name    string
		input   UpdateReservationInput
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			// StartTime/EndTime/DoctorID を含まないケース: s.db を使わないパス
			name: "updates reservation successfully (status only, no conflict check)",
			input: UpdateReservationInput{
				Status: &statusConfirmed,
			},
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateReservationInput{},
			repoErr: nil,
			wantErr: true,
			wantNF:  false,
		},
		{
			name: "returns not found error when reservation does not exist",
			input: UpdateReservationInput{
				Status: &statusConfirmed,
			},
			repoErr: apperrors.WrapNotFound("reservation", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateReservationInput{
				Status: &statusConfirmed,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
			wantNF:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
					if tt.wantNF {
						return nil, tt.repoErr
					}
					return &model.Reservation{
						ID:                id,
						ClinicID:          clinicID,
						ReservationTypeID: 1,
						Status:            model.ReservationStatusPending,
						StartTime:         time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST),
						EndTime:           time.Date(2026, 6, 1, 10, 30, 0, 0, config.JST),
					}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Reservation{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)

			reservation, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, reservation)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reservation)
			}
		})
	}
}

func TestReservationService_Update_RejectsFullReservationTypeCapacity(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	nextStart := start.Add(time.Hour)
	maxConcurrent := 2
	updateCalled := false
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:                id,
				ClinicID:          clinicID,
				ReservationTypeID: 9,
				StartTime:         start,
				EndTime:           start.Add(30 * time.Minute),
				Status:            model.ReservationStatusPending,
			}, nil
		},
		lockAndFindByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:                id,
				ClinicID:          clinicID,
				ReservationTypeID: 9,
				StartTime:         start,
				EndTime:           start.Add(30 * time.Minute),
				Status:            model.ReservationStatusPending,
			}, nil
		},
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
			return 3, nil
		},
		countConflictsFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
			return 0, nil
		},
		countByTypeAndStartTimeFn: func(_ context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(9), reservationTypeID)
			assert.Equal(t, nextStart, startTime)
			require.NotNil(t, excludeID)
			assert.Equal(t, uint64(1), *excludeID)
			return 2, nil
		},
		updateFieldsFn: func(_ context.Context, _ uint64, _ uint64, _ map[string]any) (*model.Reservation, error) {
			updateCalled = true
			return &model.Reservation{}, nil
		},
	}
	typeRepo := mockReservationTypeFinder{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{ID: id, ClinicID: clinicID, MaxConcurrent: &maxConcurrent}, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, typeRepo, &mockTransactor{}, nil, nil)
	nextEnd := nextStart.Add(30 * time.Minute)

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{StartTime: &nextStart, EndTime: &nextEnd})

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "expected conflict but got: %v", err)
	assert.Nil(t, result)
	assert.False(t, updateCalled)
}

// BUG-006: フロントはメモのみ編集でも start/end/doctor をフル送信する。
// payload の存在ではなく current との実変更で conflict/absence を再評価する。
func TestReservationService_Update_SkipsConflictCheckWhenScheduleUnchanged(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	end := start.Add(30 * time.Minute)
	doctorID := uint64(7)
	typeID := uint64(9)
	notes := "メモのみ変更"
	current := func(clinicID, id uint64) *model.Reservation {
		return &model.Reservation{
			ID:                id,
			ClinicID:          clinicID,
			ReservationTypeID: typeID,
			DoctorID:          &doctorID,
			StartTime:         start,
			EndTime:           end,
			Status:            model.ReservationStatusPending,
		}
	}
	updateCalled := false
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return current(clinicID, id), nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
			t.Fatal("unchanged schedule must not lock for conflict check")
			return nil, nil
		},
		hasDoctorConflictFn: func(_ context.Context, _, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
			t.Fatal("unchanged schedule must not call doctor conflict checker")
			return false, nil
		},
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
			t.Fatal("unchanged schedule must not call on-duty checker")
			return 0, nil
		},
		countConflictsFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
			t.Fatal("unchanged schedule must not call conflict counter")
			return 0, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
			updateCalled = true
			assert.Equal(t, notes, fields["notes"])
			return &model.Reservation{ID: 1, ClinicID: 1, Notes: notes}, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{
		StartTime:         &start,
		EndTime:           &end,
		DoctorID:          &doctorID,
		ReservationTypeID: &typeID,
		Notes:             &notes,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, updateCalled)
}

func TestReservationService_Update_ConflictChecksWhenStartChanges(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	end := start.Add(30 * time.Minute)
	nextStart := start.Add(time.Hour)
	nextEnd := nextStart.Add(30 * time.Minute)
	onDutyCalled := false
	conflictCalled := false
	current := func(clinicID, id uint64) *model.Reservation {
		return &model.Reservation{
			ID:                id,
			ClinicID:          clinicID,
			ReservationTypeID: 9,
			StartTime:         start,
			EndTime:           end,
			Status:            model.ReservationStatusPending,
		}
	}
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return current(clinicID, id), nil
		},
		lockAndFindByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return current(clinicID, id), nil
		},
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
			onDutyCalled = true
			return 1, nil
		},
		countConflictsFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
			conflictCalled = true
			return 0, nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: 1, StartTime: nextStart, EndTime: nextEnd}, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{
		StartTime: &nextStart,
		EndTime:   &nextEnd,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, onDutyCalled, "actual start change must re-run on-duty/absence check")
	assert.True(t, conflictCalled, "actual start change must re-run conflict check")
}

func TestReservationService_Update_SkipsConflictCheckWhenAlreadyCheckedIn(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	end := start.Add(30 * time.Minute)
	nextStart := start.Add(time.Hour)
	nextEnd := nextStart.Add(30 * time.Minute)
	current := func(clinicID, id uint64) *model.Reservation {
		return &model.Reservation{
			ID:                id,
			ClinicID:          clinicID,
			ReservationTypeID: 9,
			StartTime:         start,
			EndTime:           end,
			Status:            model.ReservationStatusCheckedIn,
		}
	}
	updateCalled := false
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return current(clinicID, id), nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
			t.Fatal("checked_in reservation must not re-run booking conflict lock")
			return nil, nil
		},
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
			t.Fatal("checked_in reservation must not re-run on-duty checker")
			return 0, nil
		},
		countConflictsFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
			t.Fatal("checked_in reservation must not re-run conflict counter")
			return 0, nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.Reservation, error) {
			updateCalled = true
			return &model.Reservation{ID: id, ClinicID: 1, Status: model.ReservationStatusCheckedIn, StartTime: nextStart, EndTime: nextEnd}, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{
		StartTime: &nextStart,
		EndTime:   &nextEnd,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, updateCalled)
}

// G-1: checked_in 以降でも担当者/予約区分の実変更は capability を検証する。
// BUG-006 のスロット競合/出勤チェック skip は維持する。
func TestReservationService_Update_RejectsIncapableDoctorWhenStatusCheckedInOrLater(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, config.JST)
	end := start.Add(30 * time.Minute)
	currentDoctorID := uint64(7)
	nextDoctorID := uint64(8)
	typeID := uint64(9)

	tests := []struct {
		name    string
		status  model.ReservationStatus
		capable bool
		wantErr bool
	}{
		{name: "checked_in: incapable doctor is rejected", status: model.ReservationStatusCheckedIn, capable: false, wantErr: true},
		{name: "in_consultation: incapable doctor is rejected", status: model.ReservationStatusInConsultation, capable: false, wantErr: true},
		{name: "accounting: incapable doctor is rejected", status: model.ReservationStatusAccounting, capable: false, wantErr: true},
		{name: "completed: incapable doctor is rejected", status: model.ReservationStatusCompleted, capable: false, wantErr: true},
		{name: "checked_in: capable doctor change succeeds without slot checks", status: model.ReservationStatusCheckedIn, capable: true, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := func(clinicID, id uint64) *model.Reservation {
				return &model.Reservation{
					ID:                id,
					ClinicID:          clinicID,
					ReservationTypeID: typeID,
					DoctorID:          &currentDoctorID,
					StartTime:         start,
					EndTime:           end,
					Status:            tt.status,
				}
			}
			updateCalled := false
			repo := &mockReservationRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
					return current(clinicID, id), nil
				},
				lockAndFindByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
					return current(clinicID, id), nil
				},
				hasDoctorConflictFn: func(_ context.Context, _, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
					t.Fatal("checked_in+ doctor change must not call doctor conflict checker")
					return false, nil
				},
				countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) {
					t.Fatal("checked_in+ doctor change must not call on-duty checker")
					return 0, nil
				},
				countConflictsFn: func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
					t.Fatal("checked_in+ doctor change must not call conflict counter")
					return 0, nil
				},
				updateFieldsFn: func(_ context.Context, _, id uint64, fields map[string]any) (*model.Reservation, error) {
					if tt.wantErr {
						t.Fatal("incapable doctor must not update the reservation")
						return nil, nil
					}
					updateCalled = true
					assert.Equal(t, nextDoctorID, fields["doctor_id"])
					return &model.Reservation{
						ID:                id,
						ClinicID:          1,
						ReservationTypeID: typeID,
						DoctorID:          &nextDoctorID,
						StartTime:         start,
						EndTime:           end,
						Status:            tt.status,
					}, nil
				},
			}
			staffRepo := &mockReservationStaffRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, nextDoctorID, id)
					return &model.Staff{ID: id, ClinicID: clinicID, IsActive: true}, nil
				},
				supportsReservationTypeFn: func(_ context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, nextDoctorID, staffID)
					assert.Equal(t, typeID, reservationTypeID)
					return tt.capable, nil
				},
			}
			svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, staffRepo, nil)

			result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{
				DoctorID: &nextDoctorID,
			})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input but got: %v", err)
				assert.Contains(t, err.Error(), "選択した担当者はこの予約区分に対応していません")
				assert.False(t, updateCalled)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, updateCalled)
		})
	}
}

// Q7: キャンセル(status=cancelled)への更新時は repo.Delete でソフトデリートし、
// 予約管理(FindAll の deleted_at IS NULL)から除外されることを保証する。
// RSV-06: status 更新と soft delete は同一 WithTx 内で原子的に行う。
func TestReservationService_Update_CancelledSoftDeletes(t *testing.T) {
	statusCancelled := model.ReservationStatusCancelled
	deleteCalled := false
	updateCalled := false
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID, Status: model.ReservationStatusPending}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			updateCalled = true
			return &model.Reservation{ID: 1, ClinicID: 1, Status: model.ReservationStatusCancelled}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleteCalled = true
			return nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	reservation, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{Status: &statusCancelled})

	assert.NoError(t, err)
	assert.NotNil(t, reservation)
	assert.True(t, updateCalled, "キャンセル時は status 更新が行われるべき")
	assert.True(t, deleteCalled, "キャンセル時は repo.Delete でソフトデリートされるべき")
}

// RSV-06: soft-delete 失敗時に status=cancelled だけの部分成功を残さない（同一 tx で rollback）。
func TestReservationService_Update_CancelRollsBackWhenDeleteFails(t *testing.T) {
	statusCancelled := model.ReservationStatusCancelled
	updateCalled := false
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID, Status: model.ReservationStatusPending}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			updateCalled = true
			return &model.Reservation{ID: 1, ClinicID: 1, Status: model.ReservationStatusCancelled}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("forced soft-delete failure")
		},
	}
	// mockTransactor runs fn without DB; service still requires non-nil tx and sequences update+delete.
	// Real atomicity is covered by repo DBOrTx + service WithTx structure; this asserts delete error propagates.
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	reservation, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{Status: &statusCancelled})

	assert.Error(t, err)
	assert.Nil(t, reservation)
	assert.True(t, updateCalled, "update is attempted before delete in the same tx callback")
	assert.Contains(t, err.Error(), "soft-delete")
}

// 受付ヘッダー テレメトリ（change-ui.md Phase 2）: checked_in への遷移時に checked_in_at が
// now() で記録されることを保証する。
func TestReservationService_Update_CheckedInStampsCheckedInAt(t *testing.T) {
	statusCheckedIn := model.ReservationStatusCheckedIn
	var capturedFields map[string]any
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID, Status: model.ReservationStatusPending}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
			capturedFields = fields
			return &model.Reservation{ID: 1, ClinicID: 1, Status: model.ReservationStatusCheckedIn}, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)

	before := time.Now()
	reservation, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{Status: &statusCheckedIn})
	after := time.Now()

	assert.NoError(t, err)
	assert.NotNil(t, reservation)
	require.Contains(t, capturedFields, "checked_in_at")
	stamped, ok := capturedFields["checked_in_at"].(time.Time)
	require.True(t, ok, "checked_in_at must be a time.Time")
	assert.False(t, stamped.Before(before), "checked_in_at must not be before the call")
	assert.False(t, stamped.After(after), "checked_in_at must not be after the call")
}

// 非ステータス更新（時刻変更のみ等）では checked_in_at に触れないことを保証する。
// UpdatedAt(autoUpdateTime) を待ち時間算出に流用してはならないという仕様の裏返しの検証。
func TestReservationService_Update_NonStatusUpdateLeavesCheckedInAtUntouched(t *testing.T) {
	newStart := time.Date(2026, 7, 5, 10, 0, 0, 0, config.JST)
	newEnd := newStart.Add(30 * time.Minute)
	var capturedFields map[string]any
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID, ReservationTypeID: 1, Status: model.ReservationStatusCheckedIn}, nil
		},
		lockAndFindByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID, ReservationTypeID: 1, Status: model.ReservationStatusCheckedIn}, nil
		},
		countOnDutyDoctorsFn: func(_ context.Context, _ uint64, _ time.Time) (int64, error) { return 1, nil },
		countConflictsFn:     func(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) { return 0, nil },
		updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
			capturedFields = fields
			return &model.Reservation{ID: 1, ClinicID: 1, Status: model.ReservationStatusCheckedIn}, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	reservation, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{StartTime: &newStart, EndTime: &newEnd})

	assert.NoError(t, err)
	assert.NotNil(t, reservation)
	assert.NotContains(t, capturedFields, "checked_in_at", "非ステータス更新は checked_in_at を変更してはならない")
}

// checked_in → 他ステータス → checked_in（再受付）で checked_in_at が最新遷移時刻へ上書きされる
// （待ち直しとみなす）ことを保証する。
func TestReservationService_Update_RecheckInResetsCheckedInAt(t *testing.T) {
	statusCheckedIn := model.ReservationStatusCheckedIn
	statusPending := model.ReservationStatusPending
	var capturedFields map[string]any
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID, Status: model.ReservationStatusPending}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
			capturedFields = fields
			return &model.Reservation{ID: 1, ClinicID: 1}, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)

	_, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{Status: &statusCheckedIn})
	require.NoError(t, err)
	firstStamp := capturedFields["checked_in_at"].(time.Time)

	_, err = svc.Update(context.Background(), 1, 1, &UpdateReservationInput{Status: &statusPending})
	require.NoError(t, err)
	assert.NotContains(t, capturedFields, "checked_in_at", "checked_in 以外への遷移では checked_in_at を触らない")

	time.Sleep(time.Millisecond)
	_, err = svc.Update(context.Background(), 1, 1, &UpdateReservationInput{Status: &statusCheckedIn})
	require.NoError(t, err)
	secondStamp, ok := capturedFields["checked_in_at"].(time.Time)
	require.True(t, ok)
	assert.True(t, secondStamp.After(firstStamp), "再受付時は checked_in_at が最新の遷移時刻へ上書きされるべき")
}

func TestReservationService_Update_RejectsExcludedStaffWhenTypeChanges(t *testing.T) {
	doctorID := uint64(10)
	nextTypeID := uint64(5)
	start := time.Now()
	end := start.Add(30 * time.Minute)
	transactionActive := false
	transactionCalls := 0
	capabilityChecked := false
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(1), id)
			return &model.Reservation{
				ID:                id,
				ClinicID:          clinicID,
				ReservationTypeID: 4,
				DoctorID:          &doctorID,
				StartTime:         start,
				EndTime:           end,
			}, nil
		},
		lockAndFindByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			assert.True(t, transactionActive, "reservation must be locked inside the write transaction")
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(1), id)
			return &model.Reservation{
				ID:                id,
				ClinicID:          clinicID,
				ReservationTypeID: 4,
				DoctorID:          &doctorID,
				StartTime:         start,
				EndTime:           end,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			t.Fatal("reservation must not be updated when staff cannot handle reservation type")
			return nil, nil
		},
	}
	staffRepo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
			assert.True(t, transactionActive, "staff must be validated inside the write transaction")
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, doctorID, id)
			return &model.Staff{ID: id, IsActive: true}, nil
		},
		supportsReservationTypeFn: func(_ context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
			assert.True(t, transactionActive, "staff capability must be validated inside the write transaction")
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, doctorID, staffID)
			assert.Equal(t, nextTypeID, reservationTypeID)
			capabilityChecked = true
			return false, nil
		},
	}
	tx := &mockTransactor{
		withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			transactionCalls++
			transactionActive = true
			defer func() {
				transactionActive = false
			}()
			return fn(ctx)
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, tx, staffRepo, nil)

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{
		ReservationTypeID: &nextTypeID,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "expected ErrInvalidInput but got: %v", err)
	assert.Equal(t, 1, transactionCalls)
	assert.True(t, capabilityChecked)
}

func TestReservationService_Update_FailsClosedWithoutTransactor(t *testing.T) {
	doctorID := uint64(10)
	ownerID := uint64(20)
	tests := []struct {
		name  string
		input *UpdateReservationInput
	}{
		{
			name:  "conflict check branch",
			input: &UpdateReservationInput{DoctorID: &doctorID},
		},
		{
			name:  "owner pet validation branch",
			input: &UpdateReservationInput{OwnerID: &ownerID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCalled := false
			repo := &mockReservationRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
					return &model.Reservation{
						ID:                id,
						ClinicID:          clinicID,
						ReservationTypeID: 5,
						StartTime:         time.Now(),
						EndTime:           time.Now().Add(30 * time.Minute),
					}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
					updateCalled = true
					return nil, nil
				},
			}
			svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)

			result, err := svc.Update(context.Background(), 1, 1, tt.input)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.False(t, updateCalled)
			var appErr *apperrors.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, "INTERNAL", appErr.Code)
		})
	}
}

func TestReservationService_Update_RejectsInConsultationWithoutMedicalRecord(t *testing.T) {
	status := model.ReservationStatusInConsultation
	current := func(clinicID, id uint64) *model.Reservation {
		return &model.Reservation{
			ID:                id,
			ClinicID:          clinicID,
			ReservationTypeID: 5,
			Status:            model.ReservationStatusCheckedIn,
		}
	}
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return current(clinicID, id), nil
		},
		lockAndFindByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return current(clinicID, id), nil
		},
		countMedicalRecordsByReservationID: func(_ context.Context, _ uint64) (int64, error) {
			return 0, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			t.Fatal("must not transition to in_consultation without a medical record")
			return nil, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{Status: &status})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsConflict(err), "expected conflict but got: %v", err)
}

func TestReservationService_Update_AllowsInConsultationWhenMedicalRecordExists(t *testing.T) {
	status := model.ReservationStatusInConsultation
	counted := false
	current := func(clinicID, id uint64) *model.Reservation {
		return &model.Reservation{
			ID:                id,
			ClinicID:          clinicID,
			ReservationTypeID: 5,
			Status:            model.ReservationStatusCheckedIn,
		}
	}
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return current(clinicID, id), nil
		},
		lockAndFindByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return current(clinicID, id), nil
		},
		countMedicalRecordsByReservationID: func(_ context.Context, reservationID uint64) (int64, error) {
			counted = true
			assert.Equal(t, uint64(1), reservationID)
			return 1, nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: 1, Status: model.ReservationStatusInConsultation}, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{Status: &status})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, counted)
	assert.Equal(t, model.ReservationStatusInConsultation, result.Status)
}

func TestReservationService_Update_FailsClosedWhenInConsultationRecordCountErrors(t *testing.T) {
	status := model.ReservationStatusInConsultation
	current := func(clinicID, id uint64) *model.Reservation {
		return &model.Reservation{
			ID:                id,
			ClinicID:          clinicID,
			ReservationTypeID: 5,
			Status:            model.ReservationStatusCheckedIn,
		}
	}
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return current(clinicID, id), nil
		},
		lockAndFindByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return current(clinicID, id), nil
		},
		countMedicalRecordsByReservationID: func(_ context.Context, _ uint64) (int64, error) {
			return 0, errors.New("db error")
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			t.Fatal("must not transition to in_consultation when record count fails")
			return nil, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{Status: &status})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.False(t, apperrors.IsConflict(err), "count failure must not be reported as a business conflict")
}

func TestReservationService_Update_InConsultationMedicalRecordCountRunsAfterRowLock(t *testing.T) {
	status := model.ReservationStatusInConsultation
	var events []string
	current := func(clinicID, id uint64) *model.Reservation {
		return &model.Reservation{
			ID:                id,
			ClinicID:          clinicID,
			ReservationTypeID: 5,
			Status:            model.ReservationStatusCheckedIn,
		}
	}
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			events = append(events, "find")
			return current(clinicID, id), nil
		},
		lockAndFindByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			events = append(events, "lock")
			return current(clinicID, id), nil
		},
		acquireBookingLockFn: func(_ context.Context, _ uint64) error {
			t.Fatal("status-only in_consultation must not take booking advisory lock")
			return nil
		},
		countMedicalRecordsByReservationID: func(_ context.Context, reservationID uint64) (int64, error) {
			events = append(events, "count")
			assert.Equal(t, uint64(1), reservationID)
			return 1, nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.Reservation, error) {
			events = append(events, "update")
			return &model.Reservation{ID: id, ClinicID: 1, Status: model.ReservationStatusInConsultation}, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{Status: &status})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, []string{"find", "lock", "count", "update"}, events)
}

func TestReservationService_Update_RejectsLineCheckedInWithoutOwnerPet(t *testing.T) {
	lineCustomerID := uint64(99)
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(1), id)
			return &model.Reservation{
				ID:                id,
				ClinicID:          clinicID,
				ReservationTypeID: 5,
				Source:            model.ReservationSourceLine,
				LineCustomerID:    &lineCustomerID,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			t.Fatal("line reservation must not be checked in without owner/pet link")
			return nil, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)
	status := model.ReservationStatusCheckedIn

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{
		Status: &status,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "expected ErrInvalidInput but got: %v", err)
}

func TestReservationService_Update_RejectsUnavailableTimeWhenTimeChanges(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 30, 0, 0, config.JST)
	end := start.Add(30 * time.Minute)
	specificDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(1), id)
			return &model.Reservation{
				ID:                id,
				ClinicID:          clinicID,
				ReservationTypeID: 5,
				StartTime:         time.Date(2026, 6, 1, 9, 0, 0, 0, config.JST),
				EndTime:           time.Date(2026, 6, 1, 9, 30, 0, 0, config.JST),
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			t.Fatal("reservation must not be updated during unavailable time")
			return nil, nil
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
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, unavailableRepo)

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{
		StartTime: &start,
		EndTime:   &end,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err), "expected ErrInvalidInput but got: %v", err)
}

func TestReservationService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		clinicID     uint64
		id           uint64
		recordCount  int64
		countErr     error
		repoErr      error
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name:        "deletes reservation successfully when no medical records linked",
			clinicID:    1,
			id:          10,
			recordCount: 0,
			repoErr:     nil,
			wantErr:     false,
		},
		{
			name:         "returns conflict error when medical records are linked",
			clinicID:     1,
			id:           10,
			recordCount:  2,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:     "returns error when count check fails",
			clinicID: 1,
			id:       10,
			countErr: errors.New("db error"),
			wantErr:  true,
		},
		{
			name:        "returns not found error when reservation does not exist",
			clinicID:    1,
			id:          999,
			recordCount: 0,
			repoErr:     apperrors.WrapNotFound("reservation", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:        "returns error on repository delete failure",
			clinicID:    1,
			id:          10,
			recordCount: 0,
			repoErr:     errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationRepository{
				countMedicalRecordsByReservationID: func(_ context.Context, _ uint64) (int64, error) {
					return tt.recordCount, tt.countErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReservationService_UpdateReservationRoute(t *testing.T) {
	validRoutes := []string{"line", "phone", "reception", "exam_room", "record_shortcut"}

	t.Run("success: valid routes update reservation_route", func(t *testing.T) {
		for _, route := range validRoutes {
			route := route
			t.Run(route, func(t *testing.T) {
				repo := &mockReservationRepository{
					findByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.Reservation, error) {
						return &model.Reservation{ID: 1}, nil
					},
					updateFieldsFn: func(_ context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
						assert.Equal(t, uint64(1), clinicID)
						assert.Equal(t, uint64(1), id)
						assert.Equal(t, route, fields["reservation_route"])
						return &model.Reservation{ID: 1}, nil
					},
				}
				svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)
				result, err := svc.UpdateReservationRoute(context.Background(), 1, 1, UpdateReservationRouteInput{Route: route})
				assert.NoError(t, err)
				assert.NotNil(t, result)
			})
		}
	})

	t.Run("success: empty route stores NULL", func(t *testing.T) {
		repo := &mockReservationRepository{
			findByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: 1}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
				assert.Nil(t, fields["reservation_route"])
				return &model.Reservation{ID: 1}, nil
			},
		}
		svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)
		result, err := svc.UpdateReservationRoute(context.Background(), 1, 1, UpdateReservationRouteInput{Route: ""})
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("error: invalid route 'fax' returns InvalidInput", func(t *testing.T) {
		svc := NewReservationServiceWithAvailabilityAndType(&mockReservationRepository{}, nil, nil, nil, nil)
		_, err := svc.UpdateReservationRoute(context.Background(), 1, 1, UpdateReservationRouteInput{Route: "fax"})
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("error: uppercase 'LINE' is not valid", func(t *testing.T) {
		svc := NewReservationServiceWithAvailabilityAndType(&mockReservationRepository{}, nil, nil, nil, nil)
		_, err := svc.UpdateReservationRoute(context.Background(), 1, 1, UpdateReservationRouteInput{Route: "LINE"})
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("error: reservation not found returns NotFound", func(t *testing.T) {
		repo := &mockReservationRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
				return nil, apperrors.WrapNotFound("reservation", "1")
			},
		}
		svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)
		_, err := svc.UpdateReservationRoute(context.Background(), 1, 1, UpdateReservationRouteInput{Route: "line"})
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("error: wrong clinic_id returns NotFound (P4 isolation)", func(t *testing.T) {
		repo := &mockReservationRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
				if clinicID != 1 {
					return nil, apperrors.WrapNotFound("reservation", "1")
				}
				return &model.Reservation{ID: 1}, nil
			},
		}
		svc := NewReservationServiceWithAvailabilityAndType(repo, nil, nil, nil, nil)
		_, err := svc.UpdateReservationRoute(context.Background(), 99, 1, UpdateReservationRouteInput{Route: "line"})
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// #261 P0: 死亡ペットへの予約 write は fail-closed（create / update pet 付け替え）。
func TestReservationService_Create_RejectsDeceasedPet(t *testing.T) {
	now := time.Now()
	deceasedAt := now.Add(-24 * time.Hour)
	petID := uint64(5)
	ownerID := uint64(2)
	createCalled := false
	repo := &mockReservationRepository{
		findPetOwnerInClinicFn: func(_ context.Context, _, id uint64) (uint64, error) {
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
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         now,
		EndTime:           now.Add(time.Hour),
		ReservationTypeID: 1,
		OwnerID:           &ownerID,
		PetID:             &petID,
		Status:            model.ReservationStatusConfirmed,
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "expected InvalidInput, got: %v", err)
	assert.Contains(t, err.Error(), reservationDeceasedPetMessage)
	assert.Nil(t, result)
	assert.False(t, createCalled, "repo.Create must not be called for deceased pet")
}

func TestReservationService_Create_AllowsLivingPet(t *testing.T) {
	now := time.Now()
	petID := uint64(5)
	ownerID := uint64(2)
	createCalled := false
	repo := &mockReservationRepository{
		findPetOwnerInClinicFn: func(_ context.Context, _, id uint64) (uint64, error) {
			return ownerID, nil
		},
		findPetByIDInClinicFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, OwnerID: ownerID, Status: model.PetStatusAlive}, nil
		},
		createFn: func(_ context.Context, r *model.Reservation) error {
			createCalled = true
			r.ID = 42
			return nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	result, err := svc.Create(context.Background(), &CreateManualReservationInput{
		ClinicID:          1,
		StartTime:         now,
		EndTime:           now.Add(time.Hour),
		ReservationTypeID: 1,
		OwnerID:           &ownerID,
		PetID:             &petID,
		Status:            model.ReservationStatusInConsultation,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, createCalled)
}

func TestReservationService_Update_RejectsDeceasedPetReplacement(t *testing.T) {
	deceasedAt := time.Now().Add(-24 * time.Hour)
	newPetID := uint64(9)
	ownerID := uint64(2)
	updateCalled := false
	repo := &mockReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID: id, ClinicID: 1, OwnerID: &ownerID, PetID: ptrUint64(5),
				Status: model.ReservationStatusConfirmed,
			}, nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID: id, ClinicID: 1, OwnerID: &ownerID, PetID: ptrUint64(5),
				Status: model.ReservationStatusConfirmed,
			}, nil
		},
		findPetOwnerInClinicFn: func(_ context.Context, _, id uint64) (uint64, error) {
			return ownerID, nil
		},
		findPetByIDInClinicFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, OwnerID: ownerID, DeceasedAt: &deceasedAt, Status: model.PetStatusDeceased}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			updateCalled = true
			return nil, nil
		},
	}
	svc := NewReservationServiceWithAvailabilityAndType(repo, nil, &mockTransactor{}, nil, nil)

	result, err := svc.Update(context.Background(), 1, 1, &UpdateReservationInput{PetID: &newPetID})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "expected InvalidInput, got: %v", err)
	assert.Nil(t, result)
	assert.False(t, updateCalled, "repo.update must not run when replacement pet is deceased")
}
