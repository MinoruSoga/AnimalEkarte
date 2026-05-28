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
	findAllFn                          func(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error)
	findByIDFn                         func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	createFn                           func(ctx context.Context, reservation *model.Reservation) error
	updateFieldsFn                     func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error)
	deleteFn                           func(ctx context.Context, clinicID, id uint64) error
	countMedicalRecordsByReservationID func(ctx context.Context, reservationID uint64) (int64, error)
	countOnDutyDoctorsFn               func(ctx context.Context, clinicID uint64, date time.Time) (int64, error)
	countConflictsFn                   func(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error)
}

func (m *mockReservationRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit, date, status, source, petID, ownerID)
}

func (m *mockReservationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockReservationRepository) Create(ctx context.Context, reservation *model.Reservation) error {
	return m.createFn(ctx, reservation)
}

func (m *mockReservationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
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

func (m *mockReservationRepository) CountMedicalRecordsByReservationID(ctx context.Context, reservationID uint64) (int64, error) {
	if m.countMedicalRecordsByReservationID != nil {
		return m.countMedicalRecordsByReservationID(ctx, reservationID)
	}
	return 0, nil
}

func (m *mockReservationRepository) LockAndFindByID(_ context.Context, _, _ uint64) (*model.Reservation, error) {
	return nil, nil
}

func (m *mockReservationRepository) HasDoctorConflict(_ context.Context, _, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
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

func (m *mockReservationRepository) CountByCustomerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockReservationRepository) CountByDateAndSource(_ context.Context, _ uint64, _ time.Time, _ model.ReservationSource) (int64, error) {
	return 0, nil
}

func (m *mockReservationRepository) FindAllByCategory(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}

func (m *mockReservationRepository) FindNoShowCandidates(_ context.Context, _ uint64) ([]model.Reservation, error) {
	return nil, nil
}
func (m *mockReservationRepository) HasReservationByOwnerInRange(_ context.Context, _, _ uint64, _, _ time.Time) (bool, error) {
	return false, nil
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
				findAllFn: func(_ context.Context, _ uint64, _, _ int, _ *time.Time, _ *string, _ *string, _ *uint64, _ *uint64) ([]model.Reservation, int64, error) {
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
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationRepository{}
			svc := NewReservationService(repo, nil)

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
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Reservation{ID: 1, ClinicID: 1}, nil
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

func TestReservationService_UpdateReservationRoute(t *testing.T) {
	validRoutes := []string{"line", "phone", "reception", "exam_room"}

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
				svc := NewReservationService(repo, nil)
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
		svc := NewReservationService(repo, nil)
		result, err := svc.UpdateReservationRoute(context.Background(), 1, 1, UpdateReservationRouteInput{Route: ""})
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("error: invalid route 'fax' returns InvalidInput", func(t *testing.T) {
		svc := NewReservationService(&mockReservationRepository{}, nil)
		_, err := svc.UpdateReservationRoute(context.Background(), 1, 1, UpdateReservationRouteInput{Route: "fax"})
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("error: uppercase 'LINE' is not valid", func(t *testing.T) {
		svc := NewReservationService(&mockReservationRepository{}, nil)
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
		svc := NewReservationService(repo, nil)
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
		svc := NewReservationService(repo, nil)
		_, err := svc.UpdateReservationRoute(context.Background(), 99, 1, UpdateReservationRouteInput{Route: "line"})
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
