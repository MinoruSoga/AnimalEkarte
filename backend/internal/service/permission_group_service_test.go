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
		{
			name:    "returns error when name is empty",
			input:   CreatePermissionGroupInput{Name: ""},
			wantErr: true,
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
		name       string
		input      *UpdatePermissionGroupInput
		findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
		updateErr  error
		wantErr    bool
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
		{
			name:  "returns error when group not found",
			input: &UpdatePermissionGroupInput{Name: strPtr("名前")},
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.PermissionGroup, error) {
				return nil, apperrors.WrapNotFound("permission_group", "1")
			},
			wantErr: true,
		},
		{
			name:    "returns error for invalid optional name",
			input:   &UpdatePermissionGroupInput{Name: strPtr("   ")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findByIDFn := tt.findByIDFn
			if findByIDFn == nil {
				findByIDFn = func(_ context.Context, _, _ uint64) (*model.PermissionGroup, error) {
					return existing, nil
				}
			}
			repo := &mockPermissionGroupRepository{
				findByIDFn: findByIDFn,
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

// F-3 回帰: FindAllGroupIDsByStaffID が error を返すとき、自己参照チェックを
// 素通り（fail-open）させてはならない。旧実装は error 時に staffGroupIDs を空へ
// フォールバックし、actor が自分の所属グループの master-permission edit を削除する
// 危険な編集を許していた（BUG-140 の無効化）。security control として fail-closed
// で拒否し、検証不能時に repo.UpdateRules を呼ばないことを検証する。
func TestPermissionGroupService_UpdateRules_FailClosedOnGroupLookupError(t *testing.T) {
	// 編集対象グループに actor が所属し、master-permission edit を外す入力。
	// 所属取得が成功していれば validateNotSelfReference が拒否するケース。
	inputs := []SetPermissionGroupRulesInput{
		{Resource: string(model.ResourceOwners), CanView: true},
	}

	setRulesCalled := false
	repo := &mockPermissionGroupRepository{
		getGroupIDsByStaffIDFn: func(_ context.Context, _ uint64) ([]uint64, error) {
			return nil, errors.New("db error") // 所属グループ取得が失敗
		},
		setRulesFn: func(_ context.Context, _ uint64, _ []model.PermissionGroupRule) error {
			setRulesCalled = true
			return nil
		},
	}
	svc := NewPermissionGroupService(repo)

	// groupID=1 を actorStaffID=10 が編集。所属取得失敗 → fail-closed で拒否すべき。
	err := svc.UpdateRules(context.Background(), 1, inputs, 10)

	assert.Error(t, err, "所属グループ取得が失敗したら fail-closed で拒否すべき（自己参照チェックを素通りさせない）")
	assert.False(t, setRulesCalled, "検証不能時に repo.UpdateRules を呼んではならない")
}

// ---- List ----

func TestPermissionGroupService_List(t *testing.T) {
	items := []model.PermissionGroup{{ID: 1, Name: "管理者"}}

	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{
			name:    "lists permission groups successfully",
			wantErr: false,
		},
		{
			name:    "propagates repository error",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPermissionGroupRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.PermissionGroup, error) {
					return items, tt.repoErr
				},
			}
			svc := NewPermissionGroupService(repo)
			got, err := svc.List(context.Background(), 1)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, items, got)
			}
		})
	}
}

// ---- newPermissionGroupServiceImpl / GetEffectivePermissions ----

func TestPermissionGroupService_GetEffectivePermissions(t *testing.T) {
	rules := []model.PermissionGroupRule{{Resource: string(model.ResourceOwners), CanView: true}}

	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{
			name:    "returns effective permissions successfully",
			wantErr: false,
		},
		{
			name:    "propagates repository error",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPermissionGroupRepository{
				getEffectivePermissionsByStaffID: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
					return rules, tt.repoErr
				},
			}
			// newPermissionGroupServiceImpl は service.go の DI 配線で使用される具体型コンストラクタ。
			// EffectivePermissionService と PermissionGroupService の両方を実装する。
			impl := newPermissionGroupServiceImpl(repo)
			var _ PermissionGroupService = impl
			var _ EffectivePermissionService = impl

			got, err := impl.GetEffectivePermissions(context.Background(), 5, 1)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, rules, got)
			}
		})
	}
}

// ---- buildPermissionGroupUpdate ----

