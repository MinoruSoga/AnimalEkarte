package handler

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

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock TreatmentService ----

type mockTreatmentService struct {
	listFn                func(ctx context.Context, medicalRecordID uint64) ([]model.Treatment, error)
	createFn              func(ctx context.Context, medicalRecordID uint64, input *service.CreateTreatmentInput) (*model.Treatment, error)
	updateFn              func(ctx context.Context, medicalRecordID, treatmentID uint64, input *service.UpdateTreatmentInput) (*model.Treatment, error)
	deleteFn              func(ctx context.Context, medicalRecordID, treatmentID uint64) error
	bulkUpdateSortOrderFn func(ctx context.Context, medicalRecordID uint64, input *service.BulkUpdateTreatmentsInput) error
}

func (m *mockTreatmentService) List(ctx context.Context, medicalRecordID uint64) ([]model.Treatment, error) {
	return m.listFn(ctx, medicalRecordID)
}

func (m *mockTreatmentService) Create(ctx context.Context, medicalRecordID uint64, input *service.CreateTreatmentInput) (*model.Treatment, error) {
	return m.createFn(ctx, medicalRecordID, input)
}

func (m *mockTreatmentService) Update(ctx context.Context, medicalRecordID, treatmentID uint64, input *service.UpdateTreatmentInput) (*model.Treatment, error) {
	return m.updateFn(ctx, medicalRecordID, treatmentID, input)
}

func (m *mockTreatmentService) Delete(ctx context.Context, medicalRecordID, treatmentID uint64) error {
	return m.deleteFn(ctx, medicalRecordID, treatmentID)
}

func (m *mockTreatmentService) BulkUpdateSortOrder(ctx context.Context, medicalRecordID uint64, input *service.BulkUpdateTreatmentsInput) error {
	return m.bulkUpdateSortOrderFn(ctx, medicalRecordID, input)
}

// ---- test helpers ----

func newHandlerWithTreatmentSvc(svc service.TreatmentService) *Handler {
	return &Handler{svc: &service.Services{Treatment: svc}}
}

// ---- ListTreatments ----

