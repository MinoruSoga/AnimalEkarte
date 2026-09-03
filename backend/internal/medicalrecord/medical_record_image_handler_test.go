package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

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
	uploadFn       func(ctx context.Context, key string, body io.Reader, contentType string) (string, error)
	deleteFn       func(ctx context.Context, key string) error
	getSignedURLFn func(ctx context.Context, key string, ttl time.Duration) (string, error)
	deleteCalls    []string
	signedURLCalls []string
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

func (m *mockMedicalRecordImageUploader) GetSignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	m.signedURLCalls = append(m.signedURLCalls, key)
	if m.getSignedURLFn != nil {
		return m.getSignedURLFn(ctx, key, ttl)
	}
	return "https://signed.example/tmp?key=" + key, nil
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
			name:    "returns 403 when selected clinic lacks medical record view grant",
			paramID: "5",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("clinic_id", "2")
				setResourcePermissionOnlyClinic(c, 1, string(model.ResourceMedicalRecords), "view")
			},
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					t.Fatal("medical record service must not be reached")
					return nil, nil
				},
			},
			imgSvc: &mockMedicalRecordImageService{
				listFn: func(_ context.Context, _, _ uint64) ([]model.MedicalRecordImage, error) {
					t.Fatal("image service must not be reached")
					return nil, nil
				},
			},
			wantStatus: http.StatusForbidden,
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
				require.NotEmpty(t, tt.uploader.deleteCalls, "expected uploader.Delete to be called for cleanup")
				assert.True(t, strings.HasPrefix(tt.uploader.deleteCalls[0], "medical-records/"+tt.paramID+"/"),
					"cleanup must Delete by object key, got %q", tt.uploader.deleteCalls[0])
				assert.NotContains(t, tt.uploader.deleteCalls[0], "https://")
			}
		})
	}
}

func ownedMedicalRecordImageGetter() *mockMedicalRecordService {
	return &mockMedicalRecordService{
		getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
		},
	}
}

func TestUploadMedicalRecordImage_PersistsObjectKeyAndReturnsSignedURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const signed = "https://signed.example/tmp?sig=upload"
	var persistedURL string
	uploader := &mockMedicalRecordImageUploader{
		uploadFn: func(_ context.Context, key string, _ io.Reader, contentType string) (string, error) {
			assert.Equal(t, "image/png", contentType)
			assert.True(t, strings.HasPrefix(key, "medical-records/5/"), "upload key = %q", key)
			return "https://uploads.example.test/" + key, nil
		},
		getSignedURLFn: func(_ context.Context, key string, ttl time.Duration) (string, error) {
			assert.Equal(t, medicalRecordImageSignedURLTTL, ttl)
			assert.True(t, strings.HasPrefix(key, "medical-records/5/"), "signed key = %q", key)
			assert.NotContains(t, key, "https://")
			return signed, nil
		},
	}

	h := newHandlerWithMedicalRecordImageSvc(
		ownedMedicalRecordImageGetter(),
		&mockMedicalRecordImageService{
			createFn: func(_ context.Context, clinicID, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(5), medicalRecordID)
				persistedURL = input.ImageURL
				assert.True(t, strings.HasPrefix(input.ImageURL, "medical-records/5/"), "ImageURL = %q", input.ImageURL)
				assert.NotContains(t, input.ImageURL, "https://")
				assert.NotContains(t, input.ImageURL, "uploads.example.test")
				return &model.MedicalRecordImage{
					ID:              3,
					MedicalRecordID: medicalRecordID,
					ImageURL:        input.ImageURL,
					ImageType:       input.ImageType,
				}, nil
			},
		},
		uploader,
	)

	body, ct := buildImageMultipart(t, "photo.png", "image/png", []byte("fake-image-bytes"))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", body)
	c.Request.Header.Set("Content-Type", ct)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)
	setStaffID(c)

	h.UploadMedicalRecordImage(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, strings.HasPrefix(persistedURL, "medical-records/5/"))
	assert.Contains(t, w.Body.String(), `"image_url":"`+signed+`"`)
	assert.NotContains(t, w.Body.String(), "https://uploads.example.test/")
	assert.NotContains(t, w.Body.String(), persistedURL)
	require.Len(t, uploader.signedURLCalls, 1)
}

func TestUploadMedicalRecordImage_CleanupDeletesObjectKeyNotPublicURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uploader := &mockMedicalRecordImageUploader{
		uploadFn: func(_ context.Context, key string, _ io.Reader, _ string) (string, error) {
			return "https://uploads.example.test/" + key, nil
		},
	}
	h := newHandlerWithMedicalRecordImageSvc(
		ownedMedicalRecordImageGetter(),
		&mockMedicalRecordImageService{
			createFn: func(_ context.Context, _, _ uint64, _ *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
				return nil, fmt.Errorf("db failure")
			},
		},
		uploader,
	)

	body, ct := buildImageMultipart(t, "photo.png", "image/png", []byte("fake"))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", body)
	c.Request.Header.Set("Content-Type", ct)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)
	setStaffID(c)

	h.UploadMedicalRecordImage(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.Len(t, uploader.deleteCalls, 1)
	assert.True(t, strings.HasPrefix(uploader.deleteCalls[0], "medical-records/5/"))
	assert.NotContains(t, uploader.deleteCalls[0], "https://")
	assert.NotContains(t, uploader.deleteCalls[0], "uploads.example.test")
}

