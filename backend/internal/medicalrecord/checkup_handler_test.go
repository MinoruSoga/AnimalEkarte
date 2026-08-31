package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock CheckupService ----

type mockCheckupService struct {
	listFn         func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error)
	listByClinicFn func(ctx context.Context, input ListCheckupsByClinicInput) ([]model.Checkup, int64, error)
	createFn       func(ctx context.Context, medicalRecordID uint64, input *CreateCheckupInput) (*model.Checkup, error)
	updateFn       func(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, input *UpdateCheckupInput) (*model.Checkup, error)
	deleteFn       func(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) error
}

func (m *mockCheckupService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID, medicalRecordID)
	}
	return nil, nil
}

func (m *mockCheckupService) ListByClinic(ctx context.Context, input ListCheckupsByClinicInput) ([]model.Checkup, int64, error) {
	if m.listByClinicFn != nil {
		return m.listByClinicFn(ctx, input)
	}
	return nil, 0, nil
}

func (m *mockCheckupService) Create(ctx context.Context, medicalRecordID uint64, input *CreateCheckupInput) (*model.Checkup, error) {
	if m.createFn != nil {
		return m.createFn(ctx, medicalRecordID, input)
	}
	return nil, nil
}

func (m *mockCheckupService) Update(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, input *UpdateCheckupInput) (*model.Checkup, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, medicalRecordID, checkupID, input)
	}
	return nil, nil
}

func (m *mockCheckupService) Delete(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, medicalRecordID, checkupID)
	}
	return nil
}

func (m *mockCheckupService) Wait() {}

// ---- helpers ----

func newHandlerWithCheckupSvc(svc CheckupService) *CheckupHandler {
	return NewCheckupHandler(svc, nil)
}

// ---- ListCheckups ----

