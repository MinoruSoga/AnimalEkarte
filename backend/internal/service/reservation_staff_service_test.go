package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockReservationStaffRepository は ReservationStaffRepository のテスト用モック実装
type mockReservationStaffRepository struct {
	findAllFn                              func(ctx context.Context, clinicID uint64) ([]model.Staff, error)
	findByIDFn                             func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	createFn                               func(ctx context.Context, staff *model.Staff, clinicID uint64) error
	updateFieldsFn                         func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn                               func(ctx context.Context, clinicID, id uint64) error
	countUsageByStaffIDFn                  func() (int64, error)
	swapSortOrderFn                        func(ctx context.Context, clinicID, id uint64, direction string) error
	findExcludedReservationTypesFn         func(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error)
	findExcludedReservationTypesByStaffIDs func(ctx context.Context, staffIDs []uint64) ([]model.StaffReservationExclusion, error)
	replaceExcludedReservationTypesFn      func(ctx context.Context, staffID uint64, courseIDs []uint64) error
}

func (m *mockReservationStaffRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockReservationStaffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockReservationStaffRepository) Create(ctx context.Context, staff *model.Staff, clinicID uint64) error {
	return m.createFn(ctx, staff, clinicID)
}

func (m *mockReservationStaffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return nil
}

func (m *mockReservationStaffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockReservationStaffRepository) CountUsageByStaffID(_ context.Context, _, _ uint64) (int64, error) {
	if m.countUsageByStaffIDFn != nil {
		return m.countUsageByStaffIDFn()
	}
	return 0, nil
}

func (m *mockReservationStaffRepository) UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if m.swapSortOrderFn != nil {
		return m.swapSortOrderFn(ctx, clinicID, id, direction)
	}
	return nil
}

func (m *mockReservationStaffRepository) FindAllExcludedReservationTypes(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error) {
	if m.findExcludedReservationTypesFn != nil {
		return m.findExcludedReservationTypesFn(ctx, staffID)
	}
	return []model.StaffReservationExclusion{}, nil
}

func (m *mockReservationStaffRepository) FindAllExcludedReservationTypesByStaffIDs(ctx context.Context, staffIDs []uint64) ([]model.StaffReservationExclusion, error) {
	if m.findExcludedReservationTypesByStaffIDs != nil {
		return m.findExcludedReservationTypesByStaffIDs(ctx, staffIDs)
	}
	return []model.StaffReservationExclusion{}, nil
}

func (m *mockReservationStaffRepository) UpdateExcludedReservationTypes(ctx context.Context, staffID uint64, courseIDs []uint64) error {
	if m.replaceExcludedReservationTypesFn != nil {
		return m.replaceExcludedReservationTypesFn(ctx, staffID, courseIDs)
	}
	return nil
}

// mockTransactor は trimming_service_test.go で定義済み

func newTestReservationStaffService(repo *mockReservationStaffRepository, transactor *mockTransactor) ReservationStaffService {
	return NewReservationStaffService(repo, transactor)
}

func TestReservationStaffService_List_ReturnsStaffs(t *testing.T) {
	tests := []struct {
		name       string
		repoStaffs []model.Staff
		repoErr    error
		wantLen    int
		wantErr    bool
	}{
		{
			name: "returns staff list",
			repoStaffs: []model.Staff{
				{ID: 1, Name: "田中医師", IsActive: true, StaffType: model.StaffTypeDoctor},
				{ID: 2, Name: "鈴木スタッフ", IsActive: true, StaffType: model.StaffTypeNurse},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:       "returns empty list when no staffs exist",
			repoStaffs: []model.Staff{},
			repoErr:    nil,
			wantLen:    0,
			wantErr:    false,
		},
		{
			name:       "propagates repository error",
			repoStaffs: nil,
			repoErr:    errors.New("db connection error"),
			wantLen:    0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationStaffRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
					return tt.repoStaffs, tt.repoErr
				},
			}
			transactor := &mockTransactor{}
			svc := newTestReservationStaffService(repo, transactor)

			staffs, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, staffs, tt.wantLen)
			}
		})
	}
}

func TestReservationStaffService_GetByID_Found(t *testing.T) {
	expected := &model.Staff{
		ID:        1,
		Name:      "田中医師",
		IsActive:  true,
		StaffType: model.StaffTypeDoctor,
	}
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return expected, nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, err := svc.GetByID(context.Background(), 1, 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, staff)
}

func TestReservationStaffService_GetByID_NotFound(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return nil, apperrors.WrapNotFound("reservation_staff", "999")
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, err := svc.GetByID(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, staff)
}

func TestReservationStaffService_Create_Success(t *testing.T) {
	input := &CreateReservationStaffInput{
		Name:               "新スタッフ",
		StaffType:          string(model.StaffTypeDoctor),
		ReservationVisible: true,
		ReservationComment: "担当コメント",
		SortOrder:          1,
		ExcludedTypeIDs:    []uint64{},
	}
	repo := &mockReservationStaffRepository{
		createFn: func(_ context.Context, staff *model.Staff, _ uint64) error {
			staff.ID = 10
			return nil
		},
		findExcludedReservationTypesFn: func(_ context.Context, _ uint64) ([]model.StaffReservationExclusion, error) {
			return []model.StaffReservationExclusion{}, nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	staff, excluded, err := svc.Create(context.Background(), 1, input)

	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.NotNil(t, excluded)
	assert.Equal(t, "新スタッフ", staff.Name)
}

func TestReservationStaffService_Delete_Success(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, Name: "田中医師"}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	err := svc.Delete(context.Background(), 1, 1)

	assert.NoError(t, err)
}

func TestReservationStaffService_Delete_NotFound(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return nil, apperrors.WrapNotFound("reservation_staff", "999")
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return nil
		},
	}
	transactor := &mockTransactor{}
	svc := newTestReservationStaffService(repo, transactor)

	err := svc.Delete(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}
