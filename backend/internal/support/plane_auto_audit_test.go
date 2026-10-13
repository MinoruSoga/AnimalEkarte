package support

// plane_auto_audit_test.go — 自動 Plane 起票の専用監査イベント（PR650-R1）。
//
// post-save の best-effort 起票は既存の audit_logs sink に専用アクション
// support_bug_report.plane_ticket_auto で記録する。手動起票は handler 側の
// support_bug_report.plane_ticket が担うため、service の手動経路からは
// このアクションを出さない（二重記録しない）。
// 記録内容は固定分類 outcome と安全なローカル/外部 ID のみ — 報告本文や
// upstream のエラー詳細は監査イベントにも出さない。監査書込み自体も
// best-effort であり、失敗は ERROR ログに留めて報告作成の成功を変えない
// （DB-atomic ではない点は既知の制限としてドキュメントする）。

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// mockAuditLogger はサポート層 AuditLogger の記録スタブ。
type mockAuditLogger struct {
	entries []AuditEntry
	err     error
}

func (m *mockAuditLogger) LogEntry(_ context.Context, entry AuditEntry) error {
	m.entries = append(m.entries, entry)
	return m.err
}

func planeAutoIssue() *PlaneIssue {
	return &PlaneIssue{ID: "issue-auto-1", URL: "https://app.plane.example/ws/browse/EMR-9/"}
}

// claim 成功時に provenance（clinic / reporter staff / report ID）と
// 固定 outcome=claimed + 外部 issue ID が 1 件だけ記録される。
func TestAutoPlaneAudit_LogsClaimedOutcomeAfterSuccessfulClaim(t *testing.T) {
	repo := &mockRepository{}
	savedReportMock(repo)
	repo.setTicketFn = func(context.Context, uint64, uint64, string, string) (bool, error) {
		return true, nil
	}
	tickets := &mockTicketCreator{result: planeAutoIssue()}
	auditLog := &mockAuditLogger{}
	svc := NewService(repo, tickets, nil, WithAuditLogger(auditLog))

	report, err := svc.Create(context.Background(), 1, 5, CreateBugReportInput{Title: "件名"})
	require.NoError(t, err)
	require.NotNil(t, report)

	require.Len(t, auditLog.entries, 1)
	entry := auditLog.entries[0]
	assert.Equal(t, planeAutoAuditAction, entry.Action)
	assert.Equal(t, model.AuditActorTypeStaff, entry.ActorType)
	require.NotNil(t, entry.ActorID)
	assert.Equal(t, uint64(5), *entry.ActorID)
	require.NotNil(t, entry.ClinicID)
	assert.Equal(t, uint64(1), *entry.ClinicID)
	assert.Equal(t, "support_bug_report", entry.Resource)
	require.NotNil(t, entry.ResourceID)
	assert.Equal(t, uint64(10), *entry.ResourceID)
	newValue, ok := entry.NewValue.(map[string]any)
	require.True(t, ok, "audit new_value must be a fixed-shape map")
	assert.Equal(t, "claimed", newValue["outcome"])
	assert.Equal(t, "issue-auto-1", newValue["plane_issue_id"])
}

// outcome 分類: 起票失敗（外部 ID なし）、claim 敗退・claim DB 失敗
// （孤立した外部 ID を残す）を区別できる固定値で記録する。
func TestAutoPlaneAudit_OutcomeClassification(t *testing.T) {
	tests := []struct {
		name        string
		tickets     *mockTicketCreator
		setTicketFn func(context.Context, uint64, uint64, string, string) (bool, error)
		wantOutcome string
		wantIssueID string // "" = plane_issue_id を記録しない
	}{
		{
			name:        "creation failure records creation_failed without external id",
			tickets:     &mockTicketCreator{err: &planeUpstreamError{status: 500}},
			wantOutcome: "creation_failed",
		},
		{
			name:    "lost claim keeps the orphan external id",
			tickets: &mockTicketCreator{result: planeAutoIssue()},
			setTicketFn: func(context.Context, uint64, uint64, string, string) (bool, error) {
				return false, nil
			},
			wantOutcome: "claim_lost",
			wantIssueID: "issue-auto-1",
		},
		{
			name:    "claim persistence error keeps the orphan external id",
			tickets: &mockTicketCreator{result: planeAutoIssue()},
			setTicketFn: func(context.Context, uint64, uint64, string, string) (bool, error) {
				return false, errors.New("db down")
			},
			wantOutcome: "claim_error",
			wantIssueID: "issue-auto-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			savedReportMock(repo)
			repo.setTicketFn = tt.setTicketFn
			repo.setCreationSyncErrFn = func(context.Context, uint64, uint64, string) (bool, error) {
				return true, nil
			}
			auditLog := &mockAuditLogger{}
			svc := NewService(repo, tt.tickets, nil, WithAuditLogger(auditLog))

			report, err := svc.Create(context.Background(), 1, 5, CreateBugReportInput{Title: "件名"})
			require.NoError(t, err)
			require.NotNil(t, report)

			require.Len(t, auditLog.entries, 1)
			newValue, ok := auditLog.entries[0].NewValue.(map[string]any)
			require.True(t, ok)
			assert.Equal(t, tt.wantOutcome, newValue["outcome"])
			if tt.wantIssueID == "" {
				assert.NotContains(t, newValue, "plane_issue_id")
			} else {
				assert.Equal(t, tt.wantIssueID, newValue["plane_issue_id"])
			}
		})
	}
}

