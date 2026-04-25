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

// mockReservationTypeLiffRepository は ReservationTypeLiffRepository のテスト用モック実装
type mockReservationTypeLiffRepository struct {
	findAllFn       func(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	findByIDFn      func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	createFn        func(ctx context.Context, st *model.ReservationType) error
	updateFieldsFn  func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error)
	deleteFn        func(ctx context.Context, clinicID, id uint64) error
	updateSortOrder func(ctx context.Context, clinicID, id uint64, direction string) error
}

func (m *mockReservationTypeLiffRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockReservationTypeLiffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockReservationTypeLiffRepository) Create(ctx context.Context, st *model.ReservationType) error {
	return m.createFn(ctx, st)
}

func (m *mockReservationTypeLiffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return nil, nil
}

func (m *mockReservationTypeLiffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockReservationTypeLiffRepository) UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if m.updateSortOrder != nil {
		return m.updateSortOrder(ctx, clinicID, id, direction)
	}
	return nil
}

// mockReservationAdminRepository は appointment_admin_service_test.go で定義済み

// mockReservationQueryRepository は ReservationQueryRepository のテスト用モック実装
type mockReservationQueryRepository struct {
	existsByReservationTypeIDFn func(ctx context.Context, clinicID, reservationTypeID uint64) (bool, error)
}

func (m *mockReservationQueryRepository) ExistsByReservationTypeID(ctx context.Context, clinicID, reservationTypeID uint64) (bool, error) {
	if m.existsByReservationTypeIDFn != nil {
		return m.existsByReservationTypeIDFn(ctx, clinicID, reservationTypeID)
	}
	return false, nil
}

func (m *mockReservationQueryRepository) ExistsByStaffID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationQueryRepository) CountMedicalRecordsByReservationID(_ context.Context, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockReservationQueryRepository) CountByCustomerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockReservationQueryRepository) CountByDateAndSource(_ context.Context, _ uint64, _ time.Time, _ model.ReservationSource) (int64, error) {
	return 0, nil
}

func (m *mockReservationQueryRepository) FindAllByCategory(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}

func (m *mockReservationQueryRepository) FindNoShowCandidates(_ context.Context, _ uint64) ([]model.Reservation, error) {
	return nil, nil
}

func newTestReservationTypeLiffService(
	repo *mockReservationTypeLiffRepository,
	resAdminRepo *mockReservationAdminRepository,
	resRepo *mockReservationQueryRepository,
) ReservationTypeLiffService {
	return NewReservationTypeLiffService(repo, resAdminRepo, resRepo)
}

func TestReservationTypeLiffService_List_ReturnsTypes(t *testing.T) {
	tests := []struct {
		name      string
		repoTypes []model.ReservationType
		repoErr   error
		wantLen   int
		wantErr   bool
	}{
		{
			name: "returns reservation type list",
			repoTypes: []model.ReservationType{
				{ID: 1, ClinicID: 1, Name: "一般診察", IsActive: true},
				{ID: 2, ClinicID: 1, Name: "トリミング", IsActive: true},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:      "returns empty list when no types exist",
			repoTypes: []model.ReservationType{},
			repoErr:   nil,
			wantLen:   0,
			wantErr:   false,
		},
		{
			name:      "propagates repository error",
			repoTypes: nil,
			repoErr:   errors.New("db connection error"),
			wantLen:   0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeLiffRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
					return tt.repoTypes, tt.repoErr
				},
			}
			resAdminRepo := &mockReservationAdminRepository{}
			resRepo := &mockReservationQueryRepository{}
			svc := newTestReservationTypeLiffService(repo, resAdminRepo, resRepo)

			types, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, types, tt.wantLen)
			}
		})
	}
}

func TestReservationTypeLiffService_GetByID_Found(t *testing.T) {
	// Delete + PatchStatus 以外で GetByID に相当する公開メソッドはないため、
	// Delete を経由してリポジトリの FindByID が呼ばれることを検証する代わりに、
	// Update を通じて間接的に FindByID を呼び出す動作を確認する。
	expected := &model.ReservationType{
		ID:       1,
		ClinicID: 1,
		Name:     "一般診察",
		IsActive: true,
	}
	name := "一般診察（更新）"
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return expected, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.ReservationType, error) {
			return &model.ReservationType{ID: 1, ClinicID: 1, Name: name, IsActive: true}, nil
		},
	}
	resAdminRepo := &mockReservationAdminRepository{}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resAdminRepo, resRepo)

	updated, err := svc.Update(context.Background(), 1, 1, &UpdateReservationTypeLiffInput{Name: &name})

	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, name, updated.Name)
}

func TestReservationTypeLiffService_GetByID_NotFound(t *testing.T) {
	// Update を通じて FindByID の not found エラー伝播を確認する
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return nil, apperrors.WrapNotFound("reservation_type_liff", "999")
		},
	}
	resAdminRepo := &mockReservationAdminRepository{}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resAdminRepo, resRepo)

	name := "なんでもいい"
	result, err := svc.Update(context.Background(), 1, 999, &UpdateReservationTypeLiffInput{Name: &name})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, result)
}

func TestReservationTypeLiffService_Create_Success(t *testing.T) {
	input := &CreateReservationTypeLiffInput{
		Name:            "新コース",
		Color:           "#FF5733",
		Description:     "テスト説明",
		SortOrder:       1,
		DurationMinutes: 30,
	}
	created := &model.ReservationType{
		ID:       10,
		ClinicID: 1,
		Name:     input.Name,
		Color:    input.Color,
		IsActive: true,
	}
	repo := &mockReservationTypeLiffRepository{
		createFn: func(_ context.Context, st *model.ReservationType) error {
			st.ID = 10
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return created, nil
		},
	}
	resAdminRepo := &mockReservationAdminRepository{}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resAdminRepo, resRepo)

	result, err := svc.Create(context.Background(), 1, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "新コース", result.Name)
}

func TestReservationTypeLiffService_Delete_InUse_Returns409(t *testing.T) {
	existing := &model.ReservationType{ID: 1, ClinicID: 1, Name: "使用中コース", IsActive: true}
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return existing, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return nil
		},
	}
	resAdminRepo := &mockReservationAdminRepository{}
	resRepo := &mockReservationQueryRepository{
		existsByReservationTypeIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
			// 予約データが存在する（使用中）
			return true, nil
		},
	}
	svc := newTestReservationTypeLiffService(repo, resAdminRepo, resRepo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
}

func TestReservationTypeLiffService_Delete_NotFound(t *testing.T) {
	repo := &mockReservationTypeLiffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
			return nil, apperrors.WrapNotFound("reservation_type_liff", "999")
		},
	}
	resAdminRepo := &mockReservationAdminRepository{}
	resRepo := &mockReservationQueryRepository{}
	svc := newTestReservationTypeLiffService(repo, resAdminRepo, resRepo)

	err := svc.Delete(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}
