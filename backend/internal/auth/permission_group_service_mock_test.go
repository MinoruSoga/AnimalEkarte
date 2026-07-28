package auth

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

type directPermissionTransactor struct{}

func (directPermissionTransactor) WithTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	return fn(ctx)
}

type noopPermissionAuditTxLogger struct{}

func (noopPermissionAuditTxLogger) LogEntryTx(
	context.Context,
	AuthAuditEntry,
) error {
	return nil
}

type mockPermissionGroupRepository struct {
	findAllFn      func(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
	lockByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
	createFn       func(ctx context.Context, group *model.PermissionGroup) error
	updateFieldsFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PermissionGroup, error)
	deleteFn       func(ctx context.Context, clinicID, id uint64) error
	setRulesFn     func(ctx context.Context, clinicID, groupID uint64, rules []model.PermissionGroupRule) error

	countStaffsByGroupIDFn func(ctx context.Context, clinicID, groupID uint64) (int64, error)
	countUsageByGroupIDFn  func(ctx context.Context, clinicID, groupID uint64) (int64, error)

	reorderErr error
	reorderFn  func(ctx context.Context, clinicID uint64, ids []uint64) error

	getEffectivePermissionsByStaffID func(
		ctx context.Context,
		staffID, clinicID uint64,
	) ([]model.PermissionGroupRule, error)
	getEffectivePermissionsFn func(
		ctx context.Context,
		staffID, clinicID uint64,
	) ([]model.PermissionGroupRule, error)

	getGroupIDsByStaffIDFn     func(ctx context.Context, clinicID, staffID uint64) ([]uint64, error)
	findAllGroupIDsByStaffIDFn func(ctx context.Context, clinicID, staffID uint64) ([]uint64, error)

	setStaffGroupsFn    func(ctx context.Context, staffID uint64, groupIDs []uint64) error
	updateStaffGroupsFn func(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error
}

func (m *mockPermissionGroupRepository) FindAll(
	ctx context.Context,
	clinicID uint64,
) ([]model.PermissionGroup, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockPermissionGroupRepository) FindByID(
	ctx context.Context,
	clinicID, id uint64,
) (*model.PermissionGroup, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.PermissionGroup{
		ID:       id,
		ClinicID: clinicID,
		Name:     "permission group",
	}, nil
}

func (m *mockPermissionGroupRepository) LockByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.PermissionGroup, error) {
	if m.lockByIDFn != nil {
		return m.lockByIDFn(ctx, clinicID, id)
	}
	return m.FindByID(ctx, clinicID, id)
}

func (m *mockPermissionGroupRepository) Create(
	ctx context.Context,
	group *model.PermissionGroup,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, group)
	}
	return nil
}

func (m *mockPermissionGroupRepository) Update(
	ctx context.Context,
	clinicID, id uint64,
	fields map[string]any,
) (*model.PermissionGroup, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.PermissionGroup{}, nil
}

func (m *mockPermissionGroupRepository) Delete(
	ctx context.Context,
	clinicID, id uint64,
) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockPermissionGroupRepository) DeleteSoftDeletedByClinicID(
	_ context.Context,
	_ uint64,
) error {
	return nil
}

func (m *mockPermissionGroupRepository) UpdateRules(
	ctx context.Context,
	clinicID, groupID uint64,
	rules []model.PermissionGroupRule,
) error {
	if m.setRulesFn != nil {
		return m.setRulesFn(ctx, clinicID, groupID, rules)
	}
	return nil
}

func (m *mockPermissionGroupRepository) CountUsageByGroupID(
	ctx context.Context,
	clinicID, groupID uint64,
) (int64, error) {
	if m.countStaffsByGroupIDFn != nil {
		return m.countStaffsByGroupIDFn(ctx, clinicID, groupID)
	}
	if m.countUsageByGroupIDFn != nil {
		return m.countUsageByGroupIDFn(ctx, clinicID, groupID)
	}
	return 0, nil
}

func (m *mockPermissionGroupRepository) Reorder(
	ctx context.Context,
	clinicID uint64,
	ids []uint64,
) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return m.reorderErr
}

func (m *mockPermissionGroupRepository) FindAllEffectivePermissionsByStaffID(
	ctx context.Context,
	staffID, clinicID uint64,
) ([]model.PermissionGroupRule, error) {
	if m.getEffectivePermissionsByStaffID != nil {
		return m.getEffectivePermissionsByStaffID(ctx, staffID, clinicID)
	}
	if m.getEffectivePermissionsFn != nil {
		return m.getEffectivePermissionsFn(ctx, staffID, clinicID)
	}
	return nil, nil
}

func (m *mockPermissionGroupRepository) FindAllGroupIDsByStaffID(
	ctx context.Context,
	clinicID, staffID uint64,
) ([]uint64, error) {
	if m.findAllGroupIDsByStaffIDFn != nil {
		return m.findAllGroupIDsByStaffIDFn(ctx, clinicID, staffID)
	}
	if m.getGroupIDsByStaffIDFn != nil {
		return m.getGroupIDsByStaffIDFn(ctx, clinicID, staffID)
	}
	return nil, nil
}

func (m *mockPermissionGroupRepository) UpdateStaffGroups(
	ctx context.Context,
	clinicID, staffID uint64,
	groupIDs []uint64,
) error {
	if m.updateStaffGroupsFn != nil {
		return m.updateStaffGroupsFn(ctx, clinicID, staffID, groupIDs)
	}
	if m.setStaffGroupsFn != nil {
		return m.setStaffGroupsFn(ctx, staffID, groupIDs)
	}
	return nil
}

func newPermissionGroupServiceImpl(repo PermissionGroupRepository) PermissionGroupApplication {
	return NewPermissionGroupService(
		repo,
		directPermissionTransactor{},
		noopPermissionAuditTxLogger{},
	)
}

func testPermissionMutationAudit(
	clinicID, actorStaffID uint64,
	action, resource string,
) PermissionMutationAudit {
	return PermissionMutationAudit{
		ClinicID:     clinicID,
		ActorStaffID: actorStaffID,
		Action:       action,
		Resource:     resource,
	}
}

func strPtr(value string) *string {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}