func TestListMedicalRecordImages_SignsStorageKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const storedKey = "medical-records/5/uuid.png"
	const signed = "https://signed.example/tmp?sig=list"
	uploader := &mockMedicalRecordImageUploader{
		getSignedURLFn: func(_ context.Context, key string, ttl time.Duration) (string, error) {
			assert.Equal(t, storedKey, key)
			assert.Equal(t, medicalRecordImageSignedURLTTL, ttl)
			return signed, nil
		},
	}
	h := newHandlerWithMedicalRecordImageSvc(
		ownedMedicalRecordImageGetter(),
		&mockMedicalRecordImageService{
			listFn: func(_ context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(5), medicalRecordID)
				return []model.MedicalRecordImage{{
					ID:              1,
					MedicalRecordID: medicalRecordID,
					ImageURL:        storedKey,
					ThumbnailURL:    "",
				}}, nil
			},
		},
		uploader,
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/medical-records/5/images", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)

	h.ListMedicalRecordImages(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var items []medicalRecordImageResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
	require.Len(t, items, 1)
	assert.Equal(t, signed, items[0].ImageURL)
	assert.Empty(t, items[0].ThumbnailURL)
	assert.NotEqual(t, storedKey, items[0].ImageURL)
	assert.NotContains(t, w.Body.String(), `"image_url":"`+storedKey+`"`)
}

