package lstep

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	"github.com/animal-ekarte/backend/internal/model"
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

func newPostCsvImportRouter(csvSvc LstepCsvImportService, gateOverride any, setupCtx gin.HandlerFunc) *gin.Engine {
	h := &Handler{csvImport: csvSvc, requirePermission: testPermissionMiddleware}
	gate, ok := gateOverride.(gin.HandlerFunc)
	if !ok {
		gate = newLstepCSVImportConcurrencyGate(maxConcurrentLstepCSVImports)
	}
	r := gin.New()
	r.POST("/clinics/:clinic_id/lstep/csv-imports/friend-attributes",
		setupCtx,
		h.requirePermission(string(model.ResourceLstepCsvImport), "edit"),
		gate,
		h.ImportLstepFriendAttributesCsv,
	)
	return r
}

func newGetCsvImportsRouter(csvSvc LstepCsvImportService, _ any, setupCtx gin.HandlerFunc) *gin.Engine {
	h := &Handler{csvImport: csvSvc, requirePermission: testPermissionMiddleware}
	r := gin.New()
	r.GET("/clinics/:clinic_id/lstep/csv-imports",
		setupCtx,
		h.requirePermission(string(model.ResourceLstepCsvImport), "view"),
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

func TestPostLstepCsvImport_RejectsOversizedRequestBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	serviceCalled := false
	csvSvc := &mockLstepCsvImportService{
		importFriendAttributesFn: func(_ context.Context, _ uint64, _ string, _ io.Reader, _ uint64) (*model.LstepCsvImport, error) {
			serviceCalled = true
			return &model.LstepCsvImport{ID: uuid.New(), CreatedAt: time.Now()}, nil
		},
	}
	r := newPostCsvImportRouter(csvSvc, nil, func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) })
	body, ct := buildCSVMultipart(t, "line_user_id\nU123")
	req := httptest.NewRequest(http.MethodPost, "/clinics/1/lstep/csv-imports/friend-attributes", body)
	req.Header.Set("Content-Type", ct)
	req.ContentLength = maxCSVUploadRequestBytes + 1
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, serviceCalled)
}

func TestPostLstepCsvImport_RejectsConcurrentUploadBeforeReadingBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	firstDone := make(chan struct{})
	var serviceCalls atomic.Int32
	csvSvc := &mockLstepCsvImportService{
		importFriendAttributesFn: func(_ context.Context, clinicID uint64, fileName string, _ io.Reader, _ uint64) (*model.LstepCsvImport, error) {
			if serviceCalls.Add(1) == 1 {
				close(firstEntered)
				<-releaseFirst
			}
			return &model.LstepCsvImport{
				ID: uuid.New(), ClinicID: clinicID, FileName: fileName, CreatedAt: time.Now(),
			}, nil
		},
	}
	r := newPostCsvImportRouter(
		csvSvc,
		newLstepCSVImportConcurrencyGate(1),
		func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) },
	)

	firstBody, firstContentType := buildCSVMultipart(t, "line_user_id\nU-first")
	firstReq := httptest.NewRequest(http.MethodPost, "/clinics/1/lstep/csv-imports/friend-attributes", firstBody)
	firstReq.Header.Set("Content-Type", firstContentType)
	firstRecorder := httptest.NewRecorder()
	go func() {
		defer close(firstDone)
		r.ServeHTTP(firstRecorder, firstReq)
	}()
	<-firstEntered

	secondBody, secondContentType := buildCSVMultipart(t, "line_user_id\nU-second")
	secondReader := &countingCSVRequestBody{Reader: secondBody}
	secondReq := httptest.NewRequest(http.MethodPost, "/clinics/1/lstep/csv-imports/friend-attributes", secondReader)
	secondReq.Header.Set("Content-Type", secondContentType)
	secondRecorder := httptest.NewRecorder()
	r.ServeHTTP(secondRecorder, secondReq)

	assert.Equal(t, http.StatusTooManyRequests, secondRecorder.Code)
	assert.Zero(t, secondReader.bytesRead, "busy upload must be rejected before multipart parsing")
	assert.Equal(t, int32(1), serviceCalls.Load())
	close(releaseFirst)
	<-firstDone
	assert.Equal(t, http.StatusCreated, firstRecorder.Code)
}

func TestLstepCSVImportConcurrencyGate_IsolatesCapacityPerClinic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	firstClinicRequestEntered := make(chan struct{})
	releaseFirstClinicRequest := make(chan struct{})
	firstDone := make(chan struct{})
	var clinicOneCalls atomic.Int32
	gate := newLstepCSVImportConcurrencyGate(2)
	r := gin.New()
	r.POST(
		"/clinics/:clinic_id/import",
		func(c *gin.Context) {
			c.Set("clinic_id", c.Param("clinic_id"))
			c.Next()
		},
		gate,
		func(c *gin.Context) {
			if c.Param("clinic_id") == "1" && clinicOneCalls.Add(1) == 1 {
				close(firstClinicRequestEntered)
				<-releaseFirstClinicRequest
			}
			c.Status(http.StatusNoContent)
		},
	)

	firstRecorder := httptest.NewRecorder()
	go func() {
		defer close(firstDone)
		r.ServeHTTP(firstRecorder, httptest.NewRequest(http.MethodPost, "/clinics/1/import", http.NoBody))
	}()
	<-firstClinicRequestEntered

	sameClinicRecorder := httptest.NewRecorder()
	r.ServeHTTP(sameClinicRecorder, httptest.NewRequest(http.MethodPost, "/clinics/1/import", http.NoBody))
	otherClinicRecorder := httptest.NewRecorder()
	r.ServeHTTP(otherClinicRecorder, httptest.NewRequest(http.MethodPost, "/clinics/2/import", http.NoBody))

	assert.Equal(t, http.StatusTooManyRequests, sameClinicRecorder.Code)
	assert.Equal(t, http.StatusNoContent, otherClinicRecorder.Code)
	close(releaseFirstClinicRequest)
	<-firstDone
	assert.Equal(t, http.StatusNoContent, firstRecorder.Code)
}

type countingCSVRequestBody struct {
	Reader    io.Reader
	bytesRead int
}

func (r *countingCSVRequestBody) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.bytesRead += n
	return n, err
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
		&mockLstepCsvImportService{
			listByClinicFn: func(_ context.Context, clinicID uint64, _ int) ([]*model.LstepCsvImport, error) {
				return []*model.LstepCsvImport{{
					ID: uuid.New(), ClinicID: clinicID, CreatedAt: time.Now(),
					ErrorLog: datatypes.JSON(`[{"row":2,"reason":"parse_error"}]`),
				}}, nil
			},
		},
		nil,
		func(c *gin.Context) { setSystemAdmin(c); setClinicID(c) },
	)
	req := httptest.NewRequest(http.MethodGet, "/clinics/1/lstep/csv-imports", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "error_log")
}
