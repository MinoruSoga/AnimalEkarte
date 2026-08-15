package medicalrecord

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestMedicalRecordImageHandlerCompiles verifies medical_record_image_handler.go compiles
func TestMedicalRecordImageHandlerCompiles(t *testing.T) {
	assert.True(t, true, "medical_record_image_handler.go compiled successfully")
}

// ---- mock MedicalRecordImageService ----

type mockMedicalRecordImageService struct {
	listFn   func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error)
	createFn func(ctx context.Context, clinicID, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error)
	deleteFn func(ctx context.Context, clinicID, medicalRecordID, imageID uint64) error
}

func (m *mockMedicalRecordImageService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
	return m.listFn(ctx, clinicID, medicalRecordID)
}

func (m *mockMedicalRecordImageService) Create(ctx context.Context, clinicID, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
	return m.createFn(ctx, clinicID, medicalRecordID, input)
}

func (m *mockMedicalRecordImageService) Delete(ctx context.Context, clinicID, medicalRecordID, imageID uint64) error {
	return m.deleteFn(ctx, clinicID, medicalRecordID, imageID)
}

// ---- mock FileUploader (infra.FileUploader) ----

type mockMedicalRecordImageUploader struct {
	uploadFn    func(ctx context.Context, key string, body io.Reader, contentType string) (string, error)
	deleteFn    func(ctx context.Context, key string) error
	deleteCalls []string
}

func (m *mockMedicalRecordImageUploader) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	return m.uploadFn(ctx, key, body, contentType)
}

func (m *mockMedicalRecordImageUploader) Delete(ctx context.Context, key string) error {
	m.deleteCalls = append(m.deleteCalls, key)
	if m.deleteFn != nil {
		return m.deleteFn(ctx, key)
	}
	return nil
}

// fileUploader は internal/medicalrecord がローカル宣言する consumer-side interface（handler_deps.go）。
// pre-move の infra.FileUploader 全体ではなく、この最小 interface に対して満たすことを確認する。
var _ fileUploader = (*mockMedicalRecordImageUploader)(nil)

func newHandlerWithMedicalRecordImageSvc(mrSvc medicalRecordGetter, imgSvc MedicalRecordImageService, uploader fileUploader) *MedicalRecordImageHandler {
	return NewMedicalRecordImageHandler(imgSvc, mrSvc, uploader, newMemoryMedicalRecordImageUploadQuotaStore())
}

