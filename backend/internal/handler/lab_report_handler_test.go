package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ------------------------------------
// Stub — LabReportQueryService
// ------------------------------------

type stubLabReportQueryService struct {
	listFn func(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]model.LabExamReportSummary, error)
	getFn  func(ctx context.Context, clinicID uint64, examID uint64) (*model.LabExamReportDetail, error)
}

func (s *stubLabReportQueryService) ListJobReportSummaries(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]model.LabExamReportSummary, error) {
	return s.listFn(ctx, clinicID, jobID)
}

func (s *stubLabReportQueryService) GetExamReport(ctx context.Context, clinicID uint64, examID uint64) (*model.LabExamReportDetail, error) {
	return s.getFn(ctx, clinicID, examID)
}

// ------------------------------------
// Router helpers
// ------------------------------------

// labReportTestRouter builds a router with lab-report routes and a configurable permission service.
// system_admin=false forces RequirePermission middleware to evaluate rules.
func labReportTestRouter(reportSvc service.LabReportQueryService, permSvc service.EffectivePermissionService) *gin.Engine {
	r := gin.New()
	h := &Handler{
		svc: &service.Services{
			LabReportQuery:      reportSvc,
			EffectivePermission: permSvc,
		},
	}
	g := r.Group("/api/v1")
	g.Use(func(c *gin.Context) {
		c.Set("clinic_id", "1")
		c.Set("user_id", "1")
		c.Set("is_system_admin", false)
		c.Next()
	})
	h.RegisterLabReportRoutes(g)
	return r
}

// labReportAdminRouter builds a router with system_admin=true (bypasses permission middleware).
func labReportAdminRouter(reportSvc service.LabReportQueryService) *gin.Engine {
	r := gin.New()
	h := &Handler{
		svc: &service.Services{
			LabReportQuery:      reportSvc,
			EffectivePermission: &mockEffectivePermissionService{},
		},
	}
	g := r.Group("/api/v1")
	g.Use(func(c *gin.Context) {
		c.Set("clinic_id", "1")
		c.Set("user_id", "1")
		c.Set("is_system_admin", true)
		c.Next()
	})
	h.RegisterLabReportRoutes(g)
	return r
}

