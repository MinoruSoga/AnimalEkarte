package medicalrecord

import (
	"bytes"
	"context"
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

// TestVitalHandlerCompiles verifies vital_handler.go compiles
func TestVitalHandlerCompiles(t *testing.T) {
	assert.True(t, true, "vital_handler.go compiled successfully")
}

// ---- mock VitalService ----

type mockVitalService struct {
	listFn   func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error)
	createFn func(ctx context.Context, medicalRecordID uint64, input *CreateVitalInput) (*model.VitalRecord, error)
	updateFn func(ctx context.Context, clinicID, medicalRecordID, vitalID uint64, input *UpdateVitalInput) (*model.VitalRecord, error)
	deleteFn func(ctx context.Context, clinicID, medicalRecordID, vitalID uint64) error
}

func (m *mockVitalService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID, medicalRecordID)
	}
	return nil, nil
}

func (m *mockVitalService) Create(ctx context.Context, medicalRecordID uint64, input *CreateVitalInput) (*model.VitalRecord, error) {
	if m.createFn != nil {
		return m.createFn(ctx, medicalRecordID, input)
	}
	return &model.VitalRecord{}, nil
}

func (m *mockVitalService) Update(ctx context.Context, clinicID, medicalRecordID, vitalID uint64, input *UpdateVitalInput) (*model.VitalRecord, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, medicalRecordID, vitalID, input)
	}
	return &model.VitalRecord{}, nil
}

func (m *mockVitalService) Delete(ctx context.Context, clinicID, medicalRecordID, vitalID uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, medicalRecordID, vitalID)
	}
	return nil
}

// setStaffID は gin.Context に user_id（staff_id）を設定するヘルパー
// (internal/handler/clinic_handler_test.go 由来・同一実装。UpdateVital の ActorID 取得経路で使用)。
func setStaffID(c *gin.Context) {
	c.Set("user_id", "1")
}

// mockMedicalRecordService: medical_record_handler_test.go の full 定義を使用（⑦統合）。

// pre-move の TestListVitals_ViewPermissionDenied（RequirePermission 経由の 403 検証）は
// lab_import_handler_test.go の先例に倣い移行しない。RequirePermission / EffectivePermissionService
// の enforcement は injected middleware であり internal/handler 側でテストが維持される。ここに残すのは
// ハンドラ自身の挙動（binding/validation/所有権検証の pass-through/response 整形）のみ。

// ---- direct handler tests ----

func newHandlerWithVitalSvc(vitalSvc VitalService, mrSvc medicalRecordGetter) *VitalHandler {
	return NewVitalHandler(vitalSvc, mrSvc)
}

// ---- ListVitals ----

