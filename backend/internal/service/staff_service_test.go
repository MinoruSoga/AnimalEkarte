package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockStaffRepository は StaffRepository のテスト用モック実装
type mockStaffRepository struct {
	findAllFn           func(ctx context.Context, clinicID uint64, role *string) ([]model.Staff, error)
	findByIDFn          func(ctx context.Context, id uint64) (*model.Staff, error)
	createWithAccountFn func(ctx context.Context, staff *model.Staff, account *model.UserAccount, membership *model.UserClinicMembership) error
	updateFn            func(ctx context.Context, staff *model.Staff) error
	deleteFn            func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockStaffRepository) FindAll(ctx context.Context, clinicID uint64, role *string) ([]model.Staff, error) {
	return m.findAllFn(ctx, clinicID, role)
}

func (m *mockStaffRepository) FindByID(ctx context.Context, id uint64) (*model.Staff, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockStaffRepository) CreateWithAccount(ctx context.Context, staff *model.Staff, account *model.UserAccount, membership *model.UserClinicMembership) error {
	return m.createWithAccountFn(ctx, staff, account, membership)
}

func (m *mockStaffRepository) Update(ctx context.Context, staff *model.Staff) error {
	return m.updateFn(ctx, staff)
}

func (m *mockStaffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func TestStaffService_List(t *testing.T) {
	tests := []struct {
		name       string
		clinicID   uint64
		role       *string
		repoStaffs []model.Staff
		repoErr    error
		wantLen    int
		wantErr    bool
	}{
		{
			name:     "returns all staffs without role filter",
			clinicID: 1,
			role:     nil,
			repoStaffs: []model.Staff{
				{ID: 1, ClinicID: 1, Name: "山田 太郎", StaffRole: model.StaffRoleVeterinarian},
				{ID: 2, ClinicID: 1, Name: "鈴木 花子", StaffRole: model.StaffRoleNurse},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "filters staffs by role",
			clinicID: 1,
			role:     ptrString("veterinarian"),
			repoStaffs: []model.Staff{
				{ID: 1, ClinicID: 1, Name: "山田 太郎", StaffRole: model.StaffRoleVeterinarian},
			},
			repoErr: nil,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:       "returns empty list when no staffs exist",
			clinicID:   1,
			role:       nil,
			repoStaffs: []model.Staff{},
			repoErr:    nil,
			wantLen:    0,
			wantErr:    false,
		},
		{
			name:       "propagates repository error",
			clinicID:   1,
			role:       nil,
			repoStaffs: nil,
			repoErr:    errors.New("db connection error"),
			wantLen:    0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedRole := (*string)(nil)
			repo := &mockStaffRepository{
				findAllFn: func(_ context.Context, _ uint64, role *string) ([]model.Staff, error) {
					capturedRole = role
					return tt.repoStaffs, tt.repoErr
				},
			}
			svc := NewStaffService(repo)

			staffs, err := svc.List(context.Background(), tt.clinicID, tt.role)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, staffs, tt.wantLen)
				assert.Equal(t, tt.role, capturedRole)
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
			repoStaff: &model.Staff{ID: 10, ClinicID: 1, Name: "山田 太郎", StaffRole: model.StaffRoleVeterinarian},
			repoErr:   nil,
			wantStaff: &model.Staff{ID: 10, ClinicID: 1, Name: "山田 太郎", StaffRole: model.StaffRoleVeterinarian},
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
			svc := NewStaffService(repo)

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
	svc := NewStaffService(repo)

	staff, err := svc.GetByID(context.Background(), 999)

	assert.Nil(t, staff)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

// TestStaffService_CreateWithAccount_Success は bcrypt を呼び出すため正常ケースのみテストする。
func TestStaffService_CreateWithAccount_Success(t *testing.T) {
	repo := &mockStaffRepository{
		createWithAccountFn: func(_ context.Context, staff *model.Staff, _ *model.UserAccount, _ *model.UserClinicMembership) error {
			// IDをシミュレート
			staff.ID = 1
			return nil
		},
	}
	svc := NewStaffService(repo)

	input := &CreateStaffInput{
		ClinicID:  1,
		Name:      "新規 スタッフ",
		StaffRole: model.StaffRoleVeterinarian,
		Email:     "new@example.com",
		Password:  "securepassword123",
	}

	staff, err := svc.CreateWithAccount(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.Equal(t, "新規 スタッフ", staff.Name)
	assert.Equal(t, model.StaffRoleVeterinarian, staff.StaffRole)
	assert.True(t, staff.IsActive)
}

func TestStaffService_CreateWithAccount_RepositoryError(t *testing.T) {
	repo := &mockStaffRepository{
		createWithAccountFn: func(_ context.Context, _ *model.Staff, _ *model.UserAccount, _ *model.UserClinicMembership) error {
			return errors.New("db connection error")
		},
	}
	svc := NewStaffService(repo)

	input := &CreateStaffInput{
		ClinicID:  1,
		Name:      "エラー スタッフ",
		StaffRole: model.StaffRoleNurse,
		Email:     "error@example.com",
		Password:  "password123",
	}

	staff, err := svc.CreateWithAccount(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, staff)
}

func TestStaffService_CreateWithAccount_DuplicateEmail(t *testing.T) {
	repo := &mockStaffRepository{
		createWithAccountFn: func(_ context.Context, _ *model.Staff, _ *model.UserAccount, _ *model.UserClinicMembership) error {
			return apperrors.WrapAlreadyExists("user_account", "existing@example.com")
		},
	}
	svc := NewStaffService(repo)

	input := &CreateStaffInput{
		ClinicID:  1,
		Name:      "重複 スタッフ",
		StaffRole: model.StaffRoleReception,
		Email:     "existing@example.com",
		Password:  "password123",
	}

	staff, err := svc.CreateWithAccount(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, apperrors.IsAlreadyExists(err))
}

func TestStaffService_Update(t *testing.T) {
	tests := []struct {
		name    string
		staff   *model.Staff
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name: "updates staff successfully",
			staff: &model.Staff{
				ID:        1,
				ClinicID:  1,
				Name:      "更新後 スタッフ",
				StaffRole: model.StaffRoleVeterinarian,
			},
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name: "returns not found error when staff does not exist",
			staff: &model.Staff{
				ID:        999,
				ClinicID:  1,
				Name:      "存在しないスタッフ",
				StaffRole: model.StaffRoleNurse,
			},
			repoErr: apperrors.WrapNotFound("staff", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name: "returns error on repository failure",
			staff: &model.Staff{
				ID:        1,
				ClinicID:  1,
				Name:      "エラーケース",
				StaffRole: model.StaffRoleTrimmer,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
			wantNF:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{
				updateFn: func(_ context.Context, _ *model.Staff) error {
					return tt.repoErr
				},
			}
			svc := NewStaffService(repo)

			err := svc.Update(context.Background(), tt.staff)

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

func TestStaffService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "deletes staff successfully",
			clinicID: 1,
			id:       10,
			repoErr:  nil,
			wantErr:  false,
			wantNF:   false,
		},
		{
			name:     "returns not found error when staff does not exist",
			clinicID: 1,
			id:       999,
			repoErr:  apperrors.WrapNotFound("staff", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoErr:  errors.New("db error"),
			wantErr:  true,
			wantNF:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewStaffService(repo)

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
