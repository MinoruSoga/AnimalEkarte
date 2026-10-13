package support

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---- CreateBugReport HTTP 201 × Plane 自動起票の結合 ----
//
// 実 NewService + fn-field repo/creator モック + 実ハンドラで、Plane 起票の成否や
// 連携有無にかかわらず POST が 201 を返し、同期結果がレスポンスへ反映されることを
// 固定する（EMR-340）。外部通信は発生させない（real planeClient / notifier は未使用）。

func TestCreateBugReport_PlaneAutoSyncReturns201(t *testing.T) {
	issue := &PlaneIssue{ID: "issue-uuid", URL: "https://app.plane.so/ws/browse/EMR-1/"}

	tests := []struct {
		name         string
		tickets      TicketCreator
		wantStatus   int
		wantContains string
		wantAbsent   string
	}{
		{
			name:         "201 with plane_issue_url on auto ticket success",
			tickets:      &mockTicketCreator{result: issue},
			wantStatus:   http.StatusCreated,
			wantContains: `"plane_issue_url":"https://app.plane.so/ws/browse/EMR-1/"`,
		},
		{
			name:         "201 with sanitized plane_sync_error on upstream failure",
			tickets:      &mockTicketCreator{err: &planeUpstreamError{status: 500}},
			wantStatus:   http.StatusCreated,
			wantContains: `"plane_sync_error":"plane api error (status 500)"`,
		},
		{
			name:       "201 without plane fields when integration is not configured",
			tickets:    nil,
			wantStatus: http.StatusCreated,
			wantAbsent: "plane_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			savedReportMock(repo)
			repo.setTicketFn = func(context.Context, uint64, uint64, string, string) (bool, error) {
				return true, nil
			}
			repo.setSyncErrFn = func(context.Context, uint64, uint64, string) error {
				return nil
			}
			svc := NewService(repo, tt.tickets, nil)
			h := NewHandler(svc, nil, nil, nil, nil)

			body, contentType := buildMultipartBody(t, map[string]string{"title": "受付でエラー"}, "", "", nil)
			c, rec := newRequest(t, http.MethodPost, "/api/v1/support/bug-reports", body, contentType)
			c.Request.ContentLength = int64(body.Len())
			h.CreateBugReport(c)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantContains != "" {
				assert.Contains(t, rec.Body.String(), tt.wantContains)
			}
			if tt.wantAbsent != "" {
				assert.NotContains(t, rec.Body.String(), tt.wantAbsent)
			}
		})
	}
}
