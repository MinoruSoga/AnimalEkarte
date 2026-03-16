package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- UserAccount モック ----

type mockUserAccountRepository struct {
	findByEmailFn                 func(ctx context.Context, email string) (*model.UserAccount, error)
	findByIDWithMembershipsFn     func(ctx context.Context, userID uint64) (*repository.UserAccountWithMemberships, error)
	findByClinicIDFn              func(ctx context.Context, clinicID uint64) ([]model.UserAccount, error)
	createFn                      func(ctx context.Context, account *model.UserAccount, clinicID uint64, staffID *uint64, isMain bool) error
	updateFn                      func(ctx context.Context, id uint64, fields map[string]any) error
	deleteFn                      func(ctx context.Context, id uint64) error
	findPermissionsFn             func(ctx context.Context, userID, clinicID uint64) ([]model.UserPermission, error)
	setPermissionsFn              func(ctx context.Context, userID, clinicID uint64, permissions []string) error
}

func (m *mockUserAccountRepository) FindByEmail(ctx context.Context, email string) (*model.UserAccount, error) {
	return m.findByEmailFn(ctx, email)
}

func (m *mockUserAccountRepository) FindByIDWithMemberships(ctx context.Context, userID uint64) (*repository.UserAccountWithMemberships, error) {
	return m.findByIDWithMembershipsFn(ctx, userID)
}

func (m *mockUserAccountRepository) FindByClinicID(ctx context.Context, clinicID uint64) ([]model.UserAccount, error) {
	return m.findByClinicIDFn(ctx, clinicID)
}

func (m *mockUserAccountRepository) Create(ctx context.Context, account *model.UserAccount, clinicID uint64, staffID *uint64, isMain bool) error {
	return m.createFn(ctx, account, clinicID, staffID, isMain)
}

func (m *mockUserAccountRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, id, fields)
}

func (m *mockUserAccountRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockUserAccountRepository) FindPermissions(ctx context.Context, userID, clinicID uint64) ([]model.UserPermission, error) {
	return m.findPermissionsFn(ctx, userID, clinicID)
}

func (m *mockUserAccountRepository) SetPermissions(ctx context.Context, userID, clinicID uint64, permissions []string) error {
	return m.setPermissionsFn(ctx, userID, clinicID, permissions)
}

// ---- Tests ----

func TestUserAccountService_FindByEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		repoAccount *model.UserAccount
		repoErr     error
		wantErr     bool
	}{
		{
			name:  "returns user account when found",
			email: "user@example.com",
			repoAccount: &model.UserAccount{
				ID:           1,
				Email:        "user@example.com",
				DisplayName:  "John Doe",
				UserType:     model.UserTypeStaff,
				Status:       model.AccountStatusActive,
				PasswordHash: "hashed_password",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:        "returns error when user not found",
			email:       "nonexistent@example.com",
			repoAccount: nil,
			repoErr:     errors.New("not found"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserAccountRepository{
				findByEmailFn: func(_ context.Context, _ string) (*model.UserAccount, error) {
					return tt.repoAccount, tt.repoErr
				},
			}
			svc := NewUserAccountService(repo)

			result, err := svc.FindByEmail(context.Background(), tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.email, result.Email)
			}
		})
	}
}

