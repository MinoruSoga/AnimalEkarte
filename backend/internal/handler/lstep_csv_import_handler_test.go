package handler

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock LstepCsvImportService ----

type mockLstepCsvImportService struct {
	importFriendAttributesFn func(ctx context.Context, clinicID uint64, fileName string, fileReader io.Reader, uploadedByUserID uint64) (*model.LstepCsvImport, error)
	listByClinicFn           func(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error)
}

func (m *mockLstepCsvImportService) ImportFriendAttributesCSV(ctx context.Context, clinicID uint64, fileName string, fileReader io.Reader, uploadedByUserID uint64) (*model.LstepCsvImport, error) {
	if m.importFriendAttributesFn != nil {
		return m.importFriendAttributesFn(ctx, clinicID, fileName, fileReader, uploadedByUserID)
	}
	return nil, nil
}


func (m *mockLstepCsvImportService) ListByClinic(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error) {
	if m.listByClinicFn != nil {
		return m.listByClinicFn(ctx, clinicID, limit)
	}
	return []*model.LstepCsvImport{}, nil
}

// ---- router helpers ----

func newPostCsvImportRouter(csvSvc service.LstepCsvImportService, permSvc service.EffectivePermissionService, setupCtx gin.HandlerFunc) *gin.Engine {
	h := &Handler{svc: &service.Services{LstepCsvImport: csvSvc, EffectivePermission: permSvc}}
	r := gin.New()
	r.POST("/clinics/:clinic_id/lstep/csv-imports/friend-attributes",
		setupCtx,
		h.RequirePermission(string(model.ResourceLstepCsvImport), "edit"),
		h.ImportLstepFriendAttributesCsv,
	)
	return r
}

func newGetCsvImportsRouter(csvSvc service.LstepCsvImportService, permSvc service.EffectivePermissionService, setupCtx gin.HandlerFunc) *gin.Engine {
	h := &Handler{svc: &service.Services{LstepCsvImport: csvSvc, EffectivePermission: permSvc}}
	r := gin.New()
	r.GET("/clinics/:clinic_id/lstep/csv-imports",
		setupCtx,
		h.RequirePermission(string(model.ResourceLstepCsvImport), "view"),
		h.ListLstepCsvImports,
	)
	return r
}

func buildCSVMultipart(t *testing.T, csvContent string) (body *bytes.Buffer, contentType string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", "test.csv")
	assert.NoError(t, err)
	_, err = fw.Write([]byte(csvContent))
	assert.NoError(t, err)
	w.Close()
	return &buf, w.FormDataContentType()
}

// ---- Case A: POST 403 — 権限なし ----

func TestPostLstepCsvImport_A_403_NoPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newPostCsvImportRouter(
		&mockLstepCsvImportService{},
		&mockEffectivePermissionService{},
		func(c *gin.Context) { setNonSystemAdmin(c); setClinicID(c) },
	)
	body, ct := buildCSVMultipart(t, "line_user_id,display_name\nU123,Alice")
	req := httptest.NewRequest(http.MethodPost, "/clinics/1/lstep/csv-imports/friend-attributes", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---- Case B: POST 201 — 成功、Location ヘッダ付き ----

func TestPostLstepCsvImport_B_201_Created(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	csvSvc := &mockLstepCsvImportService{
		importFriendAttributesFn: func(_ context.Context, clinicID uint64, fileName string, _ io.Reader, _ uint64) (*model.LstepCsvImport, error) {
			return &model.LstepCsvImport{
				ID:        id,
				ClinicID:  clinicID,
				CsvType:   "friend_attribute",
				FileName:  fileName,
				Status:    "completed",
				CreatedAt: time.Now(),
			}, nil
		},
	}
	r := newPostCsvImportRouter(csvSvc, nil, func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) })
	body, ct := buildCSVMultipart(t, "line_user_id,display_name\nU123,Alice")
	req := httptest.NewRequest(http.MethodPost, "/clinics/1/lstep/csv-imports/friend-attributes", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Header().Get("Location"), id.String())
}

// ---- Case C: POST 400 — ファイルなし ----

func TestPostLstepCsvImport_C_400_MissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newPostCsvImportRouter(&mockLstepCsvImportService{}, nil, func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) })
	req := httptest.NewRequest(http.MethodPost, "/clinics/1/lstep/csv-imports/friend-attributes", http.NoBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---- Case D: GET 403 — 権限なし ----

func TestListLstepCsvImports_D_403_NoPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newGetCsvImportsRouter(
		&mockLstepCsvImportService{},
		&mockEffectivePermissionService{},
		func(c *gin.Context) { setNonSystemAdmin(c); setClinicID(c) },
	)
	req := httptest.NewRequest(http.MethodGet, "/clinics/1/lstep/csv-imports", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---- Case E: GET 200 — 成功 ----

func TestListLstepCsvImports_E_200_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newGetCsvImportsRouter(
		&mockLstepCsvImportService{},
		nil,
		func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) },
	)
	req := httptest.NewRequest(http.MethodGet, "/clinics/1/lstep/csv-imports", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
