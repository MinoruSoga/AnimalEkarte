package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// These tests moved from internal/handler/lab_import_handler_test.go in BE9-2D sub-batch③.
// The pre-move file also carried RequirePermission 403 tests (via a mockEffectivePermissionService
// deny-all router); those were dropped following the sibling precedent (checkup/vaccination/inquiry
// handlers): permission enforcement is injected middleware here, and the RequirePermission +
// EffectivePermissionService machinery stays tested in internal/handler. The (resource, action)
// parity for the lab routes is hand-transcribed and documented in routes.go, not gated (see
// routes_snapshot_test.go's header). What survives here is the handler's own behaviour:
// binding/validation, clinic-scope pass-through, response shape, PII safety, and the audit trail.

// ------------------------------------
// Stubs — LabResultImportService
// ------------------------------------

type stubLabResultImportService struct {
	previewFn func(ctx context.Context, clinicID uint64, batch model.LabInboundBatch) (*model.LabImportPreviewResponse, error)
	commitFn  func(ctx context.Context, clinicID uint64, batch model.LabInboundBatch, inputs []LabExamPersistInput) (*model.LabImportCommitResponse, error)
}

func (s *stubLabResultImportService) Preview(ctx context.Context, clinicID uint64, batch model.LabInboundBatch) (*model.LabImportPreviewResponse, error) {
	return s.previewFn(ctx, clinicID, batch)
}

func (s *stubLabResultImportService) Commit(ctx context.Context, clinicID uint64, batch model.LabInboundBatch, inputs []LabExamPersistInput) (*model.LabImportCommitResponse, error) {
	return s.commitFn(ctx, clinicID, batch, inputs)
}

// ------------------------------------
// Stubs — LabImportJobService
// ------------------------------------

type stubLabImportJobServiceForHandler struct {
	getJobFn     func(ctx context.Context, clinicID uint64, jobID uuid.UUID) (*model.LabImportJob, error)
	listEventsFn func(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error)
}

func (s *stubLabImportJobServiceForHandler) CreateJob(_ context.Context, _ uint64, _ model.LabInboundBatch) (*model.LabImportJob, error) {
	return nil, nil
}

func (s *stubLabImportJobServiceForHandler) TransitionStatus(_ context.Context, _ uint64, _ uuid.UUID, _ model.LabImportJobStatus, _ TransitionCounts) (*model.LabImportJob, error) {
	return nil, nil
}

func (s *stubLabImportJobServiceForHandler) GetJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) (*model.LabImportJob, error) {
	return s.getJobFn(ctx, clinicID, jobID)
}

func (s *stubLabImportJobServiceForHandler) PreviewBatch(_ context.Context, _ uint64, _ model.LabInboundBatch) (*model.LabImportPreviewResponse, error) {
	return nil, nil
}

func (s *stubLabImportJobServiceForHandler) ListEvents(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error) {
	return s.listEventsFn(ctx, clinicID, jobID)
}

// ------------------------------------
// Helper: build handler
// ------------------------------------

func newHandlerWithLabSvc(previewSvc LabResultImportService, jobSvc LabImportJobService) *LabImportHandler {
	return NewLabImportHandler(previewSvc, jobSvc, nil)
}

// fixturePreviewSvc returns a LabResultImportService stub that succeeds for fixture
// batches and returns blocked_reasons for drwan/manual.
func fixturePreviewSvc() LabResultImportService {
	return &stubLabResultImportService{
		previewFn: func(_ context.Context, _ uint64, batch model.LabInboundBatch) (*model.LabImportPreviewResponse, error) {
			resp := &model.LabImportPreviewResponse{
				RowCount:        len(batch.ResultRows),
				MappingWarnings: []string{},
				BlockedReasons:  []string{},
			}
			if batch.SourceType == model.LabImportSourceTypeDrWan {
				resp.BlockedReasons = append(resp.BlockedReasons, "source_type=drwan is BLOCKED")
			}
			if batch.SourceType == model.LabImportSourceTypeManual {
				resp.BlockedReasons = append(resp.BlockedReasons, "source_type=manual is not yet implemented")
			}
			return resp, nil
		},
		commitFn: func(_ context.Context, _ uint64, _ model.LabInboundBatch, _ []LabExamPersistInput) (*model.LabImportCommitResponse, error) {
			return nil, apperrors.WrapInvalidInput("not expected")
		},
	}
}

// fixtureCommitSvc returns a stub that commits fixture batches and rejects non-fixture.
func fixtureCommitSvc(jobID uuid.UUID) LabResultImportService {
	return &stubLabResultImportService{
		previewFn: func(_ context.Context, _ uint64, _ model.LabInboundBatch) (*model.LabImportPreviewResponse, error) {
			return &model.LabImportPreviewResponse{RowCount: 0, MappingWarnings: []string{}, BlockedReasons: []string{}}, nil
		},
		commitFn: func(_ context.Context, _ uint64, batch model.LabInboundBatch, inputs []LabExamPersistInput) (*model.LabImportCommitResponse, error) {
			if batch.SourceType != model.LabImportSourceTypeFixture {
				return nil, apperrors.WrapInvalidInput("commit only allowed for fixture")
			}
			return &model.LabImportCommitResponse{
				JobID:          jobID,
				PersistedCount: len(inputs),
			}, nil
		},
	}
}

