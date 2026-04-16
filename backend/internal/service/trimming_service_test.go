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
	findAllByCategoryFn func(ctx context.Context, clinicID uint64, category model.ReservationTypeCategory, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Appointment, int64, error)
	findByIDFn          func(ctx context.Context, clinicID, id uint64) (*model.Appointment, error)
	createFn            func(ctx context.Context, appt *model.Appointment) error
	updateFieldsFn      func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Appointment, error)
	deleteFn            func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockTrimmingReservationRepository) FindAllByCategory(ctx context.Context, clinicID uint64, category model.ReservationTypeCategory, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Appointment, int64, error) {
	return m.findAllByCategoryFn(ctx, clinicID, category, petID, ownerID, startDate, endDate, page, limit)
}

func (m *mockTrimmingReservationRepository) FindAll(_ context.Context, _ uint64, _, _ int, _ *time.Time, _, _ *string, _, _ *uint64) ([]model.Appointment, int64, error) {
	return nil, 0, nil
}

func (m *mockTrimmingReservationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Appointment, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, apperrors.WrapNotFound("appointment", "0")
}

func (m *mockTrimmingReservationRepository) Create(ctx context.Context, appt *model.Appointment) error {
	if m.createFn != nil {
		return m.createFn(ctx, appt)
	}
	return nil
}

func (m *mockTrimmingReservationRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Appointment, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.Appointment{ID: id, ClinicID: clinicID}, nil
}

func (m *mockTrimmingReservationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockTrimmingReservationRepository) ExistsByReservationTypeID(_ context.Context, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockTrimmingReservationRepository) ExistsByStaffID(_ context.Context, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockTrimmingReservationRepository) CountMedicalRecordsByReservationID(_ context.Context, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockTrimmingReservationRepository) LockAndFindByID(_ context.Context, _, _ uint64) (*model.Appointment, error) {
	return nil, nil
}

func (m *mockTrimmingReservationRepository) HasDoctorConflict(_ context.Context, _ uint64, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
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

// コンパイル時インターフェース適合チェック
var _ repository.ReservationRepository = (*mockTrimmingReservationRepository)(nil)

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

// --- helpers ---

func newTrimmingTestService(reserv *mockTrimmingReservationRepository, detail *mockTrimmingDetailRepository) TrimmingService {
	return NewTrimmingService(reserv, detail)
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
		repoData  []model.Appointment
		repoTotal int64
		repoErr   error
		wantLen   int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:      "returns trimming list with total count",
			clinicID:  1,
			repoData:  []model.Appointment{{ID: 1, ClinicID: 1}, {ID: 2, ClinicID: 1}},
			repoTotal: 2,
			wantLen:   2,
			wantTotal: 2,
		},
		{
			name:      "filters by pet ID",
			clinicID:  1,
			petID:     &petID,
			repoData:  []model.Appointment{{ID: 1, ClinicID: 1, PetID: &petID}},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
		},
		{
			name:      "filters by owner ID",
			clinicID:  1,
			ownerID:   &ownerID,
			repoData:  []model.Appointment{{ID: 1, ClinicID: 1}},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
		},
		{
			name:      "returns empty list when no records exist",
			clinicID:  1,
			repoData:  []model.Appointment{},
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
				findAllByCategoryFn: func(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Appointment, int64, error) {
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
		repoAppt *model.Appointment
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "returns appointment when found",
			clinicID: 1,
			id:       10,
			repoAppt: &model.Appointment{ID: 10, ClinicID: 1},
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
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Appointment, error) {
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
				createFn: func(_ context.Context, a *model.Appointment) error {
					a.ID = 1
					return tt.createErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Appointment, error) {
					return &model.Appointment{ID: 1, ClinicID: 1}, nil
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
