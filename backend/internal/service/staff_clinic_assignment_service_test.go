package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mockStaffClinicAssignmentRepository ----

type mockStaffClinicAssignmentRepository struct {
	findByStaffIDFn         func(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error)
	countByStaffAndClinicFn func(ctx context.Context, staffID, clinicID uint64) (int64, error)
	createFn                func(ctx context.Context, assignment *model.StaffClinicAssignment) error
	deleteFn                func(ctx context.Context, staffID uint64) error
}

func (m *mockStaffClinicAssignmentRepository) FindByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
	if m.findByStaffIDFn != nil {
		return m.findByStaffIDFn(ctx, staffID)
	}
	return []model.StaffClinicAssignment{}, nil
}

func (m *mockStaffClinicAssignmentRepository) CountByStaffAndClinic(ctx context.Context, staffID, clinicID uint64) (int64, error) {
	if m.countByStaffAndClinicFn != nil {
		return m.countByStaffAndClinicFn(ctx, staffID, clinicID)
	}
	return 0, nil
}

func (m *mockStaffClinicAssignmentRepository) LockActiveByStaffAndClinic(
	_ context.Context,
	staffID, clinicID uint64,
) (*model.StaffClinicAssignment, error) {
	return &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}, nil
}

func (m *mockStaffClinicAssignmentRepository) LockActiveByStaff(
	_ context.Context,
	_ uint64,
) ([]model.StaffClinicAssignment, error) {
	return nil, nil
}

func (m *mockStaffClinicAssignmentRepository) Create(ctx context.Context, assignment *model.StaffClinicAssignment) error {
	if m.createFn != nil {
		return m.createFn(ctx, assignment)
	}
	return nil
}

func (m *mockStaffClinicAssignmentRepository) RestoreOrCreate(
	ctx context.Context,
	assignment *model.StaffClinicAssignment,
) error {
	return m.Create(ctx, assignment)
}

func (m *mockStaffClinicAssignmentRepository) Delete(ctx context.Context, staffID uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, staffID)
	}
	return nil
}

// ---- Tests: NewStaffClinicAssignmentService ----

func TestNewStaffClinicAssignmentService(t *testing.T) {
	repo := &mockStaffClinicAssignmentRepository{}
	svc := NewStaffClinicAssignmentService(repo)
	assert.NotNil(t, svc)
}

// ---- Tests: FindAllByStaffID ----

func TestStaffClinicAssignmentService_FindAllByStaffID(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockStaffClinicAssignmentRepository
		wantLen int
		wantErr bool
	}{
		{
			name: "正常: 紐付け一覧を返す",
			repo: &mockStaffClinicAssignmentRepository{
				findByStaffIDFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
					return []model.StaffClinicAssignment{
						{ID: 1, StaffID: staffID, ClinicID: 10},
						{ID: 2, StaffID: staffID, ClinicID: 20},
					}, nil
				},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "正常: 紐付けなし → 空リストを返す",
			repo: &mockStaffClinicAssignmentRepository{
				findByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
					return []model.StaffClinicAssignment{}, nil
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "エラー: repo.FindByStaffID がエラー",
			repo: &mockStaffClinicAssignmentRepository{
				findByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewStaffClinicAssignmentService(tt.repo)

			got, err := svc.FindAllByStaffID(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

// ---- Tests: Create ----

func TestStaffClinicAssignmentService_Create(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{
			name:    "正常: 紐付けを作成する",
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "エラー: repo.Create がエラー",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured *model.StaffClinicAssignment
			repo := &mockStaffClinicAssignmentRepository{
				createFn: func(_ context.Context, a *model.StaffClinicAssignment) error {
					captured = a
					return tt.repoErr
				},
			}
			svc := NewStaffClinicAssignmentService(repo)

			assignment := &model.StaffClinicAssignment{StaffID: 1, ClinicID: 10, IsMain: true}
			err := svc.Create(context.Background(), assignment)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, captured)
				assert.Equal(t, uint64(1), captured.StaffID)
				assert.Equal(t, uint64(10), captured.ClinicID)
			}
		})
	}
}