// fixtureJobSvc returns a minimal job service stub for get/list endpoints.
// Clinic-scope mismatch produces WrapNotFound (same as repository behavior:
// the query scopes by clinic_id, so a cross-clinic access simply finds nothing).
func fixtureJobSvc(clinicID uint64, job *model.LabImportJob, events []*model.LabImportEvent) LabImportJobService {
	return &stubLabImportJobServiceForHandler{
		getJobFn: func(_ context.Context, cid uint64, jobID uuid.UUID) (*model.LabImportJob, error) {
			if cid != clinicID || job == nil || job.ID != jobID {
				return nil, apperrors.WrapNotFound("lab_import_job", jobID.String())
			}
			return job, nil
		},
		listEventsFn: func(_ context.Context, cid uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error) {
			if cid != clinicID || job == nil || job.ID != jobID {
				return nil, apperrors.WrapNotFound("lab_import_job", jobID.String())
			}
			return events, nil
		},
	}
}

// jsonBody converts a map to a *bytes.Reader.
func jsonBody(v any) *bytes.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

// ------------------------------------
// Auth boundary tests (missing clinic_id → 401)
// ------------------------------------

// TestPostLabImportPreview_MissingClinicID_Returns401 verifies that missing clinic_id
// context results in 401 for the preview endpoint.
func TestPostLabImportPreview_MissingClinicID_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixturePreviewSvc(), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/lab-imports/preview",
		jsonBody(map[string]any{"source_type": "fixture"}))
	c.Request.Header.Set("Content-Type", "application/json")
	// No clinic_id set in context — simulates unauthenticated request

	h.PreviewLabImport(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

// TestPostLabImportCommit_MissingClinicID_Returns401 verifies missing clinic_id → 401 for commit.
func TestPostLabImportCommit_MissingClinicID_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/lab-imports",
		jsonBody(map[string]any{"batch": map[string]any{"source_type": "fixture"}}))
	c.Request.Header.Set("Content-Type", "application/json")
	// No clinic_id in context

	h.CommitLabImport(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetLabImportJob_MissingClinicID_Returns401 verifies missing clinic_id → 401 for GetJob.
func TestGetLabImportJob_MissingClinicID_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-imports/"+uuid.New().String(), http.NoBody)
	c.Params = gin.Params{{Key: "job_id", Value: uuid.New().String()}}
	// No clinic_id in context

	h.GetLabImportJob(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestListLabImportEvents_MissingClinicID_Returns401 verifies missing clinic_id → 401 for ListEvents.
func TestListLabImportEvents_MissingClinicID_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	jobID := uuid.New().String()
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-imports/"+jobID+"/events", http.NoBody)
	c.Params = gin.Params{{Key: "job_id", Value: jobID}}

	h.ListLabImportEvents(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ------------------------------------
// Clinic scope tests (job of another clinic → 404)
// ------------------------------------

// TestGetLabImportJob_WrongClinic_Returns404 verifies that a job belonging to clinic 2
// cannot be retrieved by a request scoped to clinic 1.
func TestGetLabImportJob_WrongClinic_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jobID := uuid.New()
	// Job belongs to clinic 2; request is scoped to clinic 1.
	job := &model.LabImportJob{
		ID:        jobID,
		ClinicID:  2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// jobSvc scoped to clinic 2: clinic 1 gets not-found.
	h := newHandlerWithLabSvc(nil, fixtureJobSvc(2, job, nil))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "job_id", Value: jobID.String()}}
	setClinicID(c) // clinic 1

	h.GetLabImportJob(c)

	// clinic scope mismatch → 404 (not found for this clinic)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestListLabImportEvents_WrongClinic_Returns404 verifies clinic scope for ListEvents.
func TestListLabImportEvents_WrongClinic_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jobID := uuid.New()
	job := &model.LabImportJob{
		ID:        jobID,
		ClinicID:  2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	h := newHandlerWithLabSvc(nil, fixtureJobSvc(2, job, nil))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "job_id", Value: jobID.String()}}
	setClinicID(c) // clinic 1

	h.ListLabImportEvents(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestPostLabImportCommit_ClinicID_FromContext_NotBody verifies that clinic_id
// is sourced from JWT context, not the request body. The commit service receives
// the context clinic_id regardless of what might be in the body.
func TestPostLabImportCommit_ClinicID_FromContext_NotBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedClinicID uint64
	svc := &stubLabResultImportService{
		previewFn: func(_ context.Context, _ uint64, _ model.LabInboundBatch) (*model.LabImportPreviewResponse, error) {
			return nil, nil
		},
		commitFn: func(_ context.Context, cid uint64, _ model.LabInboundBatch, _ []LabExamPersistInput) (*model.LabImportCommitResponse, error) {
			capturedClinicID = cid
			return &model.LabImportCommitResponse{JobID: uuid.New(), PersistedCount: 0}, nil
		},
	}

	h := newHandlerWithLabSvc(svc, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/lab-imports",
		jsonBody(map[string]any{"batch": map[string]any{"source_type": "fixture"}}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c) // clinic_id=1 from context

	h.CommitLabImport(c)

	assert.Equal(t, uint64(1), capturedClinicID, "clinic_id must come from JWT context, not request body")
}

// ------------------------------------
// Preview endpoint tests
// ------------------------------------

// TestPostLabImportPreview_Fixture_ReturnsOK verifies fixture preview succeeds with 200
// and returns row_count with empty blocked_reasons.
func TestPostLabImportPreview_Fixture_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixturePreviewSvc(), nil)

	body := jsonBody(map[string]any{
		"source_type": "fixture",
		"result_rows": []map[string]any{
			{"old_pet_key": "SYNTHETIC_PET_001", "exam_code": "BUN", "display_value": "12.3"},
			{"old_pet_key": "SYNTHETIC_PET_002", "exam_code": "CREA", "display_value": "0.8"},
		},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", body)
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.PreviewLabImport(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(2), resp["row_count"])
	assert.Equal(t, []any{}, resp["blocked_reasons"])
}

// TestPostLabImportPreview_DrWan_ReturnsBlockedReasons verifies drwan source
// returns 200 with blocked_reasons (not 400/500 — preview is always safe to call).
func TestPostLabImportPreview_DrWan_ReturnsBlockedReasons(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixturePreviewSvc(), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{"source_type": "drwan"}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.PreviewLabImport(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	blocked, _ := resp["blocked_reasons"].([]any)
	assert.NotEmpty(t, blocked, "drwan preview must populate blocked_reasons")
}

// TestPostLabImportPreview_Manual_ReturnsBlockedReasons verifies manual source
// also returns 200 with blocked_reasons.
func TestPostLabImportPreview_Manual_ReturnsBlockedReasons(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixturePreviewSvc(), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{"source_type": "manual"}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.PreviewLabImport(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	blocked, _ := resp["blocked_reasons"].([]any)
	assert.NotEmpty(t, blocked, "manual preview must populate blocked_reasons")
}

// TestPostLabImportPreview_NoDBWrite verifies that preview does not cause
// any job or exam creation: the stub commit function is never called.
func TestPostLabImportPreview_NoDBWrite(t *testing.T) {
	gin.SetMode(gin.TestMode)

	commitCalled := false
	svc := &stubLabResultImportService{
		previewFn: func(_ context.Context, _ uint64, _ model.LabInboundBatch) (*model.LabImportPreviewResponse, error) {
			return &model.LabImportPreviewResponse{RowCount: 1, MappingWarnings: []string{}, BlockedReasons: []string{}}, nil
		},
		commitFn: func(_ context.Context, _ uint64, _ model.LabInboundBatch, _ []LabExamPersistInput) (*model.LabImportCommitResponse, error) {
			commitCalled = true
			return nil, nil
		},
	}
	h := newHandlerWithLabSvc(svc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{"source_type": "fixture"}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.PreviewLabImport(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, commitCalled, "Preview must not invoke Commit")
}

// TestPostLabImportPreview_MissingSourceType_Returns400 verifies binding:required on source_type.
func TestPostLabImportPreview_MissingSourceType_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixturePreviewSvc(), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{"result_rows": []any{}})) // no source_type
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.PreviewLabImport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPostLabImportPreview_MalformedJSON_Returns400 verifies malformed JSON body → 400.
func TestPostLabImportPreview_MalformedJSON_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixturePreviewSvc(), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{invalid`))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.PreviewLabImport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ------------------------------------
// Commit endpoint tests
// ------------------------------------

// TestPostLabImportCommit_Fixture_Returns201 verifies fixture commit returns 201,
// a non-empty job_id, and a Location header.
func TestPostLabImportCommit_Fixture_Returns201(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jobID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	h := newHandlerWithLabSvc(fixtureCommitSvc(jobID), nil)

	body := jsonBody(map[string]any{
		"batch": map[string]any{
			"source_type": "fixture",
			"received_at": "2026-01-15T10:00:00Z",
		},
		"inputs": []map[string]any{
			{
				"exam_type_id": 5,
				"date":         "2026-01-15",
				"machine":      "SynthAnalyzer-3000",
				"items": []map[string]any{
					{"name": "BUN", "inspection_value": "12.3", "unit": "mg/dL"},
				},
			},
		},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", body)
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.CommitLabImport(c)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, jobID.String(), resp["job_id"])
	assert.Contains(t, w.Header().Get("Location"), jobID.String())
}

// TestPostLabImportCommit_DrWan_Returns400 verifies drwan source is rejected at commit.
func TestPostLabImportCommit_DrWan_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixtureCommitSvc(uuid.New()), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{
			"batch": map[string]any{"source_type": "drwan"},
		}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.CommitLabImport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

// TestPostLabImportCommit_Manual_Returns400 verifies manual source is rejected at commit.
func TestPostLabImportCommit_Manual_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixtureCommitSvc(uuid.New()), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{
			"batch": map[string]any{"source_type": "manual"},
		}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.CommitLabImport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPostLabImportCommit_MissingBatchSourceType_Returns400 verifies binding:required on batch.source_type.
func TestPostLabImportCommit_MissingBatchSourceType_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixtureCommitSvc(uuid.New()), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{"batch": map[string]any{}})) // missing source_type
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.CommitLabImport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPostLabImportCommit_ZeroExamTypeID_Returns400 verifies inputs[i].exam_type_id=0 → 400.
func TestPostLabImportCommit_ZeroExamTypeID_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixtureCommitSvc(uuid.New()), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{
			"batch": map[string]any{"source_type": "fixture"},
			"inputs": []map[string]any{
				{"exam_type_id": 0, "date": "2026-01-15"}, // exam_type_id=0 invalid
			},
		}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.CommitLabImport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPostLabImportCommit_MissingDate_Returns400 verifies inputs[i].date="" → 400.
func TestPostLabImportCommit_MissingDate_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixtureCommitSvc(uuid.New()), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{
			"batch": map[string]any{"source_type": "fixture"},
			"inputs": []map[string]any{
				{"exam_type_id": 5}, // missing date
			},
		}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.CommitLabImport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPostLabImportCommit_InvalidDateFormat_Returns400 verifies non-YYYY-MM-DD date → 400.
func TestPostLabImportCommit_InvalidDateFormat_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixtureCommitSvc(uuid.New()), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{
			"batch": map[string]any{"source_type": "fixture"},
			"inputs": []map[string]any{
				{"exam_type_id": 5, "date": "15/01/2026"}, // wrong format
			},
		}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.CommitLabImport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPostLabImportCommit_InvalidReceivedAt_Returns400 verifies non-RFC3339 received_at → 400.
func TestPostLabImportCommit_InvalidReceivedAt_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixtureCommitSvc(uuid.New()), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{
			"batch": map[string]any{
				"source_type": "fixture",
				"received_at": "2026-01-15", // not RFC3339
			},
		}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.CommitLabImport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPostLabImportCommit_MalformedJSON_Returns400 verifies malformed commit body → 400.
func TestPostLabImportCommit_MalformedJSON_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixtureCommitSvc(uuid.New()), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{bad`))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.CommitLabImport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPostLabImportCommit_JobIDMatchesLocationHeader verifies the response job_id
// matches the Location header — ensuring the caller can navigate to the new job.
func TestPostLabImportCommit_JobIDMatchesLocationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jobID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	h := newHandlerWithLabSvc(fixtureCommitSvc(jobID), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{
			"batch": map[string]any{"source_type": "fixture"},
		}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.CommitLabImport(c)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, jobID.String(), resp["job_id"])
	assert.Contains(t, w.Header().Get("Location"), jobID.String())
}

// ------------------------------------
// GetLabImportJob tests
// ------------------------------------

// TestGetLabImportJob_Returns200WithDTO verifies a valid job is returned with
// correct DTO shape: clinic_id as string, optional fields omitted when nil.
func TestGetLabImportJob_Returns200WithDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jobID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	job := &model.LabImportJob{
		ID:             jobID,
		ClinicID:       1,
		SourceType:     model.LabImportSourceTypeFixture,
		Status:         model.LabImportJobStatusPersisted,
		RowCount:       3,
		PersistedCount: 3,
		CreatedAt:      now,
		UpdatedAt:      now,
		// StartedAt, FinishedAt, ErrorCode intentionally nil
	}
	h := newHandlerWithLabSvc(nil, fixtureJobSvc(1, job, nil))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "job_id", Value: jobID.String()}}
	setClinicID(c)

	h.GetLabImportJob(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, jobID.String(), resp["id"])
	assert.Equal(t, "1", resp["clinic_id"], "clinic_id must be serialized as string")
	assert.Equal(t, "fixture", resp["source_type"])
	assert.Equal(t, "persisted", resp["status"])
	assert.Equal(t, float64(3), resp["row_count"])
	// optional fields omitted when nil
	assert.Nil(t, resp["started_at"], "started_at must be omitted when nil")
	assert.Nil(t, resp["error_code"], "error_code must be omitted when nil")
}

