package support

// plane_safe_export_test.go — Plane 送信ペイロードの安全境界テスト（PR650-01）。
//
// バグ報告の任意入力（件名・本文・ページURL・経路・UA・表示領域・バージョン・
// スクショキー）は患者・飼主・スタッフの個人情報やトークンを含みうるため、
// 外部（Plane）へは転記しない。送信境界は redaction ではなく allowlist:
// 送るのは決定論的な固定件名・固定説明文・サーバ採番の報告 ID のみ。
// 自動起票（Create 後の best-effort）と手動再送（EnsurePlaneTicket）の両経路が
// 同じクライアント境界を通ることも併せて固定する。
//
// canary にはフィールドごとに異なる機微文字列を使い、生リクエストボディに
// 混入した時点でそのフィールドの漏洩として検出する。

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

var planeExportCanaries = map[string]string{
	"Title":         "CANARY-TITLE-患者氏名タロウ",
	"Detail":        "CANARY-DETAIL-飼主ハナコ連絡先",
	"PageURL":       "CANARY-PAGEURL-token-abc123",
	"RoutePath":     "CANARY-ROUTEPATH",
	"UserAgent":     "CANARY-USERAGENT",
	"Viewport":      "CANARY-VIEWPORT",
	"AppVersion":    "CANARY-APPVERSION",
	"ScreenshotKey": "CANARY-SCREENSHOT-KEY",
}

// canaryReport は全ユーザー入力フィールドに canary を埋めた報告を返す。
// 数字の ID（報告 ID）はサーバ採番の識別子として送信が許可される唯一のメタデータ。
func canaryReport() *model.SupportBugReport {
	screenshotKey := planeExportCanaries["ScreenshotKey"]
	return &model.SupportBugReport{
		ID:              4242,
		ClinicID:        7,
		ReporterStaffID: 9,
		Title:           planeExportCanaries["Title"],
		Detail:          planeExportCanaries["Detail"],
		PageURL:         "https://example.internal/?token=" + planeExportCanaries["PageURL"],
		RoutePath:       "/reception/" + planeExportCanaries["RoutePath"],
		UserAgent:       planeExportCanaries["UserAgent"],
		Viewport:        planeExportCanaries["Viewport"],
		AppVersion:      planeExportCanaries["AppVersion"],
		ScreenshotKey:   &screenshotKey,
		CreatedAt:       time.Date(2026, 10, 13, 9, 0, 0, 0, time.UTC),
	}
}

func canaryCreateInput() CreateBugReportInput {
	report := canaryReport()
	return CreateBugReportInput{
		Title:         report.Title,
		Detail:        report.Detail,
		PageURL:       report.PageURL,
		RoutePath:     report.RoutePath,
		UserAgent:     report.UserAgent,
		Viewport:      report.Viewport,
		AppVersion:    report.AppVersion,
		ScreenshotKey: report.ScreenshotKey,
	}
}

type planeExportCapture struct {
	rawBody []byte
	calls   int
}

