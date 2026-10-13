package support

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Create: 保存成功後に Plane へ best-effort 自動起票 ----
//
// 報告の保存が主契約で、Plane 起票は後続の best-effort 副作用（EMR-340 で自動起票を復帰）。
// Plane 上流障害・DB 記録失敗でも作成自体は失敗させず、連携無効（creator nil）なら
// 何も呼ばず plane_sync_error も書かない。scheduled sync（list/backfill/cron）は
// 作成経路に関与しない。claim 前に Plane 側 issue は作成済みのため、claim 敗退時に
// 上流 orphan が残りうる既知の制限は下記ケースで contract として固定する。

func TestServiceCreate_SyncsPlaneTicketAfterSave(t *testing.T) {
	issue := &PlaneIssue{ID: "issue-uuid", URL: "https://app.plane.so/ws/browse/EMR-1/"}

	tests := []struct {
		name         string
		tickets      *mockTicketCreator
		createErr    error
		setTicketFn  func(ctx context.Context, clinicID, id uint64, issueID, issueURL string) (bool, error)
		setSyncErrFn func(ctx context.Context, clinicID, id uint64, syncErr string) error
		wantCalls    int
		wantErr      bool
		wantIssueID  bool
		wantIssueURL string
		wantSyncErr  string // SetPlaneSyncError へ渡される sanitize 済み文言（空 = 呼ばれない）
	}{
		{
			name:    "save success creates and claims plane ticket",
			tickets: &mockTicketCreator{result: issue},
			setTicketFn: func(_ context.Context, clinicID, id uint64, issueID, issueURL string) (bool, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(10), id)
				assert.Equal(t, issue.ID, issueID)
				assert.Equal(t, issue.URL, issueURL)
				return true, nil
			},
			wantCalls:    1,
			wantIssueID:  true,
			wantIssueURL: issue.URL,
		},
		{
			name:      "save failure does not call plane",
			createErr: errors.New("db write failed"),
			tickets:   &mockTicketCreator{result: issue},
			wantErr:   true,
		},
		{
			name:        "upstream failure still returns saved report with sanitized sync error",
			tickets:     &mockTicketCreator{err: &planeUpstreamError{status: 500}},
			wantCalls:   1,
			wantSyncErr: "plane api error (status 500)",
		},
		{
			name:        "non-upstream failure records generic reason without internals",
			tickets:     &mockTicketCreator{err: errors.New("connection refused")},
			wantCalls:   1,
			wantSyncErr: "plane request failed",
		},
		{
			name:    "nil creator stores report only without sync error",
			tickets: nil,
		},
		{
			name:    "claim loss keeps saved report unlinked",
			tickets: &mockTicketCreator{result: issue},
			setTicketFn: func(context.Context, uint64, uint64, string, string) (bool, error) {
				return false, nil
			},
			wantCalls: 1,
		},
		{
			name:    "ticket persistence error keeps saved report (upstream orphan is a known limit)",
			tickets: &mockTicketCreator{result: issue},
			setTicketFn: func(context.Context, uint64, uint64, string, string) (bool, error) {
				return false, errors.New("db write failed")
			},
			wantCalls: 1,
		},
		{
			name:    "sync error persistence failure stays best-effort",
			tickets: &mockTicketCreator{err: &planeUpstreamError{status: 502}},
			setSyncErrFn: func(context.Context, uint64, uint64, string) error {
				return errors.New("db write failed")
			},
			wantCalls:   1,
			wantSyncErr: "plane api error (status 502)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			if tt.createErr != nil {
				repo.createFn = func(context.Context, *model.SupportBugReport) error {
					return tt.createErr
				}
			} else {
				savedReportMock(repo)
			}
			repo.setTicketFn = tt.setTicketFn
			repo.setSyncErrFn = tt.setSyncErrFn
			if tt.wantSyncErr != "" && repo.setSyncErrFn == nil {
				repo.setSyncErrFn = func(_ context.Context, clinicID, id uint64, syncErr string) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), id)
					assert.Equal(t, tt.wantSyncErr, syncErr)
					return nil
				}
			}

			var gotReport *model.SupportBugReport
			if tt.tickets != nil {
				tt.tickets.fn = func(_ context.Context, r *model.SupportBugReport) {
					gotReport = r
				}
			}
			var tickets TicketCreator
			if tt.tickets != nil {
				tickets = tt.tickets
			}
			svc := NewService(repo, tickets, nil)

			report, err := svc.Create(context.Background(), 1, 5, CreateBugReportInput{Title: "受付でエラー"})
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, report)
			} else {
				require.NoError(t, err)
				require.NotNil(t, report)
				assert.Equal(t, uint64(10), report.ID)
			}
			if tt.tickets != nil {
				assert.Equal(t, tt.wantCalls, tt.tickets.calls)
				if tt.wantCalls > 0 {
					// repo.Create で採番された保存済み report がそのまま渡る。
					require.NotNil(t, gotReport)
					assert.Equal(t, uint64(10), gotReport.ID)
					assert.Equal(t, uint64(1), gotReport.ClinicID)
					assert.Equal(t, uint64(5), gotReport.ReporterStaffID)
				}
			}
			if report != nil {
				assert.Equal(t, tt.wantIssueID, report.PlaneIssueID != nil)
				if tt.wantIssueURL != "" {
					require.NotNil(t, report.PlaneIssueURL)
					assert.Equal(t, tt.wantIssueURL, *report.PlaneIssueURL)
				}
				assert.Equal(t, tt.wantSyncErr != "", report.PlaneSyncError != nil)
				if tt.wantSyncErr != "" {
					require.NotNil(t, report.PlaneSyncError)
					assert.Equal(t, tt.wantSyncErr, *report.PlaneSyncError)
				}
			}
		})
	}
}

func TestServiceCreate_PlaneRunsAfterDiscordAndSave(t *testing.T) {
	// 順序契約: Discord 通知（保存前 best-effort）→ DB 保存 → Plane 起票（保存後 best-effort）。
	var order []string
	repo := &mockRepository{
		setTicketFn: func(context.Context, uint64, uint64, string, string) (bool, error) {
			return true, nil
		},
	}
	repo.createFn = func(_ context.Context, report *model.SupportBugReport) error {
		order = append(order, "create")
		report.ID = 10
		return nil
	}
	notifier := &mockBugReportNotifier{order: &order}
	tickets := &mockTicketCreator{
		fn:     func(context.Context, *model.SupportBugReport) { order = append(order, "plane") },
		result: &PlaneIssue{ID: "i", URL: "u"},
	}
	svc := NewService(repo, tickets, notifier)

	report, err := svc.Create(context.Background(), 1, 5, CreateBugReportInput{Title: "件名"})
	require.NoError(t, err)
	assert.Equal(t, uint64(10), report.ID)
	assert.Equal(t, []string{"notify", "create", "plane"}, order)
}
