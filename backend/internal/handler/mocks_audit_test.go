package handler

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// mockAuditService は service.AuditService のテスト用モック実装（F-4統合正本）。
// auth_session_test.go の mockAuditService（loggedActions/logAuthLoginFn）と
// manual_article_handler_test.go の mockAuditServiceForManual（lastLogEntry/logEntryErr）、
// permission_group_handler_test.go の mockAuditServiceForPG（lastLogEntry のみ実使用、
// logFn/logAuthLoginFn は全構築箇所で未設定の死んだフィールドだったため統合時に落とした）の
// 実際に検証されている挙動を統合する。
type mockAuditService struct {
	logAuthLoginFn func(ctx context.Context, clinicID, staffID *uint64, action, ip, ua string) error
	loggedActions  []string
	lastLogEntry   *service.AuditLogInput // #122: audit 内容検証用
	logEntryErr    error
}

func (m *mockAuditService) Log(_ context.Context, _ *model.AuditLog) error { return nil }

func (m *mockAuditService) LogEntry(_ context.Context, input *service.AuditLogInput) error {
	m.lastLogEntry = input
	return m.logEntryErr
}

func (m *mockAuditService) LogAuthLogin(ctx context.Context, clinicID, staffID *uint64, action, ip, ua string) error {
	m.loggedActions = append(m.loggedActions, action)
	if m.logAuthLoginFn != nil {
		return m.logAuthLoginFn(ctx, clinicID, staffID, action, ip, ua)
	}
	return nil
}

func (m *mockAuditService) LogLstepOperation(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64) error {
	return nil
}

func (m *mockAuditService) LogLstepOperationWithMetadata(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64, _ any) error {
	return nil
}

func (m *mockAuditService) LogMedicalRecordChange(_ context.Context, _ uint64, _ *uint64, _ string, _ uint64, _, _ map[string]any) error {
	return nil
}

func (m *mockAuditService) LogVitalChange(_ context.Context, _ uint64, _ *uint64, _ string, _, _ uint64, _, _ map[string]any) error {
	return nil
}

func (m *mockAuditService) LogAddendumCreate(_ context.Context, _ uint64, _ *uint64, _, _ uint64, _ *model.MedicalRecordAddendum) error {
	return nil
}

func (m *mockAuditService) LogClinicSwitch(_ context.Context, _ *uint64, _, _ uint64, _, _ string) error {
	return nil
}
