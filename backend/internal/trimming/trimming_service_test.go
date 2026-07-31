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
	"github.com/animal-ekarte/backend/internal/reservation"
)

// --- mock: trimming が消費する reservation read + intent operation ---

type mockTrimmingReservationRepository struct {
	findAllByCategoryFn  func(ctx context.Context, clinicID uint64, category model.ReservationTypeCategory, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error)
	findByIDFn           func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	acquireBookingLockFn func(ctx context.Context, clinicID uint64) error
	lockAndFindByIDFn    func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	hasDoctorConflictFn  func(ctx context.Context, clinicID, doctorID uint64, start, end time.Time, excludeID *uint64) (bool, error)
	countOnDutyDoctorsFn func(ctx context.Context, clinicID uint64, date time.Time) (int64, error)
	countConflictsFn     func(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error)
	countByTypeFn        func(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error)
	findPetOwnerFn       func(ctx context.Context, clinicID, petID uint64) (uint64, error)
	findPetByIDFn        func(petID uint64) (*model.Pet, error)
	createFn             func(ctx context.Context, appt *model.Reservation) error
	updateFieldsFn       func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error)
	deleteFn             func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockTrimmingReservationRepository) FindAllByCategory(ctx context.Context, clinicID uint64, category model.ReservationTypeCategory, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error) {
	return m.findAllByCategoryFn(ctx, clinicID, category, petID, ownerID, startDate, endDate, page, limit)
}

func (m *mockTrimmingReservationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, apperrors.WrapNotFound("appointment", "0")
}

func (m *mockTrimmingReservationRepository) FindTrimmingByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	return m.FindByID(ctx, clinicID, id)
}

func (m *mockTrimmingReservationRepository) CreateForTrimming(ctx context.Context, clinicID uint64, input reservation.CreateTrimmingReservationInput) (*model.Reservation, error) {
	appt := &model.Reservation{
		ClinicID:          clinicID,
		ReservationTypeID: input.ReservationTypeID,
		StartTime:         input.StartTime,
		EndTime:           input.EndTime,
		PetID:             input.PetID,
		DoctorID:          input.DoctorID,
		Status:            input.Status,
		Source:            model.ReservationSourceManual,
		ReservationRoute:  input.ReservationRoute,
	}
	if m.createFn != nil {
		if err := m.createFn(ctx, appt); err != nil {
			return nil, err
		}
	}
	return appt, nil
}

func (m *mockTrimmingReservationRepository) UpdateForTrimming(ctx context.Context, clinicID, id uint64, input reservation.UpdateTrimmingReservationInput) (*model.Reservation, error) {
	fields := make(map[string]any, 5)
	if input.StartTime != nil {
		fields["start_time"] = *input.StartTime
	}
	if input.EndTime != nil {
		fields["end_time"] = *input.EndTime
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.Reservation{ID: id, ClinicID: clinicID}, nil
}

func (m *mockTrimmingReservationRepository) DeleteForTrimming(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockTrimmingReservationRepository) AcquireBookingLock(ctx context.Context, clinicID uint64) error {
	if m.acquireBookingLockFn != nil {
		return m.acquireBookingLockFn(ctx, clinicID)
	}
	return nil
}

func (m *mockTrimmingReservationRepository) LockTrimmingByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if m.lockAndFindByIDFn != nil {
		return m.lockAndFindByIDFn(ctx, clinicID, id)
	}
	return m.FindByID(ctx, clinicID, id)
}

func (m *mockTrimmingReservationRepository) HasDoctorConflict(ctx context.Context, clinicID, doctorID uint64, start, end time.Time, excludeID *uint64) (bool, error) {
	if m.hasDoctorConflictFn != nil {
		return m.hasDoctorConflictFn(ctx, clinicID, doctorID, start, end, excludeID)
	}
	return false, nil
}

func (m *mockTrimmingReservationRepository) CountOnDutyDoctors(ctx context.Context, clinicID uint64, date time.Time) (int64, error) {
	if m.countOnDutyDoctorsFn != nil {
		return m.countOnDutyDoctorsFn(ctx, clinicID, date)
	}
	return 1, nil
}

func (m *mockTrimmingReservationRepository) CountConflicts(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error) {
	if m.countConflictsFn != nil {
		return m.countConflictsFn(ctx, clinicID, start, end, excludeID)
	}
	return 0, nil
}

func (m *mockTrimmingReservationRepository) CountByTypeAndStartTime(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error) {
	if m.countByTypeFn != nil {
		return m.countByTypeFn(ctx, clinicID, reservationTypeID, startTime, excludeID)
	}
	return 0, nil
}

func (m *mockTrimmingReservationRepository) AssertOwnerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockTrimmingReservationRepository) FindPetOwnerInClinic(ctx context.Context, clinicID, petID uint64) (uint64, error) {
	if m.findPetOwnerFn != nil {
		return m.findPetOwnerFn(ctx, clinicID, petID)
	}
	return 0, nil
}

func (m *mockTrimmingReservationRepository) FindPetByIDInClinic(_ context.Context, _, petID uint64) (*model.Pet, error) {
	if m.findPetByIDFn != nil {
		return m.findPetByIDFn(petID)
	}
	return &model.Pet{ID: petID, Status: model.PetStatusAlive}, nil
}

// コンパイル時インターフェース適合チェック
var _ TrimmingReservationRepository = (*mockTrimmingReservationRepository)(nil)

// --- mock: ReservationTypeRepository ---

type mockTrimmingReservationTypeRepository struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
}

func (m *mockTrimmingReservationTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationType{
		ID:       id,
		ClinicID: clinicID,
		Category: model.ReservationTypeCategoryTrimming,
		IsActive: true,
	}, nil
}

var _ ReservationTypeRepository = (*mockTrimmingReservationTypeRepository)(nil)

type mockTrimmingUnavailableTimeRepository struct {
	findAllFn func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
}

func (m *mockTrimmingUnavailableTimeRepository) FindAll(
	ctx context.Context,
	clinicID, reservationTypeID uint64,
) ([]model.ReservationTypeUnavailableTime, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, reservationTypeID)
	}
	return []model.ReservationTypeUnavailableTime{}, nil
}

var _ ReservationTypeUnavailableTimeRepository = (*mockTrimmingUnavailableTimeRepository)(nil)

// --- mock: AppointmentTrimmingDetailRepository ---

type mockTrimmingDetailRepository struct {
	findByAppointmentIDFn func(ctx context.Context, clinicID, appointmentID uint64) (*model.AppointmentTrimmingDetail, error)
	createFn              func(ctx context.Context, detail *model.AppointmentTrimmingDetail) error
	updateFn              func(ctx context.Context, detail *model.AppointmentTrimmingDetail) error
	setOptionsFn          func(ctx context.Context, clinicID, appointmentID uint64, optionIDs []uint64) error
}

