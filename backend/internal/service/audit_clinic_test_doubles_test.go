package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// mockAuditRepository preserves the broader audit regression suite while the
// implementation under test is owned by internal/audit.
type mockAuditRepository struct {
	createFn    func(ctx context.Context, log *model.AuditLog) error
	lastLogged  *model.AuditLog
	entries     []*model.AuditLog
	createTxErr error
}

func (m *mockAuditRepository) recordLog(log *model.AuditLog) {
	m.lastLogged = log
	m.entries = append(m.entries, log)
}

func (m *mockAuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	m.recordLog(log)
	if m.createFn != nil {
		return m.createFn(ctx, log)
	}
	return nil
}

func (m *mockAuditRepository) CreateTx(ctx context.Context, log *model.AuditLog) error {
	if m.createTxErr != nil {
		return m.createTxErr
	}
	m.recordLog(log)
	if m.createFn != nil {
		return m.createFn(ctx, log)
	}
	return nil
}

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
