package support

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock Service ----

type mockService struct {
	createFn        func(ctx context.Context, clinicID, reporterStaffID uint64, input CreateBugReportInput) (*model.SupportBugReport, error)
	listFn          func(ctx context.Context) ([]BugReportWithReporter, error)
	getFn           func(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	updateStatusFn  func(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) (*model.SupportBugReport, error)
	ensureTicketFn  func(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	deleteFn        func(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	listChatFn      func(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error)
	listExchangesFn func(ctx context.Context) ([]ChatExchange, error)
	recordChatFn    func(ctx context.Context, clinicID, staffID uint64, userMessage, assistantReply string, sources []ChatSource) error
	clearChatFn     func(ctx context.Context, clinicID, staffID uint64) error
	syncPlaneFn     func(ctx context.Context) PlaneSyncResult
}

func (m *mockService) Create(ctx context.Context, clinicID, reporterStaffID uint64, input CreateBugReportInput) (*model.SupportBugReport, error) {
	return m.createFn(ctx, clinicID, reporterStaffID, input)
}
func (m *mockService) ListAll(ctx context.Context) ([]BugReportWithReporter, error) {
	return m.listFn(ctx)
}
func (m *mockService) Get(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	if m.getFn == nil {
		return nil, apperrors.WrapNotFound("support_bug_report", "0")
	}
	return m.getFn(ctx, clinicID, id)
}
func (m *mockService) UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) (*model.SupportBugReport, error) {
	return m.updateStatusFn(ctx, clinicID, id, status)
}
func (m *mockService) EnsurePlaneTicket(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	return m.ensureTicketFn(ctx, clinicID, id)
}
func (m *mockService) Delete(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	return m.deleteFn(ctx, clinicID, id)
}
func (m *mockService) ListChatHistory(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error) {
	if m.listChatFn == nil {
		return nil, nil
	}
	return m.listChatFn(ctx, clinicID, staffID)
}
func (m *mockService) RecordChatExchange(ctx context.Context, clinicID, staffID uint64, userMessage, assistantReply string, sources []ChatSource) error {
	if m.recordChatFn == nil {
		return nil
	}
	return m.recordChatFn(ctx, clinicID, staffID, userMessage, assistantReply, sources)
}
func (m *mockService) ListChatExchanges(ctx context.Context) ([]ChatExchange, error) {
	if m.listExchangesFn == nil {
		return nil, nil
	}
	return m.listExchangesFn(ctx)
}
func (m *mockService) ClearChatHistory(ctx context.Context, clinicID, staffID uint64) error {
	if m.clearChatFn == nil {
		return nil
	}
	return m.clearChatFn(ctx, clinicID, staffID)
}
func (m *mockService) SyncPlaneTicketStates(ctx context.Context) PlaneSyncResult {
	if m.syncPlaneFn == nil {
		return PlaneSyncResult{}
	}
	return m.syncPlaneFn(ctx)
}

// ---- mock fileUploader ----

type mockUploader struct {
	uploadedKey   string
	uploadedBytes []byte
	uploadErr     error
	deletedKey    string
	deleteErr     error
}

func (m *mockUploader) Upload(_ context.Context, key string, body io.Reader, _ string) (string, error) {
	if m.uploadErr != nil {
		return "", m.uploadErr
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}
	m.uploadedKey = key
	m.uploadedBytes = b
	return key, nil
}
func (m *mockUploader) Delete(_ context.Context, key string) error {
	m.deletedKey = key
	return m.deleteErr
}
func (m *mockUploader) GetSignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "/uploads/" + key, nil
}

// ---- helpers ----

func setAuthContext(c *gin.Context) {
	c.Set("clinic_id", "1")
	c.Set("user_id", "5")
}

func buildMultipartBody(t *testing.T, fields map[string]string, fileField, fileName string, fileContent []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		require.NoError(t, w.WriteField(k, v))
	}
	if fileField != "" {
		part, err := w.CreateFormFile(fileField, fileName)
		require.NoError(t, err)
		_, err = part.Write(fileContent)
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	return &buf, w.FormDataContentType()
}

func newRequest(t *testing.T, method, target string, body *bytes.Buffer, contentType string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	var reader io.Reader
	if body != nil {
		reader = body
	}
	req := httptest.NewRequest(method, target, reader)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	c.Request = req
	setAuthContext(c)
	return c, rec
}

// ---- CreateBugReport ----

func TestCreateBugReport(t *testing.T) {
	createdReport := func() *model.SupportBugReport {
		return &model.SupportBugReport{ID: 10, ClinicID: 1, ReporterStaffID: 5, Title: "受付でエラー", Status: model.SupportBugReportStatusOpen}
	}

	tests := []struct {
		name         string
		fields       map[string]string
		fileField    string
		fileName     string
		fileContent  []byte
		svcErr       error
		wantStatus   int
		wantUpload   bool
		wantDelete   bool
		wantBodyPart string
	}{
		{
			name:         "creates report without screenshot",
			fields:       map[string]string{"title": "受付でエラー", "detail": "詳細", "route_path": "/reception"},
			wantStatus:   http.StatusCreated,
			wantBodyPart: `"title":"受付でエラー"`,
		},
		{
			name:        "creates report with png screenshot",
			fields:      map[string]string{"title": "表示崩れ"},
			fileField:   "screenshot",
			fileName:    "shot.png",
			fileContent: pngBytes,
			wantStatus:  http.StatusCreated,
			wantUpload:  true,
		},
		{
			name:       "rejects missing title",
			fields:     map[string]string{"detail": "タイトルなし"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:        "rejects non-image screenshot",
			fields:      map[string]string{"title": "x"},
			fileField:   "screenshot",
			fileName:    "shot.png",
			fileContent: []byte("not an image"),
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "deletes uploaded screenshot when service fails",
			fields:      map[string]string{"title": "x"},
			fileField:   "screenshot",
			fileName:    "shot.png",
			fileContent: pngBytes,
			svcErr:      apperrors.Wrap(assert.AnError, "db failed"),
			wantStatus:  http.StatusInternalServerError,
			wantUpload:  true,
			wantDelete:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{
				createFn: func(_ context.Context, clinicID, staffID uint64, input CreateBugReportInput) (*model.SupportBugReport, error) {
					if tt.svcErr != nil {
						return nil, tt.svcErr
					}
					r := createdReport()
					r.Title = input.Title
					r.ScreenshotKey = input.ScreenshotKey
					return r, nil
				},
			}
			uploader := &mockUploader{}
			h := NewHandler(svc, uploader, nil, nil, nil)

			body, contentType := buildMultipartBody(t, tt.fields, tt.fileField, tt.fileName, tt.fileContent)
			c, rec := newRequest(t, http.MethodPost, "/api/v1/support/bug-reports", body, contentType)
			c.Request.ContentLength = int64(body.Len())
			h.CreateBugReport(c)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBodyPart != "" {
				assert.Contains(t, rec.Body.String(), tt.wantBodyPart)
			}
			assert.Equal(t, tt.wantUpload, uploader.uploadedKey != "")
			assert.Equal(t, tt.wantDelete, uploader.deletedKey != "")
		})
	}
}

// ---- ListBugReports ----

func TestListBugReports(t *testing.T) {
	svc := &mockService{
		listFn: func(_ context.Context) ([]BugReportWithReporter, error) {
			key := "support-bug-reports/clinic-1/x.png"
			return []BugReportWithReporter{{
				SupportBugReport: model.SupportBugReport{
					ID: 1, ClinicID: 1, ReporterStaffID: 5,
					Title: "エラー", Status: model.SupportBugReportStatusOpen,
					ScreenshotKey: &key,
				},
				ReporterName: "田中",
				ClinicName:   "テスト医院A",
			}}, nil
		},
	}
	h := NewHandler(svc, &mockUploader{}, nil, nil, nil)

	c, rec := newRequest(t, http.MethodGet, "/api/v1/support/bug-reports", nil, "")
	h.ListBugReports(c)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"reporter_name":"田中"`)
	assert.Contains(t, rec.Body.String(), `"clinic_name":"テスト医院A"`)
	assert.Contains(t, rec.Body.String(), `"screenshot_url":"/uploads/support-bug-reports/clinic-1/x.png"`)
}

// ---- UpdateBugReportStatus ----

func TestUpdateBugReportStatus(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		svcErr     error
		wantStatus int
	}{
		{
			name:       "updates to resolved",
			body:       `{"status":"resolved"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "rejects invalid status",
			body:       `{"status":"bogus"}`,
			svcErr:     apperrors.WrapInvalidInput("invalid status"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects empty body",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{
				updateStatusFn: func(_ context.Context, clinicID, id uint64, status model.SupportBugReportStatus) (*model.SupportBugReport, error) {
					if tt.svcErr != nil {
						return nil, tt.svcErr
					}
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), id)
					return &model.SupportBugReport{ID: id, Status: status}, nil
				},
			}
			h := NewHandler(svc, nil, nil, nil, nil)

			body := strings.NewReader(tt.body)
			rec := httptest.NewRecorder()
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(rec)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/support/bug-reports/10/status", body)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			setAuthContext(c)

			h.UpdateBugReportStatus(c)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

// ---- CreatePlaneTicket ----

func TestCreatePlaneTicket(t *testing.T) {
	issueURL := "https://app.plane.so/baritechllc/browse/EMR-42/"
	tests := []struct {
		name         string
		svcResult    *model.SupportBugReport
		svcErr       error
		wantStatus   int
		wantBodyPart string
	}{
		{
			name:         "returns report with plane url",
			svcResult:    &model.SupportBugReport{ID: 10, PlaneIssueURL: &issueURL},
			wantStatus:   http.StatusOK,
			wantBodyPart: `"plane_issue_url":"https://app.plane.so/baritechllc/browse/EMR-42/"`,
		},
		{
			name:       "returns 501 when plane is not configured",
			svcErr:     apperrors.WrapNotImplemented("plane integration is not configured"),
			wantStatus: http.StatusNotImplemented,
		},
		{
			name:       "returns 502 on upstream failure",
			svcErr:     apperrors.WrapBadGateway("plane ticket creation failed"),
			wantStatus: http.StatusBadGateway,
		},
		{
			name:       "returns 404 for unknown report",
			svcErr:     apperrors.WrapNotFound("support_bug_report", "10"),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{
				ensureTicketFn: func(_ context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), id)
					return tt.svcResult, tt.svcErr
				},
			}
			h := NewHandler(svc, nil, nil, nil, nil)

			rec := httptest.NewRecorder()
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/support/bug-reports/10/plane-ticket", nil)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			setAuthContext(c)

			h.CreatePlaneTicket(c)
			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBodyPart != "" {
				assert.Contains(t, rec.Body.String(), tt.wantBodyPart)
			}
		})
	}
}

