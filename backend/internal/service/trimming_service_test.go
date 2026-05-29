package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// --- mock: ReservationRepository（FindAllByCategory 特化） ---

type mockTrimmingReservationRepository struct {
	findAllByCategoryFn func(ctx context.Context, clinicID uint64, category model.ReservationTypeCategory, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error)
	findByIDFn          func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	createFn            func(ctx context.Context, appt *model.Reservation) error
	updateFieldsFn      func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error)
	deleteFn            func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockTrimmingReservationRepository) FindAllByCategory(ctx context.Context, clinicID uint64, category model.ReservationTypeCategory, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error) {
	return m.findAllByCategoryFn(ctx, clinicID, category, petID, ownerID, startDate, endDate, page, limit)
}

func (m *mockTrimmingReservationRepository) FindAll(_ context.Context, _ uint64, _, _ int, _ *time.Time, _, _ *string, _, _ *uint64) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}

func (m *mockTrimmingReservationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, apperrors.WrapNotFound("appointment", "0")
}

func (m *mockTrimmingReservationRepository) Create(ctx context.Context, appt *model.Reservation) error {
	if m.createFn != nil {
		return m.createFn(ctx, appt)
	}
	return nil
}

func (m *mockTrimmingReservationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.Reservation{ID: id, ClinicID: clinicID}, nil
}

func (m *mockTrimmingReservationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockTrimmingReservationRepository) ExistsByReservationTypeID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockTrimmingReservationRepository) ExistsByStaffID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockTrimmingReservationRepository) CountMedicalRecordsByReservationID(_ context.Context, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockTrimmingReservationRepository) LockAndFindByID(_ context.Context, _, _ uint64) (*model.Reservation, error) {
	return nil, nil
}

func (m *mockTrimmingReservationRepository) HasDoctorConflict(_ context.Context, _, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
	return false, nil
}

func (m *mockTrimmingReservationRepository) CountOnDutyDoctors(_ context.Context, _ uint64, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockTrimmingReservationRepository) CountConflicts(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
	return 0, nil
}

func (m *mockTrimmingReservationRepository) CountByCustomerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockTrimmingReservationRepository) CountByDateAndSource(_ context.Context, _ uint64, _ time.Time, _ model.ReservationSource) (int64, error) {
	return 0, nil
}

func (m *mockTrimmingReservationRepository) FindNoShowCandidates(_ context.Context, _ uint64) ([]model.Reservation, error) {
	return nil, nil
}
func (m *mockTrimmingReservationRepository) HasReservationByOwnerInRange(_ context.Context, _, _ uint64, _, _ time.Time) (bool, error) {
	return false, nil
}

// コンパイル時インターフェース適合チェック
var _ repository.ReservationRepository = (*mockTrimmingReservationRepository)(nil)

// --- mock: ReservationTypeRepository ---

type mockTrimmingReservationTypeRepository struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
}

func (m *mockTrimmingReservationTypeRepository) FindAll(_ context.Context, _ uint64) ([]model.ReservationType, error) {
	return nil, nil
}

func (m *mockTrimmingReservationTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationType{
		ID:       id,
		ClinicID: clinicID,
		Category: model.ReservationTypeCategoryTrimming,
	}, nil
}

func (m *mockTrimmingReservationTypeRepository) CountUsageByReservationTypeID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockTrimmingReservationTypeRepository) Create(_ context.Context, _ *model.ReservationType) error {
	return nil
}

func (m *mockTrimmingReservationTypeRepository) Update(_ context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error) {
	return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockTrimmingReservationTypeRepository) Delete(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockTrimmingReservationTypeRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return nil
}

var _ repository.ReservationTypeRepository = (*mockTrimmingReservationTypeRepository)(nil)

// --- mock: AppointmentTrimmingDetailRepository ---

type mockTrimmingDetailRepository struct {
	findByAppointmentIDFn func(ctx context.Context, clinicID, appointmentID uint64) (*model.AppointmentTrimmingDetail, error)
	createFn              func(ctx context.Context, detail *model.AppointmentTrimmingDetail) error
	updateFn              func(ctx context.Context, detail *model.AppointmentTrimmingDetail) error
	setOptionsFn          func(ctx context.Context, appointmentID uint64, optionIDs []uint64) error
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

func (m *mockTrimmingDetailRepository) SetOptions(ctx context.Context, appointmentID uint64, optionIDs []uint64) error {
	if m.setOptionsFn != nil {
		return m.setOptionsFn(ctx, appointmentID, optionIDs)
	}
	return nil
}

var _ repository.AppointmentTrimmingDetailRepository = (*mockTrimmingDetailRepository)(nil)

// --- mock: Transactor（テスト用：fn を同一コンテキストで直接実行） ---

type mockTransactor struct {
	withTxErr error
}

func (m *mockTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
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
	return NewTrimmingService(reserv, reservationType, nil, nil, nil, detail, &mockTransactor{})
}

func newTrimmingTestServiceWithAvailability(
	reserv *mockTrimmingReservationRepository,
	detail *mockTrimmingDetailRepository,
	reservationType *mockTrimmingReservationTypeRepository,
	reservationStaff repository.ReservationStaffRepository,
	unavailableTime repository.ReservationTypeUnavailableTimeRepository,
	availableSlot repository.ReservationTypeAvailableSlotRepository,
) TrimmingService {
	return NewTrimmingService(reserv, reservationType, reservationStaff, unavailableTime, availableSlot, detail, &mockTransactor{})
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
				setOptionsFn: func(_ context.Context, _ uint64, _ []uint64) error {
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
		findExcludedReservationTypesFn: func(_ context.Context, staffID uint64) ([]model.StaffReservationExclusion, error) {
			assert.Equal(t, uint64(33), staffID)
			return []model.StaffReservationExclusion{{StaffID: staffID, ReservationTypeID: 9}}, nil
		},
	}
	svc := newTrimmingTestServiceWithAvailability(
		reserv,
		&mockTrimmingDetailRepository{},
		&mockTrimmingReservationTypeRepository{},
		staffRepo,
		nil,
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

func TestTrimmingService_Create_RejectsUnavailableStartSlot(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, jstLocation)
	end := start.Add(time.Hour)
	reserv := &mockTrimmingReservationRepository{
		createFn: func(_ context.Context, _ *model.Reservation) error {
			t.Fatal("trimming appointment must not be created outside available slots")
			return nil
		},
	}
	availableSlotRepo := &mockAvailableSlotRepository{
		findAllFn: func(_ context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(9), reservationTypeID)
			dayOfWeek := int8(1)
			return []model.ReservationTypeAvailableSlot{
				{ReservationTypeID: 9, AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &dayOfWeek, StartTime: "09:45", IsActive: true},
				{ReservationTypeID: 9, AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &dayOfWeek, StartTime: "12:30", IsActive: true},
			}, nil
		},
	}
	svc := newTrimmingTestServiceWithAvailability(
		reserv,
		&mockTrimmingDetailRepository{},
		&mockTrimmingReservationTypeRepository{},
		nil,
		nil,
		availableSlotRepo,
	)

	appt, err := svc.Create(context.Background(), 1, &CreateTrimmingInput{
		ReservationTypeID: 9,
		StartTime:         start,
		EndTime:           end,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, appt)
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
	assert.False(t, appointmentUpdated)
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
				setOptionsFn: func(_ context.Context, _ uint64, _ []uint64) error {
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

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

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