func (m *mockTrimmingDetailRepository) FindByAppointmentID(ctx context.Context, clinicID, appointmentID uint64) (*model.AppointmentTrimmingDetail, error) {
	if m.findByAppointmentIDFn != nil {
		return m.findByAppointmentIDFn(ctx, clinicID, appointmentID)
	}
	return &model.AppointmentTrimmingDetail{AppointmentID: appointmentID}, nil
}

func (m *mockTrimmingDetailRepository) Create(ctx context.Context, detail *model.AppointmentTrimmingDetail) error {
	if m.createFn != nil {
		return m.createFn(ctx, detail)
	}
	return nil
}

func (m *mockTrimmingDetailRepository) Update(ctx context.Context, detail *model.AppointmentTrimmingDetail) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, detail)
	}
	return nil
}

func (m *mockTrimmingDetailRepository) SetOptions(ctx context.Context, clinicID, appointmentID uint64, optionIDs []uint64) error {
	if m.setOptionsFn != nil {
		return m.setOptionsFn(ctx, clinicID, appointmentID, optionIDs)
	}
	return nil
}

var _ AppointmentTrimmingDetailRepository = (*mockTrimmingDetailRepository)(nil)

// --- mock: Transactor（テスト用：fn を同一コンテキストで直接実行） ---

type mockTransactor struct {
	withTxErr error
	withTxFn  func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	if m.withTxErr != nil {
		return m.withTxErr
	}
	return fn(ctx)
}

// --- helpers ---

func newTrimmingTestService(reserv *mockTrimmingReservationRepository, detail *mockTrimmingDetailRepository) TrimmingService {
	return newTrimmingTestServiceWithReservationType(reserv, detail, &mockTrimmingReservationTypeRepository{})
}

func newTrimmingTestServiceWithReservationType(
	reserv *mockTrimmingReservationRepository,
	detail *mockTrimmingDetailRepository,
	reservationType *mockTrimmingReservationTypeRepository,
) TrimmingService {
	return withTrimmingTestActor(NewTrimmingServiceWithAudit(
		reserv,
		reservationType,
		newAcceptingTrimmingStaffRepository(),
		&mockTrimmingUnavailableTimeRepository{},
		detail,
		newActiveTrimmingCourseRepository(),
		newActiveTrimmingOptionRepository(),
		&mockTransactor{},
		noopTrimmingAuditTxLogger{},
	))
}

func newTrimmingTestServiceWithAvailability(
	reserv *mockTrimmingReservationRepository,
	detail *mockTrimmingDetailRepository,
	reservationType *mockTrimmingReservationTypeRepository,
	reservationStaff ReservationStaffRepository,
	unavailableTime ReservationTypeUnavailableTimeRepository,
) TrimmingService {
	return withTrimmingTestActor(NewTrimmingServiceWithAudit(
		reserv,
		reservationType,
		reservationStaff,
		unavailableTime,
		detail,
		newActiveTrimmingCourseRepository(),
		newActiveTrimmingOptionRepository(),
		&mockTransactor{},
		noopTrimmingAuditTxLogger{},
	))
}

func newAcceptingTrimmingStaffRepository() ReservationStaffRepository {
	return &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, ClinicID: clinicID, IsActive: true}, nil
		},
		supportsReservationTypeFn: func(_ context.Context, _, _, _ uint64) (bool, error) {
			return true, nil
		},
	}
}

func newActiveTrimmingCourseRepository() TrimmingCourseRepository {
	return &mockTrimmingCourseRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
			return &model.TrimmingCourse{ID: id, ClinicID: clinicID, IsActive: true}, nil
		},
	}
}

func newActiveTrimmingOptionRepository() TrimmingOptionRepository {
	return &mockTrimmingOptionRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
			return &model.TrimmingOption{ID: id, ClinicID: clinicID, IsActive: true}, nil
		},
	}
}

// --- tests ---

func TestTrimmingService_List(t *testing.T) {
	petID := uint64(10)
	ownerID := uint64(5)

	tests := []struct {
		name      string
		clinicID  uint64
		petID     *uint64
		ownerID   *uint64
		repoData  []model.Reservation
		repoTotal int64
		repoErr   error
		wantLen   int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:      "returns trimming list with total count",
			clinicID:  1,
			repoData:  []model.Reservation{{ID: 1, ClinicID: 1}, {ID: 2, ClinicID: 1}},
			repoTotal: 2,
			wantLen:   2,
			wantTotal: 2,
		},
		{
			name:      "filters by pet ID",
			clinicID:  1,
			petID:     &petID,
			repoData:  []model.Reservation{{ID: 1, ClinicID: 1, PetID: &petID}},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
		},
		{
			name:      "filters by owner ID",
			clinicID:  1,
			ownerID:   &ownerID,
			repoData:  []model.Reservation{{ID: 1, ClinicID: 1}},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
		},
		{
			name:      "returns empty list when no records exist",
			clinicID:  1,
			repoData:  []model.Reservation{},
			repoTotal: 0,
			wantLen:   0,
			wantTotal: 0,
		},
		{
			name:     "propagates repository error",
			clinicID: 1,
			repoErr:  errors.New("db connection error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reserv := &mockTrimmingReservationRepository{
				findAllByCategoryFn: func(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
					return tt.repoData, tt.repoTotal, tt.repoErr
				},
			}
			svc := newTrimmingTestService(reserv, &mockTrimmingDetailRepository{})

			appts, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, nil, nil, 1, 20)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, appts, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestTrimmingService_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoAppt *model.Reservation
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "returns appointment when found",
			clinicID: 1,
			id:       10,
			repoAppt: &model.Reservation{ID: 10, ClinicID: 1},
			wantErr:  false,
		},
		{
			name:     "returns not found error when record does not exist",
			clinicID: 1,
			id:       999,
			repoErr:  apperrors.WrapNotFound("appointment", "999"),
			wantErr:  true,
			wantNF:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reserv := &mockTrimmingReservationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					return tt.repoAppt, tt.repoErr
				},
			}
			svc := newTrimmingTestService(reserv, &mockTrimmingDetailRepository{})

			appt, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, appt)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, appt)
			}
		})
	}
}

func TestTrimmingService_GetByID_DetailFetchError(t *testing.T) {
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: 1}, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			return nil, errors.New("db connection error")
		},
	}
	svc := newTrimmingTestService(reserv, detail)

	appt, err := svc.GetByID(context.Background(), 1, 10)

	assert.Error(t, err)
	assert.Nil(t, appt)
}

