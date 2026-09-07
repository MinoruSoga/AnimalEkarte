package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type atomicPermissionGroupRepositoryStub struct {
	*mockPermissionGroupRepository
	createWithRulesFn func(
		context.Context,
		*model.PermissionGroup,
		[]model.PermissionGroupRule,
	) (*model.PermissionGroup, error)
	updateWithRulesFn func(
		context.Context,
		uint64,
		uint64,
		UpdatePermissionGroupInput,
		[]model.PermissionGroupRule,
	) (*model.PermissionGroup, error)
}

func (s *atomicPermissionGroupRepositoryStub) CreateWithRules(
	ctx context.Context,
	group *model.PermissionGroup,
	rules []model.PermissionGroupRule,
) (*model.PermissionGroup, error) {
	return s.createWithRulesFn(ctx, group, rules)
}

func (s *atomicPermissionGroupRepositoryStub) UpdateWithRules(
	ctx context.Context,
	clinicID, id uint64,
	cmd UpdatePermissionGroupInput,
	rules []model.PermissionGroupRule,
) (*model.PermissionGroup, error) {
	return s.updateWithRulesFn(ctx, clinicID, id, cmd, rules)
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
		{
			name:    "fails closed when repository returns a nil group",
			id:      1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPermissionGroupRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.PermissionGroup, error) {
					return tt.repoGroup, tt.repoErr
				},
			}
			svc := newPermissionGroupServiceImpl(repo)
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
			svc := newPermissionGroupServiceImpl(repo)
			_, err := svc.Create(
				context.Background(),
				1,
				&tt.input,
				testPermissionMutationAudit(
					1,
					10,
					model.AuditActionPermissionGroupCreate,
					"permission_group",
				),
			)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPermissionGroupService_Create_IsActiveFalsePropagates(t *testing.T) {
	var created *model.PermissionGroup
	repo := &mockPermissionGroupRepository{
		createFn: func(_ context.Context, group *model.PermissionGroup) error {
			created = group
			group.ID = 11
			return nil
		},
	}
	svc := newPermissionGroupServiceImpl(repo)

	group, err := svc.Create(
		context.Background(),
		1,
		&CreatePermissionGroupInput{
			Name:     "inactive group",
			Color:    "#112233",
			IsActive: false,
		},
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionGroupCreate,
			"permission_group",
		),
	)

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.False(t, created.IsActive, "service must pass resolved false to repository")
	assert.False(t, group.IsActive)
}

func TestPermissionGroupService_Create_IsActiveTruePropagates(t *testing.T) {
	var created *model.PermissionGroup
	repo := &mockPermissionGroupRepository{
		createFn: func(_ context.Context, group *model.PermissionGroup) error {
			created = group
			group.ID = 12
			return nil
		},
	}
	svc := newPermissionGroupServiceImpl(repo)

	group, err := svc.Create(
		context.Background(),
		1,
		&CreatePermissionGroupInput{
			Name:     "active group",
			Color:    "#112233",
			IsActive: true,
		},
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionGroupCreate,
			"permission_group",
		),
	)

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.True(t, created.IsActive)
	assert.True(t, group.IsActive)
}

func TestPermissionGroupService_CreateWithRules_AuditsFinalAggregateInTransaction(
	t *testing.T,
) {
	transactor := &permissionAuditTransactor{}
	auditLogger := &permissionAuditTxLogger{}
	baseRepo := &mockPermissionGroupRepository{}
	repo := &atomicPermissionGroupRepositoryStub{
		mockPermissionGroupRepository: baseRepo,
		createWithRulesFn: func(
			ctx context.Context,
			group *model.PermissionGroup,
			rules []model.PermissionGroupRule,
		) (*model.PermissionGroup, error) {
			assert.True(t, ctx.Value(permissionAuditTxContextKey{}) == true)
			require.Len(t, rules, 1)
			group.ID = 7
			group.Rules = []model.PermissionGroupRule{{
				ID:       9,
				GroupID:  group.ID,
				Resource: rules[0].Resource,
				CanView:  rules[0].CanView,
			}}
			return group, nil
		},
	}
	svc := NewPermissionGroupService(repo, transactor, auditLogger)

	group, err := svc.Create(
		context.Background(),
		23,
		&CreatePermissionGroupInput{
			Name:  "created with rules",
			Color: "#123456",
			Rules: []SetPermissionGroupRulesInput{{
				Resource: string(model.ResourceOwners),
				CanView:  true,
			}},
		},
		permissionAuditInput(
			model.AuditActionPermissionGroupCreate,
			"permission_group",
		),
	)

	require.NoError(t, err)
	require.Len(t, group.Rules, 1)
	require.Len(t, auditLogger.entries, 1)
	assert.True(t, auditLogger.txSeen[0])
	newValue := auditLogger.entries[0].NewValue.(map[string]any)
	assert.Equal(t, []map[string]any{{
		"resource":   "owners",
		"can_view":   true,
		"can_create": false,
		"can_edit":   false,
		"can_delete": false,
	}}, newValue["rules"])
}