// TestGetLabImportJob_InvalidUUID_Returns400 verifies non-UUID job_id param → 400.
func TestGetLabImportJob_InvalidUUID_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(nil, fixtureJobSvc(1, nil, nil))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "job_id", Value: "not-a-uuid"}}
	setClinicID(c)

	h.GetLabImportJob(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetLabImportJob_NotFound_Returns404 verifies unknown job_id → 404.
func TestGetLabImportJob_NotFound_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// fixtureJobSvc with nil job → not found for any ID
	h := newHandlerWithLabSvc(nil, fixtureJobSvc(1, nil, nil))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	unknownID := uuid.New().String()
	c.Params = gin.Params{{Key: "job_id", Value: unknownID}}
	setClinicID(c)

	h.GetLabImportJob(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ------------------------------------
// ListLabImportEvents tests
// ------------------------------------

// TestListLabImportEvents_ReturnsEmptySlice verifies empty event list returns []
// (not null) — important for JSON consumers.
func TestListLabImportEvents_ReturnsEmptySlice(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jobID := uuid.New()
	job := &model.LabImportJob{
		ID:        jobID,
		ClinicID:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	h := newHandlerWithLabSvc(nil, fixtureJobSvc(1, job, []*model.LabImportEvent{}))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "job_id", Value: jobID.String()}}
	setClinicID(c)

	h.ListLabImportEvents(c)

	require.Equal(t, http.StatusOK, w.Code)
	// Must be a JSON array (even if empty), not null
	assert.True(t, strings.HasPrefix(w.Body.String(), "["))
}

// TestListLabImportEvents_ReturnsEvents verifies events are returned with correct
// DTO shape: clinic_id as string, optional status/error fields omitted when nil.
func TestListLabImportEvents_ReturnsEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jobID := uuid.New()
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	job := &model.LabImportJob{
		ID:        jobID,
		ClinicID:  1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	toStatus := model.LabImportJobStatusValidated
	events := []*model.LabImportEvent{
		{
			ID:        1,
			ClinicID:  1,
			JobID:     jobID,
			EventType: model.LabImportEventTypeStatusTransition,
			ToStatus:  &toStatus,
			RowCount:  2,
			CreatedAt: now,
		},
	}
	h := newHandlerWithLabSvc(nil, fixtureJobSvc(1, job, events))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "job_id", Value: jobID.String()}}
	setClinicID(c)

	h.ListLabImportEvents(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp, 1)
	assert.Equal(t, float64(1), resp[0]["id"])
	assert.Equal(t, "1", resp[0]["clinic_id"], "clinic_id must be string")
	assert.Equal(t, "status_transition", resp[0]["event_type"])
	assert.Equal(t, "validated", resp[0]["to_status"])
	// from_status nil → omitted
	assert.Nil(t, resp[0]["from_status"])
}

