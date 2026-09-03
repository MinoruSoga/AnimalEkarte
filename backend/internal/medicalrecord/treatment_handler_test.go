package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock TreatmentService ----

type mockTreatmentService struct {
	listFn                func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error)
	getByIDFn             func(ctx context.Context, clinicID, id uint64) (*model.Treatment, error)
	createFn              func(ctx context.Context, clinicID, medicalRecordID uint64, input *CreateTreatmentInput) (*model.Treatment, error)
	updateFn              func(ctx context.Context, clinicID, medicalRecordID, treatmentID uint64, input *UpdateTreatmentInput) (*model.Treatment, error)
	deleteFn              func(ctx context.Context, clinicID, medicalRecordID, treatmentID uint64) error
	bulkUpdateSortOrderFn func(ctx context.Context, clinicID, medicalRecordID uint64, input *BulkUpdateTreatmentsInput) error
	listPetHistoryFn      func(ctx context.Context, clinicID, petID uint64, filter model.PetTreatmentHistoryFilter, page, limit int) ([]model.Treatment, int64, error)
}

func (m *mockTreatmentService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error) {
	return m.listFn(ctx, clinicID, medicalRecordID)
}

func (m *mockTreatmentService) ListPetHistory(ctx context.Context, clinicID, petID uint64, filter model.PetTreatmentHistoryFilter, page, limit int) ([]model.Treatment, int64, error) {
	if m.listPetHistoryFn != nil {
		return m.listPetHistoryFn(ctx, clinicID, petID, filter, page, limit)
	}
	return nil, 0, nil
}

func (m *mockTreatmentService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Treatment, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, id)
	}
	return &model.Treatment{ID: id}, nil
}

func (m *mockTreatmentService) Create(ctx context.Context, clinicID, medicalRecordID uint64, input *CreateTreatmentInput) (*model.Treatment, error) {
	return m.createFn(ctx, clinicID, medicalRecordID, input)
}

func (m *mockTreatmentService) Update(ctx context.Context, clinicID, medicalRecordID, treatmentID uint64, input *UpdateTreatmentInput) (*model.Treatment, error) {
	return m.updateFn(ctx, clinicID, medicalRecordID, treatmentID, input)
}

func (m *mockTreatmentService) Delete(ctx context.Context, clinicID, medicalRecordID, treatmentID uint64) error {
	return m.deleteFn(ctx, clinicID, medicalRecordID, treatmentID)
}

func (m *mockTreatmentService) BulkUpdateSortOrder(ctx context.Context, clinicID, medicalRecordID uint64, input *BulkUpdateTreatmentsInput) error {
	return m.bulkUpdateSortOrderFn(ctx, clinicID, medicalRecordID, input)
}

// ---- test helpers ----