func TestTrimmingService_Create_ValidationErrors(t *testing.T) {
	tests := []struct {
		name  string
		input CreateTrimmingInput
	}{
		{
			name: "returns error when start_time is zero",
			input: CreateTrimmingInput{
				ReservationTypeID: 1,
				EndTime:           time.Now().Add(time.Hour),
			},
		},
		{
			name: "returns error when end_time is zero",
			input: CreateTrimmingInput{
				ReservationTypeID: 1,
				StartTime:         time.Now(),
			},
		},
		{
			name: "returns error when reservation_route is invalid",
			input: CreateTrimmingInput{
				ReservationTypeID: 1,
				StartTime:         time.Now(),
				EndTime:           time.Now().Add(time.Hour),
				ReservationRoute:  ptrString("invalid_route"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reserv := &mockTrimmingReservationRepository{
				createFn: func(_ context.Context, _ *model.Reservation) error {
					t.Fatal("appointment must not be created when validation fails")
					return nil
				},
			}
			svc := newTrimmingTestService(reserv, &mockTrimmingDetailRepository{})

			appt, err := svc.Create(context.Background(), 1, &tt.input)

			assert.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
			assert.Nil(t, appt)
		})
	}
}

func TestTrimmingService_ValidateTrimmingReservationType_NilRepository(t *testing.T) {
	reserv := &mockTrimmingReservationRepository{
		createFn: func(_ context.Context, _ *model.Reservation) error {
			t.Fatal("appointment must not be created when reservation type repository is nil")
			return nil
		},
	}
	// reservationType repository を明示的に nil で渡し、interface が真に nil であることを検証する。
	svc := withTrimmingTestActor(NewTrimmingServiceWithAudit(reserv, nil, nil, nil, &mockTrimmingDetailRepository{}, nil, nil, &mockTransactor{}, noopTrimmingAuditTxLogger{}))

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		ReservationTypeID: 1,
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(time.Hour),
	})

	assert.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
	assert.Nil(t, appt)
}

func TestTrimmingService_Create_FailsClosedWithoutUnavailableTimeRepository(t *testing.T) {
	createCalled := false
	reserv := &mockTrimmingReservationRepository{
		createFn: func(_ context.Context, _ *model.Reservation) error {
			createCalled = true
			return nil
		},
	}
	svc := withTrimmingTestActor(NewTrimmingServiceWithAudit(
		reserv,
		&mockTrimmingReservationTypeRepository{},
		newAcceptingTrimmingStaffRepository(),
		nil,
		&mockTrimmingDetailRepository{},
		newActiveTrimmingCourseRepository(),
		newActiveTrimmingOptionRepository(),
		&mockTransactor{},
		noopTrimmingAuditTxLogger{},
	))

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		ReservationTypeID: 1,
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(time.Hour),
	})

	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
	assert.Nil(t, appt)
	assert.False(t, createCalled)
}

func TestTrimmingService_ValidateTrimmingReservationType_FindByIDError(t *testing.T) {
	reservationType := &mockTrimmingReservationTypeRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return nil, errors.New("db error")
		},
	}
	reserv := &mockTrimmingReservationRepository{
		createFn: func(_ context.Context, _ *model.Reservation) error {
			t.Fatal("appointment must not be created when reservation type lookup fails")
			return nil
		},
	}
	svc := newTrimmingTestServiceWithReservationType(reserv, &mockTrimmingDetailRepository{}, reservationType)

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		ReservationTypeID: 1,
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(time.Hour),
	})

	assert.Error(t, err)
	assert.Nil(t, appt)
}

func TestTrimmingService_Create(t *testing.T) {
	tests := []struct {
		name          string
		clinicID      uint64
		input         CreateTrimmingInput
		createErr     error
		detailErr     error
		setOptionsErr error
		wantErr       bool
	}{
		{
			name:     "creates trimming appointment successfully without options",
			clinicID: 1,
			input: CreateTrimmingInput{
				ReservationTypeID: 1,
				StartTime:         time.Now(),
				EndTime:           time.Now().Add(time.Hour),
				CourseID:          ptrUint64(1),
			},
		},
		{
			name:     "creates trimming appointment successfully with options",
			clinicID: 1,
			input: CreateTrimmingInput{
				ReservationTypeID: 1,
				StartTime:         time.Now(),
				EndTime:           time.Now().Add(time.Hour),
				OptionIDs:         []uint64{10, 20},
			},
		},
		{
			name:     "creates record shortcut trimming appointment",
			clinicID: 1,
			input: CreateTrimmingInput{
				ReservationTypeID: 1,
				StartTime:         time.Now(),
				EndTime:           time.Now().Add(time.Hour),
				Status:            model.ReservationStatusInConsultation,
				ReservationRoute:  ptrString("record_shortcut"),
			},
		},
		{
			name:      "returns error when appointment creation fails",
			clinicID:  1,
			input:     CreateTrimmingInput{ReservationTypeID: 1, StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)},
			createErr: errors.New("db error"),
			wantErr:   true,
		},
		{
			name:      "returns error when detail creation fails",
			clinicID:  1,
			input:     CreateTrimmingInput{ReservationTypeID: 1, StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)},
			detailErr: errors.New("detail error"),
			wantErr:   true,
		},
		{
			name:          "returns error when SetOptions fails",
			clinicID:      1,
			input:         CreateTrimmingInput{ReservationTypeID: 1, StartTime: time.Now(), EndTime: time.Now().Add(time.Hour), OptionIDs: []uint64{10}},
			setOptionsErr: errors.New("set options error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reserv := &mockTrimmingReservationRepository{
				createFn: func(_ context.Context, a *model.Reservation) error {
					a.ID = 1
					if tt.input.ReservationRoute != nil {
						assert.Equal(t, tt.input.ReservationRoute, a.ReservationRoute)
					}
					return tt.createErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					return &model.Reservation{ID: 1, ClinicID: 1}, nil
				},
			}
			detail := &mockTrimmingDetailRepository{
				createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
					return tt.detailErr
				},
				setOptionsFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
					return tt.setOptionsErr
				},
			}
			svc := newTrimmingTestService(reserv, detail)

			appt, err := svc.Create(context.Background(), tt.clinicID, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, appt)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, appt)
			}
		})
	}
}

func TestTrimmingService_Create_RejectsNonTrimmingReservationType(t *testing.T) {
	createCalled := false
	reserv := &mockTrimmingReservationRepository{
		createFn: func(_ context.Context, _ *model.Reservation) error {
			createCalled = true
			return nil
		},
	}
	reservationType := &mockTrimmingReservationTypeRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{
				ID:       id,
				ClinicID: clinicID,
				Category: model.ReservationTypeCategoryGeneral,
			}, nil
		},
	}
	svc := newTrimmingTestServiceWithReservationType(reserv, &mockTrimmingDetailRepository{}, reservationType)

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		ReservationTypeID: 1,
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(time.Hour),
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, appt)
	assert.False(t, createCalled)
}