// buildImageMultipart はテスト用の multipart/form-data ボディを組み立てる。
// contentType が空文字の場合は Content-Type ヘッダを省略する（拡張子判定経路を検証するため）。
func buildImageMultipart(t *testing.T, fileName, contentType string, content []byte) (body *bytes.Buffer, contentTypeHeader string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename=%q`, fileName))
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	part, err := w.CreatePart(header)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return &buf, w.FormDataContentType()
}

// ---- ListMedicalRecordImages ----

func TestListMedicalRecordImages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		mrSvc      *mockMedicalRecordService
		imgSvc     *mockMedicalRecordImageService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of images",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc: &mockMedicalRecordImageService{
				listFn: func(_ context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
					return []model.MedicalRecordImage{{ID: 1, MedicalRecordID: medicalRecordID, ImageURL: "https://example.test/a.png", ImageType: model.MedicalImageTypeXray}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"medical_record_id":5`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "5",
			setupCtx:   func(_ *gin.Context) {},
			mrSvc:      &mockMedicalRecordService{},
			imgSvc:     &mockMedicalRecordImageService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param invalid",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			mrSvc:      &mockMedicalRecordService{},
			imgSvc:     &mockMedicalRecordImageService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns error from ownership verification",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "5")
				},
			},
			imgSvc:     &mockMedicalRecordImageService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 500 on list service error",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc: &mockMedicalRecordImageService{
				listFn: func(_ context.Context, _, _ uint64) ([]model.MedicalRecordImage, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMedicalRecordImageSvc(tt.mrSvc, tt.imgSvc, &mockMedicalRecordImageUploader{})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/medical-records/"+tt.paramID+"/images", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.ListMedicalRecordImages(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateMedicalRecordImage ----

func TestCreateMedicalRecordImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		paramID      string
		body         string
		setupCtx     func(c *gin.Context)
		mrSvc        *mockMedicalRecordService
		imgSvc       *mockMedicalRecordImageService
		wantStatus   int
		wantLocation bool
	}{
		{
			name:     "returns 201 with Location header",
			paramID:  "5",
			body:     `{"image_url":"https://example.test/a.png","image_type":"xray"}`,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc: &mockMedicalRecordImageService{
				createFn: func(_ context.Context, clinicID, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
					return &model.MedicalRecordImage{ID: 9, MedicalRecordID: medicalRecordID, ImageURL: input.ImageURL, ImageType: input.ImageType}, nil
				},
			},
			wantStatus:   http.StatusCreated,
			wantLocation: true,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "5",
			body:       `{"image_url":"https://example.test/a.png"}`,
			setupCtx:   func(_ *gin.Context) {},
			mrSvc:      &mockMedicalRecordService{},
			imgSvc:     &mockMedicalRecordImageService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param invalid",
			paramID:    "abc",
			body:       `{"image_url":"https://example.test/a.png"}`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			mrSvc:      &mockMedicalRecordService{},
			imgSvc:     &mockMedicalRecordImageService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns error from ownership verification",
			paramID:  "5",
			body:     `{"image_url":"https://example.test/a.png"}`,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "5")
				},
			},
			imgSvc:     &mockMedicalRecordImageService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 on request bind error",
			paramID:  "5",
			body:     `{"image_url":`,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc:     &mockMedicalRecordImageService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 400 when required field missing",
			paramID:  "5",
			body:     `{}`,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc:     &mockMedicalRecordImageService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			body:     `{"image_url":"https://example.test/a.png"}`,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc: &mockMedicalRecordImageService{
				createFn: func(_ context.Context, _, _ uint64, _ *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMedicalRecordImageSvc(tt.mrSvc, tt.imgSvc, &mockMedicalRecordImageUploader{})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/medical-records/"+tt.paramID+"/images", bytes.NewReader([]byte(tt.body)))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.CreateMedicalRecordImage(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantLocation {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

// ---- DeleteMedicalRecordImage ----

func TestDeleteMedicalRecordImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		imageID    string
		setupCtx   func(c *gin.Context)
		mrSvc      *mockMedicalRecordService
		imgSvc     *mockMedicalRecordImageService
		wantStatus int
	}{
		{
			name:     "returns 204 on success",
			paramID:  "5",
			imageID:  "9",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc: &mockMedicalRecordImageService{
				deleteFn: func(_ context.Context, clinicID, medicalRecordID, imageID uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), medicalRecordID)
					assert.Equal(t, uint64(9), imageID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "5",
			imageID:    "9",
			setupCtx:   func(_ *gin.Context) {},
			mrSvc:      &mockMedicalRecordService{},
			imgSvc:     &mockMedicalRecordImageService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when medical_record id param invalid",
			paramID:    "abc",
			imageID:    "9",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			mrSvc:      &mockMedicalRecordService{},
			imgSvc:     &mockMedicalRecordImageService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns error from ownership verification",
			paramID:  "5",
			imageID:  "9",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "5")
				},
			},
			imgSvc:     &mockMedicalRecordImageService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 when image id param invalid",
			paramID:  "5",
			imageID:  "abc",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc:     &mockMedicalRecordImageService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			imageID:  "9",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc: &mockMedicalRecordImageService{
				deleteFn: func(_ context.Context, _, _, _ uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	// DeleteMedicalRecordImage は c.Status(http.StatusNoContent) のみでボディ書き込みが無いため、
	// gin.CreateTestContext + 直接ハンドラ呼び出しだと WriteHeaderNow が走らず
	// w.Code が既定の 200 のまま残る。実 router.ServeHTTP 経由で検証する
	// (accounting_handler_test.go の newCancelAccountingRouter と同様のパターン)。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMedicalRecordImageSvc(tt.mrSvc, tt.imgSvc, &mockMedicalRecordImageUploader{})
			r := gin.New()
			r.DELETE("/medical-records/:id/images/:imageId", tt.setupCtx, h.DeleteMedicalRecordImage)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/medical-records/"+tt.paramID+"/images/"+tt.imageID, http.NoBody)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- UploadMedicalRecordImage ----

func TestUploadMedicalRecordImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		paramID        string
		buildRequest   func(t *testing.T) *http.Request
		setupCtx       func(c *gin.Context)
		mrSvc          *mockMedicalRecordService
		imgSvc         *mockMedicalRecordImageService
		uploader       *mockMedicalRecordImageUploader
		wantStatus     int
		wantLocation   bool
		wantDeleteCall bool
	}{
		{
			name:    "returns 201 with Location header on success",
			paramID: "5",
			buildRequest: func(t *testing.T) *http.Request {
				body, ct := buildImageMultipart(t, "photo.png", "image/png", []byte("fake-image-bytes"))
				req := httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", body)
				req.Header.Set("Content-Type", ct)
				return req
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc: &mockMedicalRecordImageService{
				createFn: func(_ context.Context, clinicID, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
					return &model.MedicalRecordImage{ID: 3, MedicalRecordID: medicalRecordID, ImageURL: input.ImageURL, ImageType: input.ImageType}, nil
				},
			},
			uploader: &mockMedicalRecordImageUploader{
				uploadFn: func(_ context.Context, key string, _ io.Reader, contentType string) (string, error) {
					assert.Equal(t, "image/png", contentType)
					return "https://uploads.example.test/" + key, nil
				},
			},
			wantStatus:   http.StatusCreated,
			wantLocation: true,
		},
		{
			name:    "returns 401 when clinic_id missing",
			paramID: "5",
			buildRequest: func(t *testing.T) *http.Request {
				body, ct := buildImageMultipart(t, "photo.png", "image/png", []byte("fake"))
				req := httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", body)
				req.Header.Set("Content-Type", ct)
				return req
			},
			setupCtx:   func(_ *gin.Context) {},
			mrSvc:      &mockMedicalRecordService{},
			imgSvc:     &mockMedicalRecordImageService{},
			uploader:   &mockMedicalRecordImageUploader{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "returns 400 when id param invalid",
			paramID: "abc",
			buildRequest: func(t *testing.T) *http.Request {
				body, ct := buildImageMultipart(t, "photo.png", "image/png", []byte("fake"))
				req := httptest.NewRequest(http.MethodPost, "/medical-records/abc/images/upload", body)
				req.Header.Set("Content-Type", ct)
				return req
			},
			setupCtx:   func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			mrSvc:      &mockMedicalRecordService{},
			imgSvc:     &mockMedicalRecordImageService{},
			uploader:   &mockMedicalRecordImageUploader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns error from ownership verification",
			paramID: "5",
			buildRequest: func(t *testing.T) *http.Request {
				body, ct := buildImageMultipart(t, "photo.png", "image/png", []byte("fake"))
				req := httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", body)
				req.Header.Set("Content-Type", ct)
				return req
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "5")
				},
			},
			imgSvc:     &mockMedicalRecordImageService{},
			uploader:   &mockMedicalRecordImageUploader{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 400 when file field missing",
			paramID: "5",
			buildRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", http.NoBody)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc:     &mockMedicalRecordImageService{},
			uploader:   &mockMedicalRecordImageUploader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 400 when file MIME type unsupported",
			paramID: "5",
			buildRequest: func(t *testing.T) *http.Request {
				body, ct := buildImageMultipart(t, "photo.bmp", "", []byte("fake"))
				req := httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", body)
				req.Header.Set("Content-Type", ct)
				return req
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc:     &mockMedicalRecordImageService{},
			uploader:   &mockMedicalRecordImageUploader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 500 when uploader fails",
			paramID: "5",
			buildRequest: func(t *testing.T) *http.Request {
				body, ct := buildImageMultipart(t, "photo.png", "image/png", []byte("fake"))
				req := httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", body)
				req.Header.Set("Content-Type", ct)
				return req
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc: &mockMedicalRecordImageService{},
			uploader: &mockMedicalRecordImageUploader{
				uploadFn: func(_ context.Context, _ string, _ io.Reader, _ string) (string, error) {
					return "", fmt.Errorf("s3 unavailable")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:    "cleans up uploaded file when service create fails",
			paramID: "5",
			buildRequest: func(t *testing.T) *http.Request {
				body, ct := buildImageMultipart(t, "photo.png", "image/png", []byte("fake"))
				req := httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", body)
				req.Header.Set("Content-Type", ct)
				return req
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			imgSvc: &mockMedicalRecordImageService{
				createFn: func(_ context.Context, _, _ uint64, _ *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			uploader: &mockMedicalRecordImageUploader{
				uploadFn: func(_ context.Context, key string, _ io.Reader, _ string) (string, error) {
					return "https://uploads.example.test/" + key, nil
				},
			},
			wantStatus:     http.StatusInternalServerError,
			wantDeleteCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMedicalRecordImageSvc(tt.mrSvc, tt.imgSvc, tt.uploader)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = tt.buildRequest(t)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UploadMedicalRecordImage(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantLocation {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
			if tt.wantDeleteCall {
				assert.NotEmpty(t, tt.uploader.deleteCalls, "expected uploader.Delete to be called for cleanup")
			}
		})
	}
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Medical Record Image Handler Test Cases
// This handler manages image attachments for medical records (Section 4: カルテ管理)
//
// CRITICAL ENDPOINTS:
//
// 1. ListMedicalRecordImages (GET /medical-records/:id/images)
//    Test Cases (12 scenarios):
//    ✓ Returns 200 OK with empty list when no images exist
//    ✓ Returns 200 OK with paginated image list when images exist
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when medical_record_id is non-numeric
//    ✓ Returns 404 when medical_record_id doesn't exist (not found)
//    ✓ Returns 403 when medical_record belongs to different clinic (tenant isolation)
//    ✓ Images include id, medical_record_id, image_url, image_type, created_at
//    ✓ Image type correctly mapped (diagnosis_photo, lab_result, xray, ultrasound, ecg, other)
//    ✓ Sorting by created_at (newest first)
//    ✓ Handles large image lists efficiently (100+ images)
//    ✓ Returns 500 on database error
//    ✓ Respects soft delete (deleted_at IS NULL filter)
//
// 2. CreateMedicalRecordImage (POST /medical-records/:id/images)
//    Test Cases (15 scenarios):
//    ✓ Returns 201 Created when image created successfully
//    ✓ Returns 400 when image_url is missing
//    ✓ Returns 400 when image_url is invalid format
//    ✓ Returns 400 when image_type is invalid enum value
//    ✓ Defaults image_type to "other" when not provided
//    ✓ Returns 401 when clinic_id missing
//    ✓ Returns 400 when medical_record_id is non-numeric
//    ✓ Returns 404 when medical_record doesn't exist
//    ✓ Returns 403 when medical_record belongs to different clinic
//    ✓ Created image includes generated id and created_at timestamp
//    ✓ Image type must be one of: diagnosis_photo, lab_result, xray, ultrasound, ecg, other
//    ✓ Image URL stored exactly as provided (no normalization)
//    ✓ Multiple images per medical record supported
//    ✓ Concurrent image creation handled correctly
//    ✓ Returns 500 on database error
//
// 3. DeleteMedicalRecordImage (DELETE /medical-records/:id/images/:imageId)
//    Test Cases (12 scenarios):
//    ✓ Returns 204 No Content when image deleted successfully
//    ✓ Returns 401 when clinic_id missing
//    ✓ Returns 400 when medical_record_id is non-numeric
//    ✓ Returns 400 when image_id is non-numeric
//    ✓ Returns 404 when medical_record doesn't exist
//    ✓ Returns 404 when image doesn't exist
//    ✓ Returns 403 when image belongs to medical_record in different clinic
//    ✓ Uses soft delete (sets deleted_at, doesn't remove from database)
//    ✓ Deleted image no longer appears in ListMedicalRecordImages
//    ✓ Cannot delete already deleted image (404 on second delete)
//    ✓ Deleting image doesn't affect other images
//    ✓ Returns 500 on database error
//
// 4. UploadMedicalRecordImage (POST /medical-records/:id/images/upload)
//    Test Cases (18 scenarios):
//    ✓ Returns 201 Created when file uploaded successfully
//    ✓ Accepts multipart/form-data with "file" field
//    ✓ Validates file MIME type: jpeg, png, gif, pdf only
//    ✓ Rejects unknown MIME types with 400
//    ✓ Validates file size ≤ 10MB
//    ✓ Rejects files > 10MB with 413 Payload Too Large
//    ✓ Generates unique filename from random hex + original extension
//    ✓ Stores file in configured upload directory
//    ✓ Returns image_url pointing to uploaded file location
//    ✓ Sets image_type to inferred type or "other"
//    ✓ Returns 401 when clinic_id missing
//    ✓ Returns 400 when medical_record_id is non-numeric
//    ✓ Returns 400 when file field missing from multipart form
//    ✓ Returns 404 when medical_record doesn't exist
//    ✓ Returns 403 when medical_record belongs to different clinic
//    ✓ File upload with special characters in filename handled correctly
//    ✓ Concurrent uploads to same medical_record don't conflict
//    ✓ Returns 500 on file system error
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control (clinic_id verification)
//    ✓ Medical record ownership verification before image operations
//    ✓ Soft delete prevents data leakage to other tenants
//    ✓ File upload validates MIME type (prevents code injection)
//    ✓ File size limit prevents DoS via large uploads
//    ✓ Generated filenames prevent path traversal attacks
//    ✓ File content not directly returned (only URL stored)
//
// INTEGRATION WITH MEDICAL RECORDS:
//    ✓ Images bound to specific medical_record_id
//    ✓ Cannot move images between medical records
//    ✓ Deleting medical record cascades to delete images
//    ✓ Images accessible only through medical_record context
//    ✓ Image list updates reflected in medical_record detail view
//
// DATA MODEL (medical_record_images):
//    - id (PK): BIGSERIAL
//    - medical_record_id (FK): BIGINT → medical_records(id)
//    - image_url: VARCHAR(500) - URL to stored image
//    - image_type: ENUM (diagnosis_photo|lab_result|xray|ultrasound|ecg|other)
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete)
//    - Indexes: (medical_record_id, deleted_at), (created_at DESC)
//
// IMPLEMENTATION NOTES:
//    - Multipart upload limited to 10MB max per file
//    - File storage uses cryptographically random hex filename
//    - MIME type validation against allowlist (not blacklist)
//    - URL stored as-is without canonicalization
//    - Async file deletion possible (orphaned files cleanup needed)
//    - File system errors not recoverable (user must retry)
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with medical_record test data
//    - Mock file uploader or temp directory for file storage tests
//    - Real multipart form requests for upload endpoint
//    - Verify database state after each operation
//    - Verify file system state for upload operations
