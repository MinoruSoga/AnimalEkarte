package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock MedicalRecordService ----

type mockMedicalRecordService struct {
	listFn                       func(ctx context.Context, clinicIDs []uint64, filters MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error)
	getByIDFn                    func(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	getByIDForClinicsFn          func(ctx context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error)
	countByPetFn                 func(ctx context.Context, clinicID, petID uint64) (int64, error)
	createFn                     func(ctx context.Context, clinicID uint64, input *CreateMedicalRecordInput) (*model.MedicalRecord, error)
	updateFn                     func(ctx context.Context, clinicID, id uint64, input UpdateMedicalRecordInput) (*model.MedicalRecord, error)
	deleteFn                     func(ctx context.Context, clinicID, id uint64) error
	updateRecommendationReasonFn func(ctx context.Context, clinicID, id uint64, input UpdateRecommendationReasonInput) (*model.MedicalRecord, error)
	autoCreateFromReservationFn  func(ctx context.Context, clinicID uint64, reservation *model.Reservation)
}

func (m *mockMedicalRecordService) List(ctx context.Context, clinicIDs []uint64, filters MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error) {
	return m.listFn(ctx, clinicIDs, filters, page, limit)
}

func (m *mockMedicalRecordService) GetByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockMedicalRecordService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error) {
	if m.getByIDForClinicsFn != nil {
		return m.getByIDForClinicsFn(ctx, clinicIDs, id)
	}
	return nil, nil
}

func (m *mockMedicalRecordService) CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error) {
	if m.countByPetFn != nil {
		return m.countByPetFn(ctx, clinicID, petID)
	}
	return 0, nil
}

func (m *mockMedicalRecordService) Create(ctx context.Context, clinicID uint64, input *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.MedicalRecord{
		ID:                       1,
		ClinicID:                 clinicID,
		RecordNo:                 input.RecordNo,
		Date:                     input.Date,
		OwnerID:                  input.OwnerID,
		PetID:                    input.PetID,
		DoctorID:                 input.DoctorID,
		AppointmentID:            input.AppointmentID,
		VisitType:                &input.VisitType,
		NextVisitRecommendedDate: input.NextVisitRecommendedDate,
		RecommendationReason:     input.RecommendationReason,
		EnteredBy:                input.EnteredBy,
	}, nil
}

func (m *mockMedicalRecordService) Update(ctx context.Context, clinicID, id uint64, input UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.MedicalRecord{}, nil
}

func (m *mockMedicalRecordService) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockMedicalRecordService) CreateSubRecords(_ context.Context, _, _ uint64, _ CreateSubRecordsInput) error {
	return nil
}

func (m *mockMedicalRecordService) AutoCreateFromReservation(ctx context.Context, clinicID uint64, reservation *model.Reservation) {
	if m.autoCreateFromReservationFn != nil {
		m.autoCreateFromReservationFn(ctx, clinicID, reservation)
	}
}

func (m *mockMedicalRecordService) DeleteDraftFromReservation(_ context.Context, _, _ uint64) {}

func (m *mockMedicalRecordService) UpdateRecommendationReason(ctx context.Context, clinicID, id uint64, input UpdateRecommendationReasonInput) (*model.MedicalRecord, error) {
	if m.updateRecommendationReasonFn != nil {
		return m.updateRecommendationReasonFn(ctx, clinicID, id, input)
	}
	return &model.MedicalRecord{}, nil
}

// ---- mock ClinicalPlanService ----

// mockClinicalPlanService: clinical_plan_handler_test.go の既存定義を使用（⑦統合）。

// ---- mock InquiryService ----

// mockInquiryService moved to internal/medicalrecord/inquiry_handler_test.go (BE9-2D — the
// Inquiry service and its Services field were removed from this package).

// ---- mock LstepTagSyncService ----

// mockLstepTagSyncService: service_deps_mock_test.go の既存定義を使用（⑦統合）。

// ---- test helper ----

