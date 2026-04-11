package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockReservationRepository は ReservationRepository のテスト用モック実装
type mockReservationRepository struct {
	findAllFn                          func(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Appointment, int64, error)
	findByIDFn                         func(ctx context.Context, clinicID, id uint64) (*model.Appointment, error)
	createFn                           func(ctx context.Context, reservation *model.Appointment) error
	updateFieldsFn                     func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Appointment, error)
	deleteFn                           func(ctx context.Context, clinicID, id uint64) error
	countMedicalRecordsByReservationID func(ctx context.Context, reservationID uint64) (int64, error)
}

func (m *mockReservationRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Appointment, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit, date, status, source, petID, ownerID)
}

func (m *mockReservationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Appointment, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockReservationRepository) Create(ctx context.Context, reservation *model.Appointment) error {
	return m.createFn(ctx, reservation)
}

func (m *mockReservationRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Appointment, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockReservationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockReservationRepository) ExistsByReservationTypeID(_ context.Context, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationRepository) ExistsByStaffID(_ context.Context, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationRepository) CountMedicalRecordsByReservationID(ctx context.Context, reservationID uint64) (int64, error) {
	if m.countMedicalRecordsByReservationID != nil {
		return m.countMedicalRecordsByReservationID(ctx, reservationID)
	}
	return 0, nil
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
		repoReservations []model.Appointment
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
			repoReservations: []model.Appointment{
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
			repoReservations: []model.Appointment{
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
			repoReservations: []model.Appointment{
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
			repoReservations: []model.Appointment{
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
			repoReservations: []model.Appointment{
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
			repoReservations: []model.Appointment{},
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
				findAllFn: func(_ context.Context, _ uint64, _, _ int, _ *time.Time, _ *string, _ *string, _ *uint64, _ *uint64) ([]model.Appointment, int64, error) {
					return tt.repoReservations, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewReservationService(repo, nil)

			reservations, total, err := svc.List(context.Background(), tt.clinicID, tt.page, tt.limit, tt.date, tt.status, nil, tt.petID, tt.ownerID)

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
		repoReservation *model.Appointment
		repoErr         error
		wantReservation *model.Appointment
		wantErr         error
	}{
		{
			name:            "returns reservation when found",
			clinicID:        1,
			id:              10,
			repoReservation: &model.Appointment{ID: 10, ClinicID: 1, StartTime: now, EndTime: now.Add(time.Hour), Status: model.ReservationStatusPending},
			repoErr:         nil,
			wantReservation: &model.Appointment{ID: 10, ClinicID: 1, StartTime: now, EndTime: now.Add(time.Hour), Status: model.ReservationStatusPending},
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
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Appointment, error) {
					return tt.repoReservation, tt.repoErr
				},
			}
			svc := NewReservationService(repo, nil)

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
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Appointment, error) {
			return nil, apperrors.WrapNotFound("reservation", "999")
		},
	}
	svc := NewReservationService(repo, nil)

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
		reservation      *model.Appointment
		wantErr          bool
		wantInvalidInput bool
	}{
		{
			// BUG-034: end_time == start_time は 400 Bad Request
			name: "returns invalid input when end_time equals start_time",
			reservation: &model.Appointment{
				ClinicID:      1,
				StartTime:     now,
				EndTime:       now,
				ReservationTypeID: 1,
			},
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			// BUG-034: end_time < start_time は 400 Bad Request
			name: "returns invalid input when end_time is before start_time",
			reservation: &model.Appointment{
				ClinicID:      1,
				StartTime:     now,
				EndTime:       now.Add(-time.Minute),
				ReservationTypeID: 1,
			},
			wantErr:          true,
			wantInvalidInput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationRepository{}
			svc := NewReservationService(repo, nil)

			err := svc.Create(context.Background(), tt.reservation)

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

func TestReservationService_Update(t *testing.T) {
	now := time.Now()
	statusConfirmed := model.ReservationStatusConfirmed
	tests := []struct {
		name    string
		input   UpdateReservationInput
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name: "updates reservation successfully",
			input: UpdateReservationInput{
				Status:    &statusConfirmed,
				StartTime: &now,
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
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Appointment, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Appointment{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewReservationService(repo, nil)

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
			svc := NewReservationService(repo, nil)

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