func TestTrimmingService_Create_RejectsInactiveReservationType(t *testing.T) {
	createCalled := false
	reserv := &mockTrimmingReservationRepository{
		createFn: func(_ context.Context, _ *model.Reservation) error {
			createCalled = true
			return nil
		},
	}
	reservationType := &mockTrimmingReservationTypeRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
			return &model.ReservationType{
				ID:       id,
				ClinicID: clinicID,
				Category: model.ReservationTypeCategoryTrimming,
				IsActive: false,
			}, nil
		},
	}
	svc := newTrimmingTestServiceWithReservationType(reserv, &mockTrimmingDetailRepository{}, reservationType)

	appointment, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		ReservationTypeID: 1,
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(time.Hour),
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, appointment)
	assert.False(t, createCalled)
}

func TestTrimmingService_Create_RechecksSlotAndCapacityInsideTransaction(t *testing.T) {
	staffID := uint64(33)
	start := time.Now().UTC().Add(time.Hour)

	t.Run("doctor conflict blocks a new appointment", func(t *testing.T) {
		lockCalled := false
		reserv := &mockTrimmingReservationRepository{
			acquireBookingLockFn: func(_ context.Context, clinicID uint64) error {
				lockCalled = true
				assert.Equal(t, uint64(1), clinicID)
				return nil
			},
			hasDoctorConflictFn: func(_ context.Context, _ uint64, doctorID uint64, _, _ time.Time, excludeID *uint64) (bool, error) {
				assert.Equal(t, staffID, doctorID)
				assert.Nil(t, excludeID)
				return true, nil
			},
			createFn: func(_ context.Context, _ *model.Reservation) error {
				t.Fatal("appointment must not be created when the slot conflicts")
				return nil
			},
		}
		svc := newTrimmingTestService(reserv, &mockTrimmingDetailRepository{})

		appointment, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
			ReservationTypeID: 1,
			StartTime:         start,
			EndTime:           start.Add(time.Hour),
			StaffID:           &staffID,
		})

		assert.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Nil(t, appointment)
		assert.True(t, lockCalled)
	})

	t.Run("capacity conflict blocks a new appointment", func(t *testing.T) {
		maxConcurrent := 1
		reserv := &mockTrimmingReservationRepository{
			countByTypeFn: func(_ context.Context, _, _ uint64, _ time.Time, excludeID *uint64) (int64, error) {
				assert.Nil(t, excludeID)
				return 1, nil
			},
			createFn: func(_ context.Context, _ *model.Reservation) error {
				t.Fatal("appointment must not be created when capacity is full")
				return nil
			},
		}
		reservationType := &mockTrimmingReservationTypeRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
				return &model.ReservationType{
					ID:            id,
					ClinicID:      clinicID,
					Category:      model.ReservationTypeCategoryTrimming,
					IsActive:      true,
					MaxConcurrent: &maxConcurrent,
				}, nil
			},
		}
		svc := newTrimmingTestServiceWithReservationType(reserv, &mockTrimmingDetailRepository{}, reservationType)

		appointment, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
			ReservationTypeID: 1,
			StartTime:         start,
			EndTime:           start.Add(time.Hour),
		})

		assert.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Nil(t, appointment)
	})
}

func TestTrimmingService_Create_RejectsExcludedStaff(t *testing.T) {
	staffID := uint64(33)
	reserv := &mockTrimmingReservationRepository{
		createFn: func(_ context.Context, _ *model.Reservation) error {
			t.Fatal("trimming appointment must not be created when staff cannot handle reservation type")
			return nil
		},
	}
	staffRepo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, staffID, id)
			return &model.Staff{ID: id, IsActive: true}, nil
		},
		supportsReservationTypeFn: func(_ context.Context, clinicID, actualStaffID, reservationTypeID uint64) (bool, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, staffID, actualStaffID)
			assert.Equal(t, uint64(9), reservationTypeID)
			return false, nil
		},
	}
	svc := newTrimmingTestServiceWithAvailability(
		reserv,
		&mockTrimmingDetailRepository{},
		&mockTrimmingReservationTypeRepository{},
		staffRepo,
		nil,
	)

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		ReservationTypeID: 9,
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(time.Hour),
		StaffID:           &staffID,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, appt)
}

func TestTrimmingService_Create_RevalidatesStaffCapabilityInsideTransaction(t *testing.T) {
	type txContextKey struct{}
	staffID := uint64(33)
	var findInTx, capabilityInTx bool
	staffRepo := &mockReservationStaffRepository{
		findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, staffID, id)
			if ctx.Value(txContextKey{}) == true {
				findInTx = true
			}
			return &model.Staff{ID: id, IsActive: true}, nil
		},
		supportsReservationTypeFn: func(ctx context.Context, clinicID, actualStaffID, reservationTypeID uint64) (bool, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, staffID, actualStaffID)
			assert.Equal(t, uint64(9), reservationTypeID)
			if ctx.Value(txContextKey{}) == true {
				capabilityInTx = true
			}
			return true, nil
		},
	}
	tx := &mockTransactor{
		withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(context.WithValue(ctx, txContextKey{}, true))
		},
	}
	reservationRepo := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID}, nil
		},
	}
	svc := withTrimmingTestActor(NewTrimmingServiceWithAudit(
		reservationRepo,
		&mockTrimmingReservationTypeRepository{},
		staffRepo,
		&mockTrimmingUnavailableTimeRepository{},
		&mockTrimmingDetailRepository{},
		nil,
		nil,
		tx,
		noopTrimmingAuditTxLogger{},
	))

	appointment, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		ReservationTypeID: 9,
		StartTime:         time.Now().UTC().Add(time.Hour),
		EndTime:           time.Now().UTC().Add(2 * time.Hour),
		StaffID:           &staffID,
	})

	require.NoError(t, err)
	require.NotNil(t, appointment)
	assert.True(t, findInTx, "staff assignment must be rechecked in the write transaction")
	assert.True(t, capabilityInTx, "staff capability must be rechecked in the write transaction")
}

