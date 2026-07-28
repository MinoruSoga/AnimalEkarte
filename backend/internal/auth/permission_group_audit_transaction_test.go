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

type permissionAuditTransactor struct {
	calls     int
	committed bool
}

type permissionAuditTxContextKey struct{}

func (t *permissionAuditTransactor) WithTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	t.calls++
	err := fn(context.WithValue(ctx, permissionAuditTxContextKey{}, true))
	if err == nil {
		t.committed = true
	}
	return err
}

type permissionAuditTxLogger struct {
	entries []AuthAuditEntry
	txSeen  []bool
	err     error
}

type permissionRepositoryWithoutMutationLocker struct {
	PermissionGroupRepository
}

func (l *permissionAuditTxLogger) LogEntryTx(
	ctx context.Context,
	entry AuthAuditEntry,
) error {
	l.entries = append(l.entries, entry)
	l.txSeen = append(
		l.txSeen,
		ctx.Value(permissionAuditTxContextKey{}) == true,
	)
	return l.err
}

func permissionAuditInput(action, resource string) PermissionMutationAudit {
	return PermissionMutationAudit{
		ClinicID:     23,
		ActorStaffID: 17,
		Action:       action,
		Resource:     resource,
		IPAddress:    "127.0.0.1",
		UserAgent:    "permission-audit-test",
	}
}

