package support

// plane_safe_export_test.go — Plane 送信ペイロードの安全境界テスト（PR650-01）。
//
// バグ報告の任意入力（件名・本文・ページURL・経路・UA・表示領域・バージョン・
// スクショキー）は患者・飼主・スタッフの個人情報やトークンを含みうるため、
// 外部（Plane）へは転記しない。送信境界は redaction ではなく allowlist:
// 送るのは決定論的な固定件名・固定説明文・サーバ採番の報告 ID のみ。
// 自動起票（Create）と手動再送（EnsurePlaneTicket）は、配送契約の接続まで
// 送信自体を停止する（EMR-549 C1）。
//
// canary にはフィールドごとに異なる機微文字列を使い、生リクエストボディに
// 混入した時点でそのフィールドの漏洩として検出する。
// 送信するのは決定論的な固定件名（環境名入り）・固定説明文・環境名+サーバ採番の
// 報告 ID・設定済み label UUID のみ（PR650-R3/R4）。

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

// export 境界が送る label は EMR プロジェクトの固定 UUID（source:manual / lane:dev）。
const (
	planeExportSourceLabelID = "31530203-1a6e-445f-9952-01c6eac98fe1"
	planeExportLaneLabelID   = "7a38738a-d932-4166-a7fa-87524f29b824"
)

func newExportPlaneClient(server *httptest.Server) *planeClient {
	return &planeClient{
		baseURL:           server.URL,
		webBaseURL:        "https://app.plane.example",
		workspaceSlug:     "ws",
		projectID:         "proj-1",
		projectIdentifier: "EMR",
		apiKey:            "test-key",
		httpClient:        server.Client(),
		labelIDs:          []string{planeExportSourceLabelID, planeExportLaneLabelID},
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
			[]string{"name", "description_html", "external_source", "external_id", "labels"},
			key, "payload is an allowlist; unexpected key %q", key)
	}
	assert.Equal(t, planeExternalSource, payload["external_source"])
	assert.Equal(t, "development:4242", payload["external_id"])
	assert.Equal(t, "[バグ報告][development] #4242", payload["name"])
	assert.Equal(t, []any{planeExportSourceLabelID, planeExportLaneLabelID}, payload["labels"])
	desc, _ := payload["description_html"].(string)
	assert.Contains(t, desc, "#4242")
	assert.Contains(t, desc, "development")
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

// 自動起票経路（Create）は配送契約の接続まで外部へ送らない（EMR-549 C1）。
// ローカル保存は全文を保持し、Plane への送信だけが停止していることを検証する。
func TestSafePlaneExport_CreateDoesNotSendBeforeDeliveryContract(t *testing.T) {
	captured := &planeExportCapture{}
	server := newPlaneExportServer(t, captured)
	client := newExportPlaneClient(server)

	repo := &mockRepository{}
	repo.createFn = func(_ context.Context, r *model.SupportBugReport) error {
		r.ID = 4242
		return nil
	}
	svc := NewService(repo, client, nil)

	input := canaryCreateInput()
	report, err := svc.Create(context.Background(), 7, 9, input)
	require.NoError(t, err)
	require.NotNil(t, report)

	// ローカルの report は全文を保持する（停止するのは export だけ）。
	assert.Equal(t, input.Title, report.Title)
	assert.Equal(t, input.Detail, report.Detail)
	assert.Equal(t, input.PageURL, report.PageURL)
	assert.Equal(t, input.UserAgent, report.UserAgent)
	assert.Equal(t, input.AppVersion, report.AppVersion)

	assert.Equal(t, 0, captured.calls, "plane must not be called before the delivery contract is connected")
}

// 手動再送経路（EnsurePlaneTicket）も報告元医院の確認後に 422 で停止し、外部へ送らない。
func TestSafePlaneExport_ManualEnsureDoesNotSendBeforeDeliveryContract(t *testing.T) {
	captured := &planeExportCapture{}
	server := newPlaneExportServer(t, captured)
	client := newExportPlaneClient(server)

	repo := &mockRepository{
		findByIDForClinicF: func(context.Context, uint64, uint64) (*model.SupportBugReport, error) {
			return canaryReport(), nil
		},
	}
	svc := NewService(repo, client, nil)

	report, err := svc.EnsurePlaneTicket(context.Background(), 7, 4242)
	require.Error(t, err)
	assert.Nil(t, report)
	assert.Equal(t, 0, captured.calls, "plane must not be called before the delivery contract is connected")
}