func TestTrimmingService_Create_AcquiresBookingLockBeforeTransactionalValidation(t *testing.T) {
	type txContextKey struct{}
	staffID := uint64(33)
	petID := uint64(44)
	events := make([]string, 0, 4)
	recordInTx := func(ctx context.Context, event string) {
		if ctx.Value(txContextKey{}) == true {
			events = append(events, event)
		}
	}

	reservationRepo := &mockTrimmingReservationRepository{
		acquireBookingLockFn: func(ctx context.Context, _ uint64) error {
			recordInTx(ctx, "booking_lock")
			return nil
		},
		findPetOwnerFn: func(ctx context.Context, _, _ uint64) (uint64, error) {
			recordInTx(ctx, "pet_validation")
			return 55, nil
		},
		createFn: func(_ context.Context, appointment *model.Reservation) error {
			appointment.ID = 66
			return nil
		},
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID}, nil
		},
	}
	staffRepo := &mockReservationStaffRepository{
		findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
			recordInTx(ctx, "staff_validation")
			return &model.Staff{ID: id, ClinicID: clinicID, IsActive: true}, nil
		},
		supportsReservationTypeFn: func(ctx context.Context, _, _, _ uint64) (bool, error) {
			recordInTx(ctx, "staff_capability")
			return true, nil
		},
	}
	tx := &mockTransactor{
		withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(context.WithValue(ctx, txContextKey{}, true))
		},
	}
	svc := withTrimmingTestActor(NewTrimmingServiceWithAudit(
		reservationRepo,
		&mockTrimmingReservationTypeRepository{},
		staffRepo,
		&mockTrimmingUnavailableTimeRepository{},
		&mockTrimmingDetailRepository{},
		nil,
		nil,
		tx,
		noopTrimmingAuditTxLogger{},
	))
	start := time.Now().UTC().Add(time.Hour)

	appointment, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		ReservationTypeID: 9,
		StartTime:         start,
		EndTime:           start.Add(time.Hour),
		PetID:             &petID,
		StaffID:           &staffID,
	})

	require.NoError(t, err)
	require.NotNil(t, appointment)
	assert.Equal(t, []string{"booking_lock", "pet_validation", "staff_validation", "staff_capability"}, events)
}

func TestTrimmingService_Create_MissingCourseRepositoryFailsClosed(t *testing.T) {
	courseID := uint64(41)
	appointmentCreated := false
	detailCreated := false
	reservationRepo := &mockTrimmingReservationRepository{
		createFn: func(_ context.Context, _ *model.Reservation) error {
			appointmentCreated = true
			return nil
		},
	}
	detailRepo := &mockTrimmingDetailRepository{
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			detailCreated = true
			return nil
		},
	}
	svc := withTrimmingTestActor(NewTrimmingServiceWithAudit(
		reservationRepo,
		&mockTrimmingReservationTypeRepository{},
		nil, &mockTrimmingUnavailableTimeRepository{},
		detailRepo,
		nil,
		nil,
		&mockTransactor{},
		noopTrimmingAuditTxLogger{},
	))
	start := time.Now().UTC().Add(time.Hour)

	got, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		ReservationTypeID: 9,
		StartTime:         start,
		EndTime:           start.Add(time.Hour),
		CourseID:          &courseID,
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.False(t, appointmentCreated)
	assert.False(t, detailCreated)
}

func TestTrimmingService_Update_MissingOptionRepositoryFailsClosed(t *testing.T) {
	optionIDs := []uint64{52}
	detailUpdated := false
	optionsWritten := false
	reservationRepo := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID: id, ClinicID: clinicID, ReservationTypeID: 9,
				Status:          model.ReservationStatusPending,
				ReservationType: &model.ReservationType{ID: 9, Category: model.ReservationTypeCategoryTrimming},
			}, nil
		},
	}
	detailRepo := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, clinicID, appointmentID uint64) (*model.AppointmentTrimmingDetail, error) {
			return &model.AppointmentTrimmingDetail{ClinicID: clinicID, AppointmentID: appointmentID}, nil
		},
		updateFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			detailUpdated = true
			return nil
		},
		setOptionsFn: func(_ context.Context, _ uint64, _ uint64, _ []uint64) error {
			optionsWritten = true
			return nil
		},
	}
	svc := withTrimmingTestActor(NewTrimmingServiceWithAudit(
		reservationRepo,
		&mockTrimmingReservationTypeRepository{},
		nil, nil,
		detailRepo,
		nil,
		nil,
		&mockTransactor{},
		noopTrimmingAuditTxLogger{},
	))

	got, err := svc.Update(context.Background(), 1, 77, &UpdateTrimmingInput{OptionIDs: &optionIDs})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.False(t, detailUpdated)
	assert.False(t, optionsWritten)
}

func TestTrimmingService_Create_ExistingAppointment(t *testing.T) {
	appointmentID := uint64(77)
	petID := uint64(10)
	courseID := uint64(4)
	category := model.ReservationTypeCategoryTrimming
	detailCreated := false
	appointmentUpdated := false

	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:                id,
				ClinicID:          1,
				PetID:             &petID,
				Status:            model.ReservationStatusCheckedIn,
				ReservationTypeID: 9,
				ReservationType:   &model.ReservationType{ID: 9, Category: category},
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _ uint64, id uint64, fields map[string]any) (*model.Reservation, error) {
			appointmentUpdated = true
			assert.Equal(t, appointmentID, id)
			assert.NotContains(t, fields, "start_time")
			assert.NotContains(t, fields, "end_time")
			assert.NotContains(t, fields, "status")
			return &model.Reservation{ID: id, ClinicID: 1}, nil
		},
	}
	detailCalls := 0
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _ uint64, id uint64) (*model.AppointmentTrimmingDetail, error) {
			detailCalls++
			if detailCalls == 1 {
				return nil, apperrors.WrapNotFound("appointment_trimming_detail", "missing")
			}
			return &model.AppointmentTrimmingDetail{AppointmentID: id}, nil
		},
		createFn: func(_ context.Context, detail *model.AppointmentTrimmingDetail) error {
			detailCreated = true
			assert.Equal(t, appointmentID, detail.AppointmentID)
			assert.Equal(t, &courseID, detail.CourseID)
			return nil
		},
	}
	svc := newTrimmingTestService(reserv, detail)

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		AppointmentID:     &appointmentID,
		ReservationTypeID: 9,
		PetID:             &petID,
		CourseID:          &courseID,
	})

	assert.NoError(t, err)
	assert.NotNil(t, appt)
	assert.True(t, detailCreated)
	assert.True(t, appointmentUpdated, "detail追加writeでも既存appointmentの欠損ownerを補完する")
}

func TestTrimmingService_Create_ExistingAppointmentRechecksSlotBeforeMutation(t *testing.T) {
	appointmentID := uint64(77)
	staffID := uint64(33)
	start := time.Now().UTC().Add(time.Hour)
	category := model.ReservationTypeCategoryTrimming
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:                id,
				ClinicID:          1,
				ReservationTypeID: 9,
				ReservationType:   &model.ReservationType{ID: 9, Category: category, IsActive: true},
				StartTime:         start,
				EndTime:           start.Add(time.Hour),
				Status:            model.ReservationStatusPending,
			}, nil
		},
		hasDoctorConflictFn: func(_ context.Context, _ uint64, doctorID uint64, _, _ time.Time, excludeID *uint64) (bool, error) {
			assert.Equal(t, staffID, doctorID)
			require.NotNil(t, excludeID)
			assert.Equal(t, appointmentID, *excludeID)
			return true, nil
		},
		updateFieldsFn: func(_ context.Context, _ uint64, _ uint64, _ map[string]any) (*model.Reservation, error) {
			t.Fatal("appointment must not be updated when the slot conflicts")
			return nil, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			return nil, apperrors.WrapNotFound("appointment_trimming_detail", "missing")
		},
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			t.Fatal("detail must not be created when the slot conflicts")
			return nil
		},
	}
	svc := newTrimmingTestService(reserv, detail)

	appointment, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		AppointmentID: &appointmentID,
		StaffID:       &staffID,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Nil(t, appointment)
}