// newPlaneExportServer は送信された生 JSON ボディを記録する Plane スタブ。
func newPlaneExportServer(t *testing.T, captured *planeExportCapture) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.calls++
		captured.rawBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"issue-export-1","identifier":"EMR-77","sequence_id":77}`))
	}))
	t.Cleanup(server.Close)
	return server
}

func newExportPlaneClient(server *httptest.Server) *planeClient {
	return &planeClient{
		baseURL:           server.URL,
		webBaseURL:        "https://app.plane.example",
		workspaceSlug:     "ws",
		projectID:         "proj-1",
		projectIdentifier: "EMR",
		apiKey:            "test-key",
		httpClient:        server.Client(),
	}
}

// assertSafePlanePayload は生リクエストボディが allowlist だけを含むことを検証する。
// canary が一つでも見つかれば即座に失敗。ペイロードキーも allowlist で縛ることで
// 将来フィールドが追加されたときの黙った漏洩を防ぐ。
func assertSafePlanePayload(t *testing.T, rawBody []byte) {
	t.Helper()
	require.NotEmpty(t, rawBody)
	for field, canary := range planeExportCanaries {
		assert.NotContains(t, string(rawBody), canary,
			"user-controlled %s must never be exported to Plane", field)
	}
	var payload map[string]any
	require.NoError(t, json.Unmarshal(rawBody, &payload))
	for key := range payload {
		assert.Contains(t,
			[]string{"name", "description_html", "external_source", "external_id"},
			key, "payload is an allowlist; unexpected key %q", key)
	}
	assert.Equal(t, planeExternalSource, payload["external_source"])
	assert.Equal(t, "4242", payload["external_id"])
	assert.Equal(t, "[バグ報告] #4242", payload["name"])
	desc, _ := payload["description_html"].(string)
	assert.Contains(t, desc, "#4242")
	// 外部 URL（設定された frontend URL 含む）は一切送らない。
	assert.NotContains(t, desc, "http")
}

// クライアント境界そのもの: どんな入力が来ても送信内容は allowlist のみ。
func TestSafePlaneExport_NeverSendsUserControlledFields(t *testing.T) {
	captured := &planeExportCapture{}
	server := newPlaneExportServer(t, captured)
	client := newExportPlaneClient(server)

	issue, err := client.CreateBugReportIssue(context.Background(), canaryReport())
	require.NoError(t, err)
	require.NotNil(t, issue)
	require.Equal(t, 1, captured.calls)
	assertSafePlanePayload(t, captured.rawBody)
}

// 自動起票経路（Create → syncPlaneTicket）でも同じ境界を通る。
// ローカル保存は全文を保持し、export だけが制限されることを併せて検証する。
func TestSafePlaneExport_ViaAutoCreate(t *testing.T) {
	captured := &planeExportCapture{}
	server := newPlaneExportServer(t, captured)
	client := newExportPlaneClient(server)

	repo := &mockRepository{}
	repo.createFn = func(_ context.Context, r *model.SupportBugReport) error {
		r.ID = 4242
		return nil
	}
	repo.setTicketFn = func(context.Context, uint64, uint64, string, string) (bool, error) {
		return true, nil
	}
	svc := NewService(repo, client, nil)

	input := canaryCreateInput()
	report, err := svc.Create(context.Background(), 7, 9, input)
	require.NoError(t, err)
	require.NotNil(t, report)

	// ローカルの report は全文を保持する（export 境界だけを絞る設計の根拠）。
	assert.Equal(t, input.Title, report.Title)
	assert.Equal(t, input.Detail, report.Detail)
	assert.Equal(t, input.PageURL, report.PageURL)
	assert.Equal(t, input.UserAgent, report.UserAgent)
	assert.Equal(t, input.AppVersion, report.AppVersion)

	require.Equal(t, 1, captured.calls)
	assertSafePlanePayload(t, captured.rawBody)
}

// 手動再送経路（EnsurePlaneTicket）でも同じ境界を通る。
// DB から読み出した報告の本文が Plane に流れないことを固定する。
func TestSafePlaneExport_ViaManualEnsure(t *testing.T) {
	captured := &planeExportCapture{}
	server := newPlaneExportServer(t, captured)
	client := newExportPlaneClient(server)

	repo := &mockRepository{
		findByIDForClinicF: func(context.Context, uint64, uint64) (*model.SupportBugReport, error) {
			return canaryReport(), nil
		},
		setTicketFn: func(context.Context, uint64, uint64, string, string) (bool, error) {
			return true, nil
		},
	}
	svc := NewService(repo, client, nil)

	report, err := svc.EnsurePlaneTicket(context.Background(), 7, 4242)
	require.NoError(t, err)
	require.NotNil(t, report)
	require.Equal(t, 1, captured.calls)
	assertSafePlanePayload(t, captured.rawBody)
}
