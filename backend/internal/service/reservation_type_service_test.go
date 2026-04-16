package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockReservationTypeRepository は ReservationTypeRepository のテスト用モック実装
type mockReservationTypeRepository struct {
	findAllFn  func(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	createFn   func(ctx context.Context, st *model.ReservationType) error
	updateFn   func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
	reorderFn  func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockReservationTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return []model.ReservationType{}, nil
}

func (m *mockReservationTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockReservationTypeRepository) Create(ctx context.Context, st *model.ReservationType) error {
	if m.createFn != nil {
		return m.createFn(ctx, st)
	}
	return nil
}

func (m *mockReservationTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, fields)
	}
	return nil
}

func (m *mockReservationTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockReservationTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// mockReservationForReservationType は ReservationType テストで使用する ReservationRepository のスタブ
type mockReservationForReservationType struct {
	existsByReservationTypeIDFn func(ctx context.Context, reservationTypeID uint64) (bool, error)
	existsByStaffIDFn           func(ctx context.Context, staffID uint64) (bool, error)
}

func (m *mockReservationForReservationType) FindAll(_ context.Context, _ uint64, _, _ int, _ *time.Time, _, _ *string, _, _ *uint64) ([]model.Appointment, int64, error) {
	return nil, 0, nil
}
func (m *mockReservationForReservationType) FindByID(_ context.Context, _, _ uint64) (*model.Appointment, error) {
	return nil, nil
}
func (m *mockReservationForReservationType) Create(_ context.Context, _ *model.Appointment) error {
	return nil
}
func (m *mockReservationForReservationType) UpdateFields(_ context.Context, _, _ uint64, _ map[string]any) (*model.Appointment, error) {
	return nil, nil
}
func (m *mockReservationForReservationType) Delete(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockReservationForReservationType) ExistsByReservationTypeID(ctx context.Context, reservationTypeID uint64) (bool, error) {
	if m.existsByReservationTypeIDFn != nil {
		return m.existsByReservationTypeIDFn(ctx, reservationTypeID)
	}
	return false, nil
}
func (m *mockReservationForReservationType) ExistsByStaffID(ctx context.Context, staffID uint64) (bool, error) {
	if m.existsByStaffIDFn != nil {
		return m.existsByStaffIDFn(ctx, staffID)
	}
	return false, nil
}

func (m *mockReservationForReservationType) CountMedicalRecordsByReservationID(_ context.Context, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockReservationForReservationType) LockAndFindByID(_ context.Context, _, _ uint64) (*model.Appointment, error) {
	return nil, nil
}

func (m *mockReservationForReservationType) HasDoctorConflict(_ context.Context, _, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationForReservationType) CountOnDutyDoctors(_ context.Context, _ uint64, _ time.Time) (int64, error) {
	return 1, nil
}

func (m *mockReservationForReservationType) CountConflicts(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
	return 0, nil
}

func (m *mockReservationForReservationType) CountByCustomerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockReservationForReservationType) CountByDateAndSource(_ context.Context, _ uint64, _ time.Time, _ model.ReservationSource) (int64, error) {
	return 0, nil
}

func (m *mockReservationForReservationType) FindAllByCategory(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Appointment, int64, error) {
	return nil, 0, nil
}

func newTestReservationTypeService(repo *mockReservationTypeRepository) ReservationTypeService {
	return NewReservationTypeService(repo, &mockReservationForReservationType{})
}

// ---- List ----

func TestReservationTypeService_List(t *testing.T) {
	t.Run("returns service types with clinicID filter", func(t *testing.T) {
		const clinicID uint64 = 42
		want := []model.ReservationType{
			{ID: 1, ClinicID: clinicID, Name: "診察"},
			{ID: 2, ClinicID: clinicID, Name: "トリミング"},
		}
		repo := &mockReservationTypeRepository{
			findAllFn: func(_ context.Context, cid uint64) ([]model.ReservationType, error) {
				assert.Equal(t, clinicID, cid)
				return want, nil
			},
		}
		svc := newTestReservationTypeService(repo)

		got, err := svc.List(context.Background(), clinicID)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &mockReservationTypeRepository{
			findAllFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
				return nil, errors.New("db error")
			},
		}
		svc := newTestReservationTypeService(repo)

		_, err := svc.List(context.Background(), 1)

		assert.Error(t, err)
	})
}

// ---- Create ----

func TestReservationTypeService_Create(t *testing.T) {
	t.Run("creates service type successfully", func(t *testing.T) {
		const clinicID uint64 = 10
		input := &CreateReservationTypeInput{
			Name:        "診察",
			Color:       "#FF0000",
			IsActive:    true,
			Description: "通常診察",
			SortOrder:   1,
		}
		var captured *model.ReservationType
		repo := &mockReservationTypeRepository{
			createFn: func(_ context.Context, st *model.ReservationType) error {
				captured = st
				return nil
			},
		}
		svc := newTestReservationTypeService(repo)

		got, err := svc.Create(context.Background(), clinicID, input)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, clinicID, got.ClinicID)
		assert.Equal(t, input.Name, got.Name)
		assert.Equal(t, input.Color, got.Color)
		assert.Equal(t, input.IsActive, got.IsActive)
		assert.Equal(t, input.Description, got.Description)
		assert.Equal(t, input.SortOrder, got.SortOrder)
		assert.Same(t, captured, got)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &mockReservationTypeRepository{
			createFn: func(_ context.Context, _ *model.ReservationType) error {
				return errors.New("db error")
			},
		}
		svc := newTestReservationTypeService(repo)

		_, err := svc.Create(context.Background(), 1, &CreateReservationTypeInput{Name: "診察"})

		assert.Error(t, err)
	})
}