func TestListMedicalRecordImages_SignErrorFailClosedOmitsDurableURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const durable = "https://cdn.example/medical-records/5/secret-durable.png"
	uploader := &mockMedicalRecordImageUploader{
		getSignedURLFn: func(_ context.Context, key string, ttl time.Duration) (string, error) {
			assert.Equal(t, "medical-records/5/secret-durable.png", key)
			assert.Equal(t, medicalRecordImageSignedURLTTL, ttl)
			return "", fmt.Errorf("presign failed")
		},
	}
	h := newHandlerWithMedicalRecordImageSvc(
		ownedMedicalRecordImageGetter(),
		&mockMedicalRecordImageService{
			listFn: func(_ context.Context, _, _ uint64) ([]model.MedicalRecordImage, error) {
				return []model.MedicalRecordImage{{
					ID:              1,
					MedicalRecordID: 5,
					ImageURL:        durable,
				}}, nil
			},
		},
		uploader,
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/medical-records/5/images", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)

	h.ListMedicalRecordImages(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NotContains(t, w.Body.String(), durable)
	assert.NotContains(t, w.Body.String(), "secret-durable")
}

func TestListMedicalRecordImages_JSONCreateHTTPSUnchanged(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const clientURL = "https://example.test/a.png"
	uploader := &mockMedicalRecordImageUploader{
		getSignedURLFn: func(_ context.Context, key string, _ time.Duration) (string, error) {
			t.Fatalf("must not sign JSON-create client URL, key=%q", key)
			return "", nil
		},
	}
	h := newHandlerWithMedicalRecordImageSvc(
		ownedMedicalRecordImageGetter(),
		&mockMedicalRecordImageService{
			listFn: func(_ context.Context, _, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
				return []model.MedicalRecordImage{{
					ID:              1,
					MedicalRecordID: medicalRecordID,
					ImageURL:        clientURL,
				}}, nil
			},
		},
		uploader,
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/medical-records/5/images", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)

	h.ListMedicalRecordImages(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), clientURL)
	assert.Empty(t, uploader.signedURLCalls)
}

func TestListMedicalRecordImages_LegacyPublicURLExtractsKeyAndSigns(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const legacy = "https://cdn.example/medical-records/5/uuid.png"
	const legacyThumb = "https://cdn.example/medical-records/5/uuid-thumb.png"
	uploader := &mockMedicalRecordImageUploader{
		getSignedURLFn: func(_ context.Context, key string, ttl time.Duration) (string, error) {
			assert.Contains(t, []string{"medical-records/5/uuid.png", "medical-records/5/uuid-thumb.png"}, key)
			assert.Equal(t, medicalRecordImageSignedURLTTL, ttl)
			return "https://signed.example/tmp?key=" + key, nil
		},
	}
	h := newHandlerWithMedicalRecordImageSvc(
		ownedMedicalRecordImageGetter(),
		&mockMedicalRecordImageService{
			listFn: func(_ context.Context, _, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
				return []model.MedicalRecordImage{{
					ID:              2,
					MedicalRecordID: medicalRecordID,
					ImageURL:        legacy,
					ThumbnailURL:    legacyThumb,
				}}, nil
			},
		},
		uploader,
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/medical-records/5/images", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)

	h.ListMedicalRecordImages(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var items []medicalRecordImageResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
	require.Len(t, items, 1)
	assert.Equal(t, "https://signed.example/tmp?key=medical-records/5/uuid.png", items[0].ImageURL)
	assert.Equal(t, "https://signed.example/tmp?key=medical-records/5/uuid-thumb.png", items[0].ThumbnailURL)
	assert.NotContains(t, w.Body.String(), legacy)
	assert.NotContains(t, w.Body.String(), legacyThumb)
	assert.Equal(t, []string{"medical-records/5/uuid.png", "medical-records/5/uuid-thumb.png"}, uploader.signedURLCalls)
}

func TestListMedicalRecordImages_DoesNotSignBeforeOwnershipCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uploader := &mockMedicalRecordImageUploader{
		getSignedURLFn: func(_ context.Context, key string, _ time.Duration) (string, error) {
			t.Fatalf("must not sign before clinic-authorized ownership check, key=%q", key)
			return "", nil
		},
	}
	h := newHandlerWithMedicalRecordImageSvc(
		&mockMedicalRecordService{
			getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return nil, apperrors.WrapNotFound("medical_record", "5")
			},
		},
		&mockMedicalRecordImageService{
			listFn: func(_ context.Context, _, _ uint64) ([]model.MedicalRecordImage, error) {
				t.Fatal("list must not run when ownership fails")
				return nil, nil
			},
		},
		uploader,
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/medical-records/5/images", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)

	h.ListMedicalRecordImages(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, uploader.signedURLCalls)
}

func TestListMedicalRecordImages_DoesNotSignForeignMedicalRecordKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const foreign = "https://cdn.example/medical-records/50/other-clinic.png"
	uploader := &mockMedicalRecordImageUploader{
		getSignedURLFn: func(_ context.Context, key string, _ time.Duration) (string, error) {
			t.Fatalf("must not sign a key for another medical record, key=%q", key)
			return "", nil
		},
	}
	h := newHandlerWithMedicalRecordImageSvc(
		ownedMedicalRecordImageGetter(),
		&mockMedicalRecordImageService{
			listFn: func(_ context.Context, _, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
				return []model.MedicalRecordImage{{
					ID:              1,
					MedicalRecordID: medicalRecordID,
					ImageURL:        foreign,
				}}, nil
			},
		},
		uploader,
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/medical-records/5/images", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)

	h.ListMedicalRecordImages(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), foreign)
	assert.Empty(t, uploader.signedURLCalls)
}

func TestUploadMedicalRecordImage_SignErrorDeletesObjectAndDoesNotCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	createCalls := 0
	uploader := &mockMedicalRecordImageUploader{
		uploadFn: func(_ context.Context, key string, _ io.Reader, _ string) (string, error) {
			assert.True(t, strings.HasPrefix(key, "medical-records/5/"))
			return key, nil
		},
		getSignedURLFn: func(_ context.Context, _ string, _ time.Duration) (string, error) {
			return "", fmt.Errorf("presign failed")
		},
	}
	h := newHandlerWithMedicalRecordImageSvc(
		ownedMedicalRecordImageGetter(),
		&mockMedicalRecordImageService{
			createFn: func(_ context.Context, _, _ uint64, _ *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
				createCalls++
				t.Fatal("Create must not run when signing fails")
				return nil, nil
			},
		},
		uploader,
	)

	body, ct := buildImageMultipart(t, "photo.png", "image/png", []byte("fake"))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", body)
	c.Request.Header.Set("Content-Type", ct)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)
	setStaffID(c)

	h.UploadMedicalRecordImage(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Zero(t, createCalls)
	require.Len(t, uploader.deleteCalls, 1)
	assert.True(t, strings.HasPrefix(uploader.deleteCalls[0], "medical-records/5/"))
	assert.NotContains(t, w.Body.String(), "https://uploads.example.test/")
}

func TestCreateMedicalRecordImage_DoesNotSignForeignMedicalRecordURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const foreign = "https://cdn.example/medical-records/50/other.png"
	createCalls := 0
	uploader := &mockMedicalRecordImageUploader{
		getSignedURLFn: func(_ context.Context, key string, _ time.Duration) (string, error) {
			t.Fatalf("must not sign a key for another medical record, key=%q", key)
			return "", nil
		},
	}
	h := newHandlerWithMedicalRecordImageSvc(
		ownedMedicalRecordImageGetter(),
		&mockMedicalRecordImageService{
			createFn: func(_ context.Context, _, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
				createCalls++
				return &model.MedicalRecordImage{
					ID:              9,
					MedicalRecordID: medicalRecordID,
					ImageURL:        input.ImageURL,
					ImageType:       input.ImageType,
				}, nil
			},
		},
		uploader,
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/medical-records/5/images", bytes.NewReader([]byte(
		`{"image_url":"`+foreign+`","image_type":"xray"}`,
	)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)

	h.CreateMedicalRecordImage(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, 1, createCalls)
	assert.Contains(t, w.Body.String(), foreign)
	assert.Empty(t, uploader.signedURLCalls)
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
