package support

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- NewPlaneTicketCreator の有効/無効判定 ----

func TestNewPlaneTicketCreator_DisabledWithoutRequiredConfig(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		workspace string
		projectID string
		wantNil   bool
	}{
		{name: "nil when api key missing", apiKey: "", workspace: "ws", projectID: "p", wantNil: true},
		{name: "nil when workspace missing", apiKey: "k", workspace: "", projectID: "p", wantNil: true},
		{name: "nil when project id missing", apiKey: "k", workspace: "ws", projectID: "", wantNil: true},
		{name: "non nil when fully configured", apiKey: "k", workspace: "ws", projectID: "p", wantNil: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := NewPlaneTicketCreator("", "", tt.apiKey, tt.workspace, tt.projectID, "", "", time.Second)
			assert.Equal(t, tt.wantNil, creator == nil)
		})
	}
}

// ---- CreateBugReportIssue ----

type capturedPlaneRequest struct {
	path        string
	apiKey      string
	contentType string
	body        map[string]any
}

// newPlaneTestServer は Plane API のスタブを立て、送信リクエストを記録する。
func newPlaneTestServer(t *testing.T, status int, response string, captured *capturedPlaneRequest) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.path = r.URL.Path
		captured.apiKey = r.Header.Get("X-API-Key")
		captured.contentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured.body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
}

func newTestPlaneClient(server *httptest.Server, identifier string) *planeClient {
	return &planeClient{
		baseURL:           server.URL,
		webBaseURL:        "https://app.plane.so",
		workspaceSlug:     "baritechllc",
		projectID:         "proj-uuid",
		projectIdentifier: identifier,
		apiKey:            "secret-key",
		frontendURL:       "https://stg.noah-karte.com",
		httpClient:        server.Client(),
	}
}

func TestPlaneClient_CreateBugReportIssue_Success(t *testing.T) {
	captured := &capturedPlaneRequest{}
	server := newPlaneTestServer(t, http.StatusCreated,
		`{"id":"issue-uuid-1","identifier":"EMR-42","sequence_id":42}`, captured)
	defer server.Close()

	client := newTestPlaneClient(server, "EMR")
	report := &model.SupportBugReport{
		ID:              7,
		ClinicID:        1,
		ReporterStaffID: 5,
		Title:           "受付でエラー",
		Detail:          "詳細<script>alert(1)</script>",
		RoutePath:       "/reception",
	}

	issue, err := client.CreateBugReportIssue(context.Background(), report)
	require.NoError(t, err)
	assert.Equal(t, "issue-uuid-1", issue.ID)
	assert.Equal(t, "https://app.plane.so/baritechllc/browse/EMR-42/", issue.URL)

	// リクエスト面の検証: 宛先パス・APIキー・payload
	assert.Equal(t, "/v1/workspaces/baritechllc/projects/proj-uuid/issues/", captured.path)
	assert.Equal(t, "secret-key", captured.apiKey)
	assert.Equal(t, "application/json", captured.contentType)
	assert.Equal(t, "[バグ報告] 受付でエラー", captured.body["name"])
	assert.Equal(t, planeExternalSource, captured.body["external_source"])
	assert.Equal(t, "7", captured.body["external_id"])
	// XSS の生 HTML は埋め込まずエスケープされる
	desc, _ := captured.body["description_html"].(string)
	assert.NotContains(t, desc, "<script>")
	assert.Contains(t, desc, "&lt;script&gt;")
	assert.Contains(t, desc, "https://stg.noah-karte.com/settings/bug-reports")
}

func TestPlaneClient_CreateBugReportIssue_IdentifierFallbacks(t *testing.T) {
	tests := []struct {
		name            string
		response        string
		configuredIdent string
		wantURL         string
	}{
		{
			name:            "uses project_detail identifier when top-level missing",
			response:        `{"id":"i1","sequence_id":9,"project_detail":{"identifier":"EMR"}}`,
			configuredIdent: "",
			wantURL:         "https://app.plane.so/baritechllc/browse/EMR-9/",
		},
		{
			name:            "uses configured project identifier as last resort",
			response:        `{"id":"i2","sequence_id":11}`,
			configuredIdent: "EMR",
			wantURL:         "https://app.plane.so/baritechllc/browse/EMR-11/",
		},
		{
			name:            "empty url when identifier undeterminable",
			response:        `{"id":"i3"}`,
			configuredIdent: "EMR",
			wantURL:         "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captured := &capturedPlaneRequest{}
			server := newPlaneTestServer(t, http.StatusCreated, tt.response, captured)
			defer server.Close()

			client := newTestPlaneClient(server, tt.configuredIdent)
			issue, err := client.CreateBugReportIssue(context.Background(), &model.SupportBugReport{ID: 1, Title: "t"})
			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, issue.URL)
		})
	}
}

func TestPlaneClient_CreateBugReportIssue_Errors(t *testing.T) {
	t.Run("non-2xx becomes upstream error without leaking body", func(t *testing.T) {
		captured := &capturedPlaneRequest{}
		server := newPlaneTestServer(t, http.StatusInternalServerError,
			`{"error":"secret upstream detail <api-key>"}`, captured)
		defer server.Close()

		client := newTestPlaneClient(server, "EMR")
		issue, err := client.CreateBugReportIssue(context.Background(), &model.SupportBugReport{ID: 1, Title: "t"})
		require.Error(t, err)
		assert.Nil(t, issue)
		var upstream *planeUpstreamError
		require.ErrorAs(t, err, &upstream)
		assert.Equal(t, http.StatusInternalServerError, upstream.status)
		assert.NotContains(t, err.Error(), "secret upstream detail")
	})

	t.Run("missing id in 2xx response is an error", func(t *testing.T) {
		captured := &capturedPlaneRequest{}
		server := newPlaneTestServer(t, http.StatusOK, `{"identifier":"EMR-1"}`, captured)
		defer server.Close()

		client := newTestPlaneClient(server, "EMR")
		issue, err := client.CreateBugReportIssue(context.Background(), &model.SupportBugReport{ID: 1, Title: "t"})
		require.Error(t, err)
		assert.Nil(t, issue)
	})
}

// ---- planeSyncErrorMessage / planeIssueName ----

func TestPlaneSyncErrorMessage(t *testing.T) {
	assert.Equal(t, "plane api error (status 503)", planeSyncErrorMessage(&planeUpstreamError{status: 503}))
	assert.Equal(t, "plane request failed", planeSyncErrorMessage(assert.AnError))
}

func TestPlaneIssueName(t *testing.T) {
	assert.Equal(t, "[バグ報告] 件名", planeIssueName(&model.SupportBugReport{Title: "  件名  "}))

	longTitle := strings.Repeat("あ", 300)
	name := planeIssueName(&model.SupportBugReport{Title: longTitle})
	assert.LessOrEqual(t, len([]rune(name)), planeIssueNameMaxRunes)
	assert.True(t, strings.HasSuffix(name, "…"))
}
