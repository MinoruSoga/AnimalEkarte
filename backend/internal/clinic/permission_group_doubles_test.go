package clinic

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

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

type clinicServiceTransactorDouble struct {
	withTxErr error
	withTxFn  func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *clinicServiceTransactorDouble) WithTx(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	if m.withTxErr != nil {
		return m.withTxErr
	}
	return fn(ctx)
}

func clinicStringPtr(value string) *string {
	return &value
}

func clinicFloat64Ptr(value float64) *float64 {
	return &value
}
