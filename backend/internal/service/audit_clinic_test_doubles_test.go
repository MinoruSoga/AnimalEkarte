package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// mockPermissionGroupRepository is the minimal clinic-side writer double used
// by the retained clinic transaction tests.
type mockPermissionGroupRepository struct {
	createFn   func(ctx context.Context, group *model.PermissionGroup) error
	setRulesFn func(ctx context.Context, clinicID, groupID uint64, rules []model.PermissionGroupRule) error
}

func (m *mockPermissionGroupRepository) Create(ctx context.Context, group *model.PermissionGroup) error {
	if m.createFn != nil {
		return m.createFn(ctx, group)
	}
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