func TestTrimmingService_Create_ExistingAppointmentRejectsPetMismatch(t *testing.T) {
	appointmentID := uint64(77)
	appointmentPetID := uint64(10)
	requestPetID := uint64(99)
	category := model.ReservationTypeCategoryTrimming

	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:              id,
				ClinicID:        1,
				PetID:           &appointmentPetID,
				ReservationType: &model.ReservationType{ID: 9, Category: category},
			}, nil
		},
	}
	detailCalls := 0
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			detailCalls++
			return nil, apperrors.WrapNotFound("appointment_trimming_detail", "missing")
		},
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			t.Fatal("detail should not be created when pet_id mismatches appointment")
			return nil
		},
	}
	svc := newTrimmingTestService(reserv, detail)

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		AppointmentID: &appointmentID,
		PetID:         &requestPetID,
	})

	assert.Error(t, err)
	assert.Nil(t, appt)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Equal(t, 1, detailCalls)
}

func TestTrimmingService_Create_ExistingAppointmentFillsMissingPet(t *testing.T) {
	appointmentID := uint64(77)
	requestPetID := uint64(10)
	category := model.ReservationTypeCategoryTrimming
	var updatedFields map[string]any

	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:              id,
				ClinicID:        1,
				ReservationType: &model.ReservationType{ID: 9, Category: category},
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _ uint64, id uint64, fields map[string]any) (*model.Reservation, error) {
			assert.Equal(t, appointmentID, id)
			updatedFields = fields
			return &model.Reservation{ID: id, ClinicID: 1}, nil
		},
	}
	detailCalls := 0
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _ uint64, id uint64) (*model.AppointmentTrimmingDetail, error) {
			detailCalls++
			if detailCalls == 1 {
				return nil, apperrors.WrapNotFound("appointment_trimming_detail", "missing")
			}
			return &model.AppointmentTrimmingDetail{AppointmentID: id}, nil
		},
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			return nil
		},
	}
	svc := newTrimmingTestService(reserv, detail)

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		AppointmentID: &appointmentID,
		PetID:         &requestPetID,
	})

	assert.NoError(t, err)
	assert.NotNil(t, appt)
	assert.Equal(t, map[string]any{"pet_id": requestPetID}, updatedFields)
}

func TestTrimmingService_Create_ExistingAppointment_NotTrimmingType(t *testing.T) {
	appointmentID := uint64(77)
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:              id,
				ClinicID:        1,
				ReservationType: &model.ReservationType{ID: 9, Category: model.ReservationTypeCategoryGeneral},
			}, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			t.Fatal("detail must not be created for a non-trimming appointment")
			return nil
		},
	}
	svc := newTrimmingTestService(reserv, detail)

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		AppointmentID: &appointmentID,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, appt)
}

func TestTrimmingService_Create_ExistingAppointment_NilReservationType(t *testing.T) {
	appointmentID := uint64(77)
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: 1, ReservationType: nil}, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			t.Fatal("detail must not be created when appointment has no reservation type")
			return nil
		},
	}
	svc := newTrimmingTestService(reserv, detail)

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		AppointmentID: &appointmentID,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, appt)
}

func TestTrimmingService_Create_ExistingAppointment_AlreadyHasDetail(t *testing.T) {
	appointmentID := uint64(77)
	category := model.ReservationTypeCategoryTrimming
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:              id,
				ClinicID:        1,
				ReservationType: &model.ReservationType{ID: 9, Category: category},
			}, nil
		},
	}
	detailCreated := false
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _ uint64, id uint64) (*model.AppointmentTrimmingDetail, error) {
			return &model.AppointmentTrimmingDetail{AppointmentID: id}, nil
		},
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			detailCreated = true
			return nil
		},
	}
	svc := newTrimmingTestService(reserv, detail)

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		AppointmentID: &appointmentID,
	})

	// TRM-04: existing detail must conflict rather than silently return 201.
	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "want conflict for existing detail: %v", err)
	assert.Nil(t, appt)
	assert.False(t, detailCreated, "既に detail が存在する場合は再作成しない")
}

func TestTrimmingService_Create_ExistingAppointment_DetailLookupError(t *testing.T) {
	appointmentID := uint64(77)
	category := model.ReservationTypeCategoryTrimming
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:              id,
				ClinicID:        1,
				ReservationType: &model.ReservationType{ID: 9, Category: category},
			}, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			return nil, errors.New("db connection error")
		},
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			t.Fatal("detail must not be created when existing-detail lookup fails")
			return nil
		},
	}
	svc := newTrimmingTestService(reserv, detail)

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		AppointmentID: &appointmentID,
	})

	assert.Error(t, err)
	assert.Nil(t, appt)
}

func TestTrimmingService_Create_ExistingAppointment_StaffCannotHandleReservationType(t *testing.T) {
	appointmentID := uint64(77)
	staffID := uint64(33)
	category := model.ReservationTypeCategoryTrimming
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:                id,
				ClinicID:          1,
				ReservationTypeID: 9,
				ReservationType:   &model.ReservationType{ID: 9, Category: category},
			}, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			return nil, apperrors.WrapNotFound("appointment_trimming_detail", "missing")
		},
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			t.Fatal("detail must not be created when staff cannot handle reservation type")
			return nil
		},
	}
	staffRepo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, IsActive: true}, nil
		},
		supportsReservationTypeFn: func(_ context.Context, _, _, _ uint64) (bool, error) {
			return false, nil
		},
	}
	svc := newTrimmingTestServiceWithAvailability(
		reserv,
		detail,
		&mockTrimmingReservationTypeRepository{},
		staffRepo,
		nil,
	)

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		AppointmentID: &appointmentID,
		StaffID:       &staffID,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, appt)
}