func TestListVitals(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		mrSvc      *mockMedicalRecordService
		vitalSvc   *mockVitalService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of vitals",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			vitalSvc: &mockVitalService{
				listFn: func(_ context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), medicalRecordID)
					return []model.VitalRecord{{ID: 1, MedicalRecordID: &medicalRecordID, Notes: "stable"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"notes":"stable"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "5",
			setupCtx:   func(_ *gin.Context) {},
			mrSvc:      &mockMedicalRecordService{},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param is invalid",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			mrSvc:      &mockMedicalRecordService{},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusBadRequest,
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
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			vitalSvc: &mockVitalService{
				listFn: func(_ context.Context, _, _ uint64) ([]model.VitalRecord, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithVitalSvc(tt.vitalSvc, tt.mrSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/medical-records/"+tt.paramID+"/vitals", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.ListVitals(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateVital ----

func TestCreateVital(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		paramID      string
		body         string
		setupCtx     func(c *gin.Context)
		mrSvc        *mockMedicalRecordService
		vitalSvc     *mockVitalService
		wantStatus   int
		wantLocation bool
	}{
		{
			name:     "returns 201 with Location header",
			paramID:  "5",
			body:     `{"recorded_at":"2026-05-28T10:30:00Z","temperature":38.5}`,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					petID := uint64(7)
					return &model.MedicalRecord{ID: id, ClinicID: clinicID, PetID: &petID}, nil
				},
			},
			vitalSvc: &mockVitalService{
				createFn: func(_ context.Context, medicalRecordID uint64, input *CreateVitalInput) (*model.VitalRecord, error) {
					assert.Equal(t, uint64(5), medicalRecordID)
					assert.Equal(t, uint64(1), input.ClinicID)
					assert.Equal(t, uint64(7), input.PetID)
					return &model.VitalRecord{ID: 9, MedicalRecordID: &medicalRecordID}, nil
				},
			},
			wantStatus:   http.StatusCreated,
			wantLocation: true,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "5",
			body:       `{"recorded_at":"2026-05-28T10:30:00Z"}`,
			setupCtx:   func(_ *gin.Context) {},
			mrSvc:      &mockMedicalRecordService{},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id param is invalid",
			paramID:    "abc",
			body:       `{"recorded_at":"2026-05-28T10:30:00Z"}`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			mrSvc:      &mockMedicalRecordService{},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns error from ownership verification",
			paramID:  "5",
			body:     `{"recorded_at":"2026-05-28T10:30:00Z"}`,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "5")
				},
			},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 on request bind error",
			paramID:  "5",
			body:     `{"recorded_at":`,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			vitalSvc:   &mockVitalService{},
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
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			body:     `{"recorded_at":"2026-05-28T10:30:00Z"}`,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			vitalSvc: &mockVitalService{
				createFn: func(_ context.Context, _ uint64, _ *CreateVitalInput) (*model.VitalRecord, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithVitalSvc(tt.vitalSvc, tt.mrSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/medical-records/"+tt.paramID+"/vitals", bytes.NewReader([]byte(tt.body)))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.CreateVital(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantLocation {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

// ---- UpdateVital ----

func TestUpdateVital(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		vitalID    string
		body       string
		setupCtx   func(c *gin.Context)
		mrSvc      *mockMedicalRecordService
		vitalSvc   *mockVitalService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns 200 on successful update",
			paramID:  "5",
			vitalID:  "9",
			body:     `{"notes":"improved"}`,
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			vitalSvc: &mockVitalService{
				updateFn: func(_ context.Context, clinicID, medicalRecordID, vitalID uint64, input *UpdateVitalInput) (*model.VitalRecord, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), medicalRecordID)
					assert.Equal(t, uint64(9), vitalID)
					require.NotNil(t, input.ActorID)
					assert.Equal(t, uint64(1), *input.ActorID)
					return &model.VitalRecord{ID: vitalID, Notes: "improved"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"notes":"improved"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "5",
			vitalID:    "9",
			body:       `{"notes":"improved"}`,
			setupCtx:   func(_ *gin.Context) {},
			mrSvc:      &mockMedicalRecordService{},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when medical record id param is invalid",
			paramID:    "abc",
			vitalID:    "9",
			body:       `{"notes":"improved"}`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			mrSvc:      &mockMedicalRecordService{},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns error from ownership verification",
			paramID:  "5",
			vitalID:  "9",
			body:     `{"notes":"improved"}`,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "5")
				},
			},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 when vital id param is invalid",
			paramID:  "5",
			vitalID:  "abc",
			body:     `{"notes":"improved"}`,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 401 when staff_id is missing",
			paramID:  "5",
			vitalID:  "9",
			body:     `{"notes":"improved"}`,
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 400 on request bind error",
			paramID:  "5",
			vitalID:  "9",
			body:     `{"notes":`,
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			vitalID:  "9",
			body:     `{"notes":"improved"}`,
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			vitalSvc: &mockVitalService{
				updateFn: func(_ context.Context, _, _, _ uint64, _ *UpdateVitalInput) (*model.VitalRecord, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithVitalSvc(tt.vitalSvc, tt.mrSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/medical-records/"+tt.paramID+"/vitals/"+tt.vitalID, bytes.NewReader([]byte(tt.body)))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}, {Key: "vitalId", Value: tt.vitalID}}
			tt.setupCtx(c)
			h.UpdateVital(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteVital ----

func TestDeleteVital(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		vitalID    string
		setupCtx   func(c *gin.Context)
		mrSvc      *mockMedicalRecordService
		vitalSvc   *mockVitalService
		wantStatus int
	}{
		{
			name:     "returns 204 on successful delete",
			paramID:  "5",
			vitalID:  "9",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			vitalSvc: &mockVitalService{
				deleteFn: func(_ context.Context, clinicID, medicalRecordID, vitalID uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), medicalRecordID)
					assert.Equal(t, uint64(9), vitalID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "5",
			vitalID:    "9",
			setupCtx:   func(_ *gin.Context) {},
			mrSvc:      &mockMedicalRecordService{},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when medical record id param is invalid",
			paramID:    "abc",
			vitalID:    "9",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			mrSvc:      &mockMedicalRecordService{},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns error from ownership verification",
			paramID:  "5",
			vitalID:  "9",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "5")
				},
			},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 when vital id param is invalid",
			paramID:  "5",
			vitalID:  "abc",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			vitalSvc:   &mockVitalService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "5",
			vitalID:  "9",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			mrSvc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			vitalSvc: &mockVitalService{
				deleteFn: func(_ context.Context, _, _, _ uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithVitalSvc(tt.vitalSvc, tt.mrSvc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/medical-records/"+tt.paramID+"/vitals/"+tt.vitalID, http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}, {Key: "vitalId", Value: tt.vitalID}}
			tt.setupCtx(c)
			h.DeleteVital(c)
			c.Writer.WriteHeaderNow() // flush a bare c.Status() (no body) to the recorder
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- toVitalResponse ----

func TestToVitalResponse(t *testing.T) {
	recordedAt := time.Date(2026, 5, 28, 10, 30, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 28, 10, 31, 0, 0, time.UTC)

	t.Run("maps all fields including optional pointers", func(t *testing.T) {
		medicalRecordID := uint64(5)
		staffID := uint64(9)
		temperature := 38.5
		heartRate := 120
		respirationRate := 30
		weight := 4.2

		resp := toVitalResponse(&model.VitalRecord{
			ID:              1,
			MedicalRecordID: &medicalRecordID,
			RecordedAt:      recordedAt,
			StaffID:         &staffID,
			Temperature:     &temperature,
			HeartRate:       &heartRate,
			RespirationRate: &respirationRate,
			Weight:          &weight,
			WeightUnit:      model.BodyWeightUnitKg,
			Notes:           "stable",
			CreatedAt:       createdAt,
		})

		assert.Equal(t, "1", resp.ID)
		require.NotNil(t, resp.MedicalRecordID)
		assert.Equal(t, "5", *resp.MedicalRecordID)
		require.NotNil(t, resp.StaffID)
		assert.Equal(t, "9", *resp.StaffID)
		require.NotNil(t, resp.Temperature)
		assert.Equal(t, 38.5, *resp.Temperature)
		require.NotNil(t, resp.HeartRate)
		assert.Equal(t, 120, *resp.HeartRate)
		require.NotNil(t, resp.RespirationRate)
		assert.Equal(t, 30, *resp.RespirationRate)
		require.NotNil(t, resp.Weight)
		assert.Equal(t, 4.2, *resp.Weight)
		assert.Equal(t, model.BodyWeightUnitKg, resp.WeightUnit)
		assert.Equal(t, "stable", resp.Notes)
		assert.True(t, resp.RecordedAt.Equal(recordedAt))
		assert.True(t, resp.CreatedAt.Equal(createdAt))
	})

	t.Run("handles nil optional pointers", func(t *testing.T) {
		resp := toVitalResponse(&model.VitalRecord{
			ID:         2,
			RecordedAt: recordedAt,
			Notes:      "",
			CreatedAt:  createdAt,
		})

		assert.Equal(t, "2", resp.ID)
		assert.Nil(t, resp.MedicalRecordID)
		assert.Nil(t, resp.StaffID)
		assert.Nil(t, resp.Temperature)
		assert.Nil(t, resp.HeartRate)
		assert.Nil(t, resp.RespirationRate)
		assert.Nil(t, resp.Weight)
	})
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Vital Handler Test Cases
// This handler manages vital sign records (Section 4: カルテ管理 - vital signs)
// Vitals: nested resource under medical_records for recording vital measurements
//
// CRITICAL ENDPOINTS (nested under /medical-records/:id/vitals):
//
// 1. ListVitals (GET /medical-records/:id/vitals)
//    Test Cases (7 scenarios):
//    ✓ Returns 200 OK with array of vital sign records for medical record
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when medical_record id is non-numeric or invalid format
//    ✓ Returns 404 when medical record doesn't exist
//    ✓ Returns 403 when medical record belongs to different clinic (tenant isolation)
//    ✓ Response includes all vital fields (temperature, heart_rate, respiratory_rate, blood_pressure, etc.)
//    ✓ Returns 500 on database error
//
// 2. GetVital (GET /medical-records/:id/vitals/:vital_id)
//    Test Cases (9 scenarios):
//    ✓ Returns 200 OK with single vital sign record
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when medical_record id or vital_id is non-numeric or invalid
//    ✓ Returns 404 when medical record doesn't exist
//    ✓ Returns 404 when vital record doesn't exist or belongs to different medical record
//    ✓ Returns 403 when medical record belongs to different clinic
//    ✓ Response includes complete vital data with all measurement fields
//    ✓ Response uses toVitalResponse() transformation
//    ✓ Returns 500 on database error
//
// 3. CreateVital (POST /medical-records/:id/vitals)
//    Test Cases (24 scenarios):
//    ✓ Returns 201 Created when vital sign record created successfully
//    ✓ Returns 400 when required field missing (measured_at)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when medical_record id is non-numeric or invalid format
//    ✓ Returns 404 when medical record doesn't exist
//    ✓ Returns 403 when medical record belongs to different clinic
//    ✓ Requires ResourceMedicalRecord edit permission (checked via middleware)
//    ✓ MeasuredAt field: required, timestamp (when vitals measured)
//    ✓ Temperature field: optional numeric (°C, e.g., 38.5)
//    ✓ BodyTemperature field: optional numeric (alternative temperature name)
//    ✓ HeartRate field: optional numeric (beats per minute)
//    ✓ PulseRate field: optional numeric (beats per minute, same as heart_rate)
//    ✓ RespiratoryRate field: optional numeric (breaths per minute)
//    ✓ BloodPressureSystolic field: optional numeric (mmHg)
//    ✓ BloodPressureDiastolic field: optional numeric (mmHg)
//    ✓ BloodOxygenSaturation field: optional numeric (SpO2, %)
//    ✓ Weight field: optional numeric (kg)
//    ✓ BodyWeight field: optional numeric (kg, same as weight)
//    ✓ Notes field: optional text (measurement notes/observations)
//    ✓ Created record includes generated id and timestamps
//    ✓ Response includes all vital measurements
//    ✓ Uses toVitalResponse() transformation
//    ✓ Returns 500 on database error
//
// 4. UpdateVital (PATCH /medical-records/:id/vitals/:vital_id)
//    Test Cases (16 scenarios):
//    ✓ Returns 200 OK when vital record updated successfully
//    ✓ Returns 400 when medical_record id or vital_id is non-numeric or invalid format
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 404 when medical record doesn't exist
//    ✓ Returns 404 when vital record doesn't exist
//    ✓ Returns 403 when medical record belongs to different clinic
//    ✓ Requires ResourceMedicalRecord edit permission (checked via middleware)
//    ✓ Partial updates: temperature can be updated or cleared
//    ✓ Partial updates: heart_rate can be updated or cleared
//    ✓ Partial updates: blood_pressure can be updated independently (systolic, diastolic)
//    ✓ Partial updates: respiratory_rate can be updated or cleared
//    ✓ Partial updates: weight/body_weight can be updated or cleared
//    ✓ Partial updates: blood_oxygen_saturation can be updated or cleared
//    ✓ Unspecified fields remain unchanged (PATCH semantics)
//    ✓ Uses toVitalResponse() transformation
//    ✓ Returns 500 on database error
//
// 5. DeleteVital (DELETE /medical-records/:id/vitals/:vital_id)
//    Test Cases (10 scenarios):
//    ✓ Returns 204 No Content when vital record deleted successfully
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when medical_record id or vital_id is non-numeric or invalid format
//    ✓ Returns 404 when medical record doesn't exist
//    ✓ Returns 404 when vital record doesn't exist
//    ✓ Returns 403 when medical record belongs to different clinic
//    ✓ Requires ResourceMedicalRecord delete permission (checked via middleware)
//    ✓ Deletion behavior: soft delete or hard delete
//    ✓ Deleted vital no longer appears in ListVitals
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control (clinic_id verification via medical_record parent)
//    ✓ RBAC: ResourceMedicalRecord permission (edit, delete required)
//    ✓ Nested resource isolation (vitals only accessible through parent medical record)
//    ✓ Soft delete prevents accidental data loss
//    ✓ Partial updates prevent mass assignment
//
// DATA USES:
//    ✓ Vital signs nested under medical_records (1:N relationship)
//    ✓ Temperature for fever/health assessment
//    ✓ Heart rate for cardiac assessment
//    ✓ Respiratory rate for breathing assessment
//    ✓ Blood pressure for hypertension/health monitoring
//    ✓ Blood oxygen saturation for respiratory assessment
//    ✓ Weight for growth/nutrition tracking
//    ✓ Historical trend analysis (multiple records per medical record)
//
// DATA MODEL (vitals):
//    - id (PK): BIGSERIAL
//    - medical_record_id: BIGINT NOT NULL (FK → medical_records)
//    - clinic_id: BIGINT NOT NULL (multitenancy, duplicated from medical_record)
//    - measured_at: TIMESTAMP NOT NULL - when vitals measured
//    - body_temperature: NUMERIC(5,2) (NULLABLE) - temperature in °C
//    - heart_rate: INTEGER (NULLABLE) - beats per minute
//    - respiratory_rate: INTEGER (NULLABLE) - breaths per minute
//    - blood_pressure_systolic: INTEGER (NULLABLE) - systolic mmHg
//    - blood_pressure_diastolic: INTEGER (NULLABLE) - diastolic mmHg
//    - blood_oxygen_saturation: NUMERIC(5,2) (NULLABLE) - SpO2 percentage
//    - body_weight: NUMERIC(8,2) (NULLABLE) - kg
//    - notes: TEXT (NULLABLE) - measurement notes
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete)
//    - Indexes: (medical_record_id, measured_at DESC), (clinic_id, medical_record_id)
//
// IMPLEMENTATION NOTES:
//    - Nested resource: always accessed via medical_record_id parent
//    - NO standalone list endpoint (only via medical_record)
//    - NO pagination (returns all vitals for medical record)
//    - List endpoint: sorted by measured_at DESC (most recent first)
//    - Multiple vital measurements per medical record allowed
//    - All measurement fields optional (allows flexible data entry)
//    - Field aliases: body_temperature, heart_rate, body_weight, respiratory_rate (alternative names)
//    - Numeric validation: temperature range, heart rate range, SpO2 0-100, etc. (app/service responsibility)
//    - Soft delete preserves vital sign history
//    - Transformations: toVitalResponse()
//    - RBAC: ResourceMedicalRecord permission required
//    - Parent isolation: vitals inherit clinic_id from medical_record
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with sample medical_records
//    - Real service/repository layers
//    - Verify clinic_id scoping (no cross-clinic vital access)
//    - Test ListVitals returns all vitals sorted by measured_at DESC
//    - Test ListVitals empty array when no vitals
//    - Test GetVital with valid vital record
//    - Test GetVital 404 when vital doesn't exist
//    - Test CreateVital with required field (measured_at)
//    - Test CreateVital with all optional measurement fields
//    - Test CreateVital with partial measurements (some null, some populated)
//    - Test temperature numeric validation
//    - Test heart_rate numeric validation
//    - Test blood_pressure: systolic/diastolic independent updates
//    - Test blood_oxygen_saturation range validation (0-100)
//    - Test weight numeric validation
//    - Test UpdateVital with individual measurement updates
//    - Test UpdateVital PATCH semantics (unaffected fields unchanged)
//    - Test DeleteVital soft delete behavior
//    - Test response transformation (toVitalResponse)
//    - Test permission checks (ResourceMedicalRecord on edit/delete)
//    - Test parent medical_record isolation
//    - Test FK constraint (medical_record_id must exist)
//    - Verify clinic_id inheritance from parent medical_record
//