// ---- DeleteBugReport ----

func TestDeleteBugReport(t *testing.T) {
	screenshotKey := "support-bug-reports/clinic-1/shot.png"
	tests := []struct {
		name           string
		svcResult      *model.SupportBugReport
		svcErr         error
		deleteErr      error
		wantStatus     int
		wantObjDeleted bool
		wantSvcDelete  bool
	}{
		{
			name:           "soft deletes and removes screenshot object",
			svcResult:      &model.SupportBugReport{ID: 10, ClinicID: 1, ScreenshotKey: &screenshotKey},
			wantStatus:     http.StatusNoContent,
			wantObjDeleted: true,
			wantSvcDelete:  true,
		},
		{
			// オブジェクト削除に失敗したら 502 で報告行を残す（retry で orphan を防ぐ）。
			name:           "returns 502 and keeps the report when screenshot deletion fails",
			svcResult:      &model.SupportBugReport{ID: 10, ClinicID: 1, ScreenshotKey: &screenshotKey},
			deleteErr:      assert.AnError,
			wantStatus:     http.StatusBadGateway,
			wantObjDeleted: true,
			wantSvcDelete:  false,
		},
		{
			name:           "skips object deletion when report has no screenshot",
			svcResult:      &model.SupportBugReport{ID: 10, ClinicID: 1},
			wantStatus:     http.StatusNoContent,
			wantObjDeleted: false,
			wantSvcDelete:  true,
		},
		{
			name:           "returns 404 for unknown report",
			svcErr:         apperrors.WrapNotFound("support_bug_report", "10"),
			wantStatus:     http.StatusNotFound,
			wantObjDeleted: false,
			wantSvcDelete:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcDeleted := false
			svc := &mockService{
				getFn: func(_ context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), id)
					return tt.svcResult, tt.svcErr
				},
				deleteFn: func(_ context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
					svcDeleted = true
					return tt.svcResult, tt.svcErr
				},
			}
			uploader := &mockUploader{deleteErr: tt.deleteErr}
			h := NewHandler(svc, uploader, nil, nil, nil)

			rec := httptest.NewRecorder()
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/support/bug-reports/10", nil)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			setAuthContext(c)

			h.DeleteBugReport(c)
			// c.Status(204) は Gin の WriteHeaderNow まで recorder に反映されないため
			// ハンドラ直呼びのテストでは writer.Status() で検証する。
			assert.Equal(t, tt.wantStatus, c.Writer.Status())
			assert.Equal(t, tt.wantObjDeleted, uploader.deletedKey != "")
			assert.Equal(t, tt.wantSvcDelete, svcDeleted)
		})
	}
}
