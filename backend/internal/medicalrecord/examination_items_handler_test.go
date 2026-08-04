package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock ExaminationService ----

type mockExaminationService struct {
	listFn             func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error)
	getByIDFn          func(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
	getPrintSnapshotFn func(ctx context.Context, clinicID, examinationID uint64, version *uint64) (*ExaminationPrintSnapshot, error)
	createFn           func(ctx context.Context, clinicID uint64, input *CreateExaminationInput) (*model.Examination, error)
	updateFn           func(ctx context.Context, clinicID, id uint64, input UpdateExaminationInput) (*model.Examination, error)
	unconfirmFn        func(ctx context.Context, clinicID, id uint64, input UnconfirmExaminationInput) (*model.Examination, error)
	deleteFn           func(ctx context.Context, clinicID, id uint64, actorID *uint64) error
	listItemsFn        func(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error)
	replaceItemsFn     func(ctx context.Context, clinicID, examID uint64, actorID *uint64, inputs []UpsertExamItemInput) ([]model.ExamResult, error)
}

func (m *mockExaminationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error) {
	return m.listFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (m *mockExaminationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockExaminationService) GetPrintSnapshot(ctx context.Context, clinicID, examinationID uint64, version *uint64) (*ExaminationPrintSnapshot, error) {
	if m.getPrintSnapshotFn != nil {
		return m.getPrintSnapshotFn(ctx, clinicID, examinationID, version)
	}
	return nil, apperrors.WrapNotFound("examination_print_snapshot", "not_configured")
}

func (m *mockExaminationService) Create(ctx context.Context, clinicID uint64, input *CreateExaminationInput) (*model.Examination, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockExaminationService) Update(ctx context.Context, clinicID, id uint64, input UpdateExaminationInput) (*model.Examination, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockExaminationService) Unconfirm(ctx context.Context, clinicID, id uint64, input UnconfirmExaminationInput) (*model.Examination, error) {
	return m.unconfirmFn(ctx, clinicID, id, input)
}

func (m *mockExaminationService) Delete(ctx context.Context, clinicID, id uint64, actorID *uint64) error {
	return m.deleteFn(ctx, clinicID, id, actorID)
}

func (m *mockExaminationService) ListItems(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error) {
	return m.listItemsFn(ctx, clinicID, examID)
}

func (m *mockExaminationService) ReplaceItems(ctx context.Context, clinicID, examID uint64, actorID *uint64, inputs []UpsertExamItemInput) ([]model.ExamResult, error) {
	return m.replaceItemsFn(ctx, clinicID, examID, actorID, inputs)
}

func newHandlerWithExaminationSvc(svc ExaminationService) *ExaminationHandler {
	return NewExaminationHandler(svc)
}

// ---- ListExaminationItems ----

func TestListExaminationItems(t *testing.T) {
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
			name:     "returns items array",
			paramID:  "10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				listItemsFn: func(_ context.Context, _, examID uint64) ([]model.ExamResult, error) {
					assert.Equal(t, uint64(10), examID)
					return []model.ExamResult{
						{ID: 1, ExamID: 10, Name: "WBC", Status: model.ExaminationResultStatusNormal, IsAbnormal: false},
						{ID: 2, ExamID: 10, Name: "RBC", Status: model.ExaminationResultStatusHigh, IsAbnormal: true},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"WBC"`,
		},
		{
			name:     "returns empty items when none exist",
			paramID:  "10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				listItemsFn: func(_ context.Context, _, _ uint64) ([]model.ExamResult, error) {
					return []model.ExamResult{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"items":[]`,
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
			name:     "returns 404 when exam does not exist",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				listItemsFn: func(_ context.Context, _, _ uint64) ([]model.ExamResult, error) {
					return nil, apperrors.WrapNotFound("exam", "999")
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
			h.ListExaminationItems(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ReplaceExaminationItems ----

func TestReplaceExaminationItems(t *testing.T) {
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
			name:    "replaces items and returns saved list",
			paramID: "10",
			body: map[string]any{
				"items": []map[string]any{
					{"name": "WBC", "inspection_value": "5.0"},
					{"name": "RBC", "inspection_value": "11"},
				},
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				replaceItemsFn: func(_ context.Context, clinicID, examID uint64, _ *uint64, inputs []UpsertExamItemInput) ([]model.ExamResult, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), examID)
					assert.Len(t, inputs, 2)
					assert.Equal(t, "WBC", inputs[0].Name)
					assert.Equal(t, "5.0", inputs[0].InspectionValue)
					return []model.ExamResult{
						{ID: 1, ExamID: 10, Name: "WBC", InspectionValue: "5.0", Status: model.ExaminationResultStatusNormal},
						{ID: 2, ExamID: 10, Name: "RBC", InspectionValue: "11", Status: model.ExaminationResultStatusHigh, IsAbnormal: true},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"status":"high"`,
		},
		{
			name:    "ignores client supplied assessed state",
			paramID: "10",
			body: map[string]any{
				"items": []map[string]any{{
					"name":        "WBC",
					"is_assessed": true,
				}},
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				replaceItemsFn: func(_ context.Context, _, _ uint64, _ *uint64, inputs []UpsertExamItemInput) ([]model.ExamResult, error) {
					assert.Len(t, inputs, 1)
					return []model.ExamResult{{
						ID:     1,
						ExamID: 10,
						Name:   "WBC",
						Status: model.ExaminationResultStatusNormal,
					}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"is_assessed":false`,
		},
		{
			name:     "accepts empty items list",
			paramID:  "10",
			body:     map[string]any{"items": []map[string]any{}},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				replaceItemsFn: func(_ context.Context, _, _ uint64, _ *uint64, inputs []UpsertExamItemInput) ([]model.ExamResult, error) {
					assert.Empty(t, inputs)
					return []model.ExamResult{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"items":[]`,
		},
		{
			name:       "returns 400 for malformed JSON",
			paramID:    "10",
			bodyRaw:    `{"items": [`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 400 when item name is missing",
			paramID: "10",
			body: map[string]any{
				"items": []map[string]any{
					{"inspection_value": "5.0"}, // missing required name
				},
			},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "10",
			body:       map[string]any{"items": []map[string]any{}},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockExaminationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when id is non-numeric",
			paramID:    "abc",
			body:       map[string]any{"items": []map[string]any{}},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockExaminationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 when exam is confirmed",
			paramID:  "10",
			body:     map[string]any{"items": []map[string]any{{"name": "WBC"}}},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				replaceItemsFn: func(_ context.Context, _, _ uint64, _ *uint64, _ []UpsertExamItemInput) ([]model.ExamResult, error) {
					return nil, apperrors.WrapConflict("確定済みの検査は編集できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:     "returns 404 when exam does not exist",
			paramID:  "999",
			body:     map[string]any{"items": []map[string]any{{"name": "WBC"}}},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				replaceItemsFn: func(_ context.Context, _, _ uint64, _ *uint64, _ []UpsertExamItemInput) ([]model.ExamResult, error) {
					return nil, apperrors.WrapNotFound("exam", "999")
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

			var body []byte
			if tt.bodyRaw != "" {
				body = []byte(tt.bodyRaw)
			} else {
				var err error
				body, err = json.Marshal(tt.body)
				assert.NoError(t, err)
			}
			c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.ReplaceExaminationItems(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
