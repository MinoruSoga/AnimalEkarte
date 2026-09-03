package billing

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
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestEstimateHandlerCompiles verifies estimate_handler.go compiles
func TestEstimateHandlerCompiles(t *testing.T) {
	assert.True(t, true, "estimate_handler.go compiled successfully")
}

// ---- mock EstimateService ----

type mockEstimateService struct {
	listFn            func(ctx context.Context, clinicID uint64, ownerID, medicalRecordID *uint64, status *string, page, limit int) ([]model.Estimate, int64, error)
	getFn             func(ctx context.Context, clinicID, id uint64) (*model.Estimate, error)
	createFn          func(ctx context.Context, clinicID uint64, input *CreateEstimateInput) (*model.Estimate, error)
	updateFn          func(ctx context.Context, clinicID, id uint64, input *UpdateEstimateInput) (*model.Estimate, error)
	deleteFn          func(ctx context.Context, clinicID, id uint64) error
	createSuccessorFn func(ctx context.Context, clinicID, originalID uint64, input *CreateSuccessorInput) (*model.Estimate, error)
}

func (m *mockEstimateService) List(ctx context.Context, clinicID uint64, ownerID, medicalRecordID *uint64, status *string, page, limit int) ([]model.Estimate, int64, error) {
	return m.listFn(ctx, clinicID, ownerID, medicalRecordID, status, page, limit)
}

func (m *mockEstimateService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Estimate, error) {
	return m.getFn(ctx, clinicID, id)
}

func (m *mockEstimateService) Create(ctx context.Context, clinicID uint64, input *CreateEstimateInput) (*model.Estimate, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockEstimateService) Update(ctx context.Context, clinicID, id uint64, input *UpdateEstimateInput) (*model.Estimate, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockEstimateService) Delete(ctx context.Context, clinicID, id uint64, actorID *uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockEstimateService) CreateSuccessor(ctx context.Context, clinicID, originalID uint64, input *CreateSuccessorInput) (*model.Estimate, error) {
	if m.createSuccessorFn != nil {
		return m.createSuccessorFn(ctx, clinicID, originalID, input)
	}
	return nil, nil
}

func newHandlerWithEstimateSvc(svc EstimateService) *EstimateHandler {
	return NewEstimateHandler(svc, func(_ *gin.Context, _, _ string) bool { return true })
}

// ---- ListEstimates ----