func TestPermissionGroupService_Update(t *testing.T) {
	existing := &model.PermissionGroup{ID: 1, ClinicID: 1, Name: "既存グループ"}

	tests := []struct {
		name       string
		input      *UpdatePermissionGroupInput
		findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
		updateErr  error
		updateNil  bool
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
			name:      "fails closed when update returns a nil group",
			input:     &UpdatePermissionGroupInput{Name: strPtr("名前")},
			updateNil: true,
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
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdatePermissionGroupInput) (*model.PermissionGroup, error) {
					if tt.updateNil {
						return nil, nil
					}
					return existing, tt.updateErr
				},
			}
			svc := newPermissionGroupServiceImpl(repo)
			result, err := svc.Update(
				context.Background(),
				1,
				1,
				tt.input,
				testPermissionMutationAudit(
					1,
					10,
					model.AuditActionPermissionGroupUpdate,
					"permission_group",
				),
			)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestPermissionGroupService_UpdateWithRules_AuditsFinalAggregateInTransaction(
	t *testing.T,
) {
	transactor := &permissionAuditTransactor{}
	auditLogger := &permissionAuditTxLogger{}
	oldGroup := &model.PermissionGroup{
		ID:       7,
		ClinicID: 23,
		Name:     "before",
		Rules: []model.PermissionGroupRule{{
			Resource: "medical_record",
			CanView:  true,
		}},
	}
	baseRepo := &mockPermissionGroupRepository{
		findByIDFn: func(
			context.Context,
			uint64,
			uint64,
		) (*model.PermissionGroup, error) {
			return oldGroup, nil
		},
		lockByIDFn: func(
			context.Context,
			uint64,
			uint64,
		) (*model.PermissionGroup, error) {
			return oldGroup, nil
		},
	}
	repo := &atomicPermissionGroupRepositoryStub{
		mockPermissionGroupRepository: baseRepo,
		updateWithRulesFn: func(
			ctx context.Context,
			clinicID, id uint64,
			cmd UpdatePermissionGroupInput,
			rules []model.PermissionGroupRule,
		) (*model.PermissionGroup, error) {
			assert.True(t, ctx.Value(permissionAuditTxContextKey{}) == true)
			assert.Equal(t, uint64(23), clinicID)
			assert.Equal(t, uint64(7), id)
			require.NotNil(t, cmd.Name)
			assert.Equal(t, "after", *cmd.Name)
			require.Len(t, rules, 1)
			return &model.PermissionGroup{
				ID:       id,
				ClinicID: clinicID,
				Name:     "after",
				Rules: []model.PermissionGroupRule{{
					ID:       10,
					GroupID:  id,
					Resource: rules[0].Resource,
					CanEdit:  rules[0].CanEdit,
				}},
			}, nil
		},
	}
	svc := NewPermissionGroupService(repo, transactor, auditLogger)
	name := "after"

	group, err := svc.Update(
		context.Background(),
		23,
		7,
		&UpdatePermissionGroupInput{
			Name: &name,
			Rules: []SetPermissionGroupRulesInput{{
				Resource: string(model.ResourceOwners),
				CanEdit:  true,
			}},
		},
		permissionAuditInput(
			model.AuditActionPermissionGroupUpdate,
			"permission_group",
		),
	)

	require.NoError(t, err)
	require.Len(t, group.Rules, 1)
	require.Len(t, auditLogger.entries, 1)
	assert.True(t, auditLogger.txSeen[0])
	newValue := auditLogger.entries[0].NewValue.(map[string]any)
	assert.Equal(t, []map[string]any{{
		"resource":   "owners",
		"can_view":   false,
		"can_create": false,
		"can_edit":   true,
		"can_delete": false,
	}}, newValue["rules"])
}

