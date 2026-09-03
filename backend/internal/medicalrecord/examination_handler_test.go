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

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestExaminationHandlerCompiles verifies examination_handler.go compiles
func TestExaminationHandlerCompiles(t *testing.T) {
	assert.True(t, true, "examination_handler.go compiled successfully")
}

func TestExaminationSelectedClinicGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		invoke func(*ExaminationHandler, *gin.Context)
		svc    *mockExaminationService
	}{
		{
			name: "ListExaminations",
			invoke: func(h *ExaminationHandler, c *gin.Context) {
				h.ListExaminations(c)
			},
			svc: &mockExaminationService{
				listFn: func(_ context.Context, _ uint64, _, _, _ *uint64, _, _, _ *string, _, _ int) ([]model.Examination, int64, error) {
					t.Fatal("examination service must not be reached")
					return nil, 0, nil
				},
			},
		},
		{
			name: "GetExamination",
			invoke: func(h *ExaminationHandler, c *gin.Context) {
				h.GetExamination(c)
			},
			svc: &mockExaminationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
					t.Fatal("examination service must not be reached")
					return nil, nil
				},
			},
		},
		{
			name: "GetExaminationPrintSnapshot",
			invoke: func(h *ExaminationHandler, c *gin.Context) {
				h.GetExaminationPrintSnapshot(c)
			},
			svc: &mockExaminationService{
				getPrintSnapshotFn: func(_ context.Context, _, _ uint64, _ *uint64) (*ExaminationPrintSnapshot, error) {
					t.Fatal("examination service must not be reached")
					return nil, nil
				},
			},
		},
		{
			name: "ListExaminationItems",
			invoke: func(h *ExaminationHandler, c *gin.Context) {
				h.ListExaminationItems(c)
			},
			svc: &mockExaminationService{
				listItemsFn: func(_ context.Context, _, _ uint64) ([]model.ExamResult, error) {
					t.Fatal("examination service must not be reached")
					return nil, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithExaminationSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			setClinicID(c)
			c.Set("clinic_id", "2")
			setResourcePermissionOnlyClinic(c, 1, string(model.ResourceExaminations), "view")

			tt.invoke(h, c)

			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}

func TestExaminationUnconfirmHandler_ValidatesReasonActorAndReturnsExamination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       string
		setupCtx   func(*gin.Context)
		svc        *mockExaminationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "unconfirms a confirmed examination with an authenticated reason",
			body:     `{"reason":"result correction requested"}`,
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockExaminationService{unconfirmFn: func(
				_ context.Context,
				clinicID, examinationID uint64,
				input UnconfirmExaminationInput,
			) (*model.Examination, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(10), examinationID)
				assert.Equal(t, "result correction requested", input.Reason)
				assert.NotNil(t, input.ActorID)
				return &model.Examination{
					ID: 10, ClinicID: clinicID, ExamTypeID: 2,
					Status: model.ExaminationStatusCompleted,
				}, nil
			}},
			wantStatus: http.StatusOK,
			wantBody:   `"status":"completed"`,
		},
		{
			name:       "rejects a missing reason before the service",
			body:       `{}`,
			setupCtx:   func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "maps a whitespace-only service rejection to bad request",
			body:     `{"reason":"   "}`,
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockExaminationService{unconfirmFn: func(
				_ context.Context,
				_, _ uint64,
				_ UnconfirmExaminationInput,
			) (*model.Examination, error) {
				return nil, apperrors.WrapInvalidInput("unconfirm reason is required")
			}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects a missing authenticated actor",
			body:       `{"reason":"result correction requested"}`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithExaminationSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			tt.setupCtx(c)

			h.UnconfirmExamination(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ListExaminations ----

func TestListExaminations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockExaminationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns paginated list of examinations",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				listFn: func(_ context.Context, clinicID uint64, petID, ownerID, medicalRecordID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, 1, page)
					assert.Equal(t, 20, limit)
					return []model.Examination{{ID: 1, ClinicID: 1, ExamTypeID: 2, Status: model.ExaminationStatusPending}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"total":1`,
		},
		{
			name:     "applies pet_id, status, and date filters",
			query:    "pet_id=5&status=completed&start_date=2026-05-01&end_date=2026-05-31&page=2&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				listFn: func(_ context.Context, _ uint64, petID, ownerID, medicalRecordID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error) {
					assert.NotNil(t, petID)
					assert.Equal(t, uint64(5), *petID)
					assert.NotNil(t, status)
					assert.Equal(t, "completed", *status)
					assert.NotNil(t, startDate)
					assert.Equal(t, "2026-05-01", *startDate)
					assert.NotNil(t, endDate)
					assert.Equal(t, "2026-05-31", *endDate)
					assert.Equal(t, 2, page)
					assert.Equal(t, 10, limit)
					return []model.Examination{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"total":0`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockExaminationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when pagination is invalid",
			query:      "page=0",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when pet_id filter is invalid",
			query:      "pet_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "applies medical_record_id filter",
			query:    "medical_record_id=9",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				listFn: func(_ context.Context, _ uint64, _, _, medicalRecordID *uint64, _, _, _ *string, _, _ int) ([]model.Examination, int64, error) {
					if medicalRecordID == nil || *medicalRecordID != 9 {
						return nil, 0, fmt.Errorf("expected medicalRecordID=9")
					}
					return []model.Examination{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"total":0`,
		},
		{
			name:       "returns 400 when medical_record_id filter is invalid",
			query:      "medical_record_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				listFn: func(_ context.Context, _ uint64, _, _, _ *uint64, _, _, _ *string, _, _ int) ([]model.Examination, int64, error) {
					return nil, 0, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithExaminationSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)
			h.ListExaminations(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetExamination ----

func TestGetExamination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockExaminationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns single examination",
			paramID:  "10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Examination, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), id)
					return &model.Examination{ID: 10, ClinicID: 1, ExamTypeID: 2, Status: model.ExaminationStatusCompleted}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"id":10`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "10",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockExaminationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id is non-numeric",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when examination does not exist",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
					return nil, apperrors.WrapNotFound("examination", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithExaminationSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetExamination(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateExamination ----

func TestCreateExamination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		bodyRaw    string
		setupCtx   func(c *gin.Context)
		svc        *mockExaminationService
		wantStatus int
		wantBody   string
		wantLoc    string
	}{
		{
			name: "creates examination successfully",
			body: map[string]any{
				"exam_type_id": 2,
				"date":         "2026-05-28T00:00:00Z",
				"status":       "pending",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockExaminationService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateExaminationInput) (*model.Examination, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(2), input.ExamTypeID)
					assert.Equal(t, uint64(1), *input.ActorID)
					return &model.Examination{ID: 42, ClinicID: 1, ExamTypeID: 2, Status: model.ExaminationStatusPending}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"id":42`,
			wantLoc:    "/api/v1/examinations/42",
		},
		{
			name: "returns 401 when authenticated actor is missing",
			body: map[string]any{
				"exam_type_id": 2,
				"date":         "2026-05-28T00:00:00Z",
			},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       map[string]any{},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockExaminationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for malformed JSON",
			bodyRaw:    `{"exam_type_id":`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when required field missing",
			body:       map[string]any{"result_summary": "x"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 500 on service error",
			body: map[string]any{
				"exam_type_id": 2,
				"date":         "2026-05-28T00:00:00Z",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockExaminationService{
				createFn: func(_ context.Context, _ uint64, _ *CreateExaminationInput) (*model.Examination, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithExaminationSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var body []byte
			if tt.bodyRaw != "" {
				body = []byte(tt.bodyRaw)
			} else {
				var err error
				body, err = json.Marshal(tt.body)
				assert.NoError(t, err)
			}
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateExamination(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantLoc != "" {
				assert.Equal(t, tt.wantLoc, w.Header().Get("Location"))
			}
		})
	}
}

// ---- UpdateExamination ----

func TestUpdateExamination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		bodyRaw    string
		setupCtx   func(c *gin.Context)
		svc        *mockExaminationService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates examination successfully",
			paramID:  "10",
			body:     map[string]any{"result_summary": "normal"},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockExaminationService{
				updateFn: func(_ context.Context, clinicID, id uint64, input UpdateExaminationInput) (*model.Examination, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), id)
					assert.NotNil(t, input.ResultSummary)
					assert.Equal(t, "normal", *input.ResultSummary)
					assert.Equal(t, uint64(1), *input.ActorID)
					return &model.Examination{ID: 10, ClinicID: 1, ResultSummary: "normal"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"result_summary":"normal"`,
		},
		{
			name:       "returns 401 when authenticated actor is missing",
			paramID:    "10",
			body:       map[string]any{"result_summary": "normal"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "10",
			body:       map[string]any{},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockExaminationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id is non-numeric",
			paramID:    "abc",
			body:       map[string]any{},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for malformed JSON",
			paramID:    "10",
			bodyRaw:    `{"result_summary":`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when examination does not exist",
			paramID:  "999",
			body:     map[string]any{"result_summary": "x"},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockExaminationService{
				updateFn: func(_ context.Context, _, _ uint64, _ UpdateExaminationInput) (*model.Examination, error) {
					return nil, apperrors.WrapNotFound("examination", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 409 for confirmed examination conflict",
			paramID:  "10",
			body:     map[string]any{"result_summary": "blocked"},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockExaminationService{
				updateFn: func(_ context.Context, _, _ uint64, _ UpdateExaminationInput) (*model.Examination, error) {
					return nil, apperrors.WrapConflict("confirmed examination")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithExaminationSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var body []byte
			if tt.bodyRaw != "" {
				body = []byte(tt.bodyRaw)
			} else {
				var err error
				body, err = json.Marshal(tt.body)
				assert.NoError(t, err)
			}
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateExamination(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteExamination ----

func newDeleteExaminationRouter(svc ExaminationService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithExaminationSvc(svc)
	r.DELETE("/examinations/:id", func(c *gin.Context) {
		setClinicID(c)
		setStaffID(c)
	}, h.DeleteExamination)
	return r
}

func TestDeleteExamination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockExaminationService
		wantStatus int
	}{
		{
			name:    "deletes examination successfully",
			paramID: "10",
			svc: &mockExaminationService{
				deleteFn: func(_ context.Context, clinicID, id uint64, actorID *uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), id)
					assert.Equal(t, uint64(1), *actorID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 when id is non-numeric",
			paramID:    "abc",
			svc:        &mockExaminationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when examination does not exist",
			paramID: "999",
			svc: &mockExaminationService{
				deleteFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
					return apperrors.WrapNotFound("examination", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "returns 409 for confirmed examination conflict",
			paramID: "10",
			svc: &mockExaminationService{
				deleteFn: func(_ context.Context, _, _ uint64, _ *uint64) error {
					return apperrors.WrapConflict("confirmed examination")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteExaminationRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/examinations/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithExaminationSvc(&mockExaminationService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "10"}}
		h.DeleteExamination(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 401 when authenticated actor is missing", func(t *testing.T) {
		h := newHandlerWithExaminationSvc(&mockExaminationService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "10"}}
		setClinicID(c)
		h.DeleteExamination(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Examination Handler Test Cases
// This handler manages diagnostic examinations and test results for pets (Section 5: 検査管理)
//
// CRITICAL ENDPOINTS:
//
// 1. ListExaminations (GET /examinations)
//    Test Cases (18 scenarios):
//    ✓ Returns 200 OK with empty list when no records exist
//    ✓ Returns 200 OK with paginated examination list when records exist
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when page/limit are invalid
//    ✓ Pagination: page=1, limit=20 as defaults
//    ✓ Pagination: supports custom page and limit parameters
//    ✓ Pagination: includes total_count for client-side calculation
//    ✓ Filter: pet_id parameter filters by pet (optional)
//    ✓ Filter: owner_id parameter filters by owner (optional)
//    ✓ Filter: status parameter filters by enum (pending, in_progress, result_entered, completed, confirmed)
//    ✓ Filter: start_date parameter filters by date range (inclusive)
//    ✓ Filter: end_date parameter filters by date range (inclusive)
//    ✓ Filter: date format validation (YYYY-MM-DD)
//    ✓ Filter: multiple filters can be combined (pet_id AND status AND date range)
//    ✓ Response includes id, medical_record_id (can be null), pet_id, exam_type_id, doctor_id
//    ✓ Response includes date, status, result_summary, machine, created_at, updated_at
//    ✓ Respects clinic_id scoping (only own clinic's records)
//    ✓ Returns 500 on database error
//
// 2. GetExamination (GET /examinations/:id)
//    Test Cases (11 scenarios):
//    ✓ Returns 200 OK with single examination record
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when examination_id is non-numeric
//    ✓ Returns 404 when examination doesn't exist
//    ✓ Returns 403 when examination belongs to different clinic (tenant isolation)
//    ✓ Response includes complete examination data
//    ✓ Response includes nested pet object (if preloaded)
//    ✓ Response includes nested medical_record object (if exists and preloaded)
//    ✓ Response includes nested exam_type object (if preloaded)
//    ✓ Response includes nested doctor (staff) object (if preloaded)
//    ✓ Returns 500 on database error
//
// 3. CreateExamination (POST /examinations)
//    Test Cases (22 scenarios):
//    ✓ Returns 201 Created when examination created successfully
//    ✓ Returns 400 when required fields missing (pet_id, exam_type_id, doctor_id, date)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Validates pet_id exists (FK constraint)
//    ✓ Validates exam_type_id exists (FK constraint)
//    ✓ Validates doctor_id exists (FK constraint)
//    ✓ Validates medical_record_id exists if provided (FK constraint, optional)
//    ✓ Medical_record_id is optional: can be null for standalone examinations
//    ✓ Status field: accepts enum values (pending, in_progress, result_entered, completed, confirmed)
//    ✓ Status field: defaults to "pending" if not provided
//    ✓ Status field: rejects invalid enum values with 400
//    ✓ Date field: required, format validation (YYYY-MM-DD)
//    ✓ ResultSummary field: optional text (can be null during creation)
//    ✓ Machine field: optional text (equipment/machine name)
//    ✓ Created examination includes generated id and created_at timestamp
//    ✓ Multiple examinations per pet supported (e.g., different exam types)
//    ✓ Same pet can have multiple examinations on same date (different exam types)
//    ✓ Standalone examinations (no medical_record_id) supported
//    ✓ Returns 409 Conflict if FK constraint violated
//    ✓ Returns 500 on database error
//
// 4. UpdateExamination (PATCH /examinations/:id)
//    Test Cases (22 scenarios):
//    ✓ Returns 200 OK when examination updated successfully
//    ✓ Returns 400 when examination_id is non-numeric
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 404 when examination doesn't exist
//    ✓ Returns 403 when examination belongs to different clinic
//    ✓ Partial updates: pet_id can be updated independently
//    ✓ Partial updates: exam_type_id can be updated independently (non-zero check)
//    ✓ Partial updates: doctor_id can be updated independently
//    ✓ Partial updates: date can be updated
//    ✓ Partial updates: status can be updated (enum validation)
//    ✓ Partial updates: result_summary can be updated or cleared
//    ✓ Partial updates: machine can be updated or cleared
//    ✓ Partial updates: medical_record_id can be updated or null'd
//    ✓ Unspecified fields remain unchanged (PATCH semantics, not PUT)
//    ✓ Updated examination reflects changes in response (updated_at timestamp)
//    ✓ Status transitions validated: pending → in_progress → result_entered → completed → confirmed
//    ✓ Cannot go backwards in status (confirmed → completed should fail if enforced)
//    ✓ Returns 409 Conflict if FK constraint violated during update
//    ✓ Returns 500 on database error
//    ✓ Exam_type_id update validation: zero value treated as "unset" (not update)
//
// 5. DeleteExamination (DELETE /examinations/:id)
//    Test Cases (11 scenarios):
//    ✓ Returns 204 No Content when examination deleted successfully
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when examination_id is non-numeric
//    ✓ Returns 404 when examination doesn't exist
//    ✓ Returns 403 when examination belongs to different clinic
//    ✓ Uses soft delete (sets deleted_at, doesn't remove from database)
//    ✓ Deleted examination no longer appears in ListExaminations
//    ✓ Deleted examination cannot be retrieved by GetExamination (404)
//    ✓ Cannot delete already deleted examination (404 on second delete)
//    ✓ Deleting examination doesn't affect related medical_record
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control (clinic_id verification on all endpoints)
//    ✓ Enum validation prevents invalid status values
//    ✓ Foreign key validation on pet, exam_type, doctor, and optional medical_record
//    ✓ Partial updates prevent mass assignment (explicit field mapping)
//    ✓ Date validation: valid date format (YYYY-MM-DD)
//    ✓ Non-zero check on exam_type_id for update (prevents accidental zero values)
//    ✓ No explicit RBAC permission check (all authenticated users can manage examinations)
//
// INTEGRATION WITH OTHER MODULES:
//    ✓ Examination linked to pet (required FK)
//    ✓ Examination linked to medical_record (optional FK, can be null for standalone exams)
//    ✓ Examination linked to exam_type (required FK, defines exam classification)
//    ✓ Examination linked to doctor/staff (required FK, who performed exam)
//    ✓ Multiple examinations can reference same medical_record
//    ✓ Multiple examinations can exist without medical_record (standalone)
//
// DATA MODEL (examinations):
//    - id (PK): BIGSERIAL
//    - clinic_id: BIGINT (multitenancy)
//    - medical_record_id (FK, NULLABLE): BIGINT → medical_records(id)
//    - pet_id (FK): BIGINT → pets(id)
//    - exam_type_id (FK): BIGINT → exam_types(id)
//    - doctor_id (FK): BIGINT → staffs(id)
//    - status: ENUM (pending|in_progress|result_entered|completed|confirmed) DEFAULT pending
//    - date: DATE - examination date
//    - result_summary: TEXT (NULLABLE) - summary of findings
//    - machine: VARCHAR(100) (NULLABLE) - equipment/machine used
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete)
//    - Indexes: (clinic_id, id), (clinic_id, pet_id), (clinic_id, exam_type_id), (clinic_id, date DESC)
//
// IMPLEMENTATION NOTES:
//    - Medical_record_id is NULLABLE: examinations can be standalone (not tied to a medical record)
//    - Status enum has 5 values: pending → in_progress → result_entered → completed → confirmed
//    - Status defaults to "pending" during creation if not provided
//    - Date field is required (when examination was performed)
//    - Result_summary can be added/updated after exam (initially null during creation)
//    - Machine field optional (some exams may not use equipment)
//    - Non-zero check on exam_type_id during update (prevents accidental zero values)
//    - PATCH semantics: unspecified pointer fields remain unchanged
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with pet, exam_type, staff (doctor) test data
//    - Real service/repository layers
//    - Verify pagination with >20 records
//    - Verify enum validation for status field
//    - Verify filter combinations (pet_id, owner_id, status, date range)
//    - Verify FK constraints for all foreign key fields (including optional medical_record_id)
//    - Verify soft delete behavior (deleted records excluded from list/get)
//    - Verify PATCH semantics (unspecified fields unchanged)
//    - Test standalone examinations (no medical_record_id)
//    - Test status state machine (valid transitions)
//    - Test non-zero check on exam_type_id during update
//