func TestTrimmingService_Create_ExistingAppointment_TimeRangeInvalid(t *testing.T) {
	appointmentID := uint64(77)
	category := model.ReservationTypeCategoryTrimming
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:                id,
				ClinicID:          1,
				ReservationTypeID: 9,
				ReservationType:   &model.ReservationType{ID: 9, Category: category},
				StartTime:         time.Now(),
				EndTime:           time.Now().Add(time.Hour),
			}, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		// mockTrimmingDetailRepository.FindByAppointmentID のデフォルト実装は非nilのdetailを
		// 返すため、未設定のままだと「既存detailあり」経路に入り GetByID で早期returnしてしまい
		// 時間帯バリデーションを一切通らない。「detail未作成」を明示的に模擬する。
		findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			return nil, apperrors.WrapNotFound("appointment_trimming_detail", "missing")
		},
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			t.Fatal("detail must not be created when the requested time range is invalid")
			return nil
		},
	}
	svc := newTrimmingTestService(reserv, detail)

	// end_time が start_time より前 → reservation.ValidateReservationTypeAvailableTime 内の
	// validateTimeRange がエラーを返す。
	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		AppointmentID: &appointmentID,
		StartTime:     time.Now().Add(time.Hour),
		EndTime:       time.Now(),
	})

	assert.Error(t, err)
	assert.Nil(t, appt)
}

func TestTrimmingService_Create_ExistingAppointment_ResolvesPartialTimeRange(t *testing.T) {
	appointmentID := uint64(77)
	category := model.ReservationTypeCategoryTrimming
	existingStart := time.Now()
	existingEnd := existingStart.Add(2 * time.Hour)
	// newStart は既存の start_time より後だが、end_time (existingEnd) より前でなければならない。
	// end_time は未指定のため appt.EndTime に解決され、resolvedStart < resolvedEnd を
	// 満たす必要がある（reservation.ValidateReservationTypeAvailableTime 内の validateTimeRange）。
	newStart := existingStart.Add(30 * time.Minute)

	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:                id,
				ClinicID:          1,
				ReservationTypeID: 9,
				ReservationType:   &model.ReservationType{ID: 9, Category: category},
				StartTime:         existingStart,
				EndTime:           existingEnd,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _ uint64, id uint64, fields map[string]any) (*model.Reservation, error) {
			// start_time のみ指定 → end_time は既存予約の end_time に解決されるはず。
			assert.Equal(t, newStart, fields["start_time"])
			assert.NotContains(t, fields, "end_time")
			return &model.Reservation{ID: id, ClinicID: 1}, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			return nil, apperrors.WrapNotFound("appointment_trimming_detail", "missing")
		},
		createFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
			return nil
		},
	}
	svc := newTrimmingTestService(reserv, detail)

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		AppointmentID: &appointmentID,
		StartTime:     newStart,
	})

	assert.NoError(t, err)
	assert.NotNil(t, appt)
}

func TestTrimmingService_Update(t *testing.T) {
	ptrStr := func(s string) *string { return &s }
	optionIDs := []uint64{10, 20}
	emptyIDs := []uint64{}

	tests := []struct {
		name            string
		clinicID        uint64
		id              uint64
		input           UpdateTrimmingInput
		updateFieldsErr error
		detailFindErr   error
		detailUpdateErr error
		setOptionsErr   error
		wantErr         bool
	}{
		{
			name:     "updates appointment fields only",
			clinicID: 1,
			id:       10,
			input: UpdateTrimmingInput{
				StyleRequest: ptrStr("short cut"),
			},
		},
		{
			name:     "updates with options",
			clinicID: 1,
			id:       10,
			input: UpdateTrimmingInput{
				OptionIDs: &optionIDs,
			},
		},
		{
			name:     "clears all options with empty slice",
			clinicID: 1,
			id:       10,
			input: UpdateTrimmingInput{
				OptionIDs: &emptyIDs,
			},
		},
		{
			name:     "returns error when appointment update fails",
			clinicID: 1,
			id:       10,
			input: UpdateTrimmingInput{
				// Status を含めることで apptFields が非空になり Update が呼ばれる
				Status: func() *model.ReservationStatus { s := model.ReservationStatusConfirmed; return &s }(),
			},
			updateFieldsErr: errors.New("db error"),
			wantErr:         true,
		},
		{
			name:          "returns error when trimming detail not found",
			clinicID:      1,
			id:            10,
			input:         UpdateTrimmingInput{StyleRequest: ptrStr("x")},
			detailFindErr: apperrors.WrapNotFound("appointment_trimming_detail", "appointment_id=10"),
			wantErr:       true,
		},
		{
			name:            "returns error when trimming detail update fails",
			clinicID:        1,
			id:              10,
			input:           UpdateTrimmingInput{StyleRequest: ptrStr("x")},
			detailUpdateErr: errors.New("update error"),
			wantErr:         true,
		},
		{
			name:          "returns error when SetOptions fails",
			clinicID:      1,
			id:            10,
			input:         UpdateTrimmingInput{OptionIDs: &optionIDs},
			setOptionsErr: errors.New("set options error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reserv := &mockTrimmingReservationRepository{
				updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.Reservation, error) {
					if tt.updateFieldsErr != nil {
						return nil, tt.updateFieldsErr
					}
					return &model.Reservation{ID: id}, nil
				},
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
					return &model.Reservation{ID: id, ClinicID: clinicID}, nil
				},
			}
			detail := &mockTrimmingDetailRepository{
				findByAppointmentIDFn: func(_ context.Context, _, apptID uint64) (*model.AppointmentTrimmingDetail, error) {
					if tt.detailFindErr != nil {
						return nil, tt.detailFindErr
					}
					return &model.AppointmentTrimmingDetail{AppointmentID: apptID}, nil
				},
				updateFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
					return tt.detailUpdateErr
				},
				setOptionsFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
					return tt.setOptionsErr
				},
			}
			svc := newTrimmingTestService(reserv, detail)

			appt, err := svc.Update(context.Background(), tt.clinicID, tt.id, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, appt)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, appt)
			}
		})
	}
}

func TestTrimmingService_Update_DetailOnlyBackfillsMissingOwner(t *testing.T) {
	appointmentID := uint64(10)
	petID := uint64(20)
	styleRequest := "summer cut"
	appointmentUpdateCalls := 0
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID, PetID: &petID, ReservationTypeID: 9}, nil
		},
		lockAndFindByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID: id, ClinicID: clinicID, PetID: &petID, ReservationTypeID: 9,
				Status: model.ReservationStatusPending,
			}, nil
		},
		findPetOwnerFn: func(_ context.Context, clinicID, actualPetID uint64) (uint64, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, petID, actualPetID)
			return 30, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
			appointmentUpdateCalls++
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, appointmentID, id)
			assert.Empty(t, fields, "owner補完はreservation owner側のempty typed commandで行う")
			return &model.Reservation{ID: id, ClinicID: clinicID, PetID: &petID}, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, clinicID, id uint64) (*model.AppointmentTrimmingDetail, error) {
			return &model.AppointmentTrimmingDetail{ClinicID: clinicID, AppointmentID: id}, nil
		},
	}
	svc := newTrimmingTestService(reserv, detail)

	got, err := svc.Update(context.Background(), 1, appointmentID, &UpdateTrimmingInput{StyleRequest: &styleRequest})

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 1, appointmentUpdateCalls)
}

