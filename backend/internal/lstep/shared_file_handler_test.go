package lstep

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock SharedFileService ----

type mockSharedFileService struct {
	uploadFn         func(ctx context.Context, clinicID, uploadedBy uint64, input *UploadSharedFileInput) (*SharedFileResponse, error)
	getSignedURLFn   func(ctx context.Context, clinicID, id uint64) (string, error)
	findAllFn        func(ctx context.Context, clinicID uint64) ([]*SharedFileResponse, error)
	deleteFn         func(ctx context.Context, clinicID, id uint64) error
	cleanupExpiredFn func(ctx context.Context) error
}

func (m *mockSharedFileService) Upload(ctx context.Context, clinicID, uploadedBy uint64, input *UploadSharedFileInput) (*SharedFileResponse, error) {
	return m.uploadFn(ctx, clinicID, uploadedBy, input)
}

func (m *mockSharedFileService) GetSignedURL(ctx context.Context, clinicID, id uint64) (string, error) {
	return m.getSignedURLFn(ctx, clinicID, id)
}

func (m *mockSharedFileService) FindAll(ctx context.Context, clinicID uint64) ([]*SharedFileResponse, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockSharedFileService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockSharedFileService) CleanupExpired(ctx context.Context) error {
	if m.cleanupExpiredFn != nil {
		return m.cleanupExpiredFn(ctx)
	}
	return nil
}

var _ SharedFileService = (*mockSharedFileService)(nil)

func newHandlerWithSharedFileSvc(svc SharedFileService) *Handler {
	return &Handler{sharedFile: svc}
}

// buildSharedFileMultipart はテスト用の multipart/form-data ボディを組み立てる。
// includeFile が false の場合、file パートを省略する（欠落パラメータ検証用）。
func buildSharedFileMultipart(t *testing.T, includeFile bool, fileName string, content []byte, formFields map[string]string) (body *bytes.Buffer, contentType string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if includeFile {
		part, err := w.CreateFormFile("file", fileName)
		require.NoError(t, err)
		_, err = part.Write(content)
		require.NoError(t, err)
	}
	for k, v := range formFields {
		require.NoError(t, w.WriteField(k, v))
	}
	require.NoError(t, w.Close())
	return &buf, w.FormDataContentType()
}

// ---- ListSharedFiles ----

func TestListSharedFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockSharedFileService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of shared files",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockSharedFileService{
				findAllFn: func(_ context.Context, clinicID uint64) ([]*SharedFileResponse, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []*SharedFileResponse{{ID: 1, ClinicID: clinicID, FileName: "a.pdf"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"file_name":"a.pdf"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockSharedFileService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockSharedFileService{
				findAllFn: func(_ context.Context, _ uint64) ([]*SharedFileResponse, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithSharedFileSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/shared-files", http.NoBody)
			tt.setupCtx(c)
			h.ListSharedFiles(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UploadSharedFile ----

func TestUploadSharedFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		setupCtx     func(c *gin.Context)
		includeFile  bool
		fileName     string
		formFields   map[string]string
		svc          *mockSharedFileService
		wantStatus   int
		wantLocation bool
	}{
		{
			name:        "returns 201 with Location header",
			setupCtx:    func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			includeFile: true,
			fileName:    "file.pdf",
			formFields: map[string]string{
				"purpose":     model.SharedFilePurposeVaccineCert,
				"owner_id":    "10",
				"clinic_id":   "999",
				"uploaded_by": "999",
			},
			svc: &mockSharedFileService{
				uploadFn: func(_ context.Context, clinicID, uploadedBy uint64, input *UploadSharedFileInput) (*SharedFileResponse, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), uploadedBy)
					assert.Equal(t, model.SharedFilePurposeVaccineCert, input.Purpose)
					require.NotNil(t, input.OwnerID)
					assert.Equal(t, uint64(10), *input.OwnerID)
					return &SharedFileResponse{ID: 5, ClinicID: clinicID, FileName: "file.pdf"}, nil
				},
			},
			wantStatus:   http.StatusCreated,
			wantLocation: true,
		},
		{
			name:        "returns 401 when clinic_id is missing",
			setupCtx:    func(_ *gin.Context) {},
			includeFile: true,
			fileName:    "file.pdf",
			svc:         &mockSharedFileService{},
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "returns 401 when staff_id is missing",
			setupCtx:    func(c *gin.Context) { setClinicID(c) },
			includeFile: true,
			fileName:    "file.pdf",
			svc:         &mockSharedFileService{},
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "returns 400 when file field is missing",
			setupCtx:    func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			includeFile: false,
			formFields:  map[string]string{"purpose": model.SharedFilePurposeOther},
			svc:         &mockSharedFileService{},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "returns 400 when purpose is not an allowed enum",
			setupCtx:    func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			includeFile: true,
			fileName:    "file.pdf",
			formFields:  map[string]string{"purpose": "prescription"},
			svc:         &mockSharedFileService{},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "returns 400 when file extension is not allowed",
			setupCtx:    func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			includeFile: true,
			fileName:    "file.exe",
			svc:         &mockSharedFileService{},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "returns 400 on request bind error",
			setupCtx:    func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			includeFile: true,
			fileName:    "file.pdf",
			formFields:  map[string]string{"owner_id": "not-a-number"},
			svc:         &mockSharedFileService{},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "returns 500 on service error",
			setupCtx:    func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			includeFile: true,
			fileName:    "file.pdf",
			svc: &mockSharedFileService{
				uploadFn: func(_ context.Context, _, _ uint64, _ *UploadSharedFileInput) (*SharedFileResponse, error) {
					return nil, fmt.Errorf("upload failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithSharedFileSvc(tt.svc)
			body, contentType := buildSharedFileMultipart(t, tt.includeFile, tt.fileName, []byte("dummy"), tt.formFields)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/shared-files?clinic_id=999&uploaded_by=999", body)
			c.Request.Header.Set("Content-Type", contentType)
			tt.setupCtx(c)
			h.UploadSharedFile(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantLocation {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

// ---- GetSharedFileSignedURL ----

func TestGetSharedFileSignedURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockSharedFileService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns signed url",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockSharedFileService{
				getSignedURLFn: func(_ context.Context, clinicID, id uint64) (string, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					return "https://example.test/signed", nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"signed_url":"https://example.test/signed"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "5",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockSharedFileService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param is invalid",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockSharedFileService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockSharedFileService{
				getSignedURLFn: func(_ context.Context, _, _ uint64) (string, error) {
					return "", fmt.Errorf("storage failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithSharedFileSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/shared-files/"+tt.paramID+"/signed-url", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetSharedFileSignedURL(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteSharedFile ----

func TestDeleteSharedFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockSharedFileService
		wantStatus int
	}{
		{
			name:     "returns 204 on successful delete",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockSharedFileService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "5",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockSharedFileService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param is invalid",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockSharedFileService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockSharedFileService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return fmt.Errorf("delete failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithSharedFileSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/shared-files/"+tt.paramID, http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.DeleteSharedFile(c)
			c.Writer.WriteHeaderNow() // flush a bare c.Status() (no body) to the recorder
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- toSharedFileResponse ----

func TestToSharedFileResponse(t *testing.T) {
	expiresAt := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	ownerID := uint64(3)

	t.Run("maps all fields including expires_at", func(t *testing.T) {
		resp := toSharedFileResponse(&SharedFileResponse{
			ID:         1,
			ClinicID:   2,
			OwnerID:    &ownerID,
			UploadedBy: 4,
			FileType:   "pdf",
			FileName:   "file.pdf",
			FileSize:   1024,
			Purpose:    "prescription",
			ExpiresAt:  &expiresAt,
			CreatedAt:  expiresAt,
		})

		assert.Equal(t, uint64(1), resp.ID)
		assert.Equal(t, uint64(2), resp.ClinicID)
		require.NotNil(t, resp.OwnerID)
		assert.Equal(t, uint64(3), *resp.OwnerID)
		assert.Equal(t, uint64(4), resp.UploadedBy)
		assert.Equal(t, "pdf", resp.FileType)
		assert.Equal(t, "file.pdf", resp.FileName)
		assert.Equal(t, int64(1024), resp.FileSize)
		assert.Equal(t, "prescription", resp.Purpose)
		require.NotNil(t, resp.ExpiresAt)
		assert.True(t, resp.ExpiresAt.Equal(expiresAt))
		assert.True(t, resp.CreatedAt.Equal(expiresAt))
	})

	t.Run("handles nil owner_id and nil expires_at", func(t *testing.T) {
		resp := toSharedFileResponse(&SharedFileResponse{
			ID:         1,
			ClinicID:   2,
			OwnerID:    nil,
			UploadedBy: 4,
			FileType:   "pdf",
			FileName:   "file.pdf",
			FileSize:   1024,
			Purpose:    "other",
			ExpiresAt:  nil,
			CreatedAt:  time.Time{},
		})

		assert.Nil(t, resp.OwnerID)
		assert.Nil(t, resp.ExpiresAt)
		assert.True(t, resp.CreatedAt.IsZero())
	})
}