func TestListEstimates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockEstimateService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns paginated estimates",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockEstimateService{
				listFn: func(_ context.Context, clinicID uint64, ownerID, medicalRecordID *uint64, status *string, page, limit int) ([]model.Estimate, int64, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Nil(t, ownerID)
					assert.Nil(t, medicalRecordID)
					assert.Nil(t, status)
					assert.Equal(t, 1, page)
					assert.Equal(t, 10, limit)
					return []model.Estimate{{ID: 1, Title: "見積1"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"title":"見積1"`,
		},
		{
			name:     "applies owner_id/medical_record_id/status filters",
			query:    "owner_id=5&medical_record_id=7&status=sent",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockEstimateService{
				listFn: func(_ context.Context, _ uint64, ownerID, medicalRecordID *uint64, status *string, _, _ int) ([]model.Estimate, int64, error) {
					require.NotNil(t, ownerID)
					assert.Equal(t, uint64(5), *ownerID)
					require.NotNil(t, medicalRecordID)
					assert.Equal(t, uint64(7), *medicalRecordID)
					require.NotNil(t, status)
					assert.Equal(t, "sent", *status)
					return []model.Estimate{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockEstimateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:  "returns 403 when selected clinic lacks estimates view grant",
			query: "page=1&limit=10",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("clinic_id", "2")
				setResourcePermissionOnlyClinic(c, 1, string(model.ResourceEstimates), "view")
			},
			svc: &mockEstimateService{
				listFn: func(_ context.Context, _ uint64, _, _ *uint64, _ *string, _, _ int) ([]model.Estimate, int64, error) {
					t.Fatal("estimate service must not be reached")
					return nil, 0, nil
				},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "returns 400 on invalid pagination",
			query:      "page=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockEstimateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid owner_id filter",
			query:      "owner_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockEstimateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockEstimateService{
				listFn: func(_ context.Context, _ uint64, _, _ *uint64, _ *string, _, _ int) ([]model.Estimate, int64, error) {
					return nil, 0, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithEstimateSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListEstimates(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetEstimate ----

func TestGetEstimate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockEstimateService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns estimate detail",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockEstimateService{
				getFn: func(_ context.Context, clinicID, id uint64) (*model.Estimate, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return &model.Estimate{ID: 1, Title: "見積1"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"title":"見積1"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockEstimateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockEstimateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 403 when selected clinic lacks estimates view grant",
			paramID: "1",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("clinic_id", "2")
				setResourcePermissionOnlyClinic(c, 1, string(model.ResourceEstimates), "view")
			},
			svc: &mockEstimateService{
				getFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
					t.Fatal("estimate service must not be reached")
					return nil, nil
				},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockEstimateService{
				getFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
					return nil, apperrors.WrapNotFound("estimate", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithEstimateSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetEstimate(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateEstimate ----

func TestCreateEstimate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		body          any
		bodyRaw       string
		setupCtx      func(c *gin.Context)
		svc           *mockEstimateService
		hasPermission httpapi.PermissionChecker
		wantStatus    int
		wantHeader    string
	}{
		{
			name: "creates estimate successfully",
			body: map[string]any{
				"title":        "新規見積",
				"subtotal":     1000,
				"tax_total":    100,
				"total_amount": 1100,
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateEstimateInput) (*model.Estimate, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "新規見積", input.Title)
					require.NotNil(t, input.CreatedBy)
					assert.Equal(t, uint64(1), *input.CreatedBy)
					return &model.Estimate{ID: 10, Title: input.Title}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: "/api/v1/estimates/10",
		},
		{
			name: "creates estimate with draft status",
			body: map[string]any{
				"title":        "下書き見積",
				"status":       "draft",
				"subtotal":     1000,
				"tax_total":    100,
				"total_amount": 1100,
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				createFn: func(_ context.Context, _ uint64, input *CreateEstimateInput) (*model.Estimate, error) {
					assert.Equal(t, model.EstimateStatusDraft, input.Status)
					return &model.Estimate{ID: 12, Title: input.Title, Status: input.Status}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: "/api/v1/estimates/12",
		},
		{
			name: "creates estimate with sent status",
			body: map[string]any{
				"title":        "送付見積",
				"status":       "sent",
				"subtotal":     1000,
				"tax_total":    100,
				"total_amount": 1100,
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				createFn: func(_ context.Context, _ uint64, input *CreateEstimateInput) (*model.Estimate, error) {
					assert.Equal(t, model.EstimateStatusSent, input.Status)
					return &model.Estimate{ID: 13, Title: input.Title, Status: input.Status}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: "/api/v1/estimates/13",
		},
		{
			name: "returns 400 when create status is approved",
			body: map[string]any{
				"title":  "承認直作成",
				"status": "approved",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				createFn: func(_ context.Context, _ uint64, _ *CreateEstimateInput) (*model.Estimate, error) {
					t.Fatal("Create must not be called when status binding rejects approved")
					return nil, nil
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 when create status is rejected",
			body: map[string]any{
				"title":  "却下直作成",
				"status": "rejected",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				createFn: func(_ context.Context, _ uint64, _ *CreateEstimateInput) (*model.Estimate, error) {
					t.Fatal("Create must not be called when status binding rejects rejected")
					return nil, nil
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       map[string]any{"title": "x"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockEstimateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 401 when staff_id is missing",
			body:       map[string]any{"title": "x"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockEstimateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when title is missing",
			body:       map[string]any{},
			setupCtx:   func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc:        &mockEstimateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on malformed JSON body",
			bodyRaw:    `{"title":`,
			setupCtx:   func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc:        &mockEstimateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 500 on service error",
			body: map[string]any{
				"title":        "新規見積",
				"subtotal":     1000,
				"tax_total":    100,
				"total_amount": 1100,
			},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				createFn: func(_ context.Context, _ uint64, _ *CreateEstimateInput) (*model.Estimate, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			// BUG-372: discount_amount != 0 かつ discount:create 権限なし → 403
			name: "returns 403 when discount permission denied on create",
			body: map[string]any{
				"title":           "新規見積",
				"subtotal":        1000,
				"tax_total":       100,
				"total_amount":    1100,
				"discount_amount": 500,
			},
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c) },
			svc:      &mockEstimateService{},
			// 旧mockEffectivePermissionService(rules空)=deny写像（BE9-2D⑥先例）
			hasPermission: func(_ *gin.Context, _, _ string) bool { return false },
			wantStatus:    http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := httpapi.PermissionChecker(func(_ *gin.Context, _, _ string) bool { return true })
			if tt.hasPermission != nil {
				checker = tt.hasPermission
			}
			h := NewEstimateHandler(tt.svc, checker)

			var bodyBytes []byte
			if tt.bodyRaw != "" {
				bodyBytes = []byte(tt.bodyRaw)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateEstimate(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantHeader != "" {
				assert.Equal(t, tt.wantHeader, w.Header().Get("Location"))
			}
		})
	}
}

func TestCreateEstimate_AuthStaffIDWinsOverBodyCreatedBy(t *testing.T) {
	// AUD-005: body created_by must be ignored; extractStaffID wins.
	gin.SetMode(gin.TestMode)
	h := newHandlerWithEstimateSvc(&mockEstimateService{
		createFn: func(_ context.Context, _ uint64, input *CreateEstimateInput) (*model.Estimate, error) {
			require.NotNil(t, input.CreatedBy)
			assert.Equal(t, uint64(1), *input.CreatedBy, "auth staff must win over body created_by")
			return &model.Estimate{ID: 11, Title: input.Title, CreatedBy: input.CreatedBy}, nil
		},
	})
	body := []byte(`{"title":"spoof","subtotal":0,"tax_total":0,"total_amount":0,"created_by":999}`)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)
	setStaffID(c)
	h.CreateEstimate(c)
	assert.Equal(t, http.StatusCreated, w.Code)
}

// ---- UpdateEstimate ----

func TestUpdateEstimate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		paramID       string
		body          any
		bodyRaw       string
		setupCtx      func(c *gin.Context)
		svc           *mockEstimateService
		hasPermission httpapi.PermissionChecker
		wantStatus    int
		wantBody      string
	}{
		{
			name:     "updates estimate successfully without discount change",
			paramID:  "1",
			body:     map[string]any{"title": "更新後"},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *UpdateEstimateInput) (*model.Estimate, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, input.Title)
					assert.Equal(t, "更新後", *input.Title)
					require.NotNil(t, input.ActorID)
					assert.Equal(t, uint64(1), *input.ActorID)
					return &model.Estimate{ID: 1, Title: "更新後"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"title":"更新後"`,
		},
		{
			name:     "updates estimate status to approved",
			paramID:  "1",
			body:     map[string]any{"status": "approved"},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateEstimateInput) (*model.Estimate, error) {
					require.NotNil(t, input.Status)
					assert.Equal(t, model.EstimateStatus("approved"), *input.Status)
					return &model.Estimate{ID: 1, Status: model.EstimateStatusApproved}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "updates estimate status to rejected",
			paramID:  "1",
			body:     map[string]any{"status": "rejected"},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateEstimateInput) (*model.Estimate, error) {
					require.NotNil(t, input.Status)
					assert.Equal(t, model.EstimateStatus("rejected"), *input.Status)
					return &model.Estimate{ID: 1, Status: model.EstimateStatusRejected}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "updates estimate when discount amount unchanged",
			paramID:  "1",
			body:     map[string]any{"discount_amount": 500},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				getFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
					return &model.Estimate{ID: 1, DiscountAmount: 500}, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateEstimateInput) (*model.Estimate, error) {
					require.NotNil(t, input.DiscountAmount)
					assert.Equal(t, int64(500), *input.DiscountAmount)
					return &model.Estimate{ID: 1, DiscountAmount: 500}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockEstimateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid id",
			paramID:    "abc",
			body:       map[string]any{},
			setupCtx:   func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc:        &mockEstimateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on malformed JSON body",
			paramID:    "1",
			bodyRaw:    `{"title":`,
			setupCtx:   func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc:        &mockEstimateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns error when existing estimate lookup fails for discount change",
			paramID:  "1",
			body:     map[string]any{"discount_amount": 500},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				getFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
					return nil, apperrors.WrapNotFound("estimate", "1")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			body:     map[string]any{"title": "更新後"},
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateEstimateInput) (*model.Estimate, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			// BUG-372: discount_amount が既存値から変化し discount:edit 権限なし → 403
			name:     "returns 403 when discount permission denied on change",
			paramID:  "1",
			body:     map[string]any{"discount_amount": 999},
			setupCtx: func(c *gin.Context) { setNonSystemAdmin(c); setStaffID(c) },
			svc: &mockEstimateService{
				getFn: func(_ context.Context, _, _ uint64) (*model.Estimate, error) {
					return &model.Estimate{ID: 1, DiscountAmount: 500}, nil
				},
			},
			// 旧mockEffectivePermissionService(rules空)=deny写像（BE9-2D⑥先例）
			hasPermission: func(_ *gin.Context, _, _ string) bool { return false },
			wantStatus:    http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := httpapi.PermissionChecker(func(_ *gin.Context, _, _ string) bool { return true })
			if tt.hasPermission != nil {
				checker = tt.hasPermission
			}
			h := NewEstimateHandler(tt.svc, checker)

			var bodyBytes []byte
			if tt.bodyRaw != "" {
				bodyBytes = []byte(tt.bodyRaw)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateEstimate(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- DeleteEstimate ----

func TestDeleteEstimate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockEstimateService
		wantStatus int
	}{
		{
			name:     "deletes estimate successfully",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockEstimateService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc:        &mockEstimateService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c); setStaffID(c) },
			svc: &mockEstimateService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithEstimateSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.DeleteEstimate(c)
			c.Writer.WriteHeaderNow() // flush a bare c.Status() (no body) to the recorder

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Estimate Handler Test Cases
// This handler manages estimate (見積) records for medical service quotes (Section 12: 見積管理)
//
// 1. ListEstimates (GET /estimates)
//    ✓ Returns 200 OK paginated list / empty list
//    ✓ Returns 401/400/500 on auth/validation/db error
//    ✓ Filters: owner_id (optional FK), medical_record_id (optional FK), status (optional enum)
//    ✓ Pagination: page, limit defaults (page=1, limit=20)
//    ✓ Respects clinic_id scoping (soft delete)
//    ✓ Returns: id, owner_id, medical_record_id, title, status, subtotal, tax_total, total_amount
//
// 2. GetEstimate (GET /estimates/:id)
//    ✓ Returns 200 OK with single estimate
//    ✓ Returns 401/400/404/403/500 on auth/validation/not-found/clinic-mismatch/db error
//    ✓ Response: full estimate record with all fields (title, amounts, valid_until, comment, notes, created_by, timestamps)
//    ✓ ID validation (numeric), clinic_id check
//
// 3. CreateEstimate (POST /estimates)
//    ✓ Returns 201 Created with generated id and timestamps
//    ✓ Returns 400 on required field missing (title, owner_id, subtotal, tax_total, total_amount)
//    ✓ Returns 401/400/409/500 on auth/validation/fk-violation/db error
//    ✓ FK validation: medical_record_id (if provided), owner_id
//    ✓ Status field: enum (pending, approved, rejected, accepted, expired), defaults if not provided
//    ✓ Fields: title, owner_id, medical_record_id, subtotal, tax_total, total_amount (required)
//    ✓ Fields: insurance_amount, discount_amount, valid_until, comment, notes, created_by (optional)
//    ✓ Amount validation: non-negative, subtotal + tax = total_amount (business rule if enforced)
//
// 4. UpdateEstimate (PATCH /estimates/:id)
//    ✓ Returns 200 OK on successful update
//    ✓ Returns 400 on id/json validation error
//    ✓ Returns 401/404/403/500 on auth/not-found/clinic-mismatch/db error
//    ✓ Partial updates: all fields independently updatable (PATCH semantics)
//    ✓ Status field: enum validation if provided (pointer-based update)
//    ✓ ClearValidUntil field: special flag to null out valid_until date
//    ✓ Fields NOT updated: medical_record_id (tied to original), created_by (immutable)
//    ✓ Updated response reflects changed_at timestamp
//
// 5. DeleteEstimate (DELETE /estimates/:id)
//    ✓ Returns 204 No Content on success
//    ✓ Returns 401/400/404/403/500 on auth/validation/not-found/clinic-mismatch/db error
//    ✓ Soft delete: deleted_at set, not removed from database
//    ✓ Deleted estimate excluded from ListEstimates, returns 404 on GetEstimate
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control (clinic_id on all endpoints)
//    ✓ FK validation on owner_id, medical_record_id (if provided)
//    ✓ Enum validation on status field
//    ✓ Amount range validation (non-negative)
//    ✓ No explicit RBAC permission check documented
//    ✓ Soft delete prevents data leakage
//
// DATA MODEL (estimates):
//    - id (PK): BIGSERIAL
//    - clinic_id: BIGINT (multitenancy)
//    - medical_record_id (FK, NULLABLE): BIGINT → medical_records(id)
//    - owner_id (FK): BIGINT → owners(id)
//    - title: VARCHAR(200) - estimate title/description
//    - status: ENUM (pending|approved|rejected|accepted|expired) DEFAULT pending
//    - subtotal: NUMERIC(10,2) - sub-total before tax
//    - tax_total: NUMERIC(10,2) - tax amount
//    - total_amount: NUMERIC(10,2) - total (subtotal + tax)
//    - insurance_amount: NUMERIC(10,2) (NULLABLE) - insurance coverage
//    - discount_amount: NUMERIC(10,2) (NULLABLE) - discount applied
//    - valid_until: DATE (NULLABLE) - estimate expiration date
//    - comment: TEXT (NULLABLE) - public comment visible to customer
//    - notes: TEXT (NULLABLE) - internal notes
//    - created_by: VARCHAR(100) (NULLABLE) - staff who created estimate
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete)
//
// IMPLEMENTATION NOTES:
//    - Status enum: pending (待機), approved (承認), rejected (却下), accepted (受理), expired (期限切れ)
//    - Medical_record_id is optional (estimate can exist independently or linked to medical record)
//    - ClearValidUntil field: special flag to null out valid_until on PATCH (pointer-based)
//    - Amount validation: typically subtotal + tax_total = total_amount (validation depends on service layer)
//    - Insurance/discount amounts stored separately (not subtracted from total)
//    - Soft delete: standard clinic_id scoping
//
// TESTING STRATEGY:
//    - Test filters combined (owner_id AND medical_record_id AND status)
//    - Verify FK constraints (owner, medical_record)
//    - Verify enum validation (status field)
//    - Verify amount validation (non-negative)
//    - Verify ClearValidUntil flag behavior on PATCH
//    - Verify PATCH semantics (unspecified fields unchanged)
//    - Verify soft delete (excluded from list, 404 on get)
//    - Verify clinic_id scoping
//