func TestTrimmingService_Update_RejectsTerminalDetailMutation(t *testing.T) {
	statuses := []model.ReservationStatus{
		model.ReservationStatusCompleted,
		model.ReservationStatusCancelled,
		model.ReservationStatusNoShow,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			appointmentID := uint64(10)
			reserv := &mockTrimmingReservationRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
					return &model.Reservation{
						ID:                id,
						ClinicID:          clinicID,
						ReservationTypeID: 9,
						Status:            status,
					}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
					t.Fatal("terminal appointment must not be updated through trimming")
					return nil, nil
				},
			}
			detail := &mockTrimmingDetailRepository{
				findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
					t.Fatal("terminal trimming detail must not be read for mutation")
					return nil, nil
				},
				updateFn: func(_ context.Context, _ *model.AppointmentTrimmingDetail) error {
					t.Fatal("terminal trimming detail must not be updated")
					return nil
				},
			}
			svc := newTrimmingTestService(reserv, detail)
			style := "must not change"

			appointment, err := svc.Update(context.Background(), 1, appointmentID, &UpdateTrimmingInput{
				StyleRequest: &style,
			})

			assert.Error(t, err)
			assert.True(t, apperrors.IsConflict(err))
			assert.Nil(t, appointment)
		})
	}
}

func TestTrimmingService_Update_FindByIDError(t *testing.T) {
	ptrStr := func(s string) *string { return &s }
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
			return nil, apperrors.WrapNotFound("appointment", "999")
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			t.Fatal("appointment must not be updated when the parent lookup fails")
			return nil, nil
		},
	}
	svc := newTrimmingTestService(reserv, &mockTrimmingDetailRepository{})

	appt, err := svc.Update(context.Background(), 1, 999, &UpdateTrimmingInput{StyleRequest: ptrStr("x")})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, appt)
}

func TestTrimmingService_Update_RejectsSlotConflictBeforeWrite(t *testing.T) {
	doctorID := uint64(31)
	appointmentID := uint64(44)
	start := time.Date(2026, 7, 22, 1, 0, 0, 0, time.UTC)
	newStart := start.Add(time.Hour)
	newEnd := newStart.Add(30 * time.Minute)
	conflictChecked := false
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{
				ID:                id,
				ClinicID:          clinicID,
				ReservationTypeID: 9,
				DoctorID:          &doctorID,
				StartTime:         start,
				EndTime:           start.Add(30 * time.Minute),
				Status:            model.ReservationStatusConfirmed,
			}, nil
		},
		hasDoctorConflictFn: func(_ context.Context, clinicID, actualDoctorID uint64, actualStart, actualEnd time.Time, excludeID *uint64) (bool, error) {
			conflictChecked = true
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, doctorID, actualDoctorID)
			assert.Equal(t, newStart, actualStart)
			assert.Equal(t, newEnd, actualEnd)
			if assert.NotNil(t, excludeID) {
				assert.Equal(t, appointmentID, *excludeID)
			}
			return true, nil
		},
		updateFieldsFn: func(_ context.Context, _ uint64, _ uint64, _ map[string]any) (*model.Reservation, error) {
			t.Fatal("appointment must not be updated when the requested slot conflicts")
			return nil, nil
		},
	}
	detail := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			t.Fatal("trimming detail must not be touched when the appointment slot conflicts")
			return nil, nil
		},
	}
	svc := newTrimmingTestServiceWithAvailability(
		reserv,
		detail,
		&mockTrimmingReservationTypeRepository{},
		nil,
		&mockTrimmingUnavailableTimeRepository{},
	)

	appointment, err := svc.Update(context.Background(), 1, appointmentID, &UpdateTrimmingInput{
		StartTime: &newStart,
		EndTime:   &newEnd,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Nil(t, appointment)
	assert.True(t, conflictChecked)
}

func TestTrimmingService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "deletes trimming appointment successfully",
			clinicID: 1,
			id:       10,
		},
		{
			name:     "returns not found error",
			clinicID: 1,
			id:       999,
			repoErr:  apperrors.WrapNotFound("appointment", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reserv := &mockTrimmingReservationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Reservation{ID: 1, ClinicID: 1}, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := newTrimmingTestService(reserv, &mockTrimmingDetailRepository{})

			err := svc.Delete(context.Background(), tt.clinicID, tt.id, ptrUint64(42))

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestTrimmingService_Delete_RepositoryFailureAfterFindSucceeds は typed delete operation の
// 非NotFoundエラーをNotFoundへ誤分類しないことを検証する（旧テスト名は履歴互換で維持）。
func TestTrimmingService_Delete_RepositoryFailureAfterFindSucceeds(t *testing.T) {
	reserv := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: 1}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("db error")
		},
	}
	svc := newTrimmingTestService(reserv, &mockTrimmingDetailRepository{})

	err := svc.Delete(context.Background(), 1, 10, ptrUint64(42))

	assert.Error(t, err)
	assert.False(t, apperrors.IsNotFound(err))
}

func TestTrimmingService_Delete_RunsTypedDeleteInsideTransaction(t *testing.T) {
	type txMarkerKey struct{}
	marker := &struct{}{}
	lockCalled := false
	deleteCalled := false
	reserv := &mockTrimmingReservationRepository{
		acquireBookingLockFn: func(ctx context.Context, clinicID uint64) error {
			lockCalled = true
			assert.Equal(t, uint64(3), clinicID)
			assert.Same(t, marker, ctx.Value(txMarkerKey{}))
			return nil
		},
		lockAndFindByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
			assert.Equal(t, uint64(3), clinicID)
			assert.Equal(t, uint64(27), id)
			assert.Same(t, marker, ctx.Value(txMarkerKey{}))
			return &model.Reservation{ID: id, ClinicID: clinicID}, nil
		},
		deleteFn: func(ctx context.Context, clinicID, id uint64) error {
			deleteCalled = true
			assert.Equal(t, uint64(3), clinicID)
			assert.Equal(t, uint64(27), id)
			assert.Same(t, marker, ctx.Value(txMarkerKey{}))
			return nil
		},
	}
	transactor := &mockTransactor{
		withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(context.WithValue(ctx, txMarkerKey{}, marker))
		},
	}
	svc := withTrimmingTestActor(NewTrimmingServiceWithAudit(
		reserv,
		&mockTrimmingReservationTypeRepository{},
		nil,
		nil,
		&mockTrimmingDetailRepository{},
		nil,
		nil,
		transactor,
		noopTrimmingAuditTxLogger{},
	))

	err := svc.Delete(context.Background(), 3, 27, ptrUint64(42))

	assert.NoError(t, err)
	assert.True(t, lockCalled)
	assert.True(t, deleteCalled)
}
