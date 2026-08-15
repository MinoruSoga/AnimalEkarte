package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func uniquePermissionGroupNameErr() error {
	return apperrors.FromGORM(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: apperrors.ConstraintPermissionGroupName,
	}, "permission_group", "")
}

func uniquePermissionGroupRulesErr() error {
	return apperrors.FromGORM(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: apperrors.ConstraintPermissionGroupRules,
	}, "permission_group_rule", "")
}

func TestPermissionGroupService_Create_NameConflictMapsDomainCode(t *testing.T) {
	repo := &mockPermissionGroupRepository{
		createFn: func(_ context.Context, _ *model.PermissionGroup) error {
			return uniquePermissionGroupNameErr()
		},
	}
	svc := newPermissionGroupServiceImpl(repo)

	group, err := svc.Create(
		context.Background(),
		1,
		&CreatePermissionGroupInput{Name: "執行"},
		testPermissionMutationAudit(
			1, 10, model.AuditActionPermissionGroupCreate, "permission_group",
		),
	)

	require.Error(t, err)
	assert.Nil(t, group)
	assert.True(t, apperrors.IsNameConflict(err, apperrors.CodePermissionGroupNameConflict))
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "執行", appErr.Params["name"])
	assert.NotContains(t, appErr.Message, "permission_group '' already exists")
	assert.NotContains(t, err.Error(), "uk_permission_groups")
}

func TestPermissionGroupService_Update_RenameNameConflictMapsDomainCode(t *testing.T) {
	existing := &model.PermissionGroup{ID: 1, ClinicID: 1, Name: "旧名"}
	repo := &mockPermissionGroupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.PermissionGroup, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.PermissionGroup, error) {
			return nil, uniquePermissionGroupNameErr()
		},
	}
	svc := newPermissionGroupServiceImpl(repo)

	result, err := svc.Update(
		context.Background(),
		1,
		1,
		&UpdatePermissionGroupInput{Name: strPtr("執行")},
		testPermissionMutationAudit(
			1, 10, model.AuditActionPermissionGroupUpdate, "permission_group",
		),
	)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsNameConflict(err, apperrors.CodePermissionGroupNameConflict))
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "執行", appErr.Params["name"])
}

func TestPermissionGroupService_CreateWithRules_RulesUniqueNotNameConflict(t *testing.T) {
	base := &mockPermissionGroupRepository{}
	repo := &atomicPermissionGroupRepositoryStub{
		mockPermissionGroupRepository: base,
		createWithRulesFn: func(
			_ context.Context,
			_ *model.PermissionGroup,
			_ []model.PermissionGroupRule,
		) (*model.PermissionGroup, error) {
			return nil, uniquePermissionGroupRulesErr()
		},
	}
	svc := newPermissionGroupServiceImpl(repo)

	group, err := svc.Create(
		context.Background(),
		1,
		&CreatePermissionGroupInput{
			Name: "執行",
			Rules: []SetPermissionGroupRulesInput{{
				Resource: string(model.ResourceOwners),
				CanView:  true,
			}},
		},
		testPermissionMutationAudit(
			1, 10, model.AuditActionPermissionGroupCreate, "permission_group",
		),
	)

	require.Error(t, err)
	assert.Nil(t, group)
	assert.False(t, apperrors.IsNameConflict(err, apperrors.CodePermissionGroupNameConflict),
		"permission_group_rules 23505 must not map to name conflict")
	assert.True(t, apperrors.IsAlreadyExists(err))
}
