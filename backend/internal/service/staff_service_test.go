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

// mockStaffRepository は StaffRepository のテスト用モック実装
type mockStaffRepository struct {
	findAllFn         func(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error)
	findByIDFn        func(ctx context.Context, id uint64) (*model.Staff, error)
	findByAccountIDFn func(ctx context.Context, accountID uint64) (*model.Staff, error)
	createFn          func(ctx context.Context, staff *model.Staff) error
	updateFn          func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn          func(ctx context.Context, clinicID, id uint64) error
	reorderErr        error
}

func (m *mockStaffRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error) {
	return m.findAllFn(ctx, clinicID, page, limit)
}

func (m *mockStaffRepository) FindByID(ctx context.Context, id uint64) (*model.Staff, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockStaffRepository) FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error) {
	if m.findByAccountIDFn != nil {
		return m.findByAccountIDFn(ctx, accountID)
	}
	return nil, apperrors.WrapNotFound("staff", "account_id")
}

func (m *mockStaffRepository) Create(ctx context.Context, staff *model.Staff) error {
	return m.createFn(ctx, staff)
}

func (m *mockStaffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockStaffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockStaffRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

// mockReservationForStaff は Staff テストで使用する ReservationRepository のスタブ
type mockReservationForStaff struct {
	existsByStaffIDFn func(ctx context.Context, staffID uint64) (bool, error)
}

func (m *mockReservationForStaff) FindAll(_ context.Context, _ uint64, _, _ int, _ *time.Time, _ *string, _, _ *uint64) ([]model.ReservationAppointment, int64, error) {
	return nil, 0, nil
}
func (m *mockReservationForStaff) FindByID(_ context.Context, _, _ uint64) (*model.ReservationAppointment, error) {
	return nil, nil
}
func (m *mockReservationForStaff) Create(_ context.Context, _ *model.ReservationAppointment) error {
	return nil
}
func (m *mockReservationForStaff) UpdateFields(_ context.Context, _, _ uint64, _ map[string]any) (*model.ReservationAppointment, error) {
	return nil, nil
}
func (m *mockReservationForStaff) Delete(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockReservationForStaff) ExistsByServiceTypeID(_ context.Context, _ uint64) (bool, error) {
	return false, nil
}
func (m *mockReservationForStaff) ExistsByStaffID(ctx context.Context, staffID uint64) (bool, error) {
	if m.existsByStaffIDFn != nil {
		return m.existsByStaffIDFn(ctx, staffID)
	}
	return false, nil
}

func (m *mockReservationForStaff) FindByStaffAndTimeSlot(_ context.Context, _, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationForStaff) CountMedicalRecordsByReservationID(_ context.Context, _ uint64) (int64, error) {
	return 0, nil
}

// mockShiftEntryForStaff は Staff テストで使用する ShiftEntryRepository のスタブ
type mockShiftEntryForStaff struct {
	existsByStaffIDFn func(ctx context.Context, staffID uint64) (bool, error)
}

func (m *mockShiftEntryForStaff) List(_ context.Context, _ uint64, _ repository.ShiftEntryFilter) ([]model.ShiftEntry, error) {
	return nil, nil
}
func (m *mockShiftEntryForStaff) FindByID(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
	return nil, nil
}
func (m *mockShiftEntryForStaff) Create(_ context.Context, _ *model.ShiftEntry) error {
	return nil
}
func (m *mockShiftEntryForStaff) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *mockShiftEntryForStaff) Delete(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockShiftEntryForStaff) ExistsByStaffID(ctx context.Context, staffID uint64) (bool, error) {
	if m.existsByStaffIDFn != nil {
		return m.existsByStaffIDFn(ctx, staffID)
	}
	return false, nil
}

func newTestStaffService(repo *mockStaffRepository) StaffService {
	return NewStaffService(repo, &mockReservationForStaff{}, &mockShiftEntryForStaff{})
}