func TestPermissionGroupAuditedMutations_RollBackWhenAuditFails(t *testing.T) {
	auditFailure := errors.New("audit write failed")
	tests := []struct {
		name         string
		wantAction   string
		wantResource string
		wantOld      any
		wantNew      any
		invoke       func(PermissionGroupService) error
	}{
		{
			name:         "create",
			wantAction:   model.AuditActionPermissionGroupCreate,
			wantResource: "permission_group",
			wantNew: map[string]any{
				"name":        "created",
				"description": "",
				"color":       "#123456",
				"is_active":   false,
				"sort_order":  0,
				"rules":       []map[string]any{},
			},
			invoke: func(service PermissionGroupService) error {
				_, err := service.Create(
					context.Background(),
					23,
					&CreatePermissionGroupInput{Name: "created", Color: "#123456"},
					permissionAuditInput(
						model.AuditActionPermissionGroupCreate,
						"permission_group",
					),
				)
				return err
			},
		},
		{
			name:         "update",
			wantAction:   model.AuditActionPermissionGroupUpdate,
			wantResource: "permission_group",
			wantOld: map[string]any{
				"name":        "old",
				"description": "",
				"color":       "",
				"is_active":   false,
				"sort_order":  0,
				"rules": []map[string]any{{
					"resource":   "owners",
					"can_view":   true,
					"can_create": false,
					"can_edit":   false,
					"can_delete": false,
				}},
			},
			wantNew: map[string]any{
				"name":        "updated",
				"description": "",
				"color":       "",
				"is_active":   false,
				"sort_order":  0,
				"rules": []map[string]any{{
					"resource":   "owners",
					"can_view":   true,
					"can_create": false,
					"can_edit":   false,
					"can_delete": false,
				}},
			},
			invoke: func(service PermissionGroupService) error {
				name := "updated"
				_, err := service.Update(
					context.Background(),
					23,
					7,
					&UpdatePermissionGroupInput{Name: &name},
					permissionAuditInput(
						model.AuditActionPermissionGroupUpdate,
						"permission_group",
					),
				)
				return err
			},
		},
		{
			name:         "delete",
			wantAction:   model.AuditActionPermissionGroupDelete,
			wantResource: "permission_group",
			wantOld: map[string]any{
				"name":        "old",
				"description": "",
				"color":       "",
				"is_active":   false,
				"sort_order":  0,
				"rules": []map[string]any{{
					"resource":   "owners",
					"can_view":   true,
					"can_create": false,
					"can_edit":   false,
					"can_delete": false,
				}},
			},
			invoke: func(service PermissionGroupService) error {
				return service.Delete(
					context.Background(),
					23,
					7,
					permissionAuditInput(
						model.AuditActionPermissionGroupDelete,
						"permission_group",
					),
				)
			},
		},
		{
			name:         "rules",
			wantAction:   model.AuditActionPermissionRulesUpdate,
			wantResource: "permission_group_rules",
			wantOld: map[string]any{"rules": []map[string]any{{
				"resource":   "owners",
				"can_view":   true,
				"can_create": false,
				"can_edit":   false,
				"can_delete": false,
			}}},
			wantNew: map[string]any{"rules": []map[string]any{{
				"resource":   "reservations",
				"can_view":   false,
				"can_create": false,
				"can_edit":   true,
				"can_delete": false,
			}}},
			invoke: func(service PermissionGroupService) error {
				_, err := service.UpdateRules(
					context.Background(),
					23,
					7,
					[]SetPermissionGroupRulesInput{{
						Resource: "reservations",
						CanEdit:  true,
					}},
					17,
					permissionAuditInput(
						model.AuditActionPermissionRulesUpdate,
						"permission_group_rules",
					),
				)
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutationAttempted := false
			lockSeen := false
			repo := &mockPermissionGroupRepository{
				lockByIDFn: func(
					ctx context.Context,
					_ uint64,
					_ uint64,
				) (*model.PermissionGroup, error) {
					assert.Equal(
						t,
						true,
						ctx.Value(permissionAuditTxContextKey{}),
					)
					lockSeen = true
					return &model.PermissionGroup{
						ID:       7,
						ClinicID: 23,
						Name:     "old",
						Rules: []model.PermissionGroupRule{{
							Resource: "owners",
							CanView:  true,
						}},
					}, nil
				},
				findByIDFn: func(
					context.Context,
					uint64,
					uint64,
				) (*model.PermissionGroup, error) {
					rules := []model.PermissionGroupRule{{
						Resource: "owners",
						CanView:  true,
					}}
					if test.name == "rules" && mutationAttempted {
						rules = []model.PermissionGroupRule{{
							Resource: "reservations",
							CanEdit:  true,
						}}
					}
					return &model.PermissionGroup{
						ID:       7,
						ClinicID: 23,
						Name:     "old",
						Rules:    rules,
					}, nil
				},
				createFn: func(
					ctx context.Context,
					group *model.PermissionGroup,
				) error {
					assert.Equal(
						t,
						true,
						ctx.Value(permissionAuditTxContextKey{}),
					)
					assert.False(t, lockSeen)
					mutationAttempted = true
					group.ID = 7
					return nil
				},
				updateFieldsFn: func(
					ctx context.Context,
					_ uint64,
					_ uint64,
					_ map[string]any,
				) (*model.PermissionGroup, error) {
					assert.Equal(
						t,
						true,
						ctx.Value(permissionAuditTxContextKey{}),
					)
					assert.True(t, lockSeen)
					mutationAttempted = true
					return &model.PermissionGroup{
						ID:       7,
						ClinicID: 23,
						Name:     "updated",
						Rules: []model.PermissionGroupRule{{
							Resource: "owners",
							CanView:  true,
						}},
					}, nil
				},
				deleteFn: func(ctx context.Context, _ uint64, _ uint64) error {
					assert.Equal(
						t,
						true,
						ctx.Value(permissionAuditTxContextKey{}),
					)
					assert.True(t, lockSeen)
					mutationAttempted = true
					return nil
				},
				setRulesFn: func(
					ctx context.Context,
					_ uint64,
					_ uint64,
					_ []model.PermissionGroupRule,
				) error {
					assert.Equal(
						t,
						true,
						ctx.Value(permissionAuditTxContextKey{}),
					)
					assert.True(t, lockSeen)
					mutationAttempted = true
					return nil
				},
			}
			transactor := &permissionAuditTransactor{}
			audit := &permissionAuditTxLogger{err: auditFailure}
			service := NewPermissionGroupService(
				repo,
				transactor,
				audit,
			)

			err := test.invoke(service)

			require.Error(t, err)
			assert.ErrorIs(t, err, auditFailure)
			assert.True(t, mutationAttempted)
			assert.Equal(t, test.name != "create", lockSeen)
			assert.Equal(t, 1, transactor.calls)
			assert.False(t, transactor.committed)
			require.Len(t, audit.entries, 1)
			assert.Equal(t, uint64(23), *audit.entries[0].ClinicID)
			assert.Equal(t, uint64(17), *audit.entries[0].ActorID)
			assert.Equal(t, "127.0.0.1", audit.entries[0].IPAddress)
			assert.Equal(t, "permission-audit-test", audit.entries[0].UserAgent)
			assert.Equal(t, test.wantAction, audit.entries[0].Action)
			assert.Equal(t, test.wantResource, audit.entries[0].Resource)
			assert.Equal(t, test.wantOld, audit.entries[0].OldValue)
			assert.Equal(t, test.wantNew, audit.entries[0].NewValue)
			assert.Equal(t, []bool{true}, audit.txSeen)
		})
	}
}

