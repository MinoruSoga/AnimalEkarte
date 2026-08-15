package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestTreatmentPlanHandlerCompiles verifies treatment_plan_handler.go compiles
func TestTreatmentPlanHandlerCompiles(t *testing.T) {
	assert.True(t, true, "treatment_plan_handler.go compiled successfully")
}

// ---- mock TreatmentPlanService ----

type mockTreatmentPlanService struct {
	listByMedicalRecordFn   func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error)
	listByHospitalizationFn func(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.TreatmentPlan, error)
	getByIDFn               func(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error)
	createFn                func(ctx context.Context, clinicID uint64, medicalRecordID, hospitalizationID *uint64, input *CreateTreatmentPlanInput) (*model.TreatmentPlan, error)
	updateFn                func(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64, input *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error)
	deleteFn                func(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64) error
}

func (m *mockTreatmentPlanService) ListByMedicalRecord(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error) {
	return m.listByMedicalRecordFn(ctx, clinicID, medicalRecordID)
}

func (m *mockTreatmentPlanService) ListByHospitalization(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.TreatmentPlan, error) {
	return m.listByHospitalizationFn(ctx, clinicID, hospitalizationID)
}

func (m *mockTreatmentPlanService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, id)
	}
	return &model.TreatmentPlan{ID: id, ClinicID: clinicID}, nil
}

func (m *mockTreatmentPlanService) Create(ctx context.Context, clinicID uint64, medicalRecordID, hospitalizationID *uint64, input *CreateTreatmentPlanInput) (*model.TreatmentPlan, error) {
	return m.createFn(ctx, clinicID, medicalRecordID, hospitalizationID, input)
}

func (m *mockTreatmentPlanService) Update(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64, input *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error) {
	return m.updateFn(ctx, clinicID, id, medicalRecordID, hospitalizationID, input)
}

func (m *mockTreatmentPlanService) Delete(ctx context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64) error {
	return m.deleteFn(ctx, clinicID, id, medicalRecordID, hospitalizationID)
}

// ---- test helpers ----
// (mockHospitalizationService is already defined in hospitalization_handler_test.go and is
// reused here; note that its getByIDFn field has no nil-safe default, so every test case in
// this file that reaches Hospitalization.GetByID must set it explicitly.)

// BE9-2D ⑥: 旧 harness の Services+EffectivePermission 注入を、TreatmentPlanHandler の
// PermissionChecker（bool 判定）注入へ変換。旧 mockEffectivePermissionService のデフォルト
// （rules なし=deny）は denyAllPermission が、discount:create 付与ケースは
// grantDiscountCreatePermission が等価に写像する。
func newHandlerWithTreatmentPlanSvc(tpSvc TreatmentPlanService, hospSvc HospitalizationService, mrSvc medicalRecordGetter, check PermissionChecker) *TreatmentPlanHandler {
	return NewTreatmentPlanHandler(tpSvc, hospSvc, mrSvc, check)
}

// setNonSystemAdmin は internal/handler/clinic_handler_test.go の同名ヘルパーの最小複製
// （⑥移動・原本は旧 package の残存テストが使用）。
func setNonSystemAdmin(c *gin.Context) {
	c.Set("is_system_admin", false)
	c.Set("user_id", "1")
	c.Set("clinic_id", "1")
}

// grantDiscountCreatePermission は discount:create のみ許可する PermissionChecker。
func grantDiscountCreatePermission(_ *gin.Context, resource, action string) bool {
	return resource == "discount" && action == "create"
}

// ---- ListTreatmentPlansByMedicalRecord ----

func TestListTreatmentPlansByMedicalRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		mrSvc      *mockMedicalRecordService
		tpSvc      *mockTreatmentPlanService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of treatment plans",
			paramID:  "7",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(7), id)
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				listByMedicalRecordFn: func(_ context.Context, _, _ uint64) ([]model.TreatmentPlan, error) {
					return []model.TreatmentPlan{{ID: 1, TreatmentContent: "点滴"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"treatment_content":"点滴"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "7",
			setupCtx:   func(_ *gin.Context) {},
			mrSvc:      &mockMedicalRecordService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric medical record id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			mrSvc:      &mockMedicalRecordService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when medical record not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "999")
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "7",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				listByMedicalRecordFn: func(_ context.Context, _, _ uint64) ([]model.TreatmentPlan, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTreatmentPlanSvc(tt.tpSvc, &mockHospitalizationService{}, tt.mrSvc, denyAllPermission)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.ListTreatmentPlansByMedicalRecord(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateTreatmentPlanForMedicalRecord ----

func TestCreateTreatmentPlanForMedicalRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := map[string]any{
		"treatment_content": "点滴",
		"unit_price":        1000,
		"quantity":          1,
	}

	tests := []struct {
		name       string
		paramID    string
		body       any
		malformed  bool
		setupCtx   func(c *gin.Context)
		mrSvc      *mockMedicalRecordService
		tpSvc      *mockTreatmentPlanService
		effPerm    PermissionChecker
		wantStatus int
		wantHeader string
	}{
		{
			name:     "creates plan with valid body",
			paramID:  "7",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				createFn: func(_ context.Context, clinicID uint64, medicalRecordID, hospitalizationID *uint64, input *CreateTreatmentPlanInput) (*model.TreatmentPlan, error) {
					assert.Equal(t, uint64(1), clinicID)
					require.NotNil(t, medicalRecordID)
					assert.Equal(t, uint64(7), *medicalRecordID)
					assert.Nil(t, hospitalizationID)
					assert.Equal(t, "点滴", input.TreatmentContent)
					return &model.TreatmentPlan{ID: 1, TreatmentContent: input.TreatmentContent}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: "/api/v1/medical-records/7/treatment-plans/1",
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "7",
			body:       validBody,
			setupCtx:   func(_ *gin.Context) {},
			mrSvc:      &mockMedicalRecordService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric medical record id",
			paramID:    "abc",
			body:       validBody,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			mrSvc:      &mockMedicalRecordService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when medical record not found",
			paramID:  "999",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "999")
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "returns 400 when body is malformed",
			paramID:   "7",
			malformed: true,
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 400 when required field missing",
			paramID:  "7",
			body:     map[string]any{"unit_price": 1000},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 403 when discount rate specified without permission",
			paramID: "7",
			body: map[string]any{
				"treatment_content": "点滴",
				"quantity":          1.0,
				"discount_rate":     10.0,
			},
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			effPerm:    denyAllPermission,
			wantStatus: http.StatusForbidden,
		},
		{
			name:    "returns 403 when discount amount specified without permission",
			paramID: "7",
			body: map[string]any{
				"treatment_content": "点滴",
				"quantity":          1.0,
				"discount_amount":   500,
			},
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			effPerm:    denyAllPermission,
			wantStatus: http.StatusForbidden,
		},
		{
			name:    "allows discount when permission granted",
			paramID: "7",
			body: map[string]any{
				"treatment_content": "点滴",
				"quantity":          1.0,
				"discount_rate":     10.0,
			},
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				createFn: func(_ context.Context, _ uint64, _, _ *uint64, input *CreateTreatmentPlanInput) (*model.TreatmentPlan, error) {
					assert.InDelta(t, 10.0, input.DiscountRate, 0.0001)
					return &model.TreatmentPlan{ID: 2}, nil
				},
			},
			effPerm:    grantDiscountCreatePermission,
			wantStatus: http.StatusCreated,
			wantHeader: "/api/v1/medical-records/7/treatment-plans/2",
		},
		{
			name:     "returns 500 on service error",
			paramID:  "7",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				createFn: func(_ context.Context, _ uint64, _, _ *uint64, _ *CreateTreatmentPlanInput) (*model.TreatmentPlan, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effPerm := tt.effPerm
			if effPerm == nil {
				effPerm = denyAllPermission
			}
			h := newHandlerWithTreatmentPlanSvc(tt.tpSvc, &mockHospitalizationService{}, tt.mrSvc, effPerm)

			bodyBytes := []byte("{invalid")
			if !tt.malformed {
				b, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyBytes = b
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.CreateTreatmentPlanForMedicalRecord(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantHeader != "" {
				assert.Equal(t, tt.wantHeader, w.Header().Get("Location"))
			}
		})
	}
}

// ---- ListTreatmentPlansByHospitalization ----

func TestListTreatmentPlansByHospitalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		hospSvc    *mockHospitalizationService
		tpSvc      *mockTreatmentPlanService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of treatment plans",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), id)
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				listByHospitalizationFn: func(_ context.Context, _, _ uint64) ([]model.TreatmentPlan, error) {
					return []model.TreatmentPlan{{ID: 1, TreatmentContent: "点滴"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"treatment_content":"点滴"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "5",
			setupCtx:   func(_ *gin.Context) {},
			hospSvc:    &mockHospitalizationService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric hospitalization id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			hospSvc:    &mockHospitalizationService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when hospitalization not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return nil, apperrors.WrapNotFound("hospitalization", "999")
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				listByHospitalizationFn: func(_ context.Context, _, _ uint64) ([]model.TreatmentPlan, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTreatmentPlanSvc(tt.tpSvc, tt.hospSvc, &mockMedicalRecordService{}, denyAllPermission)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.ListTreatmentPlansByHospitalization(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateTreatmentPlanForHospitalization ----

func TestCreateTreatmentPlanForHospitalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := map[string]any{
		"treatment_content": "点滴",
		"quantity":          1.0,
	}

	tests := []struct {
		name       string
		paramID    string
		body       any
		malformed  bool
		setupCtx   func(c *gin.Context)
		hospSvc    *mockHospitalizationService
		tpSvc      *mockTreatmentPlanService
		effPerm    PermissionChecker
		wantStatus int
		wantHeader string
	}{
		{
			name:     "creates plan with valid body",
			paramID:  "5",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				createFn: func(_ context.Context, _ uint64, medicalRecordID, hospitalizationID *uint64, _ *CreateTreatmentPlanInput) (*model.TreatmentPlan, error) {
					assert.Nil(t, medicalRecordID)
					require.NotNil(t, hospitalizationID)
					assert.Equal(t, uint64(5), *hospitalizationID)
					return &model.TreatmentPlan{ID: 3}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: "/api/v1/hospitalizations/5/treatment-plans/3",
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "5",
			body:       validBody,
			setupCtx:   func(_ *gin.Context) {},
			hospSvc:    &mockHospitalizationService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric hospitalization id",
			paramID:    "abc",
			body:       validBody,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			hospSvc:    &mockHospitalizationService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when hospitalization not found",
			paramID:  "999",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return nil, apperrors.WrapNotFound("hospitalization", "999")
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "returns 400 when body is malformed",
			paramID:   "5",
			malformed: true,
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 400 when required field missing",
			paramID:  "5",
			body:     map[string]any{"unit_price": 1000},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 403 when discount rate specified without permission",
			paramID: "5",
			body: map[string]any{
				"treatment_content": "点滴",
				"quantity":          1.0,
				"discount_rate":     15.0,
			},
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			effPerm:    denyAllPermission,
			wantStatus: http.StatusForbidden,
		},
		{
			name:    "returns 403 when discount amount specified without permission",
			paramID: "5",
			body: map[string]any{
				"treatment_content": "点滴",
				"quantity":          1.0,
				"discount_amount":   300,
			},
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			effPerm:    denyAllPermission,
			wantStatus: http.StatusForbidden,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				createFn: func(_ context.Context, _ uint64, _, _ *uint64, _ *CreateTreatmentPlanInput) (*model.TreatmentPlan, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effPerm := tt.effPerm
			if effPerm == nil {
				effPerm = denyAllPermission
			}
			h := newHandlerWithTreatmentPlanSvc(tt.tpSvc, tt.hospSvc, &mockMedicalRecordService{}, effPerm)

			bodyBytes := []byte("{invalid")
			if !tt.malformed {
				b, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyBytes = b
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.CreateTreatmentPlanForHospitalization(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantHeader != "" {
				assert.Equal(t, tt.wantHeader, w.Header().Get("Location"))
			}
		})
	}
}

// ---- UpdateTreatmentPlanInMedicalRecord (covers checkTreatmentPlanDiscountPermission) ----

func TestUpdateTreatmentPlanInMedicalRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		mrID       string
		planID     string
		body       any
		malformed  bool
		setupCtx   func(c *gin.Context)
		mrSvc      *mockMedicalRecordService
		tpSvc      *mockTreatmentPlanService
		effPerm    PermissionChecker
		wantStatus int
	}{
		{
			name:     "updates plan successfully",
			mrID:     "7",
			planID:   "1",
			body:     map[string]any{"memo": "経過良好"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				updateFn: func(_ context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64, input *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, input.Memo)
					assert.Equal(t, "経過良好", *input.Memo)
					return &model.TreatmentPlan{ID: 1, Memo: "経過良好"}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			mrID:       "7",
			planID:     "1",
			body:       map[string]any{"memo": "x"},
			setupCtx:   func(_ *gin.Context) {},
			mrSvc:      &mockMedicalRecordService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric medical record id",
			mrID:       "abc",
			planID:     "1",
			body:       map[string]any{"memo": "x"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			mrSvc:      &mockMedicalRecordService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when medical record not found",
			mrID:     "999",
			planID:   "1",
			body:     map[string]any{"memo": "x"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "999")
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 for non-numeric plan id",
			mrID:     "7",
			planID:   "abc",
			body:     map[string]any{"memo": "x"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "returns 400 when body is malformed",
			mrID:      "7",
			planID:    "1",
			malformed: true,
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns error from GetByID when checking discount permission",
			mrID:     "7",
			planID:   "1",
			body:     map[string]any{"discount_rate": 5.0},
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.TreatmentPlan, error) {
					return nil, apperrors.WrapNotFound("treatment_plan", "1")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 403 when discount rate changed without permission",
			mrID:     "7",
			planID:   "1",
			body:     map[string]any{"discount_rate": 5.0},
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.TreatmentPlan, error) {
					return &model.TreatmentPlan{ID: 1, DiscountRate: 0}, nil
				},
			},
			effPerm:    denyAllPermission,
			wantStatus: http.StatusForbidden,
		},
		{
			name:     "returns 403 when discount amount changed without permission",
			mrID:     "7",
			planID:   "1",
			body:     map[string]any{"discount_amount": 200},
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.TreatmentPlan, error) {
					return &model.TreatmentPlan{ID: 1, DiscountAmount: 0}, nil
				},
			},
			effPerm:    denyAllPermission,
			wantStatus: http.StatusForbidden,
		},
		{
			name:     "allows update when discount rate unchanged",
			mrID:     "7",
			planID:   "1",
			body:     map[string]any{"discount_rate": 5.0},
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.TreatmentPlan, error) {
					return &model.TreatmentPlan{ID: 1, DiscountRate: 5.0}, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, _, _ *uint64, _ *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error) {
					return &model.TreatmentPlan{ID: 1, DiscountRate: 5.0}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "returns 500 on service error",
			mrID:     "7",
			planID:   "1",
			body:     map[string]any{"memo": "x"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				updateFn: func(_ context.Context, _, _ uint64, _, _ *uint64, _ *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effPerm := tt.effPerm
			if effPerm == nil {
				effPerm = denyAllPermission
			}
			h := newHandlerWithTreatmentPlanSvc(tt.tpSvc, &mockHospitalizationService{}, tt.mrSvc, effPerm)

			bodyBytes := []byte("{invalid")
			if !tt.malformed {
				b, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyBytes = b
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.mrID}, {Key: "planId", Value: tt.planID}}
			tt.setupCtx(c)

			h.UpdateTreatmentPlanInMedicalRecord(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteTreatmentPlanInMedicalRecord ----

func TestDeleteTreatmentPlanInMedicalRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		mrID       string
		planID     string
		setupCtx   func(c *gin.Context)
		mrSvc      *mockMedicalRecordService
		tpSvc      *mockTreatmentPlanService
		wantStatus int
	}{
		{
			name:     "deletes plan successfully",
			mrID:     "7",
			planID:   "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				deleteFn: func(_ context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			mrID:       "7",
			planID:     "1",
			setupCtx:   func(_ *gin.Context) {},
			mrSvc:      &mockMedicalRecordService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric medical record id",
			mrID:       "abc",
			planID:     "1",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			mrSvc:      &mockMedicalRecordService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when medical record not found",
			mrID:     "999",
			planID:   "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "999")
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 for non-numeric plan id",
			mrID:     "7",
			planID:   "abc",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			mrID:     "7",
			planID:   "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 7, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				deleteFn: func(_ context.Context, _, _ uint64, _, _ *uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTreatmentPlanSvc(tt.tpSvc, &mockHospitalizationService{}, tt.mrSvc, denyAllPermission)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.mrID}, {Key: "planId", Value: tt.planID}}
			tt.setupCtx(c)
			h.DeleteTreatmentPlanInMedicalRecord(c)
			c.Writer.WriteHeaderNow() // flush a bare c.Status() (no body) to the recorder
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- UpdateTreatmentPlanInHospitalization ----

func TestUpdateTreatmentPlanInHospitalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		hospID     string
		planID     string
		body       any
		malformed  bool
		setupCtx   func(c *gin.Context)
		hospSvc    *mockHospitalizationService
		tpSvc      *mockTreatmentPlanService
		effPerm    PermissionChecker
		wantStatus int
	}{
		{
			// W-002: hospitalization-nested plans are create-time snapshots — service returns Conflict.
			name:     "returns 409 when service rejects hospitalization-nested update",
			hospID:   "5",
			planID:   "1",
			body:     map[string]any{"memo": "経過良好"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				updateFn: func(_ context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64, _ *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					assert.Nil(t, medicalRecordID)
					require.NotNil(t, hospitalizationID)
					assert.Equal(t, uint64(5), *hospitalizationID)
					return nil, apperrors.WrapConflict("入院に紐づく治療プランは登録時スナップショットのため変更・削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			hospID:     "5",
			planID:     "1",
			body:       map[string]any{"memo": "x"},
			setupCtx:   func(_ *gin.Context) {},
			hospSvc:    &mockHospitalizationService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric hospitalization id",
			hospID:     "abc",
			planID:     "1",
			body:       map[string]any{"memo": "x"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			hospSvc:    &mockHospitalizationService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when hospitalization not found",
			hospID:   "999",
			planID:   "1",
			body:     map[string]any{"memo": "x"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return nil, apperrors.WrapNotFound("hospitalization", "999")
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 for non-numeric plan id",
			hospID:   "5",
			planID:   "abc",
			body:     map[string]any{"memo": "x"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "returns 400 when body is malformed",
			hospID:    "5",
			planID:    "1",
			malformed: true,
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 403 when discount rate changed without permission",
			hospID:   "5",
			planID:   "1",
			body:     map[string]any{"discount_rate": 8.0},
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.TreatmentPlan, error) {
					return &model.TreatmentPlan{ID: 1, DiscountRate: 0}, nil
				},
			},
			effPerm:    denyAllPermission,
			wantStatus: http.StatusForbidden,
		},
		{
			name:     "returns 500 on service error",
			hospID:   "5",
			planID:   "1",
			body:     map[string]any{"memo": "x"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				updateFn: func(_ context.Context, _, _ uint64, _, _ *uint64, _ *UpdateTreatmentPlanInput) (*model.TreatmentPlan, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effPerm := tt.effPerm
			if effPerm == nil {
				effPerm = denyAllPermission
			}
			h := newHandlerWithTreatmentPlanSvc(tt.tpSvc, tt.hospSvc, &mockMedicalRecordService{}, effPerm)

			bodyBytes := []byte("{invalid")
			if !tt.malformed {
				b, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyBytes = b
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.hospID}, {Key: "planId", Value: tt.planID}}
			tt.setupCtx(c)

			h.UpdateTreatmentPlanInHospitalization(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteTreatmentPlanInHospitalization ----

func TestDeleteTreatmentPlanInHospitalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		hospID     string
		planID     string
		setupCtx   func(c *gin.Context)
		hospSvc    *mockHospitalizationService
		tpSvc      *mockTreatmentPlanService
		wantStatus int
	}{
		{
			// W-002: hospitalization-nested plans are create-time snapshots — service returns Conflict.
			name:     "returns 409 when service rejects hospitalization-nested delete",
			hospID:   "5",
			planID:   "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				deleteFn: func(_ context.Context, clinicID, id uint64, medicalRecordID, hospitalizationID *uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					assert.Nil(t, medicalRecordID)
					require.NotNil(t, hospitalizationID)
					assert.Equal(t, uint64(5), *hospitalizationID)
					return apperrors.WrapConflict("入院に紐づく治療プランは登録時スナップショットのため変更・削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			hospID:     "5",
			planID:     "1",
			setupCtx:   func(_ *gin.Context) {},
			hospSvc:    &mockHospitalizationService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric hospitalization id",
			hospID:     "abc",
			planID:     "1",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			hospSvc:    &mockHospitalizationService{},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when hospitalization not found",
			hospID:   "999",
			planID:   "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return nil, apperrors.WrapNotFound("hospitalization", "999")
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 for non-numeric plan id",
			hospID:   "5",
			planID:   "abc",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc:      &mockTreatmentPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			hospID:   "5",
			planID:   "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			hospSvc: &mockHospitalizationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{ID: 5, ClinicID: 1}, nil
				},
			},
			tpSvc: &mockTreatmentPlanService{
				deleteFn: func(_ context.Context, _, _ uint64, _, _ *uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithTreatmentPlanSvc(tt.tpSvc, tt.hospSvc, &mockMedicalRecordService{}, denyAllPermission)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.hospID}, {Key: "planId", Value: tt.planID}}
			tt.setupCtx(c)
			h.DeleteTreatmentPlanInHospitalization(c)
			c.Writer.WriteHeaderNow() // flush a bare c.Status() (no body) to the recorder
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- toTreatmentPlanResponse (pure conversion) ----

func TestToTreatmentPlanResponse(t *testing.T) {
	now := time.Now()
	mrID := uint64(7)
	hospID := uint64(5)

	tests := []struct {
		name string
		plan *model.TreatmentPlan
	}{
		{
			name: "maps all fields with medical record id",
			plan: &model.TreatmentPlan{
				ID:               1,
				MedicalRecordID:  &mrID,
				TreatmentContent: "点滴",
				Memo:             "経過良好",
				IsInsurance:      true,
				UnitPrice:        1000,
				Quantity:         2,
				DiscountRate:     10,
				DiscountAmount:   100,
				Subtotal:         1800,
				SortOrder:        1,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
		{
			name: "maps hospitalization id and nil medical record id",
			plan: &model.TreatmentPlan{
				ID:                2,
				HospitalizationID: &hospID,
				TreatmentContent:  "投薬",
			},
		},
		{
			name: "both parent ids nil (zero value plan)",
			plan: &model.TreatmentPlan{
				ID:               3,
				TreatmentContent: "検査",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toTreatmentPlanResponse(tt.plan)
			assert.Equal(t, strconv.FormatUint(tt.plan.ID, 10), got.ID)
			assert.Equal(t, tt.plan.TreatmentContent, got.TreatmentContent)
			assert.Equal(t, tt.plan.Memo, got.Memo)
			assert.Equal(t, tt.plan.IsInsurance, got.IsInsurance)
			assert.Equal(t, tt.plan.UnitPrice, got.UnitPrice)
			assert.Equal(t, tt.plan.Quantity, got.Quantity)
			assert.Equal(t, tt.plan.DiscountRate, got.DiscountRate)
			assert.Equal(t, tt.plan.DiscountAmount, got.DiscountAmount)
			assert.Equal(t, tt.plan.Subtotal, got.Subtotal)
			assert.Equal(t, tt.plan.SortOrder, got.SortOrder)

			if tt.plan.MedicalRecordID != nil {
				require.NotNil(t, got.MedicalRecordID)
				assert.Equal(t, strconv.FormatUint(*tt.plan.MedicalRecordID, 10), *got.MedicalRecordID)
			} else {
				assert.Nil(t, got.MedicalRecordID)
			}

			if tt.plan.HospitalizationID != nil {
				require.NotNil(t, got.HospitalizationID)
				assert.Equal(t, strconv.FormatUint(*tt.plan.HospitalizationID, 10), *got.HospitalizationID)
			} else {
				assert.Nil(t, got.HospitalizationID)
			}
		})
	}
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Treatment Plan Handler Test Cases
// This handler manages medical treatment plans (Section 4: カルテ管理 - treatment planning)
// TreatmentPlan: nested resource under medical_records for recording planned treatments
//
// CRITICAL ENDPOINTS (nested under /medical-records/:id/treatment-plans):
//
// 1. ListTreatmentPlans (GET /medical-records/:id/treatment-plans)
//    Test Cases (7 scenarios):
//    ✓ Returns 200 OK with array of treatment plans for medical record
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when medical_record id is non-numeric or invalid format
//    ✓ Returns 404 when medical record doesn't exist
//    ✓ Returns 403 when medical record belongs to different clinic (tenant isolation)
//    ✓ Response includes all treatment plan fields with transformations
//    ✓ Returns 500 on database error
//
// 2. GetTreatmentPlan (GET /medical-records/:id/treatment-plans/:plan_id)
//    Test Cases (9 scenarios):
//    ✓ Returns 200 OK with single treatment plan record
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when medical_record id or plan_id is non-numeric or invalid
//    ✓ Returns 404 when medical record doesn't exist
//    ✓ Returns 404 when treatment plan doesn't exist or belongs to different medical record
//    ✓ Returns 403 when medical record belongs to different clinic
//    ✓ Response includes complete treatment plan data with all fields
//    ✓ Uses toTreatmentPlanResponse() transformation for response
//    ✓ Returns 500 on database error
//
// 3. CreateTreatmentPlan (POST /medical-records/:id/treatment-plans)
//    Test Cases (18 scenarios):
//    ✓ Returns 201 Created when treatment plan created successfully
//    ✓ Returns 400 when required field missing (treatment_name or planned_date)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when medical_record id is non-numeric or invalid format
//    ✓ Returns 404 when medical record doesn't exist
//    ✓ Returns 403 when medical record belongs to different clinic
//    ✓ Requires ResourceMedicalRecord edit permission (checked via middleware)
//    ✓ TreatmentName field: required, text (name of planned treatment)
//    ✓ PlannedDate field: required, date (when treatment is planned)
//    ✓ PlannedCost field: optional numeric (estimated cost)
//    ✓ Notes field: optional text (treatment details/instructions)
//    ✓ IsCompleted field: optional boolean, defaults to false
//    ✓ CompletedDate field: optional date (when treatment was actually done)
//    ✓ Created plan includes generated id and timestamps
//    ✓ Response includes related treatment details if populated
//    ✓ Uses toTreatmentPlanResponse() transformation
//    ✓ Returns 500 on database error
//
// 4. UpdateTreatmentPlan (PATCH /medical-records/:id/treatment-plans/:plan_id)
//    Test Cases (16 scenarios):
//    ✓ Returns 200 OK when treatment plan updated successfully
//    ✓ Returns 400 when medical_record id or plan_id is non-numeric or invalid format
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 404 when medical record doesn't exist
//    ✓ Returns 404 when treatment plan doesn't exist
//    ✓ Returns 403 when medical record belongs to different clinic
//    ✓ Requires ResourceMedicalRecord edit permission (checked via middleware)
//    ✓ Partial updates: treatment_name can be updated independently
//    ✓ Partial updates: planned_date can be updated independently
//    ✓ Partial updates: planned_cost can be updated or cleared
//    ✓ Partial updates: is_completed can be toggled
//    ✓ Partial updates: completed_date can be updated (when marking done)
//    ✓ Partial updates: notes can be updated or cleared
//    ✓ Unspecified fields remain unchanged (PATCH semantics)
//    ✓ Uses toTreatmentPlanResponse() transformation
//    ✓ Returns 500 on database error
//
// 5. DeleteTreatmentPlan (DELETE /medical-records/:id/treatment-plans/:plan_id)
//    Test Cases (10 scenarios):
//    ✓ Returns 204 No Content when treatment plan deleted successfully
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when medical_record id or plan_id is non-numeric or invalid format
//    ✓ Returns 404 when medical record doesn't exist
//    ✓ Returns 404 when treatment plan doesn't exist
//    ✓ Returns 403 when medical record belongs to different clinic
//    ✓ Requires ResourceMedicalRecord delete permission (checked via middleware)
//    ✓ Deletion behavior: soft delete or hard delete
//    ✓ Deleted plan no longer appears in ListTreatmentPlans
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control (clinic_id verification via medical_record parent)
//    ✓ RBAC: ResourceMedicalRecord permission (edit, delete required)
//    ✓ Nested resource isolation (plans only accessible through parent medical record)
//    ✓ Soft delete prevents accidental data loss
//    ✓ Partial updates prevent mass assignment
//
// DATA USES:
//    ✓ Treatment plan nested under medical_records (1:N relationship)
//    ✓ Planned treatments guide clinical care workflow
//    ✓ Cost tracking for billing integration
//    ✓ Completion tracking for treatment follow-up
//    ✓ Notes document treatment rationale and instructions
//
// DATA MODEL (treatment_plans):
//    - id (PK): BIGSERIAL
//    - medical_record_id: BIGINT NOT NULL (FK → medical_records)
//    - clinic_id: BIGINT NOT NULL (multitenancy, duplicated from medical_record)
//    - treatment_name: VARCHAR(255) NOT NULL - name of planned treatment
//    - planned_date: DATE NOT NULL - date treatment is planned
//    - planned_cost: NUMERIC(10,2) (NULLABLE) - estimated cost
//    - is_completed: BOOLEAN DEFAULT false - completion flag
//    - completed_date: DATE (NULLABLE) - actual date completed
//    - notes: TEXT (NULLABLE) - treatment details/instructions
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete)
//    - Indexes: (medical_record_id, planned_date), (clinic_id, medical_record_id)
//
// IMPLEMENTATION NOTES:
//    - Nested resource: always accessed via medical_record_id parent
//    - NO standalone list endpoint (only via medical_record)
//    - NO pagination (returns all plans for medical record)
//    - List endpoint: sorted by planned_date (chronological order)
//    - Completion tracking: is_completed flag with optional completed_date
//    - Cost field: optional, for billing/estimate integration
//    - Soft delete preserves treatment history
//    - Transformations: toTreatmentPlanResponse()
//    - RBAC: ResourceMedicalRecord permission required
//    - Parent isolation: plans inherit clinic_id from medical_record
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with sample medical_records
//    - Real service/repository layers
//    - Verify clinic_id scoping (no cross-clinic plan access)
//    - Test ListTreatmentPlans returns all plans sorted by planned_date
//    - Test ListTreatmentPlans empty array when no plans
//    - Test GetTreatmentPlan with valid plan
//    - Test GetTreatmentPlan 404 when plan doesn't exist
//    - Test CreateTreatmentPlan with required fields
//    - Test CreateTreatmentPlan with optional fields (cost, notes)
//    - Test CreateTreatmentPlan default values (is_completed=false)
//    - Test UpdateTreatmentPlan marking as completed (is_completed=true, completed_date)
//    - Test UpdateTreatmentPlan with cost update
//    - Test UpdateTreatmentPlan PATCH semantics
//    - Test DeleteTreatmentPlan soft delete behavior
//    - Test response transformation (toTreatmentPlanResponse)
//    - Test permission checks (ResourceMedicalRecord on edit/delete)
//    - Test parent medical_record isolation
//    - Test FK constraint (medical_record_id must exist)
//    - Verify clinic_id inheritance from parent medical_record
//
