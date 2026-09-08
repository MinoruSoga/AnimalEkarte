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

func TestPermissionGroupService_Update_RejectsSelfLockoutWhenDeactivatingLastAdminGroup(t *testing.T) {
	existing := &model.PermissionGroup{ID: 1, ClinicID: 1, Name: "管理者", IsActive: true}
	falseVal := false
	repo := &mockPermissionGroupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.PermissionGroup, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdatePermissionGroupInput) (*model.PermissionGroup, error) {
			return existing, nil
		},
		getEffectivePermissionsByStaffID: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
			return []model.PermissionGroupRule{{
				Resource: string(model.ResourceMasterPermission),
				CanView:  true,
			}}, nil
		},
	}
	svc := newPermissionGroupServiceImpl(repo)
	_, err := svc.Update(
		context.Background(),
		1,
		1,
		&UpdatePermissionGroupInput{IsActive: &falseVal},
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionGroupUpdate,
			"permission_group",
		),
	)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))
}

func TestPermissionGroupService_Update_AllowsSystemAdminSelfLockoutPath(t *testing.T) {
	existing := &model.PermissionGroup{ID: 1, ClinicID: 1, Name: "管理者", IsActive: true}
	falseVal := false
	repo := &mockPermissionGroupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.PermissionGroup, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdatePermissionGroupInput) (*model.PermissionGroup, error) {
			return existing, nil
		},
		getEffectivePermissionsByStaffID: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
			return nil, nil
		},
	}
	svc := newPermissionGroupServiceImpl(repo)
	audit := testPermissionMutationAudit(
		1,
		10,
		model.AuditActionPermissionGroupUpdate,
		"permission_group",
	)
	audit.ActorIsSystemAdmin = true
	result, err := svc.Update(
		context.Background(),
		1,
		1,
		&UpdatePermissionGroupInput{IsActive: &falseVal},
		audit,
	)
	require.NoError(t, err)
	assert.True(t, result != nil)
}

func TestPermissionGroupService_Update_LookupFailureFailsClosed(t *testing.T) {
	existing := &model.PermissionGroup{ID: 1, ClinicID: 1, Name: "管理者", IsActive: true}
	falseVal := false
	repo := &mockPermissionGroupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.PermissionGroup, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdatePermissionGroupInput) (*model.PermissionGroup, error) {
			return existing, nil
		},
		getEffectivePermissionsByStaffID: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
			return nil, errors.New("lookup failed")
		},
	}
	svc := newPermissionGroupServiceImpl(repo)
	_, err := svc.Update(
		context.Background(),
		1,
		1,
		&UpdatePermissionGroupInput{IsActive: &falseVal},
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionGroupUpdate,
			"permission_group",
		),
	)
	require.Error(t, err)
	assert.False(t, errors.Is(err, apperrors.ErrForbidden))
}

func TestPermissionGroupService_Update_AllowsWhenAnotherGroupKeepsViewAndEdit(t *testing.T) {
	existing := &model.PermissionGroup{ID: 1, ClinicID: 1, Name: "管理者", IsActive: true}
	falseVal := false
	repo := &mockPermissionGroupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.PermissionGroup, error) {
			return existing, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdatePermissionGroupInput) (*model.PermissionGroup, error) {
			return existing, nil
		},
		getEffectivePermissionsByStaffID: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
			return []model.PermissionGroupRule{{
				Resource: string(model.ResourceMasterPermission),
				CanView:  true,
				CanEdit:  true,
			}}, nil
		},
	}
	svc := newPermissionGroupServiceImpl(repo)
	result, err := svc.Update(
		context.Background(),
		1,
		1,
		&UpdatePermissionGroupInput{IsActive: &falseVal},
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionGroupUpdate,
			"permission_group",
		),
	)
	require.NoError(t, err)
	assert.NotNil(t, result)
}
