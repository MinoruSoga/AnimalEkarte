package medicalrecord

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

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// These tests moved from internal/handler/lab_report_handler_test.go in BE9-2D sub-batch③.
// The pre-move RequirePermission 403 tests (mockEffectivePermissionService deny-all router) were
// dropped following the sibling precedent — permission enforcement is injected middleware here,
// tested in internal/handler; the report routes' ResourceLabImport "view" parity is documented in
// routes.go. The admin-bypass routers were replaced by direct handler calls (equivalent: no
// permission middleware to bypass). What survives is the handler behaviour: clinic-scope
// pass-through, param validation, response shape, and the PII field allowlists.

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

func (s *stubLabReportQueryService) GetExamReport(ctx context.Context, clinicID, examID uint64) (*model.LabExamReportDetail, error) {
	return s.getFn(ctx, clinicID, examID)
}

// ------------------------------------
// Fixtures
// ------------------------------------

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

// newReportContext builds a gin.Context with clinic_id=1 and the given URL param, for direct
// handler invocation (the handlers write via c.JSON, so no gin.Engine round-trip is needed).
func newReportContext(w *httptest.ResponseRecorder, paramKey, paramVal string) *gin.Context {
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: paramKey, Value: paramVal}}
	setClinicID(c)
	return c
}

// ------------------------------------
// Missing clinic_id tests (401)
// ------------------------------------

// TestGetLabJobReportSummaries_MissingClinicID_Returns401 verifies missing clinic_id → 401.
func TestGetLabJobReportSummaries_MissingClinicID_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLabReportHandler(nil)

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
	h := NewLabReportHandler(nil)

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
	h := NewLabReportHandler(&stubLabReportQueryService{
		listFn: func(_ context.Context, _ uint64, _ uuid.UUID) ([]model.LabExamReportSummary, error) {
			return nil, nil
		},
	})

	w := httptest.NewRecorder()
	c := newReportContext(w, "job_id", "not-a-uuid")

	h.GetLabJobReportSummaries(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetLabExamReport_NonNumericExamID_Returns400 verifies that a non-integer exam_id returns 400.
func TestGetLabExamReport_NonNumericExamID_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLabReportHandler(&stubLabReportQueryService{
		getFn: func(_ context.Context, _ uint64, _ uint64) (*model.LabExamReportDetail, error) {
			return nil, nil
		},
	})

	w := httptest.NewRecorder()
	c := newReportContext(w, "exam_id", "abc")

	h.GetLabExamReport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetLabExamReport_ZeroExamID_Returns400 verifies that exam_id=0 is rejected.
func TestGetLabExamReport_ZeroExamID_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLabReportHandler(&stubLabReportQueryService{
		getFn: func(_ context.Context, _ uint64, _ uint64) (*model.LabExamReportDetail, error) {
			return nil, nil
		},
	})

	w := httptest.NewRecorder()
	c := newReportContext(w, "exam_id", "0")

	h.GetLabExamReport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ------------------------------------
// Happy-path tests
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
	h := NewLabReportHandler(svc)

	w := httptest.NewRecorder()
	c := newReportContext(w, "job_id", jobID.String())

	h.GetLabJobReportSummaries(c)

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
	h := NewLabReportHandler(svc)

	w := httptest.NewRecorder()
	c := newReportContext(w, "job_id", uuid.New().String())

	h.GetLabJobReportSummaries(c)

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
	h := NewLabReportHandler(svc)

	w := httptest.NewRecorder()
	c := newReportContext(w, "exam_id", "10")

	h.GetLabExamReport(c)

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
	h := NewLabReportHandler(svc)

	w := httptest.NewRecorder()
	c := newReportContext(w, "exam_id", "999")

	h.GetLabExamReport(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ------------------------------------
// Clinic scope isolation
// ------------------------------------

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
	h := NewLabReportHandler(svc)

	w := httptest.NewRecorder()
	c := newReportContext(w, "job_id", jobID.String()) // request as clinic 1

	h.GetLabJobReportSummaries(c)

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
	h := NewLabReportHandler(svc)

	w := httptest.NewRecorder()
	c := newReportContext(w, "exam_id", "5") // request as clinic 1

	h.GetLabExamReport(c)

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
	h := NewLabReportHandler(svc)

	w := httptest.NewRecorder()
	c := newReportContext(w, "exam_id", "1")

	h.GetLabExamReport(c)

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
	h := NewLabReportHandler(svc)

	w := httptest.NewRecorder()
	c := newReportContext(w, "job_id", jobID.String())

	h.GetLabJobReportSummaries(c)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	assert.NotContains(t, body, "owner_name")
	assert.NotContains(t, body, "pet_name")
	assert.NotContains(t, body, "result_summary")
}

// ------------------------------------
// Exact JSON field allowlist tests (Phase 4B.3 hardening)
// ------------------------------------

// summarySafeKeys is the exact set of JSON keys allowed in a LabExamReportSummary response.
// Any key outside this set is a contract violation.
var summarySafeKeys = map[string]struct{}{
	"exam_id":        {},
	"clinic_id":      {},
	"job_id":         {},
	"date":           {},
	"exam_type_name": {},
	"status":         {},
	"result_count":   {},
	"abnormal_count": {},
	"machine":        {},
	"created_at":     {},
}

// detailSafeKeys is the exact set of JSON keys allowed in a LabExamReportDetail response.
var detailSafeKeys = map[string]struct{}{
	"exam_id":           {},
	"clinic_id":         {},
	"job_id":            {},
	"pet_id":            {},
	"medical_record_id": {},
	"doctor_id":         {},
	"date":              {},
	"exam_type_name":    {},
	"status":            {},
	"machine":           {},
	"items":             {},
	"created_at":        {},
	"updated_at":        {},
}

// itemSafeKeys is the exact set of JSON keys allowed in each LabExamResultItem.
var itemSafeKeys = map[string]struct{}{
	"name":             {},
	"inspection_value": {},
	"normal_value":     {},
	"unit":             {},
	"reference_value":  {},
	"ref_min":          {},
	"ref_max":          {},
	"is_assessed":      {},
	"is_abnormal":      {},
	"status":           {},
	"sort_order":       {},
}

// TestGetLabJobReportSummaries_ExactFieldAllowlist verifies that the summary
// response contains exactly the approved keys and no others.
// Keys absent from summarySafeKeys indicate a contract violation (e.g., PII leak).
func TestGetLabJobReportSummaries_ExactFieldAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jobID := uuid.New()
	svc := &stubLabReportQueryService{
		listFn: func(_ context.Context, _ uint64, jid uuid.UUID) ([]model.LabExamReportSummary, error) {
			return []model.LabExamReportSummary{syntheticSummary(1, jid)}, nil
		},
	}
	h := NewLabReportHandler(svc)

	w := httptest.NewRecorder()
	c := newReportContext(w, "job_id", jobID.String())

	h.GetLabJobReportSummaries(c)

	require.Equal(t, http.StatusOK, w.Code)

	var body []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body, 1)

	for key := range body[0] {
		_, allowed := summarySafeKeys[key]
		assert.True(t, allowed, "unexpected key in summary response: %q (not in allowlist)", key)
	}
}