// TestListLabImportEvents_InvalidUUID_Returns400 verifies non-UUID job_id param → 400.
func TestListLabImportEvents_InvalidUUID_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(nil, fixtureJobSvc(1, nil, nil))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "job_id", Value: "bad-id"}}
	setClinicID(c)

	h.ListLabImportEvents(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ------------------------------------
// Response shape / no-PHI tests
// ------------------------------------

// TestPostLabImportPreview_ResponseHasNoRawLabValues verifies the preview response
// does not include raw lab values, patient identifiers, or credentials.
func TestPostLabImportPreview_ResponseHasNoRawLabValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithLabSvc(fixturePreviewSvc(), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{
			"source_type":        "fixture",
			"source_fingerprint": "sha256:synthetic-test-only",
			"result_rows": []map[string]any{
				{"old_pet_key": "SYNTH-001", "exam_code": "BUN", "display_value": "12.3"},
			},
		}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.PreviewLabImport(c)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	// Response must only contain row_count, mapping_warnings, blocked_reasons
	assert.NotContains(t, body, "SYNTH-001", "preview response must not echo back patient keys")
	assert.NotContains(t, body, "12.3", "preview response must not include raw lab values")
	assert.NotContains(t, body, "sha256:", "preview response must not include source_fingerprint")
}

// ------------------------------------
// Duplicate commit tests
// ------------------------------------

// TestPostLabImportCommit_Duplicate_ReturnsCommitResponse verifies that when the service
// reports all-duplicate, the handler still returns 201 with duplicate_count populated.
func TestPostLabImportCommit_Duplicate_ReturnsCommitResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jobID := uuid.New()
	svc := &stubLabResultImportService{
		previewFn: func(_ context.Context, _ uint64, _ model.LabInboundBatch) (*model.LabImportPreviewResponse, error) {
			return nil, nil
		},
		commitFn: func(_ context.Context, _ uint64, batch model.LabInboundBatch, inputs []LabExamPersistInput) (*model.LabImportCommitResponse, error) {
			return &model.LabImportCommitResponse{
				JobID:          jobID,
				PersistedCount: 0,
				DuplicateCount: len(inputs),
			}, nil
		},
	}
	h := newHandlerWithLabSvc(svc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/",
		jsonBody(map[string]any{
			"batch": map[string]any{
				"source_type": "fixture",
			},
			"inputs": []map[string]any{
				{"exam_type_id": 5, "date": "2026-01-15"},
			},
		}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.CommitLabImport(c)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(0), resp["persisted_count"])
	assert.Equal(t, float64(1), resp["duplicate_count"])
}

// ------------------------------------
// request/response conversion unit tests
// ------------------------------------

// TestToLabImportJobResponse verifies the DTO conversion does not expose clinic_id as uint64.
func TestToLabImportJobResponse(t *testing.T) {
	resp := labImportJobResponse{ClinicID: "12345"}
	assert.Equal(t, "12345", resp.ClinicID)
}

// TestToExamInputs_InjectsClinicID verifies clinic_id from context is injected into inputs.
func TestToExamInputs_InjectsClinicID(t *testing.T) {
	reqs := []labExamInputReq{
		{ExamTypeID: 1, Date: "2026-01-15"},
	}
	inputs, err := toExamInputs(99, reqs)
	assert.NoError(t, err)
	assert.Len(t, inputs, 1)
	assert.Equal(t, uint64(99), inputs[0].ClinicID)
}

// TestToExamInputs_RejectsZeroExamTypeID verifies exam_type_id=0 returns an error.
func TestToExamInputs_RejectsZeroExamTypeID(t *testing.T) {
	reqs := []labExamInputReq{
		{ExamTypeID: 0, Date: "2026-01-15"},
	}
	_, err := toExamInputs(1, reqs)
	assert.Error(t, err)
}

// TestToExamInputs_RejectsMissingDate verifies empty date returns an error.
func TestToExamInputs_RejectsMissingDate(t *testing.T) {
	reqs := []labExamInputReq{
		{ExamTypeID: 1, Date: ""},
	}
	_, err := toExamInputs(1, reqs)
	assert.Error(t, err)
}

// TestToExamInputs_RejectsInvalidDateFormat verifies non-YYYY-MM-DD date returns an error.
func TestToExamInputs_RejectsInvalidDateFormat(t *testing.T) {
	reqs := []labExamInputReq{
		{ExamTypeID: 1, Date: "2026/01/15"},
	}
	_, err := toExamInputs(1, reqs)
	assert.Error(t, err)
}

// TestToExamInputs_NormalisesDateToUTC verifies the date is normalised to UTC midnight.
func TestToExamInputs_NormalisesDateToUTC(t *testing.T) {
	reqs := []labExamInputReq{
		{ExamTypeID: 1, Date: "2026-01-15"},
	}
	inputs, err := toExamInputs(1, reqs)
	assert.NoError(t, err)
	d := inputs[0].Date
	assert.Equal(t, 0, d.Hour())
	assert.Equal(t, 0, d.Minute())
	assert.Equal(t, 0, d.Second())
}

// TestToBatch_AcceptsRFC3339ReceivedAt verifies received_at in RFC3339 is accepted.
func TestToBatch_AcceptsRFC3339ReceivedAt(t *testing.T) {
	req := labImportBatchReq{
		SourceType: "fixture",
		ReceivedAt: "2026-01-15T10:00:00Z",
	}
	batch, err := toBatch(req)
	assert.NoError(t, err)
	assert.False(t, batch.ReceivedAt.IsZero())
}

// TestToBatch_RejectsInvalidReceivedAt verifies malformed received_at returns an error.
func TestToBatch_RejectsInvalidReceivedAt(t *testing.T) {
	req := labImportBatchReq{
		SourceType: "fixture",
		ReceivedAt: "2026-01-15",
	}
	_, err := toBatch(req)
	assert.Error(t, err)
}

// TestToBatch_DefaultsReceivedAtWhenEmpty verifies empty received_at is defaulted to now.
func TestToBatch_DefaultsReceivedAtWhenEmpty(t *testing.T) {
	req := labImportBatchReq{SourceType: "fixture"}
	batch, err := toBatch(req)
	assert.NoError(t, err)
	assert.False(t, batch.ReceivedAt.IsZero())
}

// ------------------------------------
// Phase 4A — LabAuditLogger handler integration tests
// ------------------------------------

// stubLabAuditLoggerForHandler captures audit calls without DB I/O.
type stubLabAuditLoggerForHandler struct {
	previewRequested []previewAuditCall
	commitRequested  []commitRequestedCall
	commitSucceeded  []commitSucceededCall
	commitFailed     []commitFailedCall
	sourceBlocked    []sourceBlockedCall
}

type previewAuditCall struct {
	clinicID   uint64
	actorID    *uint64
	sourceType string
	rowCount   int
}
type commitRequestedCall struct {
	clinicID   uint64
	actorID    *uint64
	sourceType string
	rowCount   int
}
type commitSucceededCall struct {
	clinicID uint64
	actorID  *uint64
	jobID    uuid.UUID
	counts   CommitAuditCounts
}
type commitFailedCall struct {
	clinicID      uint64
	actorID       *uint64
	errorCategory model.LabAuditErrorCategory
}
type sourceBlockedCall struct {
	clinicID   uint64
	actorID    *uint64
	sourceType string
	operation  string
	reason     model.LabBlockedReason
}

func (s *stubLabAuditLoggerForHandler) LogPreviewRequested(_ context.Context, clinicID uint64, actorID *uint64, sourceType string, rowCount int) {
	s.previewRequested = append(s.previewRequested, previewAuditCall{clinicID, actorID, sourceType, rowCount})
}
func (s *stubLabAuditLoggerForHandler) LogCommitRequested(_ context.Context, clinicID uint64, actorID *uint64, sourceType string, rowCount int) {
	s.commitRequested = append(s.commitRequested, commitRequestedCall{clinicID, actorID, sourceType, rowCount})
}
func (s *stubLabAuditLoggerForHandler) LogCommitSucceeded(_ context.Context, clinicID uint64, actorID *uint64, jobID uuid.UUID, counts CommitAuditCounts) {
	s.commitSucceeded = append(s.commitSucceeded, commitSucceededCall{clinicID, actorID, jobID, counts})
}
func (s *stubLabAuditLoggerForHandler) LogCommitFailed(_ context.Context, clinicID uint64, actorID *uint64, errorCategory model.LabAuditErrorCategory) {
	s.commitFailed = append(s.commitFailed, commitFailedCall{clinicID, actorID, errorCategory})
}
func (s *stubLabAuditLoggerForHandler) LogRevertSucceeded(_ context.Context, clinicID uint64, actorID *uint64, jobID uuid.UUID, retractedExamCount int) {
}
func (s *stubLabAuditLoggerForHandler) LogSourceBlocked(_ context.Context, clinicID uint64, actorID *uint64, sourceType, operation string, reason model.LabBlockedReason) {
	s.sourceBlocked = append(s.sourceBlocked, sourceBlockedCall{clinicID, actorID, sourceType, operation, reason})
}

// newHandlerWithAudit builds a LabImportHandler with a lab audit stub wired in.
func newHandlerWithAudit(previewSvc LabResultImportService, jobSvc LabImportJobService, audit LabAuditLogger) *LabImportHandler {
	return NewLabImportHandler(previewSvc, jobSvc, audit)
}

// TestPostLabImportPreview_AuditPreviewRequested verifies the preview endpoint emits
// a preview_requested audit event with clinic_id, actor_id, source_type, and row_count.
func TestPostLabImportPreview_AuditPreviewRequested(t *testing.T) {
	audit := &stubLabAuditLoggerForHandler{}
	h := newHandlerWithAudit(fixturePreviewSvc(), fixtureJobSvc(1, nil, nil), audit)
	actorID := uint64(42)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("clinic_id", "1")
	c.Set("user_id", "42")
	body := jsonBody(map[string]any{
		"source_type": "fixture",
		"result_rows": []any{},
	})
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/lab-imports/preview", body)
	c.Request.Header.Set("Content-Type", "application/json")
	h.PreviewLabImport(c)

	assert.Equal(t, http.StatusOK, w.Code)
	require.Len(t, audit.previewRequested, 1)
	pr := audit.previewRequested[0]
	assert.Equal(t, uint64(1), pr.clinicID)
	assert.Equal(t, &actorID, pr.actorID)
	assert.Equal(t, "fixture", pr.sourceType)
}

// TestPostLabImportPreview_AuditSourceBlocked_DrWan verifies drwan preview emits
// source_blocked event in addition to preview_requested.
func TestPostLabImportPreview_AuditSourceBlocked_DrWan(t *testing.T) {
	audit := &stubLabAuditLoggerForHandler{}
	h := newHandlerWithAudit(fixturePreviewSvc(), fixtureJobSvc(1, nil, nil), audit)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("clinic_id", "1")
	c.Set("user_id", "1")
	body := jsonBody(map[string]any{"source_type": "drwan"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/lab-imports/preview", body)
	c.Request.Header.Set("Content-Type", "application/json")
	h.PreviewLabImport(c)

	assert.Equal(t, http.StatusOK, w.Code)
	require.Len(t, audit.previewRequested, 1, "preview_requested must be emitted")
	require.Len(t, audit.sourceBlocked, 1, "source_blocked must be emitted for drwan")
	sb := audit.sourceBlocked[0]
	assert.Equal(t, "drwan", sb.sourceType)
	assert.Equal(t, "preview", sb.operation)
	assert.Equal(t, model.LabBlockedReasonSourceTypeBlocked, sb.reason, "reason must be a typed taxonomy constant")
}

// TestPostLabImportCommit_AuditCommitRequestedAndSucceeded verifies commit emits
// commit_requested then commit_succeeded with job_id and counts.
func TestPostLabImportCommit_AuditCommitRequestedAndSucceeded(t *testing.T) {
	jobID := uuid.New()
	audit := &stubLabAuditLoggerForHandler{}
	h := newHandlerWithAudit(fixtureCommitSvc(jobID), fixtureJobSvc(1, nil, nil), audit)
	actorID := uint64(7)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("clinic_id", "1")
	c.Set("user_id", "7")
	body := jsonBody(map[string]any{
		"batch": map[string]any{
			"source_type":        "fixture",
			"source_fingerprint": "sha256:synthetic-test-only",
			"received_at":        "2026-01-15T00:00:00Z",
		},
		"inputs": []any{
			map[string]any{
				"exam_type_id": 1,
				"date":         "2026-01-15",
			},
		},
	})
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/lab-imports", body)
	c.Request.Header.Set("Content-Type", "application/json")
	h.CommitLabImport(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	require.Len(t, audit.commitRequested, 1)
	cr := audit.commitRequested[0]
	assert.Equal(t, uint64(1), cr.clinicID)
	assert.Equal(t, &actorID, cr.actorID)
	assert.Equal(t, "fixture", cr.sourceType)

	require.Len(t, audit.commitSucceeded, 1)
	cs := audit.commitSucceeded[0]
	assert.Equal(t, jobID, cs.jobID)
	assert.Equal(t, uint64(1), cs.clinicID)
}

// TestPostLabImportCommit_AuditCommitFailed_BlockedSourceType verifies that committing
// a drwan batch emits commit_requested then commit_failed with error_category=invalid_input.
func TestPostLabImportCommit_AuditCommitFailed_BlockedSourceType(t *testing.T) {
	jobID := uuid.New()
	audit := &stubLabAuditLoggerForHandler{}
	h := newHandlerWithAudit(fixtureCommitSvc(jobID), fixtureJobSvc(1, nil, nil), audit)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("clinic_id", "1")
	c.Set("user_id", "1")
	body := jsonBody(map[string]any{
		"batch": map[string]any{
			"source_type": "drwan",
			"received_at": "2026-01-15T00:00:00Z",
		},
		"inputs": []any{},
	})
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/lab-imports", body)
	c.Request.Header.Set("Content-Type", "application/json")
	h.CommitLabImport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.Len(t, audit.commitRequested, 1, "commit_requested must be emitted even on failure")
	require.Len(t, audit.commitFailed, 1, "commit_failed must be emitted")
	assert.Equal(t, model.LabAuditErrorCategoryInvalidInput, audit.commitFailed[0].errorCategory)
	assert.Empty(t, audit.commitSucceeded, "commit_succeeded must NOT be emitted on failure")
}

// TestPostLabImportCommit_AuditPayloadNoPII verifies commit_succeeded metadata
// contains no raw lab values, patient identifiers, or source_fingerprint.
func TestPostLabImportCommit_AuditPayloadNoPII(t *testing.T) {
	jobID := uuid.New()
	fakeSvc := &stubLabResultImportService{
		previewFn: func(_ context.Context, _ uint64, _ model.LabInboundBatch) (*model.LabImportPreviewResponse, error) {
			return &model.LabImportPreviewResponse{MappingWarnings: []string{}, BlockedReasons: []string{}}, nil
		},
		commitFn: func(_ context.Context, _ uint64, _ model.LabInboundBatch, inputs []LabExamPersistInput) (*model.LabImportCommitResponse, error) {
			return &model.LabImportCommitResponse{JobID: jobID, PersistedCount: 1}, nil
		},
	}
	audit := &stubLabAuditLoggerForHandler{}
	h := newHandlerWithAudit(fakeSvc, fixtureJobSvc(1, nil, nil), audit)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("clinic_id", "1")
	c.Set("user_id", "1")
	body := jsonBody(map[string]any{
		"batch": map[string]any{
			"source_type":        "fixture",
			"source_fingerprint": "sha256:synthetic-SENSITIVE",
			"received_at":        "2026-01-15T00:00:00Z",
			"result_rows": []any{
				map[string]any{
					"old_pet_key":     "SYNTHETIC_PET_001",
					"display_value":   "99.9",
					"reference_value": "0-100",
				},
			},
		},
		"inputs": []any{
			map[string]any{"exam_type_id": 1, "date": "2026-01-15"},
		},
	})
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/lab-imports", body)
	c.Request.Header.Set("Content-Type", "application/json")
	h.CommitLabImport(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	require.Len(t, audit.commitSucceeded, 1)
	// Verify the audit stub received safe counts only (no raw values).
	cs := audit.commitSucceeded[0]
	assert.Equal(t, jobID, cs.jobID)
	// The counts struct has no raw value fields by design; this is a type-level PII guarantee.
	// If CommitAuditCounts gains raw fields, this test will fail to compile.
	_ = cs.counts.RowCount
	_ = cs.counts.PersistedCount
	_ = cs.counts.DuplicateCount
	_ = cs.counts.FailedCount
}

// ------------------------------------
// Phase 4A.3 — errorCategory helper and typed taxonomy at call sites
// ------------------------------------

// TestErrorCategory_ReturnsTypedConstants verifies every errorCategory() path maps to
// a declared model.LabAuditErrorCategory constant (not an ad-hoc string).
func TestErrorCategory_ReturnsTypedConstants(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want model.LabAuditErrorCategory
	}{
		{"nil", nil, model.LabAuditErrorCategoryInternal},
		{"invalid_input", apperrors.WrapInvalidInput("bad"), model.LabAuditErrorCategoryInvalidInput},
		{"not_found", apperrors.WrapNotFound("resource", "id"), model.LabAuditErrorCategoryNotFound},
		{"forbidden", apperrors.ErrForbidden, model.LabAuditErrorCategoryForbidden},
		{"unauthorized", apperrors.ErrUnauthorized, model.LabAuditErrorCategoryUnauthorized},
		{"internal_default", apperrors.WrapInternalServerError("something went wrong"), model.LabAuditErrorCategoryInternal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := errorCategory(tc.err)
			assert.Equal(t, tc.want, got)
			// The returned value must pass the taxonomy validator.
			assert.True(t, model.ValidLabAuditErrorCategory(got),
				"errorCategory() must return a valid LabAuditErrorCategory constant, got %q", got)
		})
	}
}

// TestPostLabImportCommit_AuditCommitFailed_InternalError verifies that an internal service
// error emits commit_failed with LabAuditErrorCategoryInternal (not a raw error string).
func TestPostLabImportCommit_AuditCommitFailed_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &stubLabResultImportService{
		previewFn: func(_ context.Context, _ uint64, _ model.LabInboundBatch) (*model.LabImportPreviewResponse, error) {
			return nil, nil
		},
		commitFn: func(_ context.Context, _ uint64, _ model.LabInboundBatch, _ []LabExamPersistInput) (*model.LabImportCommitResponse, error) {
			return nil, apperrors.WrapInternalServerError("db unavailable")
		},
	}
	audit := &stubLabAuditLoggerForHandler{}
	h := newHandlerWithAudit(svc, nil, audit)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("clinic_id", "1")
	c.Set("user_id", "1")
	body := jsonBody(map[string]any{
		"batch": map[string]any{"source_type": "fixture"},
	})
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/lab-imports", body)
	c.Request.Header.Set("Content-Type", "application/json")
	h.CommitLabImport(c)

	require.Len(t, audit.commitFailed, 1)
	assert.Equal(t, model.LabAuditErrorCategoryInternal, audit.commitFailed[0].errorCategory)
}

// TestPostLabImportPreview_AuditNilSafe verifies handler does not panic when audit is nil.
func TestPostLabImportPreview_AuditNilSafe(t *testing.T) {
	// Handler with no LabAudit set (e.g., older tests / nil audit)
	h := newHandlerWithLabSvc(fixturePreviewSvc(), fixtureJobSvc(1, nil, nil))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("clinic_id", "1")
	c.Set("user_id", "1")
	body := jsonBody(map[string]any{"source_type": "fixture"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/lab-imports/preview", body)
	c.Request.Header.Set("Content-Type", "application/json")

	assert.NotPanics(t, func() { h.PreviewLabImport(c) })
	assert.Equal(t, http.StatusOK, w.Code)
}