func TestListTreatments(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockTreatmentService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns treatments for valid medical_record_id",
			paramID:  "10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTreatmentService{
				listFn: func(_ context.Context, medicalRecordID uint64) ([]model.Treatment, error) {
					assert.Equal(t, uint64(10), medicalRecordID)
					return []model.Treatment{{ID: 1, Content: "血液検査"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"content":"血液検査"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockTreatmentService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockTreatmentService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTreatmentService{
				listFn: func(_ context.Context, _ uint64) ([]model.Treatment, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "returns empty list when no treatments",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTreatmentService{
				listFn: func(_ context.Context, _ uint64) ([]model.Treatment, error) {
					return []model.Treatment{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTreatmentSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.ListTreatments(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateTreatment ----

func TestCreateTreatment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"item_type":  "procedure",
			"unit_price": 5000,
			"quantity":   1.0,
			"content":    "狂犬病ワクチン接種",
		}
	}

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockTreatmentService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates treatment successfully",
			paramID:  "10",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTreatmentService{
				createFn: func(_ context.Context, medicalRecordID uint64, input *service.CreateTreatmentInput) (*model.Treatment, error) {
					assert.Equal(t, uint64(10), medicalRecordID)
					assert.Equal(t, model.TreatmentItemType("procedure"), input.ItemType)
					assert.Equal(t, "狂犬病ワクチン接種", input.Content)
					return &model.Treatment{ID: 1, Content: "狂犬病ワクチン接種"}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"content":"狂犬病ワクチン接種"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockTreatmentService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric medical_record_id",
			paramID:    "abc",
			body:       validBody(),
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockTreatmentService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when item_type is missing",
			paramID:    "1",
			body:       map[string]any{"unit_price": 1000},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockTreatmentService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTreatmentService{
				createFn: func(_ context.Context, _ uint64, _ *service.CreateTreatmentInput) (*model.Treatment, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTreatmentSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.CreateTreatment(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateTreatment ----

func TestUpdateTreatment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		paramID     string
		treatmentID string
		body        any
		setupCtx    func(c *gin.Context)
		svc         *mockTreatmentService
		wantStatus  int
	}{
		{
			name:        "updates treatment successfully",
			paramID:     "10",
			treatmentID: "5",
			body:        map[string]any{"unit_price": 8000, "quantity": 2.0},
			setupCtx:    func(c *gin.Context) { setClinicID(c) },
			svc: &mockTreatmentService{
				updateFn: func(_ context.Context, medicalRecordID, treatmentID uint64, input *service.UpdateTreatmentInput) (*model.Treatment, error) {
					assert.Equal(t, uint64(10), medicalRecordID)
					assert.Equal(t, uint64(5), treatmentID)
					require.NotNil(t, input.UnitPrice)
					assert.Equal(t, int64(8000), *input.UnitPrice)
					return &model.Treatment{ID: 5}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "returns 401 when clinic_id is missing",
			paramID:     "1",
			treatmentID: "1",
			body:        map[string]any{},
			setupCtx:    func(_ *gin.Context) {},
			svc:         &mockTreatmentService{},
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "returns 400 for non-numeric medical_record_id",
			paramID:     "abc",
			treatmentID: "1",
			body:        map[string]any{"unit_price": 100},
			setupCtx:    func(c *gin.Context) { setClinicID(c) },
			svc:         &mockTreatmentService{},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "returns 400 for non-numeric treatment_id",
			paramID:     "1",
			treatmentID: "xyz",
			body:        map[string]any{"unit_price": 100},
			setupCtx:    func(c *gin.Context) { setClinicID(c) },
			svc:         &mockTreatmentService{},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "returns 404 when treatment not found",
			paramID:     "1",
			treatmentID: "999",
			body:        map[string]any{"content": "更新"},
			setupCtx:    func(c *gin.Context) { setClinicID(c) },
			svc: &mockTreatmentService{
				updateFn: func(_ context.Context, _, _ uint64, _ *service.UpdateTreatmentInput) (*model.Treatment, error) {
					return nil, apperrors.WrapNotFound("treatment", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTreatmentSvc(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{
				{Key: "id", Value: tt.paramID},
				{Key: "treatmentId", Value: tt.treatmentID},
			}
			tt.setupCtx(c)
			h.UpdateTreatment(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteTreatment ----

func newDeleteTreatmentRouter(svc service.TreatmentService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithTreatmentSvc(svc)
	r.DELETE("/medical-records/:id/treatments/:treatmentId", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteTreatment)
	return r
}

func TestDeleteTreatment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		paramID     string
		treatmentID string
		svc         *mockTreatmentService
		wantStatus  int
	}{
		{
			name:        "deletes treatment successfully",
			paramID:     "10",
			treatmentID: "5",
			svc: &mockTreatmentService{
				deleteFn: func(_ context.Context, medicalRecordID, treatmentID uint64) error {
					assert.Equal(t, uint64(10), medicalRecordID)
					assert.Equal(t, uint64(5), treatmentID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:        "returns 400 for non-numeric medical_record_id",
			paramID:     "abc",
			treatmentID: "1",
			svc:         &mockTreatmentService{},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "returns 400 for non-numeric treatment_id",
			paramID:     "1",
			treatmentID: "xyz",
			svc:         &mockTreatmentService{},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "returns 404 when treatment not found",
			paramID:     "1",
			treatmentID: "999",
			svc: &mockTreatmentService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("treatment", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteTreatmentRouter(tt.svc)
			w := httptest.NewRecorder()
			url := "/medical-records/" + tt.paramID + "/treatments/" + tt.treatmentID
			req := httptest.NewRequest(http.MethodDelete, url, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithTreatmentSvc(&mockTreatmentService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "treatmentId", Value: "1"},
		}
		h.DeleteTreatment(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- BulkUpdateTreatments ----

func newBulkUpdateRouter(svc service.TreatmentService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithTreatmentSvc(svc)
	r.PUT("/medical-records/:id/treatments", func(c *gin.Context) {
		setClinicID(c)
	}, h.BulkUpdateTreatments)
	return r
}

func TestBulkUpdateTreatments(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"treatments": []map[string]any{
				{"id": 1, "sort_order": 0},
				{"id": 2, "sort_order": 1},
				{"id": 3, "sort_order": 2},
			},
		}
	}

	tests := []struct {
		name       string
		paramID    string
		body       any
		svc        *mockTreatmentService
		wantStatus int
	}{
		{
			name:    "updates sort order successfully",
			paramID: "10",
			body:    validBody(),
			svc: &mockTreatmentService{
				bulkUpdateSortOrderFn: func(_ context.Context, medicalRecordID uint64, input *service.BulkUpdateTreatmentsInput) error {
					assert.Equal(t, uint64(10), medicalRecordID)
					assert.Len(t, input.Treatments, 3)
					assert.Equal(t, uint64(1), input.Treatments[0].ID)
					assert.Equal(t, 2, input.Treatments[2].SortOrder)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 when treatments field is missing",
			paramID:    "1",
			body:       map[string]any{},
			svc:        &mockTreatmentService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 500 on service error",
			paramID: "1",
			body:    validBody(),
			svc: &mockTreatmentService{
				bulkUpdateSortOrderFn: func(_ context.Context, _ uint64, _ *service.BulkUpdateTreatmentsInput) error {
					return fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newBulkUpdateRouter(tt.svc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			url := "/medical-records/" + tt.paramID + "/treatments"
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 400 for non-numeric medical_record_id", func(t *testing.T) {
		h := newHandlerWithTreatmentSvc(&mockTreatmentService{})
		bodyBytes, _ := json.Marshal(validBody())
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "abc"}}
		setClinicID(c)
		h.BulkUpdateTreatments(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithTreatmentSvc(&mockTreatmentService{})
		bodyBytes, _ := json.Marshal(validBody())
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.BulkUpdateTreatments(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