// TestGetLabExamReport_ExactFieldAllowlist verifies that the detail response
// contains exactly the approved top-level keys and no others.
func TestGetLabExamReport_ExactFieldAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jobID := uuid.New()
	svc := &stubLabReportQueryService{
		getFn: func(_ context.Context, _ uint64, _ uint64) (*model.LabExamReportDetail, error) {
			return syntheticDetail(1, jobID), nil
		},
	}
	h := NewLabReportHandler(svc)

	w := httptest.NewRecorder()
	c := newReportContext(w, "exam_id", "1")

	h.GetLabExamReport(c)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	for key := range body {
		_, allowed := detailSafeKeys[key]
		assert.True(t, allowed, "unexpected key in detail response: %q (not in allowlist)", key)
	}

	// Verify items keys are also within allowlist.
	items, ok := body["items"].([]any)
	require.True(t, ok, "items must be an array")
	for _, rawItem := range items {
		item, ok := rawItem.(map[string]any)
		require.True(t, ok)
		for key := range item {
			_, allowed := itemSafeKeys[key]
			assert.True(t, allowed, "unexpected key in item: %q (not in allowlist)", key)
		}
	}
}

// TestGetLabExamReport_ExactFieldAllowlist_NilOptionalFields verifies that
// optional omitempty fields (job_id, pet_id, medical_record_id, doctor_id)
// are absent from the JSON body when nil, and the remaining keys still satisfy
// the allowlist.
func TestGetLabExamReport_ExactFieldAllowlist_NilOptionalFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &stubLabReportQueryService{
		getFn: func(_ context.Context, _ uint64, _ uint64) (*model.LabExamReportDetail, error) {
			return &model.LabExamReportDetail{
				ExamID:       1,
				ClinicID:     "1",
				JobID:        nil, // omitempty — job was deleted (ON DELETE SET NULL)
				PetID:        nil,
				Date:         "2026-01-15",
				ExamTypeName: "血液化学",
				Status:       string(model.ExaminationStatusResultEntered),
				Machine:      "Fuji DRI-CHEM",
				Items:        []model.LabExamResultItem{},
				CreatedAt:    "2026-01-15T09:00:00+09:00",
				UpdatedAt:    "2026-01-15T09:00:00+09:00",
			}, nil
		},
	}
	h := NewLabReportHandler(svc)

	w := httptest.NewRecorder()
	c := newReportContext(w, "exam_id", "1")

	h.GetLabExamReport(c)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	// job_id, pet_id, medical_record_id, doctor_id must be absent when nil (omitempty).
	assert.NotContains(t, body, "job_id", "nil job_id must be absent (omitempty) — models ON DELETE SET NULL state")
	assert.NotContains(t, body, "pet_id", "nil pet_id must be absent (omitempty)")

	// All present keys must be in allowlist.
	for key := range body {
		_, allowed := detailSafeKeys[key]
		assert.True(t, allowed, "unexpected key in detail response (nil-optional case): %q", key)
	}
}

// SEC-CODEX-UHQPM2 selected-clinic grant
func TestLabReportSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		invoke func(*LabReportHandler, *gin.Context)
		query  *stubLabReportQueryService
	}{
		{
			name: "GetLabJobReportSummaries returns 403 when selected clinic lacks lab-import view grant",
			invoke: func(h *LabReportHandler, c *gin.Context) {
				h.GetLabJobReportSummaries(c)
			},
			query: &stubLabReportQueryService{
				listFn: func(_ context.Context, _ uint64, _ uuid.UUID) ([]model.LabExamReportSummary, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
		{
			name: "GetLabExamReport returns 403 when selected clinic lacks lab-import view grant",
			invoke: func(h *LabReportHandler, c *gin.Context) {
				h.GetLabExamReport(c)
			},
			query: &stubLabReportQueryService{
				getFn: func(_ context.Context, _, _ uint64) (*model.LabExamReportDetail, error) {
					t.Fatal("service must not be reached")
					return nil, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewLabReportHandler(tt.query)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "job_id", Value: uuid.New().String()}, {Key: "exam_id", Value: "10"}}
			setClinicID(c)
			c.Set("clinic_id", "2")
			c.Set("is_system_admin", false)
			setResourcePermissionOnlyClinic(c, 1, string(model.ResourceLabImport), "view")
			tt.invoke(h, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}