func TestPermissionGroupService_Update_NilInput(t *testing.T) {
	repo := &mockPermissionGroupRepository{}
	svc := newPermissionGroupServiceImpl(repo)
	result, err := svc.Update(
		context.Background(),
		1,
		1,
		nil,
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionGroupUpdate,
			"permission_group",
		),
	)
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
				findByIDFn: func(
					_ context.Context,
					clinicID, id uint64,
				) (*model.PermissionGroup, error) {
					return &model.PermissionGroup{
						ID:       id,
						ClinicID: clinicID,
					}, nil
				},
				countStaffsByGroupIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.staffCount, tt.countErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := newPermissionGroupServiceImpl(repo)
			err := svc.Delete(
				context.Background(),
				1,
				1,
				testPermissionMutationAudit(
					1,
					10,
					model.AuditActionPermissionGroupDelete,
					"permission_group",
				),
			)
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
			svc := newPermissionGroupServiceImpl(repo)
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
				setRulesFn: func(_ context.Context, clinicID, _ uint64, _ []model.PermissionGroupRule) error {
					assert.Equal(t, uint64(1), clinicID)
					return tt.setRulesErr
				},
			}
			svc := newPermissionGroupServiceImpl(repo)
			_, err := svc.UpdateRules(
				context.Background(),
				1,
				1,
				tt.inputs,
				10,
				testPermissionMutationAudit(
					1,
					10,
					model.AuditActionPermissionRulesUpdate,
					"permission_group_rules",
				),
			)
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
		getGroupIDsByStaffIDFn: func(_ context.Context, clinicID, staffID uint64) ([]uint64, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(10), staffID)
			return nil, errors.New("db error") // 所属グループ取得が失敗
		},
		setRulesFn: func(_ context.Context, _, _ uint64, _ []model.PermissionGroupRule) error {
			setRulesCalled = true
			return nil
		},
	}
	svc := newPermissionGroupServiceImpl(repo)

	// groupID=1 を actorStaffID=10 が編集。所属取得失敗 → fail-closed で拒否すべき。
	_, err := svc.UpdateRules(
		context.Background(),
		1,
		1,
		inputs,
		10,
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionRulesUpdate,
			"permission_group_rules",
		),
	)

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
			svc := newPermissionGroupServiceImpl(repo)
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
			setRulesFn: func(_ context.Context, _, _ uint64, _ []model.PermissionGroupRule) error {
				setRulesCalled = true
				return nil
			},
		}
		svc := newPermissionGroupServiceImpl(repo)

		inputs := []SetPermissionGroupRulesInput{
			{Resource: string(model.ResourceOwners)},
			{Resource: string(model.ResourceOwners)},
		}
		_, err := svc.UpdateRules(
			context.Background(),
			1,
			1,
			inputs,
			10,
			testPermissionMutationAudit(
				1,
				10,
				model.AuditActionPermissionRulesUpdate,
				"permission_group_rules",
			),
		)

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.False(t, setRulesCalled)
	})

	t.Run("rejects removing own master-permission edit", func(t *testing.T) {
		setRulesCalled := false
		repo := &mockPermissionGroupRepository{
			getGroupIDsByStaffIDFn: func(_ context.Context, clinicID, staffID uint64) ([]uint64, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(10), staffID)
				return []uint64{1}, nil
			},
			setRulesFn: func(_ context.Context, _, _ uint64, _ []model.PermissionGroupRule) error {
				setRulesCalled = true
				return nil
			},
		}
		svc := newPermissionGroupServiceImpl(repo)

		inputs := []SetPermissionGroupRulesInput{
			{Resource: string(model.ResourceOwners), CanView: true},
		}
		_, err := svc.UpdateRules(
			context.Background(),
			1,
			1,
			inputs,
			10,
			testPermissionMutationAudit(
				1,
				10,
				model.AuditActionPermissionRulesUpdate,
				"permission_group_rules",
			),
		)

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.False(t, setRulesCalled)
	})

	t.Run("allows self-referencing group when master-permission edit is retained", func(t *testing.T) {
		repo := &mockPermissionGroupRepository{
			getGroupIDsByStaffIDFn: func(_ context.Context, clinicID, staffID uint64) ([]uint64, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(10), staffID)
				return []uint64{1}, nil
			},
			setRulesFn: func(_ context.Context, clinicID, _ uint64, _ []model.PermissionGroupRule) error {
				assert.Equal(t, uint64(1), clinicID)
				return nil
			},
		}
		svc := newPermissionGroupServiceImpl(repo)

		inputs := []SetPermissionGroupRulesInput{
			{Resource: string(model.ResourceMasterPermission), CanEdit: true},
		}
		_, err := svc.UpdateRules(
			context.Background(),
			1,
			1,
			inputs,
			10,
			testPermissionMutationAudit(
				1,
				10,
				model.AuditActionPermissionRulesUpdate,
				"permission_group_rules",
			),
		)

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
