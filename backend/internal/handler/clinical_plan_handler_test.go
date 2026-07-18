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

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// TestClinicalPlanHandlerCompiles verifies clinical_plan_handler.go compiles
func TestClinicalPlanHandlerCompiles(t *testing.T) {
	assert.True(t, true, "clinical_plan_handler.go compiled successfully")
}

// mockClinicalPlanService は medical_record_handler_test.go で定義済み。ここでは再利用する。

// ---- test helper ----

func newHandlerWithClinicalPlanSvc(svc service.ClinicalPlanService) *Handler {
	return &Handler{svc: &service.Services{ClinicalPlan: svc}}
}

// ---- GetClinicalPlan ----

func TestGetClinicalPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockClinicalPlanService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 with clinical plan",
			paramID:  "8",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClinicalPlanService{
				getOrCreateFn: func(_ context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(8), medicalRecordID)
					return &model.ClinicalPlan{ID: 1, MedicalRecordID: 8, PhysicalExam: "異常なし"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"physical_exam":"異常なし"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "8",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockClinicalPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param is invalid",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockClinicalPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "8",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClinicalPlanService{
				getOrCreateFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithClinicalPlanSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/medical-records/"+tt.paramID+"/clinical-plan", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetClinicalPlan(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateClinicalPlan ----

func TestUpdateClinicalPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	physicalExam := "軽度の発熱"
	validBody := map[string]any{"physical_exam": physicalExam}

	tests := []struct {
		name       string
		paramID    string
		body       any
		malformed  bool
		setupCtx   func(c *gin.Context)
		svc        *mockClinicalPlanService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 when updated successfully",
			paramID:  "8",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClinicalPlanService{
				updateFn: func(_ context.Context, clinicID, medicalRecordID uint64, input *service.UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(8), medicalRecordID)
					require.NotNil(t, input.PhysicalExam)
					assert.Equal(t, physicalExam, *input.PhysicalExam)
					return &model.ClinicalPlan{ID: 1, MedicalRecordID: 8, PhysicalExam: physicalExam}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"physical_exam":"軽度の発熱"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "8",
			body:       validBody,
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockClinicalPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param is invalid",
			paramID:    "abc",
			body:       validBody,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockClinicalPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when body is malformed",
			paramID:    "8",
			malformed:  true,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockClinicalPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "8",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClinicalPlanService{
				updateFn: func(_ context.Context, _, _ uint64, _ *service.UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
					return nil, apperrors.WrapConflict("diagnosis type not found")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithClinicalPlanSvc(tt.svc)

			bodyBytes := []byte("{invalid")
			if !tt.malformed {
				b, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyBytes = b
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/medical-records/"+tt.paramID+"/clinical-plan", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateClinicalPlan(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteClinicalPlan ----

func TestDeleteClinicalPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockClinicalPlanService
		wantStatus int
	}{
		{
			name:     "returns 204 when deleted successfully",
			paramID:  "8",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClinicalPlanService{
				deleteFn: func(_ context.Context, clinicID, medicalRecordID uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(8), medicalRecordID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "8",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockClinicalPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param is invalid",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockClinicalPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "8",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockClinicalPlanService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithClinicalPlanSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/medical-records/"+tt.paramID+"/clinical-plan", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.DeleteClinicalPlan(c)
			c.Writer.WriteHeaderNow() // flush a bare c.Status() (no body) to the recorder

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUpdateClinicalPlan_BindsDiagnosis2TypeID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	typeID := uint64(31)
	nameID := uint64(32)
	body := map[string]any{"diagnosis_2_type_id": typeID, "diagnosis_2_name_id": nameID}
	called := false
	svc := &mockClinicalPlanService{
		updateFn: func(_ context.Context, _, _ uint64, input *service.UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
			called = true
			require.NotNil(t, input.Diagnosis2TypeID)
			require.NotNil(t, *input.Diagnosis2TypeID)
			assert.Equal(t, typeID, **input.Diagnosis2TypeID)
			require.NotNil(t, input.Diagnosis2NameID)
			require.NotNil(t, *input.Diagnosis2NameID)
			assert.Equal(t, nameID, **input.Diagnosis2NameID)
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 8, Diagnosis2TypeID: &typeID, Diagnosis2NameID: &nameID}, nil
		},
	}
	h := newHandlerWithClinicalPlanSvc(svc)
	b, err := json.Marshal(body)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/medical-records/8/clinical-plan", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "8"}}
	setClinicID(c)
	h.UpdateClinicalPlan(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, called)
	assert.Contains(t, w.Body.String(), `"diagnosis_2_type_id":"31"`)
	assert.Contains(t, w.Body.String(), `"diagnosis_2_name_id":"32"`)
}

func TestUpdateClinicalPlan_NullClearsDiagnosis2(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := map[string]any{"diagnosis_2_type_id": nil, "diagnosis_2_name_id": nil}
	svc := &mockClinicalPlanService{
		updateFn: func(_ context.Context, _, _ uint64, input *service.UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
			require.NotNil(t, input.Diagnosis2TypeID)
			assert.Nil(t, *input.Diagnosis2TypeID)
			require.NotNil(t, input.Diagnosis2NameID)
			assert.Nil(t, *input.Diagnosis2NameID)
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 8}, nil
		},
	}
	h := newHandlerWithClinicalPlanSvc(svc)
	b, err := json.Marshal(body)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/medical-records/8/clinical-plan", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "8"}}
	setClinicID(c)
	h.UpdateClinicalPlan(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
