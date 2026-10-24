package support

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---- CreateBugReport HTTP 201 × Plane 自動起票の停止 ----
//
// 配送契約（送信前確認 intent・宛先ごとの試行 claim・結果記録）の接続まで、保存後の Plane 自動起票は
// 外部へ送らない（EMR-549 C1）。連携の有無にかかわらず POST は 201 を返し、応答は Plane の
// 送信有無を推測せず plane_delivery_state=unavailable を返す。旧 plane_issue_url /
// plane_sync_error は列ごと削除済みで、応答にも含めない。外部通信は発生させない。

func TestCreateBugReport_PlaneAutoSyncIsStopped(t *testing.T) {
	configured := &mockTicketCreator{result: &PlaneIssue{ID: "issue-uuid", URL: "https://app.plane.so/ws/browse/EMR-1/"}}

	tests := []struct {
		name    string
		tickets TicketCreator
	}{
		{name: "201 without calling plane when integration is configured", tickets: configured},
		{name: "201 when integration is not configured", tickets: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			savedReportMock(repo)
			svc := NewService(repo, tt.tickets, nil)
			h := NewHandler(svc, nil, nil, nil, nil)

			body, contentType := buildMultipartBody(t, map[string]string{"title": "受付でエラー"}, "", "", nil)
			c, rec := newRequest(t, http.MethodPost, "/api/v1/support/bug-reports", body, contentType)
			c.Request.ContentLength = int64(body.Len())
			h.CreateBugReport(c)

			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"plane_delivery_state":"unavailable"`)
			assert.NotContains(t, rec.Body.String(), "plane_issue_url")
			assert.NotContains(t, rec.Body.String(), "plane_sync_error")
		})
	}
	assert.Equal(t, 0, configured.calls, "plane must not be called before the delivery contract is connected")
}
