package audit

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