func TestPermissionGroupAuditedMutations_FailClosedWithoutAuditLogger(
	t *testing.T,
) {
	tests := []struct {
		name       string
		transactor Transactor
		audit      PermissionAuditTxLogger
		input      PermissionMutationAudit
	}{
		{
			name:       "audit logger missing",
			transactor: &permissionAuditTransactor{},
			input: permissionAuditInput(
				model.AuditActionPermissionGroupCreate,
				"permission_group",
			),
		},
		{
			name:  "transactor missing",
			audit: &permissionAuditTxLogger{},
			input: permissionAuditInput(
				model.AuditActionPermissionGroupCreate,
				"permission_group",
			),
		},
		{
			name:       "actor missing",
			transactor: &permissionAuditTransactor{},
			audit:      &permissionAuditTxLogger{},
			input: PermissionMutationAudit{
				ClinicID: 23,
				Action:   model.AuditActionPermissionGroupCreate,
				Resource: "permission_group",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &mockPermissionGroupRepository{
				createFn: func(context.Context, *model.PermissionGroup) error {
					t.Fatal("repository mutation must not run without audit preconditions")
					return nil
				},
			}
			service := NewPermissionGroupService(
				repo,
				test.transactor,
				test.audit,
			)

			_, err := service.Create(
				context.Background(),
				23,
				&CreatePermissionGroupInput{
					Name:  "created",
					Color: "#123456",
				},
				test.input,
			)

			require.Error(t, err)
			if transactor, ok := test.transactor.(*permissionAuditTransactor); ok {
				assert.Zero(t, transactor.calls)
				assert.False(t, transactor.committed)
			}
		})
	}
}

func TestPermissionGroupAuditedMutation_RejectsMismatchedTypedAuditInput(
	t *testing.T,
) {
	repo := &mockPermissionGroupRepository{}
	transactor := &permissionAuditTransactor{}
	service := NewPermissionGroupService(
		repo,
		transactor,
		&permissionAuditTxLogger{},
	)
	input := permissionAuditInput(
		model.AuditActionPermissionGroupDelete,
		"permission_group",
	)
	input.ClinicID = 99

	_, err := service.Create(
		context.Background(),
		23,
		&CreatePermissionGroupInput{Name: "created", Color: "#123456"},
		input,
	)

	require.Error(t, err)
	assert.Zero(t, transactor.calls)
}

func TestPermissionGroupAuditedMutation_FailsClosedWithoutRowLocker(
	t *testing.T,
) {
	repo := permissionRepositoryWithoutMutationLocker{
		PermissionGroupRepository: &mockPermissionGroupRepository{},
	}
	transactor := &permissionAuditTransactor{}
	audit := &permissionAuditTxLogger{}
	service := NewPermissionGroupService(repo, transactor, audit)
	name := "updated"

	result, err := service.Update(
		context.Background(),
		23,
		7,
		&UpdatePermissionGroupInput{Name: &name},
		permissionAuditInput(
			model.AuditActionPermissionGroupUpdate,
			"permission_group",
		),
	)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 1, transactor.calls)
	assert.False(t, transactor.committed)
	assert.Empty(t, audit.entries)
}

func TestPermissionGroupRepository_LockByIDForUpdateRequiresAmbientTransaction(
	t *testing.T,
) {
	repo := NewPermissionGroupRepository(nil)
	locker, ok := repo.(PermissionGroupMutationLocker)
	require.True(t, ok)

	group, err := locker.LockByIDForUpdate(context.Background(), 23, 7)

	require.Error(t, err)
	assert.Nil(t, group)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
}