func TestBuildPermissionGroupUpdate(t *testing.T) {
	name := "新名称"
	desc := "説明"
	color := "#FFFFFF"
	sortOrder := 3
	active := true

	tests := []struct {
		name  string
		input *UpdatePermissionGroupInput
		want  map[string]any
	}{
		{
			name: "all fields set",
			input: &UpdatePermissionGroupInput{
				Name:        &name,
				Description: &desc,
				Color:       &color,
				SortOrder:   &sortOrder,
				IsActive:    &active,
			},
			want: map[string]any{
				colPermissionGroupName:        name,
				colPermissionGroupDescription: desc,
				colPermissionGroupColor:       color,
				colPermissionGroupSortOrder:   sortOrder,
				colPermissionGroupIsActive:    active,
			},
		},
		{
			name:  "no fields set returns empty map",
			input: &UpdatePermissionGroupInput{},
			want:  map[string]any{},
		},
		{
			name:  "only name set",
			input: &UpdatePermissionGroupInput{Name: &name},
			want:  map[string]any{colPermissionGroupName: name},
		},
		{
			name:  "only is_active set to false",
			input: &UpdatePermissionGroupInput{IsActive: boolPtr(false)},
			want:  map[string]any{colPermissionGroupIsActive: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPermissionGroupUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---- UpdateRules validation branches (経由: service.UpdateRules) ----

func TestPermissionGroupService_UpdateRules_ValidationErrors(t *testing.T) {
	t.Run("rejects duplicate rules before checking self reference", func(t *testing.T) {
		setRulesCalled := false
		repo := &mockPermissionGroupRepository{
			setRulesFn: func(_ context.Context, _ uint64, _ []model.PermissionGroupRule) error {
				setRulesCalled = true
				return nil
			},
		}
		svc := NewPermissionGroupService(repo)

		inputs := []SetPermissionGroupRulesInput{
			{Resource: string(model.ResourceOwners)},
			{Resource: string(model.ResourceOwners)},
		}
		err := svc.UpdateRules(context.Background(), 1, inputs, 10)

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.False(t, setRulesCalled)
	})

	t.Run("rejects removing own master-permission edit", func(t *testing.T) {
		setRulesCalled := false
		repo := &mockPermissionGroupRepository{
			getGroupIDsByStaffIDFn: func(_ context.Context, _ uint64) ([]uint64, error) {
				return []uint64{1}, nil
			},
			setRulesFn: func(_ context.Context, _ uint64, _ []model.PermissionGroupRule) error {
				setRulesCalled = true
				return nil
			},
		}
		svc := NewPermissionGroupService(repo)

		inputs := []SetPermissionGroupRulesInput{
			{Resource: string(model.ResourceOwners), CanView: true},
		}
		err := svc.UpdateRules(context.Background(), 1, inputs, 10)

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.False(t, setRulesCalled)
	})

	t.Run("allows self-referencing group when master-permission edit is retained", func(t *testing.T) {
		repo := &mockPermissionGroupRepository{
			getGroupIDsByStaffIDFn: func(_ context.Context, _ uint64) ([]uint64, error) {
				return []uint64{1}, nil
			},
			setRulesFn: func(_ context.Context, _ uint64, _ []model.PermissionGroupRule) error {
				return nil
			},
		}
		svc := NewPermissionGroupService(repo)

		inputs := []SetPermissionGroupRulesInput{
			{Resource: string(model.ResourceMasterPermission), CanEdit: true},
		}
		err := svc.UpdateRules(context.Background(), 1, inputs, 10)

		assert.NoError(t, err)
	})
}

// ---- validateNoDuplicateRules ----

func TestValidateNoDuplicateRules(t *testing.T) {
	tests := []struct {
		name    string
		rules   []model.PermissionGroupRule
		wantErr bool
	}{
		{
			name:    "no rules is valid",
			rules:   nil,
			wantErr: false,
		},
		{
			name: "unique valid resources",
			rules: []model.PermissionGroupRule{
				{Resource: string(model.ResourceOwners)},
				{Resource: string(model.ResourceMasterPermission)},
			},
			wantErr: false,
		},
		{
			name:    "empty resource name is rejected",
			rules:   []model.PermissionGroupRule{{Resource: ""}},
			wantErr: true,
		},
		{
			name:    "invalid resource name is rejected",
			rules:   []model.PermissionGroupRule{{Resource: "bogus-resource"}},
			wantErr: true,
		},
		{
			name: "duplicate resource name is rejected",
			rules: []model.PermissionGroupRule{
				{Resource: string(model.ResourceOwners)},
				{Resource: string(model.ResourceOwners)},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNoDuplicateRules(tt.rules)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- validateNotSelfReference ----

func TestValidateNotSelfReference(t *testing.T) {
	const groupID uint64 = 5

	tests := []struct {
		name          string
		staffGroupIDs []uint64
		rules         []model.PermissionGroupRule
		wantErr       bool
	}{
		{
			name:          "not a self-referencing group",
			staffGroupIDs: []uint64{1, 2},
			rules:         []model.PermissionGroupRule{{Resource: string(model.ResourceMasterPermission), CanEdit: false}},
			wantErr:       false,
		},
		{
			name:          "self-referencing group retains master-permission edit",
			staffGroupIDs: []uint64{groupID},
			rules:         []model.PermissionGroupRule{{Resource: string(model.ResourceMasterPermission), CanEdit: true}},
			wantErr:       false,
		},
		{
			name:          "self-referencing group removes master-permission edit",
			staffGroupIDs: []uint64{groupID},
			rules:         []model.PermissionGroupRule{{Resource: string(model.ResourceOwners), CanEdit: true}},
			wantErr:       true,
		},
		{
			name:          "self-referencing group with empty rules",
			staffGroupIDs: []uint64{groupID},
			rules:         nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNotSelfReference(groupID, tt.rules, tt.staffGroupIDs)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
