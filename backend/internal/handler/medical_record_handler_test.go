package handler

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

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock MedicalRecordService ----

type mockMedicalRecordService struct {
	listFn       func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error)
	getByIDFn    func(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	countByPetFn func(ctx context.Context, clinicID, petID uint64) (int64, error)
	createFn     func(ctx context.Context, record *model.MedicalRecord) error
	updateFn     func(ctx context.Context, clinicID, id uint64, input service.UpdateMedicalRecordInput) (*model.MedicalRecord, error)
	deleteFn     func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockMedicalRecordService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
	return m.listFn(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

func (m *mockMedicalRecordService) GetByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockMedicalRecordService) CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error) {
	if m.countByPetFn != nil {
		return m.countByPetFn(ctx, clinicID, petID)
	}
	return 0, nil
}

func (m *mockMedicalRecordService) Create(ctx context.Context, record *model.MedicalRecord) error {
	return m.createFn(ctx, record)
}

func (m *mockMedicalRecordService) Update(ctx context.Context, clinicID, id uint64, input service.UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.MedicalRecord{}, nil
}

func (m *mockMedicalRecordService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockMedicalRecordService) CreateSubRecords(_ context.Context, _, _ uint64, _ service.CreateSubRecordsInput) {
}

func (m *mockMedicalRecordService) AutoCreateFromReservation(_ context.Context, _ uint64, _ *model.Appointment) {
}

// ---- mock ClinicalPlanService ----

type mockClinicalPlanService struct {
	getOrCreateFn func(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error)
	updateFn      func(ctx context.Context, clinicID, medicalRecordID uint64, input *service.UpdateClinicalPlanInput) (*model.ClinicalPlan, error)
	deleteFn      func(ctx context.Context, clinicID, medicalRecordID uint64) error
}

func (m *mockClinicalPlanService) GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error) {
	if m.getOrCreateFn != nil {
		return m.getOrCreateFn(ctx, clinicID, medicalRecordID)
	}
	return &model.ClinicalPlan{}, nil
}

func (m *mockClinicalPlanService) Update(ctx context.Context, clinicID, medicalRecordID uint64, input *service.UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, medicalRecordID, input)
	}
	return &model.ClinicalPlan{}, nil
}

func (m *mockClinicalPlanService) Delete(ctx context.Context, clinicID, medicalRecordID uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, medicalRecordID)
	}
	return nil
}

// ---- mock InquiryService ----

type mockInquiryService struct {
	upsertFn func(ctx context.Context, input service.UpsertInquiryInput) (*model.Inquiry, error)
}

func (m *mockInquiryService) Upsert(ctx context.Context, input service.UpsertInquiryInput) (*model.Inquiry, error) {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, input)
	}
	return &model.Inquiry{}, nil
}

// ---- test helper ----

func newHandlerWithMedicalRecordSvc(mrSvc service.MedicalRecordService, cpSvc service.ClinicalPlanService) *Handler {
	return &Handler{
		svc: &service.Services{
			MedicalRecord: mrSvc,
			ClinicalPlan:  cpSvc,
			Inquiry:       &mockInquiryService{},
		},
	}
}

// ---- ListMedicalRecords ----

func TestListMedicalRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockMedicalRecordService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns paginated records",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordService{
				listFn: func(_ context.Context, clinicID uint64, _, _ *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.MedicalRecord{{ID: 1, RecordNo: "MR-20260324-1-ABCDEF"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"record_no":"MR-20260324-1-ABCDEF"`,
		},
		{
			name:     "passes pet_id filter to service",
			query:    "page=1&limit=10&pet_id=5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordService{
				listFn: func(_ context.Context, _ uint64, petID, _ *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
					require.NotNil(t, petID)
					assert.Equal(t, uint64(5), *petID)
					return []model.MedicalRecord{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid pet_id",
			query:      "page=1&limit=10&pet_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMedicalRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "page=1&limit=10",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMedicalRecordService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordService{
				listFn: func(_ context.Context, _ uint64, _, _ *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
					return nil, 0, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMedicalRecordSvc(tt.svc, &mockClinicalPlanService{})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.ListMedicalRecords(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetMedicalRecord ----

func TestGetMedicalRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockMedicalRecordService
		wantStatus int
	}{
		{
			name:     "returns record for valid id",
			paramID:  "7",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(7), id)
					return &model.MedicalRecord{ID: 7}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMedicalRecordService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMedicalRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMedicalRecordSvc(tt.svc, &mockClinicalPlanService{})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetMedicalRecord(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- CreateMedicalRecord ----

func TestCreateMedicalRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ptrStr := func(s string) *string { return &s }

	validBody := func() map[string]any {
		return map[string]any{
			"owner_id":   "10",
			"pet_id":     "20",
			"visit_date": "2026-03-24",
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		mrSvc      *mockMedicalRecordService
		cpSvc      *mockClinicalPlanService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates record with valid body",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			mrSvc: &mockMedicalRecordService{
				createFn: func(_ context.Context, record *model.MedicalRecord) error {
					assert.Equal(t, uint64(1), record.ClinicID)
					require.NotNil(t, record.OwnerID)
					assert.Equal(t, uint64(10), *record.OwnerID)
					require.NotNil(t, record.PetID)
					assert.Equal(t, uint64(20), *record.PetID)
					assert.Equal(t, 2026, record.Date.Year())
					require.NotNil(t, record.EnteredBy)
					assert.Equal(t, uint64(1), *record.EnteredBy) // extractStaffID from user_id="1"
					return nil
				},
			},
			cpSvc:      &mockClinicalPlanService{},
			wantStatus: http.StatusCreated,
		},
		{
			// RecordNo の自動生成は service 層の責務。
			// handler は RecordNo を空のまま service に渡し、service が生成する。
			name:     "auto-generates record_no when not provided",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			mrSvc: &mockMedicalRecordService{
				createFn: func(_ context.Context, record *model.MedicalRecord) error {
					assert.Empty(t, record.RecordNo) // handler は空で渡す（service が生成）
					return nil
				},
			},
			cpSvc:      &mockClinicalPlanService{},
			wantStatus: http.StatusCreated,
		},
		{
			name: "uses provided record_no when given",
			body: func() map[string]any {
				b := validBody()
				b["record_no"] = "MR-CUSTOM-001"
				return b
			}(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			mrSvc: &mockMedicalRecordService{
				createFn: func(_ context.Context, record *model.MedicalRecord) error {
					assert.Equal(t, "MR-CUSTOM-001", record.RecordNo)
					return nil
				},
			},
			cpSvc:      &mockClinicalPlanService{},
			wantStatus: http.StatusCreated,
		},
		{
			name: "also updates clinical_plan when plan is provided",
			body: func() map[string]any {
				b := validBody()
				b["plan"] = "経過観察"
				return b
			}(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			mrSvc: &mockMedicalRecordService{
				createFn: func(_ context.Context, _ *model.MedicalRecord) error { return nil },
			},
			cpSvc: &mockClinicalPlanService{
				updateFn: func(_ context.Context, _, _ uint64, input *service.UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
					require.NotNil(t, input.TreatmentPolicy)
					assert.Equal(t, "経過観察", *input.TreatmentPolicy)
					return &model.ClinicalPlan{}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			mrSvc:      &mockMedicalRecordService{},
			cpSvc:      &mockClinicalPlanService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when owner_id is missing",
			body:       map[string]any{"pet_id": "20"}, // owner_id missing (binding required)
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			mrSvc:      &mockMedicalRecordService{},
			cpSvc:      &mockClinicalPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 when owner_id is non-numeric",
			body: func() map[string]any {
				b := validBody()
				b["owner_id"] = ptrStr("not-a-number")
				return b
			}(),
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			mrSvc:      &mockMedicalRecordService{},
			cpSvc:      &mockClinicalPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for invalid visit_date format",
			body: func() map[string]any {
				b := validBody()
				b["visit_date"] = ptrStr("03/24/2026") // wrong format
				return b
			}(),
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			mrSvc:      &mockMedicalRecordService{},
			cpSvc:      &mockClinicalPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for invalid status",
			body: func() map[string]any {
				b := validBody()
				b["status"] = "unknown_status"
				return b
			}(),
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			mrSvc:      &mockMedicalRecordService{},
			cpSvc:      &mockClinicalPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			mrSvc: &mockMedicalRecordService{
				createFn: func(_ context.Context, _ *model.MedicalRecord) error {
					return fmt.Errorf("db error")
				},
			},
			cpSvc:      &mockClinicalPlanService{},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMedicalRecordSvc(tt.mrSvc, tt.cpSvc)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreateMedicalRecord(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateMedicalRecord ----

func TestUpdateMedicalRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now()

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockMedicalRecordService
		wantStatus int
	}{
		{
			name:     "updates record successfully",
			paramID:  "1",
			body:     map[string]any{"status": "finalized"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordService{
				updateFn: func(_ context.Context, _, _ uint64, input service.UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
					require.NotNil(t, input.Status)
					assert.Equal(t, model.MedicalRecordStatus("finalized"), *input.Status)
					return &model.MedicalRecord{ID: 1}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "updates date field",
			paramID:  "1",
			body:     map[string]any{"date": now.Format(time.RFC3339)},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordService{
				updateFn: func(_ context.Context, _, _ uint64, input service.UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
					require.NotNil(t, input.Date)
					assert.False(t, input.Date.IsZero())
					return &model.MedicalRecord{ID: 1}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid status",
			paramID:    "1",
			body:       map[string]any{"status": "bad_status"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMedicalRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockMedicalRecordService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			body:       map[string]any{},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMedicalRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"status": "finalized"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordService{
				updateFn: func(_ context.Context, _, _ uint64, _ service.UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithMedicalRecordSvc(tt.svc, &mockClinicalPlanService{})
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UpdateMedicalRecord(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteMedicalRecord ----

func newDeleteMedicalRecordRouter(mrSvc service.MedicalRecordService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithMedicalRecordSvc(mrSvc, &mockClinicalPlanService{})
	r.DELETE("/medical-records/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteMedicalRecord)
	return r
}

func TestDeleteMedicalRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockMedicalRecordService
		wantStatus int
	}{
		{
			name:    "deletes record successfully",
			paramID: "1",
			svc: &mockMedicalRecordService{
				deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockMedicalRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			svc: &mockMedicalRecordService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("medical_record", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteMedicalRecordRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/medical-records/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithMedicalRecordSvc(&mockMedicalRecordService{}, &mockClinicalPlanService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteMedicalRecord(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
