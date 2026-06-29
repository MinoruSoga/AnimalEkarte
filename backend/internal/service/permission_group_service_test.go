package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- PermissionGroup モック ----

type mockPermissionGroupRepository struct {
	findAllFn                        func(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error)
	findByIDFn                       func(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
	createFn                         func(ctx context.Context, group *model.PermissionGroup) error
	updateFieldsFn                   func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PermissionGroup, error)
	deleteFn                         func(ctx context.Context, clinicID, id uint64) error
	setRulesFn                       func(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error
	countStaffsByGroupIDFn           func(ctx context.Context, clinicID, groupID uint64) (int64, error)
	reorderErr                       error
	getEffectivePermissionsByStaffID func(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error)
	getGroupIDsByStaffIDFn           func(ctx context.Context, staffID uint64) ([]uint64, error)
	setStaffGroupsFn                 func(ctx context.Context, staffID uint64, groupIDs []uint64) error
}

func (m *mockPermissionGroupRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
	return m.findAllFn(ctx, clinicID)
}
func (m *mockPermissionGroupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
	return m.findByIDFn(ctx, clinicID, id)
}
func (m *mockPermissionGroupRepository) Create(ctx context.Context, group *model.PermissionGroup) error {
	return m.createFn(ctx, group)
}
func (m *mockPermissionGroupRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PermissionGroup, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}
func (m *mockPermissionGroupRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}
func (m *mockPermissionGroupRepository) UpdateRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error {
	return m.setRulesFn(ctx, groupID, rules)
}
func (m *mockPermissionGroupRepository) CountUsageByGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error) {
	return m.countStaffsByGroupIDFn(ctx, clinicID, groupID)
}
func (m *mockPermissionGroupRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}
func (m *mockPermissionGroupRepository) FindAllEffectivePermissionsByStaffID(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error) {
	return m.getEffectivePermissionsByStaffID(ctx, staffID, clinicID)
}
func (m *mockPermissionGroupRepository) FindAllGroupIDsByStaffID(ctx context.Context, staffID uint64) ([]uint64, error) {
	if m.getGroupIDsByStaffIDFn != nil {
		return m.getGroupIDsByStaffIDFn(ctx, staffID)
	}
	return nil, nil
}
func (m *mockPermissionGroupRepository) UpdateStaffGroups(ctx context.Context, _, staffID uint64, groupIDs []uint64) error {
	if m.setStaffGroupsFn != nil {
		return m.setStaffGroupsFn(ctx, staffID, groupIDs)
	}
	return nil
}

// ---- Tests ----

func TestPermissionGroupService_GetByID(t *testing.T) {
	existing := &model.PermissionGroup{ID: 1, ClinicID: 1, Name: "管理者"}

	tests := []struct {
		name         string
		id           uint64
		repoGroup    *model.PermissionGroup
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:      "returns group when found",
			id:        1,
			repoGroup: existing,
			wantErr:   false,
		},
		{
			name:         "returns not found error when group does not exist",
			id:           999,
			repoErr:      apperrors.WrapNotFound("permission_group", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPermissionGroupRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.PermissionGroup, error) {
					return tt.repoGroup, tt.repoErr
				},
			}
			svc := NewPermissionGroupService(repo)
			result, err := svc.GetByID(context.Background(), 1, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoGroup, result)
			}
		})
	}
}

func TestPermissionGroupService_Create(t *testing.T) {
	tests := []struct {
		name      string
		input     CreatePermissionGroupInput
		createErr error
		wantErr   bool
	}{
		{
			name:    "creates group successfully",
			input:   CreatePermissionGroupInput{Name: "管理者"},
			wantErr: false,
		},
		{
			name:      "propagates repository error",
			input:     CreatePermissionGroupInput{Name: "管理者"},
			createErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPermissionGroupRepository{
				createFn: func(_ context.Context, _ *model.PermissionGroup) error {
					return tt.createErr
				},
			}
			svc := NewPermissionGroupService(repo)
			_, err := svc.Create(context.Background(), 1, &tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPermissionGroupService_Update(t *testing.T) {
	existing := &model.PermissionGroup{ID: 1, ClinicID: 1, Name: "既存グループ"}

	tests := []struct {
		name      string
		input     *UpdatePermissionGroupInput
		updateErr error
		wantErr   bool
	}{
		{
			name:    "updates group successfully",
			input:   &UpdatePermissionGroupInput{Name: strPtr("新グループ名")},
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   &UpdatePermissionGroupInput{},
			wantErr: true,
		},
		{
			name:      "propagates update error",
			input:     &UpdatePermissionGroupInput{Name: strPtr("名前")},
			updateErr: errors.New("update failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPermissionGroupRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.PermissionGroup, error) {
					return existing, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.PermissionGroup, error) {
					return existing, tt.updateErr
				},
			}
			svc := NewPermissionGroupService(repo)
			result, err := svc.Update(context.Background(), 1, 1, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestPermissionGroupService_Update_NilInput(t *testing.T) {
	repo := &mockPermissionGroupRepository{}
	svc := NewPermissionGroupService(repo)
	result, err := svc.Update(context.Background(), 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestPermissionGroupService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		staffCount   int64
		countErr     error
		deleteErr    error
		wantErr      bool
		wantConflict bool
	}{
		{
			name:    "deletes group successfully",
			wantErr: false,
		},
		{
			name:         "returns conflict when group has assigned staffs",
			staffCount:   2,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:     "propagates count error",
			countErr: errors.New("db error"),
			wantErr:  true,
		},
		{
			name:      "propagates delete error",
			deleteErr: errors.New("delete failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPermissionGroupRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.PermissionGroup, error) {
					return &model.PermissionGroup{ID: id}, nil
				},
				countStaffsByGroupIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.staffCount, tt.countErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := NewPermissionGroupService(repo)
			err := svc.Delete(context.Background(), 1, 1)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPermissionGroupService_Reorder(t *testing.T) {
	tests := []struct {
		name       string
		ids        []uint64
		reorderErr error
		wantErr    bool
	}{
		{
			name:    "reorders groups successfully",
			ids:     []uint64{3, 1, 2},
			wantErr: false,
		},
		{
			name:    "returns error for empty ids",
			ids:     []uint64{},
			wantErr: true,
		},
		{
			name:       "propagates repository error",
			ids:        []uint64{1, 2},
			reorderErr: errors.New("reorder failed"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPermissionGroupRepository{
				reorderErr: tt.reorderErr,
			}
			svc := NewPermissionGroupService(repo)
			err := svc.Reorder(context.Background(), 1, tt.ids)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPermissionGroupService_SetRules(t *testing.T) {
	inputs := []SetPermissionGroupRulesInput{
		{Resource: string(model.ResourceOwners), CanView: true},
	}

	tests := []struct {
		name        string
		inputs      []SetPermissionGroupRulesInput
		setRulesErr error
		wantErr     bool
	}{
		{
			name:    "sets rules successfully",
			inputs:  inputs,
			wantErr: false,
		},
		{
			name:        "propagates repository error",
			inputs:      inputs,
			setRulesErr: errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPermissionGroupRepository{
				setRulesFn: func(_ context.Context, _ uint64, _ []model.PermissionGroupRule) error {
					return tt.setRulesErr
				},
			}
			svc := NewPermissionGroupService(repo)
			err := svc.UpdateRules(context.Background(), 1, tt.inputs, 0)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