func TestStaffService_List(t *testing.T) {
	tests := []struct {
		name       string
		clinicID   uint64
		repoStaffs []model.Staff
		repoTotal  int64
		repoErr    error
		wantLen    int
		wantTotal  int64
		wantErr    bool
	}{
		{
			name:     "returns all staffs",
			clinicID: 1,
			repoStaffs: []model.Staff{
				{ID: 1, Name: "山田 太郎"},
				{ID: 2, Name: "鈴木 花子"},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:       "returns empty list when no staffs exist",
			clinicID:   1,
			repoStaffs: []model.Staff{},
			repoTotal:  0,
			repoErr:    nil,
			wantLen:    0,
			wantTotal:  0,
			wantErr:    false,
		},
		{
			name:       "propagates repository error",
			clinicID:   1,
			repoStaffs: nil,
			repoTotal:  0,
			repoErr:    errors.New("db connection error"),
			wantLen:    0,
			wantTotal:  0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedPage := 0
			capturedLimit := 0
			repo := &mockStaffRepository{
				findAllFn: func(_ context.Context, _ uint64, page, limit int) ([]model.Staff, int64, error) {
					capturedPage = page
					capturedLimit = limit
					return tt.repoStaffs, tt.repoTotal, tt.repoErr
				},
			}
			svc := newTestStaffService(repo)

			staffs, total, err := svc.List(context.Background(), tt.clinicID, 1, 20)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, int64(0), total)
			} else {
				assert.NoError(t, err)
				assert.Len(t, staffs, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
				assert.Equal(t, 1, capturedPage)
				assert.Equal(t, 20, capturedLimit)
			}
		})
	}
}

func TestStaffService_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		id        uint64
		repoStaff *model.Staff
		repoErr   error
		wantStaff *model.Staff
		wantErr   error
	}{
		{
			name:      "returns staff when found",
			id:        10,
			repoStaff: &model.Staff{ID: 10, Name: "山田 太郎"},
			repoErr:   nil,
			wantStaff: &model.Staff{ID: 10, Name: "山田 太郎"},
			wantErr:   nil,
		},
		{
			name:      "returns not found error when staff does not exist",
			id:        999,
			repoStaff: nil,
			repoErr:   apperrors.WrapNotFound("staff", "999"),
			wantStaff: nil,
			wantErr:   apperrors.ErrNotFound,
		},
		{
			name:      "returns error on repository failure",
			id:        10,
			repoStaff: nil,
			repoErr:   errors.New("db error"),
			wantStaff: nil,
			wantErr:   errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					return tt.repoStaff, tt.repoErr
				},
			}
			svc := newTestStaffService(repo)

			staff, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, apperrors.ErrNotFound) {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStaff, staff)
			}
		})
	}
}

func TestStaffService_GetByID_NotFound(t *testing.T) {
	repo := &mockStaffRepository{
		findByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
			return nil, apperrors.WrapNotFound("staff", "999")
		},
	}
	svc := newTestStaffService(repo)

	staff, err := svc.GetByID(context.Background(), 999)

	assert.Nil(t, staff)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

// TestStaffService_Create_Success はスタッフ作成の正常ケースをテストする。
func TestStaffService_Create_Success(t *testing.T) {
	repo := &mockStaffRepository{
		createFn: func(_ context.Context, staff *model.Staff) error {
			// IDをシミュレート
			staff.ID = 1
			return nil
		},
	}
	svc := newTestStaffService(repo)

	input := &CreateStaffInput{
		Name: "新規 スタッフ",
	}

	staff, err := svc.Create(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.Equal(t, "新規 スタッフ", staff.Name)
	assert.True(t, staff.IsActive)
}

func TestStaffService_Create_RepositoryError(t *testing.T) {
	repo := &mockStaffRepository{
		createFn: func(_ context.Context, _ *model.Staff) error {
			return errors.New("db connection error")
		},
	}
	svc := newTestStaffService(repo)

	input := &CreateStaffInput{
		Name: "エラー スタッフ",
	}

	staff, err := svc.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, staff)
}

func TestStaffService_Create_DuplicateName(t *testing.T) {
	repo := &mockStaffRepository{
		createFn: func(_ context.Context, _ *model.Staff) error {
			return apperrors.WrapAlreadyExists("staff", "existing@example.com")
		},
	}
	svc := newTestStaffService(repo)

	input := &CreateStaffInput{
		Name: "重複 スタッフ",
	}

	staff, err := svc.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, apperrors.IsAlreadyExists(err))
}