// clinical_plan handler methods moved to internal/medicalrecord (BE9-2D sub-batch④a); the parent
// medical-record handler never reads svc.ClinicalPlan, so cpSvc is retained only to keep the
// existing table-driven call sites unchanged and is intentionally unused here.
func newHandlerWithMedicalRecordSvc(mrSvc MedicalRecordService, _ ClinicalPlanService) *MedicalRecordHandler {
	return NewMedicalRecordHandler(mrSvc)
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
				listFn: func(_ context.Context, clinicIDs []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
					assert.Equal(t, []uint64{1}, clinicIDs)
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
				listFn: func(_ context.Context, _ []uint64, filters MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
					require.NotNil(t, filters.PetID)
					assert.Equal(t, uint64(5), *filters.PetID)
					return []model.MedicalRecord{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "passes search/status/doctor_id/animal_species_id filters to service",
			query:    "page=1&limit=10&search=" + "%E7%94%B0%E4%B8%AD" + "&status=draft&doctor_id=7&animal_species_id=3",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordService{
				listFn: func(_ context.Context, _ []uint64, filters MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
					assert.Equal(t, "田中", filters.Search)
					require.NotNil(t, filters.Status)
					assert.Equal(t, model.MedicalRecordStatusDraft, *filters.Status)
					require.NotNil(t, filters.DoctorID)
					assert.Equal(t, uint64(7), *filters.DoctorID)
					require.NotNil(t, filters.AnimalSpeciesID)
					assert.Equal(t, uint64(3), *filters.AnimalSpeciesID)
					return []model.MedicalRecord{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "passes allowed sort/order to service (B-1 follow-up)",
			query:    "page=1&limit=10&sort=owner_name&order=asc",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordService{
				listFn: func(_ context.Context, _ []uint64, filters MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
					assert.Equal(t, "owner_name", filters.Sort)
					assert.Equal(t, "asc", filters.Order)
					return []model.MedicalRecord{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "unknown sort key falls back to default (empty Sort/Order)",
			query:    "page=1&limit=10&sort=unknown_column&order=asc",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordService{
				listFn: func(_ context.Context, _ []uint64, filters MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
					assert.Empty(t, filters.Sort, "許可外の sort キーは無視され既定順にフォールバックする")
					assert.Empty(t, filters.Order)
					return []model.MedicalRecord{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "sort without order defaults to desc",
			query:    "page=1&limit=10&sort=date",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockMedicalRecordService{
				listFn: func(_ context.Context, _ []uint64, filters MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
					assert.Equal(t, "date", filters.Sort)
					assert.Equal(t, "desc", filters.Order)
					return []model.MedicalRecord{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid status",
			query:      "page=1&limit=10&status=unknown",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockMedicalRecordService{},
			wantStatus: http.StatusBadRequest,
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
				listFn: func(_ context.Context, _ []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
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
				getByIDForClinicsFn: func(_ context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error) {
					assert.Equal(t, []uint64{1}, clinicIDs)
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
				getByIDForClinicsFn: func(_ context.Context, _ []uint64, _ uint64) (*model.MedicalRecord, error) {
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
				createFn: func(_ context.Context, clinicID uint64, input *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
					assert.Equal(t, uint64(1), clinicID)
					require.NotNil(t, input.OwnerID)
					assert.Equal(t, uint64(10), *input.OwnerID)
					require.NotNil(t, input.PetID)
					assert.Equal(t, uint64(20), *input.PetID)
					assert.Equal(t, 2026, input.Date.Year())
					require.NotNil(t, input.EnteredBy)
					assert.Equal(t, uint64(1), *input.EnteredBy) // extractStaffID from user_id="1"
					return &model.MedicalRecord{ID: 1, ClinicID: clinicID, Date: input.Date, OwnerID: input.OwnerID, PetID: input.PetID, EnteredBy: input.EnteredBy}, nil
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
				createFn: func(_ context.Context, clinicID uint64, input *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
					assert.Empty(t, input.RecordNo) // handler は空で渡す（service が生成）
					return &model.MedicalRecord{ID: 1, ClinicID: clinicID, Date: input.Date}, nil
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
				createFn: func(_ context.Context, clinicID uint64, input *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
					assert.Equal(t, "MR-CUSTOM-001", input.RecordNo)
					return &model.MedicalRecord{ID: 1, ClinicID: clinicID, Date: input.Date, RecordNo: input.RecordNo}, nil
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
				createFn: func(_ context.Context, clinicID uint64, input *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 1, ClinicID: clinicID, Date: input.Date}, nil
				},
			},
			cpSvc: &mockClinicalPlanService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
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
				createFn: func(_ context.Context, _ uint64, _ *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			cpSvc:      &mockClinicalPlanService{},
			wantStatus: http.StatusInternalServerError,
		},
		// FEAT-382-2 supplement: recommendation_reason on create
		{
			name: "sets recommendation_reason when valid value provided",
			body: func() map[string]any {
				b := validBody()
				b["recommendation_reason"] = "checkup"
				return b
			}(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			mrSvc: &mockMedicalRecordService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
					require.NotNil(t, input.RecommendationReason)
					assert.Equal(t, "checkup", *input.RecommendationReason)
					return &model.MedicalRecord{ID: 1, ClinicID: clinicID, Date: input.Date, RecommendationReason: input.RecommendationReason}, nil
				},
			},
			cpSvc:      &mockClinicalPlanService{},
			wantStatus: http.StatusCreated,
		},
		{
			name: "returns 400 when recommendation_reason is invalid",
			body: func() map[string]any {
				b := validBody()
				b["recommendation_reason"] = "invalid_value"
				return b
			}(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			mrSvc: &mockMedicalRecordService{
				createFn: func(_ context.Context, _ uint64, _ *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapInvalidInput("recommendation_reason must be one of: revisit, checkup, prevention, exam")
				},
			},
			cpSvc:      &mockClinicalPlanService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "stores recommendation_reason as NULL when empty string provided",
			body: func() map[string]any {
				b := validBody()
				b["recommendation_reason"] = ""
				return b
			}(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			mrSvc: &mockMedicalRecordService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
					assert.Nil(t, input.RecommendationReason)
					return &model.MedicalRecord{ID: 1, ClinicID: clinicID, Date: input.Date, RecommendationReason: input.RecommendationReason}, nil
				},
			},
			cpSvc:      &mockClinicalPlanService{},
			wantStatus: http.StatusCreated,
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
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc: &mockMedicalRecordService{
				updateFn: func(_ context.Context, _, _ uint64, input UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
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
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc: &mockMedicalRecordService{
				updateFn: func(_ context.Context, _, _ uint64, input UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
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
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc: &mockMedicalRecordService{
				updateFn: func(_ context.Context, _, _ uint64, _ UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "updates visit_type=first",
			paramID:  "1",
			body:     map[string]any{"visit_type": "first"},
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc: &mockMedicalRecordService{
				updateFn: func(_ context.Context, _, _ uint64, input UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
					require.NotNil(t, input.VisitType)
					assert.Equal(t, model.VisitTypeFirst, *input.VisitType)
					return &model.MedicalRecord{ID: 1}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "updates visit_type=revisit",
			paramID:  "1",
			body:     map[string]any{"visit_type": "revisit"},
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc: &mockMedicalRecordService{
				updateFn: func(_ context.Context, _, _ uint64, input UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
					require.NotNil(t, input.VisitType)
					assert.Equal(t, model.VisitTypeRevisit, *input.VisitType)
					return &model.MedicalRecord{ID: 1}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid visit_type",
			paramID:    "1",
			body:       map[string]any{"visit_type": "invalid"},
			setupCtx:   func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc:        &mockMedicalRecordService{},
			wantStatus: http.StatusBadRequest,
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

func newDeleteMedicalRecordRouter(mrSvc MedicalRecordService) *gin.Engine {
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

// ---- UpdateMedicalRecordRecommendationReason ----

func newPatchRecommendationReasonRouter(mrSvc MedicalRecordService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithMedicalRecordSvc(mrSvc, &mockClinicalPlanService{})
	r.PATCH("/medical-records/:id/recommendation-reason", func(c *gin.Context) {
		setClinicID(c)
	}, h.UpdateMedicalRecordRecommendationReason)
	return r
}

func TestPatchMedicalRecordRecommendationReason(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       string
		svc        *mockMedicalRecordService
		wantStatus int
	}{
		{
			name:    "200 success with valid reason revisit",
			paramID: "1",
			body:    `{"reason":"revisit"}`,
			svc: &mockMedicalRecordService{
				updateRecommendationReasonFn: func(_ context.Context, _, _ uint64, _ UpdateRecommendationReasonInput) (*model.MedicalRecord, error) {
					reason := "revisit"
					return &model.MedicalRecord{ID: 1, RecommendationReason: &reason}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "200 success with empty reason clears field",
			paramID: "1",
			body:    `{"reason":""}`,
			svc: &mockMedicalRecordService{
				updateRecommendationReasonFn: func(_ context.Context, _, _ uint64, _ UpdateRecommendationReasonInput) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: 1}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "400 for malformed JSON body",
			paramID:    "1",
			body:       `{invalid}`,
			svc:        &mockMedicalRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 for reason exceeding 100 chars",
			paramID:    "1",
			body:       `{"reason":"` + string(make([]byte, 101)) + `"}`,
			svc:        &mockMedicalRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "404 when record not found",
			paramID: "999",
			body:    `{"reason":"checkup"}`,
			svc: &mockMedicalRecordService{
				updateRecommendationReasonFn: func(_ context.Context, _, _ uint64, _ UpdateRecommendationReasonInput) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapNotFound("medical_record", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "500 on service internal error",
			paramID: "1",
			body:    `{"reason":"exam"}`,
			svc: &mockMedicalRecordService{
				updateRecommendationReasonFn: func(_ context.Context, _, _ uint64, _ UpdateRecommendationReasonInput) (*model.MedicalRecord, error) {
					return nil, apperrors.WrapInternalServerError("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPatchRecommendationReasonRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/medical-records/"+tt.paramID+"/recommendation-reason", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithMedicalRecordSvc(&mockMedicalRecordService{}, &mockClinicalPlanService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"reason":"revisit"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.UpdateMedicalRecordRecommendationReason(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- createMedicalRecordRequest conversion unit tests ----

// TestCreateMedicalRecordRequest_ToServiceInput_VisitType は Path B（visit_type 未送信）で 'revisit' がデフォルト設定され、
// "first" 送信時は 'first' になることを検証する。
func TestCreateMedicalRecordRequest_ToServiceInput_VisitType(t *testing.T) {
	ownerID := "10"
	petID := "20"
	visitDate := "2026-03-24"

	tests := []struct {
		name           string
		inputVisitType string
		wantVisitType  model.VisitType
	}{
		{
			name:           "Path B: visit_type 未送信 → revisit デフォルト",
			inputVisitType: "",
			wantVisitType:  model.VisitTypeRevisit,
		},
		{
			name:           "visit_type=first 送信 → first",
			inputVisitType: "first",
			wantVisitType:  model.VisitTypeFirst,
		},
		{
			name:           "visit_type=revisit 送信 → revisit",
			inputVisitType: "revisit",
			wantVisitType:  model.VisitTypeRevisit,
		},
		{
			name:           "visit_type=unknown 送信 → revisit フォールバック",
			inputVisitType: "unknown",
			wantVisitType:  model.VisitTypeRevisit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &createMedicalRecordRequest{
				OwnerID:   &ownerID,
				PetID:     &petID,
				VisitDate: &visitDate,
				VisitType: tt.inputVisitType,
			}
			input, err := req.toServiceInput(9)
			require.NoError(t, err)
			assert.Equal(t, tt.wantVisitType, input.VisitType)
		})
	}
}

// ---- next_visit_recommended_date clear/set tests ----

// TestUpdateMedicalRecordRequest_NextVisitDate は次回来院推奨日の更新・クリア・省略を確認する。
func TestUpdateMedicalRecordRequest_NextVisitDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		body            any
		wantStatus      int
		wantClear       bool
		wantDatePresent bool
	}{
		{
			name:            "next_visit_recommended_date='' → ClearNextVisitRecommendedDate=true",
			body:            map[string]any{"next_visit_recommended_date": ""},
			wantStatus:      http.StatusOK,
			wantClear:       true,
			wantDatePresent: false,
		},
		{
			name:            "next_visit_recommended_date=valid date → set",
			body:            map[string]any{"next_visit_recommended_date": "2030-01-15"},
			wantStatus:      http.StatusOK,
			wantClear:       false,
			wantDatePresent: true,
		},
		{
			name:            "next_visit_recommended_date omitted → no update (mock returns 200)",
			body:            map[string]any{},
			wantStatus:      http.StatusOK,
			wantClear:       false,
			wantDatePresent: false,
		},
		{
			name:       "next_visit_recommended_date=invalid → 400",
			body:       map[string]any{"next_visit_recommended_date": "not-a-date"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotInput UpdateMedicalRecordInput
			h := newHandlerWithMedicalRecordSvc(&mockMedicalRecordService{
				updateFn: func(_ context.Context, _, _ uint64, input UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
					gotInput = input
					return &model.MedicalRecord{ID: 1}, nil
				},
			}, &mockClinicalPlanService{})

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: "1"}}
			setClinicID(c)
			c.Set("user_id", "1")
			h.UpdateMedicalRecord(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, tt.wantClear, gotInput.ClearNextVisitRecommendedDate)
				if tt.wantDatePresent {
					assert.NotNil(t, gotInput.NextVisitRecommendedDate)
				} else {
					assert.Nil(t, gotInput.NextVisitRecommendedDate)
				}
			}
		})
	}
}

// ---- view permission gate tests ----

// TestListMedicalRecords_ViewPermissionDenied（RequirePermission 経由の 403 検証）は
// lab/vital 先例に倣い移行しない。RequirePermission/EffectivePermission の enforcement は
// injected middleware であり internal/handler 側でテストが維持される（⑤の方針踏襲）。
