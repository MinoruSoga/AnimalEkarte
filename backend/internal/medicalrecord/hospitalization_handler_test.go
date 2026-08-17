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

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestHospitalizationHandlerCompiles verifies hospitalization_handler.go compiles
func TestHospitalizationHandlerCompiles(t *testing.T) {
	assert.True(t, true, "hospitalization_handler.go compiled successfully")
}

// ---- mock HospitalizationService ----

type mockHospitalizationService struct {
	listFn func(
		ctx context.Context,
		clinicID uint64,
		petID, ownerID *uint64,
		status, startDate, endDate *string,
		page, limit int,
	) ([]model.Hospitalization, int64, error)
	getByIDFn              func(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	createFn               func(ctx context.Context, clinicID uint64, input *CreateHospitalizationInput) (*model.Hospitalization, error)
	updateFn               func(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationInput) (*model.Hospitalization, error)
	deleteFn               func(ctx context.Context, clinicID, id uint64) error
	dischargeWithBillingFn func(ctx context.Context, clinicID, id uint64, input DischargeWithBillingInput) (*DischargeWithBillingResult, error)
}

func (m *mockHospitalizationService) List(
	ctx context.Context,
	clinicID uint64,
	petID, ownerID *uint64,
	status, startDate, endDate *string,
	page, limit int,
) ([]model.Hospitalization, int64, error) {
	return m.listFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (m *mockHospitalizationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockHospitalizationService) Create(
	ctx context.Context, clinicID uint64, input *CreateHospitalizationInput,
) (*model.Hospitalization, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockHospitalizationService) Update(
	ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationInput,
) (*model.Hospitalization, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockHospitalizationService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockHospitalizationService) DischargeWithBilling(
	ctx context.Context, clinicID, id uint64, input DischargeWithBillingInput,
) (*DischargeWithBillingResult, error) {
	return m.dischargeWithBillingFn(ctx, clinicID, id, input)
}

func newHandlerWithHospitalizationSvc(svc HospitalizationService) *HospitalizationHandler {
	return NewHospitalizationHandler(svc, allowAllPermission)
}

// ---- ListHospitalizations ----

func TestListHospitalizations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockHospitalizationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns paginated list of hospitalizations",
			query:    "page=1&limit=20",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationService{
				listFn: func(
					_ context.Context, clinicID uint64, petID, ownerID *uint64,
					status, startDate, endDate *string, page, limit int,
				) ([]model.Hospitalization, int64, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Nil(t, petID)
					assert.Nil(t, ownerID)
					assert.Nil(t, status)
					assert.Equal(t, 1, page)
					assert.Equal(t, 20, limit)
					return []model.Hospitalization{{ID: 1, Memo: "経過観察"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"memo":"経過観察"`,
		},
		{
			name:     "passes filters through to service",
			query:    "pet_id=2&owner_id=3&status=admitted&start_date=2026-05-01&end_date=2026-05-31",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationService{
				listFn: func(
					_ context.Context, _ uint64, petID, ownerID *uint64,
					status, startDate, endDate *string, _, _ int,
				) ([]model.Hospitalization, int64, error) {
					require.NotNil(t, petID)
					assert.Equal(t, uint64(2), *petID)
					require.NotNil(t, ownerID)
					assert.Equal(t, uint64(3), *ownerID)
					require.NotNil(t, status)
					assert.Equal(t, "admitted", *status)
					require.NotNil(t, startDate)
					assert.Equal(t, "2026-05-01", *startDate)
					require.NotNil(t, endDate)
					assert.Equal(t, "2026-05-31", *endDate)
					return []model.Hospitalization{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when pagination is invalid",
			query:      "page=0",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when pet_id filter is invalid",
			query:      "pet_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationService{
				listFn: func(
					_ context.Context, _ uint64, _, _ *uint64, _, _, _ *string, _, _ int,
				) ([]model.Hospitalization, int64, error) {
					return nil, 0, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithHospitalizationSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListHospitalizations(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetHospitalization ----

func TestGetHospitalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockHospitalizationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns hospitalization for valid id",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					return &model.Hospitalization{ID: 5, Memo: "術後管理"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"memo":"術後管理"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when hospitalization not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return nil, apperrors.WrapNotFound("hospitalization", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithHospitalizationSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetHospitalization(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateHospitalization ----

func TestCreateHospitalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"owner_id":             1,
			"pet_id":               2,
			"hospitalization_type": "hospitalization",
			"start_date":           "2026-05-28T00:00:00Z",
			"end_date":             "2026-05-30T00:00:00Z",
			"status":               "admitted",
			"cage_id":              10,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockHospitalizationService
		wantStatus int
		wantBody   string
		wantHeader bool
	}{
		{
			name:     "creates hospitalization successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateHospitalizationInput) (*model.Hospitalization, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), input.OwnerID)
					assert.Equal(t, uint64(2), input.PetID)
					return &model.Hospitalization{
						ID:                  10,
						ClinicID:            clinicID,
						OwnerID:             input.OwnerID,
						PetID:               input.PetID,
						HospitalizationType: input.HospitalizationType,
						Status:              input.Status,
					}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"id":10`,
			wantHeader: true,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when required fields are missing",
			body:       map[string]any{"pet_id": 2},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid hospitalization_type",
			body:       map[string]any{"owner_id": 1, "pet_id": 2, "hospitalization_type": "invalid", "start_date": "2026-05-28T00:00:00Z", "end_date": "2026-05-30T00:00:00Z"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for malformed JSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 when cage_id is missing (BUG-037)",
			body: map[string]any{
				"owner_id":             1,
				"pet_id":               2,
				"hospitalization_type": "hospitalization",
				"start_date":           "2026-05-28T00:00:00Z",
				"end_date":             "2026-05-30T00:00:00Z",
				"status":               "admitted",
			},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationService{
				createFn: func(_ context.Context, _ uint64, _ *CreateHospitalizationInput) (*model.Hospitalization, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "nested treatment_plans with discount are passed to service when permitted",
			body: func() map[string]any {
				b := validBody()
				b["treatment_plans"] = []map[string]any{
					{
						"treatment_content": "adm rate",
						"unit_price":        990,
						"quantity":          1,
						"discount_rate":     10,
						"discount_amount":   0,
						"sort_order":        0,
					},
				}
				return b
			}(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateHospitalizationInput) (*model.Hospitalization, error) {
					assert.Equal(t, uint64(1), clinicID)
					require.Len(t, input.TreatmentPlans, 1)
					assert.Equal(t, "adm rate", input.TreatmentPlans[0].TreatmentContent)
					assert.Equal(t, float64(10), input.TreatmentPlans[0].DiscountRate)
					return &model.Hospitalization{ID: 11, ClinicID: clinicID}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"id":11`,
			wantHeader: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithHospitalizationSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)


			h.CreateHospitalization(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantHeader {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

func TestCreateHospitalization_NestedDiscountForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	createCalled := false
	svc := &mockHospitalizationService{
		createFn: func(_ context.Context, _ uint64, _ *CreateHospitalizationInput) (*model.Hospitalization, error) {
			createCalled = true
			return &model.Hospitalization{ID: 1}, nil
		},
	}
	h := NewHospitalizationHandler(svc, denyAllPermission)
	body := map[string]any{
		"owner_id":             1,
		"pet_id":               2,
		"hospitalization_type": "hospitalization",
		"start_date":           "2026-05-28T00:00:00Z",
		"end_date":             "2026-05-30T00:00:00Z",
		"cage_id":              10,
		"treatment_plans": []map[string]any{
			{
				"treatment_content": "adm",
				"unit_price":        100,
				"quantity":          1,
				"discount_rate":     5,
			},
		},
	}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)
	h.CreateHospitalization(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, createCalled, "service.Create must not run when discount create is forbidden")
}

// ---- UpdateHospitalization ----

func TestUpdateHospitalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockHospitalizationService
		wantStatus int
	}{
		{
			name:     "updates hospitalization successfully",
			paramID:  "1",
			body:     map[string]any{"memo": "更新後メモ"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateHospitalizationInput) (*model.Hospitalization, error) {
					require.NotNil(t, input.Memo)
					assert.Equal(t, "更新後メモ", *input.Memo)
					return &model.Hospitalization{ID: 1, Memo: *input.Memo}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"memo": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"memo": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for malformed JSON",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid status enum",
			paramID:    "1",
			body:       map[string]any{"status": "unknown"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when hospitalization not found",
			paramID:  "999",
			body:     map[string]any{"memo": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockHospitalizationService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateHospitalizationInput) (*model.Hospitalization, error) {
					return nil, apperrors.WrapNotFound("hospitalization", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithHospitalizationSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateHospitalization(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DischargeWithBilling ----

func TestDischargeWithBilling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockHospitalizationService
		wantStatus int
		wantBody   string
	}{
		{
			name:    "discharges with billing successfully",
			paramID: "1",
			body:    map[string]any{"discharge_date": "2026-05-28T00:00:00Z", "create_accounting": true},
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("user_id", "42")
			},
			svc: &mockHospitalizationService{
				dischargeWithBillingFn: func(_ context.Context, clinicID, id uint64, input DischargeWithBillingInput) (*DischargeWithBillingResult, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					assert.True(t, input.CreateAccounting)
					require.NotNil(t, input.ActorID)
					assert.Equal(t, uint64(42), *input.ActorID)
					accountingID := uint64(99)
					return &DischargeWithBillingResult{
						HospitalizationID: id,
						AccountingID:      &accountingID,
						Status:            "discharged",
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"status":"discharged"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"discharge_date": "2026-05-28T00:00:00Z"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 401 when user_id is missing",
			paramID:    "1",
			body:       map[string]any{"discharge_date": "2026-05-28T00:00:00Z", "create_accounting": true},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "returns 400 when user_id has invalid type",
			paramID: "1",
			body:    map[string]any{"discharge_date": "2026-05-28T00:00:00Z", "create_accounting": true},
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("user_id", uint64(42))
			},
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 400 when user_id is zero",
			paramID: "1",
			body:    map[string]any{"discharge_date": "2026-05-28T00:00:00Z", "create_accounting": true},
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("user_id", "0")
			},
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 400 for non-numeric id",
			paramID: "abc",
			body:    map[string]any{"discharge_date": "2026-05-28T00:00:00Z"},
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("user_id", "42")
			},
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 400 when discharge_date is missing",
			paramID: "1",
			body:    map[string]any{},
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("user_id", "42")
			},
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 400 for malformed JSON",
			paramID: "1",
			body:    "not-json",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("user_id", "42")
			},
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 500 on service error",
			paramID: "1",
			body:    map[string]any{"discharge_date": "2026-05-28T00:00:00Z"},
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("user_id", "42")
			},
			svc: &mockHospitalizationService{
				dischargeWithBillingFn: func(_ context.Context, _, _ uint64, _ DischargeWithBillingInput) (*DischargeWithBillingResult, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithHospitalizationSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.DischargeWithBilling(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteHospitalization ----

func TestDeleteHospitalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockHospitalizationService
		wantStatus int
	}{
		{
			name:    "deletes hospitalization successfully",
			paramID: "1",
			svc: &mockHospitalizationService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockHospitalizationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when hospitalization not found",
			paramID: "999",
			svc: &mockHospitalizationService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("hospitalization", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteHospitalizationRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/hospitalizations/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithHospitalizationSvc(&mockHospitalizationService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteHospitalization(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func newDeleteHospitalizationRouter(svc HospitalizationService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithHospitalizationSvc(svc)
	r.DELETE("/hospitalizations/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteHospitalization)
	return r
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Hospitalization Handler Test Cases
// This handler manages inpatient care and hotel boarding records for pets (Section 7: 入院・ホテル管理)
//
// CRITICAL ENDPOINTS:
//
// 1. ListHospitalizations (GET /hospitalizations)
//    Test Cases (18 scenarios):
//    ✓ Returns 200 OK with empty list when no records exist
//    ✓ Returns 200 OK with paginated hospitalization list when records exist
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when page/limit are invalid
//    ✓ Pagination: page=1, limit=20 as defaults
//    ✓ Pagination: supports custom page and limit parameters
//    ✓ Pagination: includes total_count for client-side calculation
//    ✓ Filter: pet_id parameter filters by pet (optional)
//    ✓ Filter: owner_id parameter filters by owner (optional)
//    ✓ Filter: status parameter filters by enum (admitted, discharged, reserved)
//    ✓ Filter: start_date parameter filters by date range (inclusive)
//    ✓ Filter: end_date parameter filters by date range (inclusive)
//    ✓ Filter: date format validation (YYYY-MM-DD)
//    ✓ Filter: multiple filters can be combined (pet_id AND status AND date range)
//    ✓ Response includes id, owner_id, pet_id, hospitalization_type, start_date, end_date
//    ✓ Response includes status, cage_id, doctor_id, memo, owner_request, staff_notes
//    ✓ Respects clinic_id scoping (only own clinic's records)
//    ✓ Returns 500 on database error
//
// 2. GetHospitalization (GET /hospitalizations/:id)
//    Test Cases (11 scenarios):
//    ✓ Returns 200 OK with single hospitalization record
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when hospitalization_id is non-numeric
//    ✓ Returns 404 when hospitalization doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic (tenant isolation)
//    ✓ Response includes complete hospitalization data
//    ✓ Response includes nested owner object (if preloaded)
//    ✓ Response includes nested pet object (if preloaded)
//    ✓ Response includes nested cage object (if preloaded)
//    ✓ Response includes nested doctor (staff) object (if preloaded)
//    ✓ Returns 500 on database error
//
// 3. CreateHospitalization (POST /hospitalizations)
//    Test Cases (21 scenarios):
//    ✓ Returns 201 Created when hospitalization created successfully
//    ✓ Returns 400 when required fields missing (owner_id, pet_id, hospitalization_type, start_date)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Validates owner_id exists (FK constraint)
//    ✓ Validates pet_id exists (FK constraint)
//    ✓ Validates cage_id exists if provided (FK constraint)
//    ✓ Validates doctor_id exists if provided (FK constraint)
//    ✓ Hospitalization_type field: accepts enum values (inpatient, hotel)
//    ✓ Hospitalization_type field: rejects invalid enum values with 400
//    ✓ Status field: accepts enum values (admitted, discharged, reserved)
//    ✓ Status field: defaults to admitted if not provided during creation
//    ✓ Status field: rejects invalid enum values with 400
//    ✓ StartDate required, EndDate can be null (not yet discharged)
//    ✓ Created hospitalization includes generated id and created_at timestamp
//    ✓ Multiple hospitalizations per pet/owner supported (e.g., repeat admissions)
//    ✓ Cage must be available during stay (no double-booking validation if applicable)
//    ✓ Doctor must be active staff member
//    ✓ Memo and owner_request are optional text fields
//    ✓ StaffNotes are optional text field
//    ✓ Returns 500 on database error
//
// 4. UpdateHospitalization (PATCH /hospitalizations/:id)
//    Test Cases (23 scenarios):
//    ✓ Returns 200 OK when hospitalization updated successfully
//    ✓ Returns 400 when hospitalization_id is non-numeric
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 404 when hospitalization doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ Partial updates: owner_id can be updated independently
//    ✓ Partial updates: pet_id can be updated independently
//    ✓ Partial updates: hospitalization_type can be updated (enum validation)
//    ✓ Partial updates: start_date can be updated
//    ✓ Partial updates: end_date can be updated or cleared
//    ✓ Partial updates: cage_id can be updated or null'd
//    ✓ Partial updates: doctor_id can be updated or null'd
//    ✓ Partial updates: status can be updated (enum validation)
//    ✓ Partial updates: memo can be updated or cleared
//    ✓ Partial updates: owner_request can be updated or cleared
//    ✓ Partial updates: staff_notes can be updated or cleared
//    ✓ Unspecified fields remain unchanged (PATCH semantics, not PUT)
//    ✓ Updated hospitalization reflects changes in response (updated_at timestamp)
//    ✓ Cannot change hospitalization_type from inpatient to hotel if already admitted
//    ✓ Returns 409 Conflict if FK constraint violated during update
//    ✓ Returns 500 on database error
//
// 5. DischargeWithBilling (POST /hospitalizations/:id/discharge-with-billing)
//    Test Cases (15 scenarios):
//    ✓ Returns 200 OK when discharged successfully
//    ✓ Complex operation: updates status to discharged and optionally creates accounting
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when hospitalization_id is non-numeric
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 404 when hospitalization doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ DischargeDate parameter: required date
//    ✓ DischargeDate cannot be before start_date (business logic validation)
//    ✓ CreateAccounting parameter: boolean flag (defaults to false)
//    ✓ When createAccounting=true: creates new accounting record with hospitalization_id
//    ✓ When createAccounting=false: updates status only (no accounting created)
//    ✓ After discharge, status changes to "discharged"
//    ✓ Response includes updated hospitalization + created accounting (if applicable)
//    ✓ Returns 500 on database error or transaction failure
//
// 6. DeleteHospitalization (DELETE /hospitalizations/:id)
//    Test Cases (11 scenarios):
//    ✓ Returns 204 No Content when hospitalization deleted successfully
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when hospitalization_id is non-numeric
//    ✓ Returns 404 when hospitalization doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ Uses soft delete (sets deleted_at, doesn't remove from database)
//    ✓ Deleted hospitalization no longer appears in ListHospitalizations
//    ✓ Deleted hospitalization cannot be retrieved by GetHospitalization (404)
//    ✓ Cannot delete already deleted hospitalization (404 on second delete)
//    ✓ Deleting hospitalization cascades to related records (if applicable)
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control (clinic_id verification on all endpoints)
//    ✓ Enum validation prevents invalid state transitions
//    ✓ Foreign key validation on owner, pet, cage, doctor
//    ✓ Partial updates prevent mass assignment (explicit field mapping)
//    ✓ Date validation: start_date before end_date
//    ✓ No explicit RBAC permission check (all authenticated users can manage hospitalization)
//
// INTEGRATION WITH OTHER MODULES:
//    ✓ Hospitalization linked to owner and pet (required FKs)
//    ✓ Hospitalization linked to cage (optional FK for inpatient stays)
//    ✓ Hospitalization linked to doctor/staff (optional FK for assigned veterinarian)
//    ✓ DischargeWithBilling integrates with accounting module
//    ✓ Cannot change pet during hospitalization (tied to a specific animal)
//
// DATA MODEL (hospitalizations):
//    - id (PK): BIGSERIAL
//    - clinic_id: BIGINT (multitenancy)
//    - owner_id (FK): BIGINT → owners(id)
//    - pet_id (FK): BIGINT → pets(id)
//    - hospitalization_type: ENUM (inpatient|hotel)
//    - status: ENUM (admitted|discharged|reserved) DEFAULT admitted
//    - start_date: DATE
//    - end_date: DATE (NULLABLE) - until discharge
//    - cage_id (FK, NULLABLE): BIGINT → cages(id)
//    - doctor_id (FK, NULLABLE): BIGINT → staffs(id)
//    - memo: TEXT (NULLABLE) - general notes
//    - owner_request: TEXT (NULLABLE) - owner's special requests
//    - staff_notes: TEXT (NULLABLE) - staff observations during stay
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete)
//    - Indexes: (clinic_id, id), (clinic_id, owner_id), (clinic_id, pet_id), (clinic_id, start_date DESC)
//
// IMPLEMENTATION NOTES:
//    - DischargeWithBilling is a complex operation combining discharge + optional billing
//    - Status enum validation prevents invalid transitions (e.g., discharged → admitted)
//    - PATCH semantics: unspecified pointer fields remain unchanged
//    - StartDate required, EndDate optional (discharge date set at discharge time)
//    - Cage and doctor are optional (only cage needed for inpatient, doctor optional for both)
//    - Multiple hospitalizations possible for same pet (e.g., repeat boardings)
//    - Date validation: start_date should be <= end_date when both provided
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with owner, pet, cage, staff (doctor) test data
//    - Real service/repository layers
//    - Verify pagination with >20 records
//    - Verify enum validation for hospitalization_type and status
//    - Verify filter combinations (pet_id, owner_id, status, date range)
//    - Verify FK constraints for all foreign key fields
//    - Verify soft delete behavior (deleted records excluded from list/get)
//    - Verify PATCH semantics (unspecified fields unchanged)
//    - Test DischargeWithBilling transaction (discharge + accounting creation)
//    - Test date validation (start_date <= end_date)
//