// ---- Update ----

func TestReservationTypeService_Update(t *testing.T) {
	t.Run("returns existing record when no fields provided", func(t *testing.T) {
		const (
			clinicID uint64 = 1
			id       uint64 = 5
		)
		existing := &model.ReservationType{ID: id, ClinicID: clinicID, Name: "既存"}
		updateCalled := false
		repo := &mockReservationTypeRepository{
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				updateCalled = true
				return nil
			},
			findByIDFn: func(_ context.Context, cid, rid uint64) (*model.ReservationType, error) {
				assert.Equal(t, clinicID, cid)
				assert.Equal(t, id, rid)
				return existing, nil
			},
		}
		svc := newTestReservationTypeService(repo)

		got, err := svc.Update(context.Background(), clinicID, id, &UpdateReservationTypeInput{})

		require.NoError(t, err)
		assert.Equal(t, existing, got)
		assert.False(t, updateCalled)
	})

	t.Run("updates fields and returns updated record", func(t *testing.T) {
		const (
			clinicID uint64 = 1
			id       uint64 = 5
		)
		newName := "新しい名前"
		newActive := false
		updated := &model.ReservationType{ID: id, ClinicID: clinicID, Name: newName, IsActive: newActive}

		var capturedFields map[string]any
		repo := &mockReservationTypeRepository{
			updateFn: func(_ context.Context, cid, rid uint64, fields map[string]any) error {
				assert.Equal(t, clinicID, cid)
				assert.Equal(t, id, rid)
				capturedFields = fields
				return nil
			},
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
				return updated, nil
			},
		}
		svc := newTestReservationTypeService(repo)

		got, err := svc.Update(context.Background(), clinicID, id, &UpdateReservationTypeInput{
			Name:     &newName,
			IsActive: &newActive,
		})

		require.NoError(t, err)
		assert.Equal(t, updated, got)
		assert.Equal(t, newName, capturedFields[colReservationTypeName])
		assert.Equal(t, newActive, capturedFields[colReservationTypeIsActive])
		assert.NotContains(t, capturedFields, colReservationTypeColor)
	})

	t.Run("returns not found when record does not exist", func(t *testing.T) {
		name := "test"
		repo := &mockReservationTypeRepository{
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return apperrors.WrapNotFound("reservation_type", "5")
			},
		}
		svc := newTestReservationTypeService(repo)

		_, err := svc.Update(context.Background(), 1, 5, &UpdateReservationTypeInput{Name: &name})

		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// ---- Delete ----

func TestReservationTypeService_Delete(t *testing.T) {
	t.Run("deletes successfully", func(t *testing.T) {
		const (
			clinicID uint64 = 1
			id       uint64 = 5
		)
		repo := &mockReservationTypeRepository{
			deleteFn: func(_ context.Context, cid, rid uint64) error {
				assert.Equal(t, clinicID, cid)
				assert.Equal(t, id, rid)
				return nil
			},
		}
		svc := newTestReservationTypeService(repo)

		err := svc.Delete(context.Background(), clinicID, id)

		assert.NoError(t, err)
	})

	t.Run("returns not found when record does not exist", func(t *testing.T) {
		repo := &mockReservationTypeRepository{
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return apperrors.WrapNotFound("reservation_type", "5")
			},
		}
		svc := newTestReservationTypeService(repo)

		err := svc.Delete(context.Background(), 1, 5)

		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// ---- Reorder ----

func TestReservationTypeService_Reorder(t *testing.T) {
	tests := []struct {
		name        string
		ids         []uint64
		repoErr     error
		wantErr     bool
		wantInvalid bool
	}{
		{
			name:    "reorders successfully",
			ids:     []uint64{3, 1, 2},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:        "returns error when ids is empty",
			ids:         []uint64{},
			repoErr:     nil,
			wantErr:     true,
			wantInvalid: true,
		},
		{
			name:        "returns error when ids is nil",
			ids:         nil,
			repoErr:     nil,
			wantErr:     true,
			wantInvalid: true,
		},
		{
			name:    "propagates repository error",
			ids:     []uint64{1, 2, 3},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:    "returns invalid error when id not found in clinic",
			ids:     []uint64{1, 999},
			repoErr: apperrors.WrapInvalidInput("reservation_type id 999 not found in this clinic"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeRepository{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					return tt.repoErr
				},
			}
			svc := newTestReservationTypeService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantInvalid {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