func TestUserAccountService_GetMemberships(t *testing.T) {
	tests := []struct {
		name            string
		userID          uint64
		repoData        *repository.UserAccountWithMemberships
		repoErr         error
		wantLen         int
		wantErr         bool
	}{
		{
			name:   "returns user memberships",
			userID: 1,
			repoData: &repository.UserAccountWithMemberships{
				UserAccount: model.UserAccount{ID: 1, Email: "user@example.com"},
				Memberships: []model.UserClinicMembership{
					{UserID: 1, ClinicID: 1, IsMain: true},
					{UserID: 1, ClinicID: 2, IsMain: false},
				},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:   "returns empty memberships when user has none",
			userID: 1,
			repoData: &repository.UserAccountWithMemberships{
				UserAccount: model.UserAccount{ID: 1},
				Memberships: []model.UserClinicMembership{},
			},
			repoErr: nil,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:     "returns error when user not found",
			userID:   999,
			repoData: nil,
			repoErr:  errors.New("not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserAccountRepository{
				findByIDWithMembershipsFn: func(_ context.Context, _ uint64) (*repository.UserAccountWithMemberships, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewUserAccountService(repo)

			memberships, err := svc.GetMemberships(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, memberships, tt.wantLen)
			}
		})
	}
}

func TestUserAccountService_GetWithMemberships(t *testing.T) {
	tests := []struct {
		name      string
		userIDStr string
		repoData  *repository.UserAccountWithMemberships
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "returns user with memberships when found",
			userIDStr: "1",
			repoData: &repository.UserAccountWithMemberships{
				UserAccount: model.UserAccount{ID: 1, Email: "user@example.com"},
				Memberships: []model.UserClinicMembership{},
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:      "returns error on invalid user ID",
			userIDStr: "not_a_number",
			repoData:  nil,
			repoErr:   nil,
			wantErr:   true,
		},
		{
			name:      "returns error when user not found",
			userIDStr: "999",
			repoData:  nil,
			repoErr:   errors.New("not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserAccountRepository{
				findByIDWithMembershipsFn: func(_ context.Context, _ uint64) (*repository.UserAccountWithMemberships, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewUserAccountService(repo)

			result, err := svc.GetWithMemberships(context.Background(), tt.userIDStr)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestUserAccountService_ListUsers(t *testing.T) {
	tests := []struct {
		name        string
		clinicID    uint64
		repoUsers   []model.UserAccount
		repoErr     error
		wantLen     int
		wantErr     bool
	}{
		{
			name:     "returns users for clinic",
			clinicID: 1,
			repoUsers: []model.UserAccount{
				{ID: 1, Email: "doctor@example.com", UserType: model.UserTypeStaff},
				{ID: 2, Email: "nurse@example.com", UserType: model.UserTypeClinicAdmin},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:      "returns empty list when no users exist",
			clinicID:  999,
			repoUsers: []model.UserAccount{},
			repoErr:   nil,
			wantLen:   0,
			wantErr:   false,
		},
		{
			name:      "propagates repository error",
			clinicID:  1,
			repoUsers: nil,
			repoErr:   errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserAccountRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) ([]model.UserAccount, error) {
					return tt.repoUsers, tt.repoErr
				},
			}
			svc := NewUserAccountService(repo)

			users, err := svc.ListUsers(context.Background(), tt.clinicID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, users, tt.wantLen)
			}
		})
	}
}

func TestUserAccountService_CreateUser(t *testing.T) {
	staffID := uint64(5)

	tests := []struct {
		name    string
		clinicID uint64
		input   *CreateUserAccountInput
		repoErr error
		wantErr bool
	}{
		{
			name:    "creates user account successfully",
			clinicID: 1,
			input: &CreateUserAccountInput{
				Email:       "doctor@example.com",
				Password:    "SecurePass123!",
				DisplayName: "Dr. Smith",
				UserType:    model.UserTypeStaff,
				StaffID:     &staffID,
				IsMain:      true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "creates nurse account",
			clinicID: 1,
			input: &CreateUserAccountInput{
				Email:       "nurse@example.com",
				Password:    "NursePass123!",
				DisplayName: "Nurse Jane",
				UserType:    model.UserTypeClinicAdmin,
				IsMain:      false,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when repository fails",
			clinicID: 1,
			input: &CreateUserAccountInput{
				Email:       "user@example.com",
				Password:    "Password123!",
				DisplayName: "User",
				UserType:    model.UserTypeSystemAdmin,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserAccountRepository{
				createFn: func(_ context.Context, _ *model.UserAccount, _ uint64, _ *uint64, _ bool) error {
					return tt.repoErr
				},
			}
			svc := NewUserAccountService(repo)

			account, err := svc.CreateUser(context.Background(), tt.clinicID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, account)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, account)
				assert.Equal(t, tt.input.Email, account.Email)
			}
		})
	}
}

func TestUserAccountService_UpdateUser(t *testing.T) {
	newDisplayName := "Updated Name"
	newUserType := model.UserTypeClinicAdmin

	tests := []struct {
		name          string
		id            uint64
		input         *UpdateUserAccountInput
		repoUpdateErr error
		wantErr       bool
	}{
		{
			name: "updates user successfully",
			id:   1,
			input: &UpdateUserAccountInput{
				DisplayName: &newDisplayName,
				UserType:    &newUserType,
			},
			repoUpdateErr: nil,
			wantErr:       false,
		},
		{
			name:          "returns error when no fields provided",
			id:            1,
			input:         &UpdateUserAccountInput{},
			repoUpdateErr: nil,
			wantErr:       true,
		},
		{
			name: "returns error when update fails",
			id:   1,
			input: &UpdateUserAccountInput{
				DisplayName: &newDisplayName,
			},
			repoUpdateErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserAccountRepository{
				updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
					return tt.repoUpdateErr
				},
			}
			svc := NewUserAccountService(repo)

			err := svc.UpdateUser(context.Background(), tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserAccountService_DeleteUser(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
	}{
		{
			name:    "deletes user successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when user not found",
			id:      999,
			repoErr: errors.New("not found"),
			wantErr: true,
		},
		{
			name:    "returns error when delete fails",
			id:      1,
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserAccountRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewUserAccountService(repo)

			err := svc.DeleteUser(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserAccountService_GetPermissions(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint64
		clinicID      uint64
		repoPermissions []model.UserPermission
		repoErr       error
		wantLen       int
		wantErr       bool
	}{
		{
			name:     "returns permissions for user clinic",
			userID:   1,
			clinicID: 1,
			repoPermissions: []model.UserPermission{
				{ID: 1, UserID: 1, ClinicID: 1, Permission: "view_records"},
				{ID: 2, UserID: 1, ClinicID: 1, Permission: "edit_records"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:            "returns empty permissions when user has none",
			userID:          1,
			clinicID:        1,
			repoPermissions: []model.UserPermission{},
			repoErr:         nil,
			wantLen:         0,
			wantErr:         false,
		},
		{
			name:            "returns error when repository fails",
			userID:          1,
			clinicID:        1,
			repoPermissions: nil,
			repoErr:         errors.New("db error"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserAccountRepository{
				findPermissionsFn: func(_ context.Context, _, _ uint64) ([]model.UserPermission, error) {
					return tt.repoPermissions, tt.repoErr
				},
			}
			svc := NewUserAccountService(repo)

			permissions, err := svc.GetPermissions(context.Background(), tt.userID, tt.clinicID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, permissions, tt.wantLen)
			}
		})
	}
}

func TestUserAccountService_SetPermissions(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint64
		clinicID    uint64
		input       *SetPermissionsInput
		repoErr     error
		wantErr     bool
	}{
		{
			name:     "sets permissions successfully",
			userID:   1,
			clinicID: 1,
			input: &SetPermissionsInput{
				Permissions: []string{"view_records", "edit_records"},
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "sets empty permissions",
			userID:   1,
			clinicID: 1,
			input: &SetPermissionsInput{
				Permissions: []string{},
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when repository fails",
			userID:   1,
			clinicID: 1,
			input: &SetPermissionsInput{
				Permissions: []string{"view_records"},
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserAccountRepository{
				setPermissionsFn: func(_ context.Context, _, _ uint64, _ []string) error {
					return tt.repoErr
				},
			}
			svc := NewUserAccountService(repo)

			err := svc.SetPermissions(context.Background(), tt.userID, tt.clinicID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