func TestStaffService_Update(t *testing.T) {
	name := "更新後 スタッフ"
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		input    *UpdateStaffInput
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "updates staff successfully",
			clinicID: 1,
			id:       1,
			input:    &UpdateStaffInput{Name: &name},
			repoErr:  nil,
			wantErr:  false,
		},
		{
			name:     "returns not found error when staff does not exist",
			clinicID: 1,
			id:       999,
			input:    &UpdateStaffInput{Name: &name},
			repoErr:  apperrors.WrapNotFound("staff", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error when no field provided",
			clinicID: 1,
			id:       1,
			input:    &UpdateStaffInput{},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       1,
			input:    &UpdateStaffInput{Name: &name},
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id}, nil
				},
			}
			svc := newTestStaffService(repo)

			staff, err := svc.Update(context.Background(), tt.clinicID, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, staff)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, staff)
			}
		})
	}
}

func TestStaffService_Reorder(t *testing.T) {
	tests := []struct {
		name             string
		ids              []uint64
		repoErr          error
		wantErr          bool
		wantInvalidInput bool
	}{
		{
			name:    "reorders successfully",
			ids:     []uint64{3, 1, 2},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:             "returns invalid input when ids is empty",
			ids:              []uint64{},
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			name:    "propagates repository error",
			ids:     []uint64{1, 2},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{reorderErr: tt.repoErr}
			svc := newTestStaffService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantInvalidInput {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStaffService_Delete(t *testing.T) {
	tests := []struct {
		name                string
		clinicID            uint64
		id                  uint64
		reservationExists   bool
		shiftExists         bool
		checkReservationErr error
		checkShiftErr       error
		repoErr             error
		wantErr             bool
		wantNF              bool
		wantConflict        bool
	}{
		{
			name:                "deletes staff successfully when no dependencies exist",
			clinicID:            1,
			id:                  10,
			reservationExists:   false,
			shiftExists:         false,
			checkReservationErr: nil,
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             false,
		},
		{
			name:                "returns conflict error when staff has reservations",
			clinicID:            1,
			id:                  10,
			reservationExists:   true,
			shiftExists:         false,
			checkReservationErr: nil,
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             true,
			wantConflict:        true,
		},
		{
			name:                "returns conflict error when staff has shift entries",
			clinicID:            1,
			id:                  10,
			reservationExists:   false,
			shiftExists:         true,
			checkReservationErr: nil,
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             true,
			wantConflict:        true,
		},
		{
			name:                "returns conflict error when staff has both reservations and shifts",
			clinicID:            1,
			id:                  10,
			reservationExists:   true,
			shiftExists:         true,
			checkReservationErr: nil,
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             true,
			wantConflict:        true,
		},
		{
			name:                "returns error when reservation check fails",
			clinicID:            1,
			id:                  10,
			reservationExists:   false,
			shiftExists:         false,
			checkReservationErr: errors.New("db error"),
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             true,
		},
		{
			name:                "returns error when shift check fails",
			clinicID:            1,
			id:                  10,
			reservationExists:   false,
			shiftExists:         false,
			checkReservationErr: nil,
			checkShiftErr:       errors.New("db error"),
			repoErr:             nil,
			wantErr:             true,
		},
		{
			name:                "returns not found error when staff does not exist",
			clinicID:            1,
			id:                  999,
			reservationExists:   false,
			shiftExists:         false,
			checkReservationErr: nil,
			checkShiftErr:       nil,
			repoErr:             apperrors.WrapNotFound("staff", "999"),
			wantErr:             true,
			wantNF:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			reservationRepo := &mockReservationForStaff{
				existsByStaffIDFn: func(_ context.Context, _ uint64) (bool, error) {
					return tt.reservationExists, tt.checkReservationErr
				},
			}
			shiftRepo := &mockShiftEntryForStaff{
				existsByStaffIDFn: func(_ context.Context, _ uint64) (bool, error) {
					return tt.shiftExists, tt.checkShiftErr
				},
			}
			svc := NewStaffService(repo, reservationRepo, shiftRepo)

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