// newHandlerWithTreatmentSvc は TreatmentHandler を allow-all の PermissionChecker で構築する
// （BUG-372 discount ガードの 403 分岐は権限 stub の責務・本 package では pass-through のみ検証）。
func newHandlerWithTreatmentSvc(svc TreatmentService) *TreatmentHandler {
	return NewTreatmentHandler(svc, func(_ *gin.Context, _, _ string) bool { return true })
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
				listFn: func(_ context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error) {
					assert.Equal(t, uint64(1), clinicID)
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
				listFn: func(_ context.Context, _, _ uint64) ([]model.Treatment, error) {
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
				listFn: func(_ context.Context, _, _ uint64) ([]model.Treatment, error) {
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

// ---- ListPetTreatmentHistory (#158 飼主レポート) ----

func TestListPetTreatmentHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	medicineID := uint64(3)
	historyRow := func() []model.Treatment {
		return []model.Treatment{{
			ID:              5,
			MedicalRecordID: 9,
			ItemType:        model.TreatmentItemTypeMedicine,
			Content:         "投薬",
			MedicineID:      &medicineID,
			MedicalRecord:   &model.MedicalRecord{ID: 9, Date: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
			Medicine:        &model.Medicine{ID: medicineID, Name: "アモキシシリン"},
		}}
	}

	tests := []struct {
		name       string
		paramID    string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockTreatmentService
		wantStatus int
		wantBody   []string
	}{
		{
			name:     "returns paginated history with medical_records.date and medicine name",
			paramID:  "7",
			query:    "?item_type=medicine&limit=100",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTreatmentService{
				listPetHistoryFn: func(_ context.Context, clinicID, petID uint64, f model.PetTreatmentHistoryFilter, _, _ int) ([]model.Treatment, int64, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(7), petID)
					if assert.NotNil(t, f.ItemType) {
						assert.Equal(t, model.TreatmentItemTypeMedicine, *f.ItemType)
					}
					return historyRow(), 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   []string{`"medicine_name":"アモキシシリン"`, `"2026-06-01`, `"total":1`, `"item_type":"medicine"`},
		},
		{
			name:     "item_type=all passes nil filter",
			paramID:  "7",
			query:    "?item_type=all",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTreatmentService{
				listPetHistoryFn: func(_ context.Context, _, _ uint64, f model.PetTreatmentHistoryFilter, _, _ int) ([]model.Treatment, int64, error) {
					assert.Nil(t, f.ItemType)
					return []model.Treatment{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   []string{`"data":[]`, `"total":0`},
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "7",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockTreatmentService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric pet id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockTreatmentService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid item_type",
			paramID:    "7",
			query:      "?item_type=bogus",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockTreatmentService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "7",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockTreatmentService{
				listPetHistoryFn: func(_ context.Context, _, _ uint64, _ model.PetTreatmentHistoryFilter, _, _ int) ([]model.Treatment, int64, error) {
					return nil, 0, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTreatmentSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/"+tt.query, http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.ListPetTreatmentHistory(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			for _, want := range tt.wantBody {
				assert.Contains(t, w.Body.String(), want)
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
				createFn: func(_ context.Context, clinicID, medicalRecordID uint64, input *CreateTreatmentInput) (*model.Treatment, error) {
					assert.Equal(t, uint64(1), clinicID)
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
				createFn: func(_ context.Context, _, _ uint64, _ *CreateTreatmentInput) (*model.Treatment, error) {
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
				updateFn: func(_ context.Context, clinicID, medicalRecordID, treatmentID uint64, input *UpdateTreatmentInput) (*model.Treatment, error) {
					assert.Equal(t, uint64(1), clinicID)
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
			// SEC-CS-F09: handler must pass discount:edit into service for TX recheck.
			name:        "passes DiscountEditAllowed=true when discount:edit granted",
			paramID:     "10",
			treatmentID: "5",
			body:        map[string]any{"memo": "m"},
			setupCtx:    func(c *gin.Context) { setClinicID(c) },
			svc: &mockTreatmentService{
				updateFn: func(_ context.Context, _, _, _ uint64, input *UpdateTreatmentInput) (*model.Treatment, error) {
					assert.True(t, input.DiscountEditAllowed)
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
				updateFn: func(_ context.Context, _, _, _ uint64, _ *UpdateTreatmentInput) (*model.Treatment, error) {
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

func newDeleteTreatmentRouter(svc TreatmentService) *gin.Engine {
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
				deleteFn: func(_ context.Context, clinicID, medicalRecordID, treatmentID uint64) error {
					assert.Equal(t, uint64(1), clinicID)
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
				deleteFn: func(_ context.Context, _, _, _ uint64) error {
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

func newBulkUpdateRouter(svc TreatmentService) *gin.Engine {
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
				bulkUpdateSortOrderFn: func(_ context.Context, _, medicalRecordID uint64, input *BulkUpdateTreatmentsInput) error {
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
				bulkUpdateSortOrderFn: func(_ context.Context, _, _ uint64, _ *BulkUpdateTreatmentsInput) error {
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

// SEC-CODEX-UHQPM2 selected-clinic grant
func TestTreatmentSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		invoke func(*TreatmentHandler, *gin.Context)
		svc    *mockTreatmentService
	}{
		{
			name: "ListTreatments returns 403 when selected clinic lacks medical record view grant",
			invoke: func(h *TreatmentHandler, c *gin.Context) {
				h.ListTreatments(c)
			},
			svc: &mockTreatmentService{
				listFn: func(_ context.Context, _, _ uint64) ([]model.Treatment, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
		{
			name: "ListPetTreatmentHistory returns 403 when selected clinic lacks medical record view grant",
			invoke: func(h *TreatmentHandler, c *gin.Context) {
				h.ListPetTreatmentHistory(c)
			},
			svc: &mockTreatmentService{
				listPetHistoryFn: func(_ context.Context, _, _ uint64, _ model.PetTreatmentHistoryFilter, _, _ int) ([]model.Treatment, int64, error) {
					t.Fatal("service must not be reached")
					return nil, 0, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTreatmentSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			setClinicID(c)
			c.Set("clinic_id", "2")
			c.Set("is_system_admin", false)
			setResourcePermissionOnlyClinic(c, 1, string(model.ResourceMedicalRecords), "view")
			tt.invoke(h, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}