// 連携無効（tickets=nil）では監査も何もしない。
func TestAutoPlaneAudit_NotEmittedWhenIntegrationDisabled(t *testing.T) {
	repo := &mockRepository{}
	savedReportMock(repo)
	auditLog := &mockAuditLogger{}
	svc := NewService(repo, nil, nil, WithAuditLogger(auditLog))

	report, err := svc.Create(context.Background(), 1, 5, CreateBugReportInput{Title: "件名"})
	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Empty(t, auditLog.entries)
}

// 監査ロガー未注入でも従来どおり報告作成は成功する（後方互換の optional seam）。
func TestAutoPlaneAudit_NotEmittedWithoutLogger(t *testing.T) {
	repo := &mockRepository{}
	savedReportMock(repo)
	repo.setTicketFn = func(context.Context, uint64, uint64, string, string) (bool, error) {
		return true, nil
	}
	svc := NewService(repo, &mockTicketCreator{result: planeAutoIssue()}, nil)

	report, err := svc.Create(context.Background(), 1, 5, CreateBugReportInput{Title: "件名"})
	require.NoError(t, err)
	require.NotNil(t, report)
	require.NotNil(t, report.PlaneIssueID)
}

// 手動起票（EnsurePlaneTicket）は専用アクションを出さない —
// 手動の監査は handler 側の既存経路が担う（二重記録しない）。
func TestAutoPlaneAudit_ManualEnsureUsesSeparateAuditPath(t *testing.T) {
	repo := &mockRepository{
		findByIDForClinicF: func(context.Context, uint64, uint64) (*model.SupportBugReport, error) {
			return &model.SupportBugReport{ID: 10, ClinicID: 1, ReporterStaffID: 5}, nil
		},
		setTicketFn: func(context.Context, uint64, uint64, string, string) (bool, error) {
			return true, nil
		},
	}
	auditLog := &mockAuditLogger{}
	svc := NewService(repo, &mockTicketCreator{result: planeAutoIssue()}, nil, WithAuditLogger(auditLog))

	_, err := svc.EnsurePlaneTicket(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Empty(t, auditLog.entries)
}

// 監査書込み失敗は ERROR ログで可視化し、報告作成の成功は変えない。
func TestAutoPlaneAudit_FailureIsVisibleButKeepsReportCreated(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	repo := &mockRepository{}
	savedReportMock(repo)
	repo.setTicketFn = func(context.Context, uint64, uint64, string, string) (bool, error) {
		return true, nil
	}
	auditLog := &mockAuditLogger{err: errors.New("audit sink down")}
	svc := NewService(repo, &mockTicketCreator{result: planeAutoIssue()}, nil, WithAuditLogger(auditLog))

	report, err := svc.Create(context.Background(), 1, 5, CreateBugReportInput{Title: "件名"})
	require.NoError(t, err)
	require.NotNil(t, report)
	require.Len(t, auditLog.entries, 1)
	assert.Contains(t, buf.String(), "level=ERROR")
}

// 監査イベントに報告のユーザー入力・upstream エラー詳細は転記しない。
func TestAutoPlaneAudit_NeverExportsReportFields(t *testing.T) {
	repo := &mockRepository{}
	savedReportMock(repo)
	repo.setTicketFn = func(context.Context, uint64, uint64, string, string) (bool, error) {
		return true, nil
	}
	auditLog := &mockAuditLogger{}
	svc := NewService(repo, &mockTicketCreator{result: planeAutoIssue()}, nil, WithAuditLogger(auditLog))

	_, err := svc.Create(context.Background(), 1, 5, CreateBugReportInput{
		Title:   "CANARY-患者氏名",
		Detail:  "CANARY-詳細本文",
		PageURL: "https://example.internal/?token=CANARY-token",
	})
	require.NoError(t, err)
	require.Len(t, auditLog.entries, 1)

	serialized, err := json.Marshal(auditLog.entries[0])
	require.NoError(t, err)
	assert.NotContains(t, string(serialized), "CANARY")
}