// syntheticSummary builds a PII-safe LabExamReportSummary for handler tests.
func syntheticSummary(examID uint64, jobID uuid.UUID) model.LabExamReportSummary {
	return model.LabExamReportSummary{
		ExamID:        examID,
		ClinicID:      "1",
		JobID:         &jobID,
		Date:          "2026-01-15",
		ExamTypeName:  "血液化学(合成)",
		Status:        string(model.ExaminationStatusResultEntered),
		ResultCount:   2,
		AbnormalCount: 1,
		Machine:       "Fuji DRI-CHEM",
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
}

// syntheticDetail builds a PII-safe LabExamReportDetail for handler tests.
func syntheticDetail(examID uint64, jobID uuid.UUID) *model.LabExamReportDetail {
	petID := uint64(42)
	return &model.LabExamReportDetail{
		ExamID:       examID,
		ClinicID:     "1",
		JobID:        &jobID,
		PetID:        &petID,
		Date:         "2026-01-15",
		ExamTypeName: "血液化学(合成)",
		Status:       string(model.ExaminationStatusResultEntered),
		Machine:      "Fuji DRI-CHEM",
		Items: []model.LabExamResultItem{
			{Name: "BUN", InspectionValue: "12.3", IsAbnormal: false, Status: string(model.ExaminationResultStatusNormal), SortOrder: 1},
		},
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
}

// ------------------------------------
// Permission enforcement tests
// ------------------------------------

// TestGetLabJobReportSummaries_MissingPermission_Returns403 verifies ResourceLabImport "view"
// is enforced for the summaries endpoint.
func TestGetLabJobReportSummaries_MissingPermission_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := labReportTestRouter(nil, denyAllPermSvc())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/jobs/"+uuid.New().String()+"/summaries", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestGetLabExamReport_MissingPermission_Returns403 verifies ResourceLabImport "view"
// is enforced for the exam detail endpoint.
func TestGetLabExamReport_MissingPermission_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := labReportTestRouter(nil, denyAllPermSvc())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/exams/1", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ------------------------------------
// Missing clinic_id tests (via direct handler call, no router)
// ------------------------------------

// TestGetLabJobReportSummaries_MissingClinicID_Returns401 verifies missing clinic_id → 401.
func TestGetLabJobReportSummaries_MissingClinicID_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{svc: &service.Services{LabReportQuery: nil}}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/jobs/"+uuid.New().String()+"/summaries", http.NoBody)
	c.Params = gin.Params{{Key: "job_id", Value: uuid.New().String()}}
	// No clinic_id in context — simulates unauthenticated request

	h.GetLabJobReportSummaries(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetLabExamReport_MissingClinicID_Returns401 verifies missing clinic_id → 401.
func TestGetLabExamReport_MissingClinicID_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{svc: &service.Services{LabReportQuery: nil}}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/exams/1", http.NoBody)
	c.Params = gin.Params{{Key: "exam_id", Value: "1"}}

	h.GetLabExamReport(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ------------------------------------
// Param validation tests
// ------------------------------------

// TestGetLabJobReportSummaries_InvalidUUID_Returns400 verifies that a non-UUID job_id returns 400.
func TestGetLabJobReportSummaries_InvalidUUID_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{svc: &service.Services{LabReportQuery: &stubLabReportQueryService{
		listFn: func(_ context.Context, _ uint64, _ uuid.UUID) ([]model.LabExamReportSummary, error) {
			return nil, nil
		},
	}}}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/jobs/not-a-uuid/summaries", http.NoBody)
	c.Params = gin.Params{{Key: "job_id", Value: "not-a-uuid"}}
	setClinicID(c)

	h.GetLabJobReportSummaries(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetLabExamReport_NonNumericExamID_Returns400 verifies that a non-integer exam_id returns 400.
func TestGetLabExamReport_NonNumericExamID_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{svc: &service.Services{LabReportQuery: &stubLabReportQueryService{
		getFn: func(_ context.Context, _ uint64, _ uint64) (*model.LabExamReportDetail, error) {
			return nil, nil
		},
	}}}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/exams/abc", http.NoBody)
	c.Params = gin.Params{{Key: "exam_id", Value: "abc"}}
	setClinicID(c)

	h.GetLabExamReport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetLabExamReport_ZeroExamID_Returns400 verifies that exam_id=0 is rejected.
func TestGetLabExamReport_ZeroExamID_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{svc: &service.Services{LabReportQuery: &stubLabReportQueryService{
		getFn: func(_ context.Context, _ uint64, _ uint64) (*model.LabExamReportDetail, error) {
			return nil, nil
		},
	}}}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/exams/0", http.NoBody)
	c.Params = gin.Params{{Key: "exam_id", Value: "0"}}
	setClinicID(c)

	h.GetLabExamReport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ------------------------------------
// Happy-path tests (via router — permission bypassed with admin flag)
// ------------------------------------

// TestGetLabJobReportSummaries_Happy_Returns200 verifies a valid request returns 200
// with a non-empty summaries array.
func TestGetLabJobReportSummaries_Happy_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jobID := uuid.New()
	svc := &stubLabReportQueryService{
		listFn: func(_ context.Context, clinicID uint64, jid uuid.UUID) ([]model.LabExamReportSummary, error) {
			if clinicID != 1 || jid != jobID {
				return nil, apperrors.WrapNotFound("job", jid.String())
			}
			return []model.LabExamReportSummary{syntheticSummary(1, jobID)}, nil
		},
	}
	r := labReportAdminRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/jobs/"+jobID.String()+"/summaries", http.NoBody)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body, 1)
	assert.Equal(t, float64(1), body[0]["exam_id"])
	assert.Equal(t, "血液化学(合成)", body[0]["exam_type_name"])
	assert.Equal(t, "2026-01-15", body[0]["date"])
}

// TestGetLabJobReportSummaries_EmptyResult_Returns200 verifies that an unknown job_id
// returns 200 with an empty array (not 404).
func TestGetLabJobReportSummaries_EmptyResult_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubLabReportQueryService{
		listFn: func(_ context.Context, _ uint64, _ uuid.UUID) ([]model.LabExamReportSummary, error) {
			return []model.LabExamReportSummary{}, nil
		},
	}
	r := labReportAdminRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/jobs/"+uuid.New().String()+"/summaries", http.NoBody)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Len(t, body, 0)
}

// TestGetLabExamReport_Happy_Returns200 verifies a valid exam_id returns 200 with a detail DTO.
func TestGetLabExamReport_Happy_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jobID := uuid.New()
	svc := &stubLabReportQueryService{
		getFn: func(_ context.Context, clinicID uint64, examID uint64) (*model.LabExamReportDetail, error) {
			if clinicID != 1 || examID != 10 {
				return nil, apperrors.WrapNotFound("exam", "")
			}
			return syntheticDetail(10, jobID), nil
		},
	}
	r := labReportAdminRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/exams/10", http.NoBody)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(10), body["exam_id"])
	assert.Equal(t, "血液化学(合成)", body["exam_type_name"])
	items, ok := body["items"].([]any)
	require.True(t, ok, "items must be an array")
	require.Len(t, items, 1)
}

// TestGetLabExamReport_NotFound_Returns404 verifies that a missing exam returns 404.
func TestGetLabExamReport_NotFound_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubLabReportQueryService{
		getFn: func(_ context.Context, _ uint64, _ uint64) (*model.LabExamReportDetail, error) {
			return nil, apperrors.WrapNotFound("exam", "999")
		},
	}
	r := labReportAdminRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/exams/999", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ------------------------------------
// Clinic scope isolation via router
// ------------------------------------

// labReportClinicScopeRouter builds a router that injects a specific clinic_id.
func labReportClinicScopeRouter(clinicID string, reportSvc service.LabReportQueryService) *gin.Engine {
	r := gin.New()
	h := &Handler{
		svc: &service.Services{
			LabReportQuery:      reportSvc,
			EffectivePermission: &mockEffectivePermissionService{},
		},
	}
	g := r.Group("/api/v1")
	g.Use(func(c *gin.Context) {
		c.Set("clinic_id", clinicID)
		c.Set("user_id", "1")
		c.Set("is_system_admin", true) // bypass permission; focus on clinic scope
		c.Next()
	})
	h.RegisterLabReportRoutes(g)
	return r
}

// TestGetLabJobReportSummaries_WrongClinic_ReturnsEmpty verifies that an exam belonging
// to clinic 2 is not returned when the request is scoped to clinic 1.
func TestGetLabJobReportSummaries_WrongClinic_ReturnsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jobID := uuid.New()

	// Service scoped to clinic 2: clinic 1 gets empty result (not an error).
	svc := &stubLabReportQueryService{
		listFn: func(_ context.Context, clinicID uint64, jid uuid.UUID) ([]model.LabExamReportSummary, error) {
			if clinicID != 2 {
				return []model.LabExamReportSummary{}, nil
			}
			return []model.LabExamReportSummary{syntheticSummary(1, jid)}, nil
		},
	}
	r := labReportClinicScopeRouter("1", svc) // request as clinic 1

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/jobs/"+jobID.String()+"/summaries", http.NoBody)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Len(t, body, 0, "clinic 2 exam must not be visible to clinic 1")
}

// TestGetLabExamReport_WrongClinic_Returns404 verifies that an exam belonging to clinic 2
// cannot be retrieved when the request is scoped to clinic 1.
func TestGetLabExamReport_WrongClinic_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Service returns not-found when clinic doesn't match.
	svc := &stubLabReportQueryService{
		getFn: func(_ context.Context, clinicID uint64, examID uint64) (*model.LabExamReportDetail, error) {
			if clinicID != 2 {
				return nil, apperrors.WrapNotFound("exam", "")
			}
			return syntheticDetail(examID, uuid.New()), nil
		},
	}
	r := labReportClinicScopeRouter("1", svc) // request as clinic 1

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/exams/5", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ------------------------------------
// PII-safe response field allowlist
// ------------------------------------

// TestGetLabExamReport_ResponseDoesNotLeakPII verifies that the JSON response
// does not include owner_name, pet_name, or result_summary.
func TestGetLabExamReport_ResponseDoesNotLeakPII(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jobID := uuid.New()
	svc := &stubLabReportQueryService{
		getFn: func(_ context.Context, _ uint64, _ uint64) (*model.LabExamReportDetail, error) {
			return syntheticDetail(1, jobID), nil
		},
	}
	r := labReportAdminRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/exams/1", http.NoBody)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	// PII fields must not appear in the response body.
	assert.NotContains(t, body, "owner_name", "owner_name must not be in response")
	assert.NotContains(t, body, "pet_name", "pet_name must not be in response")
	assert.NotContains(t, body, "result_summary", "result_summary must not be in response")
	assert.NotContains(t, body, "owner_email", "owner_email must not be in response")
	assert.NotContains(t, body, "owner_phone", "owner_phone must not be in response")
	assert.NotContains(t, body, "owner_address", "owner_address must not be in response")
}

// TestGetLabJobReportSummaries_ResponseDoesNotLeakPII verifies that summaries
// do not include PII fields.
func TestGetLabJobReportSummaries_ResponseDoesNotLeakPII(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jobID := uuid.New()
	svc := &stubLabReportQueryService{
		listFn: func(_ context.Context, _ uint64, jid uuid.UUID) ([]model.LabExamReportSummary, error) {
			return []model.LabExamReportSummary{syntheticSummary(1, jid)}, nil
		},
	}
	r := labReportAdminRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lab-reports/jobs/"+jobID.String()+"/summaries", http.NoBody)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	assert.NotContains(t, body, "owner_name")
	assert.NotContains(t, body, "pet_name")
	assert.NotContains(t, body, "result_summary")
}