func TestListCheckups(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of checkups",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupService{
				listFn: func(_ context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), medicalRecordID)
					return []model.Checkup{{ID: 1, MedicalRecordID: 1, CheckupTypeID: 2, Result: "normal"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"result":"normal"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid id param",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupService{
				listFn: func(_ context.Context, _, _ uint64) ([]model.Checkup, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.ListCheckups(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateCheckup ----

func TestCreateCheckup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"checkup_type_id": 2,
			"date":            "2026-05-28",
			"result":          "normal",
		}
	}

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupService
		wantStatus int
		wantHeader bool
	}{
		{
			name:     "creates checkup successfully",
			paramID:  "1",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupService{
				createFn: func(_ context.Context, medicalRecordID uint64, input *CreateCheckupInput) (*model.Checkup, error) {
					assert.Equal(t, uint64(1), medicalRecordID)
					assert.Equal(t, uint64(1), input.ClinicID)
					return &model.Checkup{ID: 5, MedicalRecordID: medicalRecordID, CheckupTypeID: input.CheckupTypeID, Result: input.Result}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: true,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid id param",
			paramID:    "abc",
			body:       validBody(),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when checkup_type_id is missing",
			paramID:    "1",
			body:       map[string]any{"date": "2026-05-28"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when date is invalid format",
			paramID:    "1",
			body:       map[string]any{"checkup_type_id": 2, "date": "2026/05/28"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupService{
				createFn: func(_ context.Context, _ uint64, _ *CreateCheckupInput) (*model.Checkup, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.CreateCheckup(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantHeader {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

// ---- UpdateCheckup ----

func TestUpdateCheckup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		params     gin.Params
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates checkup successfully",
			params:   gin.Params{{Key: "id", Value: "1"}, {Key: "checkupId", Value: "2"}},
			body:     map[string]any{"result": "abnormal"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupService{
				updateFn: func(_ context.Context, clinicID, medicalRecordID, checkupID uint64, input *UpdateCheckupInput) (*model.Checkup, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), medicalRecordID)
					assert.Equal(t, uint64(2), checkupID)
					return &model.Checkup{ID: checkupID, MedicalRecordID: medicalRecordID, Result: *input.Result}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"result":"abnormal"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			params:     gin.Params{{Key: "id", Value: "1"}, {Key: "checkupId", Value: "2"}},
			body:       map[string]any{"result": "abnormal"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid medical record id",
			params:     gin.Params{{Key: "id", Value: "abc"}, {Key: "checkupId", Value: "2"}},
			body:       map[string]any{"result": "abnormal"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid checkup id",
			params:     gin.Params{{Key: "id", Value: "1"}, {Key: "checkupId", Value: "abc"}},
			body:       map[string]any{"result": "abnormal"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON body",
			params:     gin.Params{{Key: "id", Value: "1"}, {Key: "checkupId", Value: "2"}},
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when date is invalid format",
			params:     gin.Params{{Key: "id", Value: "1"}, {Key: "checkupId", Value: "2"}},
			body:       map[string]any{"date": "2026/05/28"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			params:   gin.Params{{Key: "id", Value: "1"}, {Key: "checkupId", Value: "2"}},
			body:     map[string]any{"result": "abnormal"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupService{
				updateFn: func(_ context.Context, _, _, _ uint64, _ *UpdateCheckupInput) (*model.Checkup, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = tt.params
			tt.setupCtx(c)

			h.UpdateCheckup(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteCheckup ----

// newDeleteCheckupRouter は c.Status() のみを書く DeleteCheckup 用の router ヘルパー。
// c.Status() は WriteHeaderNow() を伴わないため、gin.Engine.ServeHTTP を経由しないと
// httptest.ResponseRecorder に実際のステータスコードが反映されない。
func newDeleteCheckupRouter(svc CheckupService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithCheckupSvc(svc)
	r.DELETE("/medical-records/:id/checkups/:checkupId", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteCheckup)
	return r
}

func TestDeleteCheckup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		path       string
		svc        *mockCheckupService
		wantStatus int
	}{
		{
			name: "deletes checkup successfully",
			path: "/medical-records/1/checkups/2",
			svc: &mockCheckupService{
				deleteFn: func(_ context.Context, clinicID, medicalRecordID, checkupID uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), medicalRecordID)
					assert.Equal(t, uint64(2), checkupID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 on invalid medical record id",
			path:       "/medical-records/abc/checkups/2",
			svc:        &mockCheckupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid checkup id",
			path:       "/medical-records/1/checkups/abc",
			svc:        &mockCheckupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 500 on service error",
			path: "/medical-records/1/checkups/2",
			svc: &mockCheckupService{
				deleteFn: func(_ context.Context, _, _, _ uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteCheckupRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, tt.path, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithCheckupSvc(&mockCheckupService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}, {Key: "checkupId", Value: "2"}}
		h.DeleteCheckup(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ListGlobalCheckups ----

func TestListGlobalCheckups(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns global checkups with default pagination and real total",
			query:    "pet_id=42&start_date=2026-01-01",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupService{
				listByClinicFn: func(_ context.Context, input ListCheckupsByClinicInput) ([]model.Checkup, int64, error) {
					assert.Equal(t, uint64(1), input.ClinicID)
					require.NotNil(t, input.PetID)
					assert.Equal(t, uint64(42), *input.PetID)
					assert.NotNil(t, input.StartDate)
					assert.Equal(t, 1, input.Page, "デフォルト page=1")
					assert.Equal(t, 20, input.Limit, "デフォルト limit=20")
					return []model.Checkup{{ID: 1, MedicalRecordID: 1, CheckupTypeID: 2, Result: "normal"}}, 11, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"result":"normal"`,
		},
		{
			name:     "passes page/limit query through to service and reflects real total in envelope",
			query:    "page=2&limit=5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupService{
				listByClinicFn: func(_ context.Context, input ListCheckupsByClinicInput) ([]model.Checkup, int64, error) {
					assert.Equal(t, 2, input.Page)
					assert.Equal(t, 5, input.Limit)
					return []model.Checkup{{ID: 6, MedicalRecordID: 1, CheckupTypeID: 2, Result: "page2"}}, 11, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"total":11,"page":2,"limit":5`,
		},
		{
			name:       "returns 400 for invalid pet_id",
			query:      "pet_id=invalid",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid pagination",
			query:      "page=0",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupService{
				listByClinicFn: func(_ context.Context, _ ListCheckupsByClinicInput) ([]model.Checkup, int64, error) {
					return nil, 0, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.ListGlobalCheckups(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
