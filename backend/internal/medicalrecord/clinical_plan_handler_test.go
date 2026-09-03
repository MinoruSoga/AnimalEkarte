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

// TestClinicalPlanHandlerCompiles verifies clinical_plan_handler.go compiles
func TestClinicalPlanHandlerCompiles(t *testing.T) {
	assert.True(t, true, "clinical_plan_handler.go compiled successfully")
}

// mockClinicalPlanService は移行前 internal/handler/medical_record_handler_test.go で定義され
// clinical_plan テストが再利用していた ClinicalPlanService のテストダブル。medicalrecord パッケージ
// はその _test.go 定義を import できないため、ここへ最小コピーを再宣言する。
type mockClinicalPlanService struct {
	getOrCreateFn func(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error)
	updateFn      func(ctx context.Context, clinicID, medicalRecordID uint64, input *UpdateClinicalPlanInput) (*model.ClinicalPlan, error)
	deleteFn      func(ctx context.Context, clinicID, medicalRecordID uint64, actorID *uint64) error
}

func (m *mockClinicalPlanService) GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error) {
	if m.getOrCreateFn != nil {
		return m.getOrCreateFn(ctx, clinicID, medicalRecordID)
	}
	return &model.ClinicalPlan{}, nil
}

func (m *mockClinicalPlanService) Update(ctx context.Context, clinicID, medicalRecordID uint64, input *UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, medicalRecordID, input)
	}
	return &model.ClinicalPlan{}, nil
}

func (m *mockClinicalPlanService) Delete(ctx context.Context, clinicID, medicalRecordID uint64, actorID *uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, medicalRecordID, actorID)
	}
	return nil
}

// ---- test helper ----

func newHandlerWithClinicalPlanSvc(svc ClinicalPlanService) *ClinicalPlanHandler {
	return NewClinicalPlanHandler(svc)
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
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "42") },
			svc: &mockClinicalPlanService{
				updateFn: func(_ context.Context, clinicID, medicalRecordID uint64, input *UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(8), medicalRecordID)
					require.NotNil(t, input.PhysicalExam)
					assert.Equal(t, physicalExam, *input.PhysicalExam)
					require.NotNil(t, input.ActorID)
					assert.Equal(t, uint64(42), *input.ActorID)
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
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "42") },
			svc:        &mockClinicalPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when body is malformed",
			paramID:    "8",
			malformed:  true,
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "42") },
			svc:        &mockClinicalPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "8",
			body:     validBody,
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "42") },
			svc: &mockClinicalPlanService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
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
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "42") },
			svc: &mockClinicalPlanService{
				deleteFn: func(_ context.Context, clinicID, medicalRecordID uint64, actorID *uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(8), medicalRecordID)
					require.NotNil(t, actorID)
					assert.Equal(t, uint64(42), *actorID)
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
			name:       "returns 401 when staff actor missing",
			paramID:    "8",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockClinicalPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param is invalid",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "42") },
			svc:        &mockClinicalPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "8",
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "42") },
			svc: &mockClinicalPlanService{
				deleteFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
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
		updateFn: func(_ context.Context, _, _ uint64, input *UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
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
	c.Set("user_id", "42")
	h.UpdateClinicalPlan(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, called)
	assert.Contains(t, w.Body.String(), `"diagnosis_2_type_id":"31"`)
	assert.Contains(t, w.Body.String(), `"diagnosis_2_name_id":"32"`)
}

// TestUpdateClinicalPlan_BindsVersion は BUG-416③ の楽観的ロックについて、PATCH body の
// "version" フィールドが updateClinicalPlanRequest.Version にバインドされ、
// medicalrecord.UpdateClinicalPlanInput.Version までそのまま届くことを検証する。
func TestUpdateClinicalPlan_BindsVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	physicalExam := "経過観察"
	claimedVersion := 3
	body := map[string]any{"physical_exam": physicalExam, "version": claimedVersion}
	called := false
	svc := &mockClinicalPlanService{
		updateFn: func(_ context.Context, _, _ uint64, input *UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
			called = true
			require.NotNil(t, input.Version)
			assert.Equal(t, claimedVersion, *input.Version)
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 8, PhysicalExam: physicalExam, Version: claimedVersion + 1}, nil
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
	c.Set("user_id", "42")
	h.UpdateClinicalPlan(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, called)
	assert.Contains(t, w.Body.String(), `"version":4`)
}

// TestUpdateClinicalPlan_OmittedVersionBindsNil は version 未送信時に
// input.Version が nil のまま（照合スキップ）であることを検証する。
func TestUpdateClinicalPlan_OmittedVersionBindsNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	physicalExam := "経過観察"
	body := map[string]any{"physical_exam": physicalExam}
	svc := &mockClinicalPlanService{
		updateFn: func(_ context.Context, _, _ uint64, input *UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
			assert.Nil(t, input.Version, "version未送信時はnilのままであるべき")
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 8, PhysicalExam: physicalExam}, nil
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
	c.Set("user_id", "42")
	h.UpdateClinicalPlan(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateClinicalPlan_NullClearsDiagnosis2(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := map[string]any{"diagnosis_2_type_id": nil, "diagnosis_2_name_id": nil}
	svc := &mockClinicalPlanService{
		updateFn: func(_ context.Context, _, _ uint64, input *UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
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
	c.Set("user_id", "42")
	h.UpdateClinicalPlan(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

// SEC-CODEX-UHQPM2 selected-clinic grant
func TestGetClinicalPlanSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		invoke func(*ClinicalPlanHandler, *gin.Context)
		svc    *mockClinicalPlanService
	}{
		{
			name: "GetClinicalPlan returns 403 when selected clinic lacks medical record view grant",
			invoke: func(h *ClinicalPlanHandler, c *gin.Context) {
				h.GetClinicalPlan(c)
			},
			svc: &mockClinicalPlanService{
				getOrCreateFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithClinicalPlanSvc(tt.svc)
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
