package billing

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

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock CampaignService ----

type mockCampaignService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]model.Campaign, error)
	getByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Campaign, error)
	createFn  func(ctx context.Context, clinicID uint64, input *CreateCampaignInput) (*model.Campaign, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, input *UpdateCampaignInput) (*model.Campaign, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockCampaignService) List(ctx context.Context, clinicID uint64) ([]model.Campaign, error) {
	return m.listFn(ctx, clinicID)
}

func (m *mockCampaignService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Campaign, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockCampaignService) Create(ctx context.Context, clinicID uint64, input *CreateCampaignInput) (*model.Campaign, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockCampaignService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCampaignInput) (*model.Campaign, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockCampaignService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockCampaignService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return m.reorderFn(ctx, clinicID, ids)
}

// ---- test helper ----

func newHandlerWithCampaignSvc(svc CampaignService) *CampaignHandler {
	return NewCampaignHandler(svc)
}

func sampleCampaign(id uint64) *model.Campaign {
	return &model.Campaign{
		ID:            id,
		ClinicID:      1,
		Name:          "夏の割引キャンペーン",
		StartDate:     time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		EndDate:       time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
		DiscountType:  model.CampaignDiscountTypeRate,
		DiscountValue: 10,
		IsActive:      true,
	}
}

// ---- ListCampaigns ----

func TestListCampaigns(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockCampaignService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of campaigns",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCampaignService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.Campaign, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.Campaign{*sampleCampaign(1)}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"夏の割引キャンペーン"`,
		},
		{
			name:     "returns empty list when no campaigns exist",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCampaignService{
				listFn: func(_ context.Context, _ uint64) ([]model.Campaign, error) {
					return []model.Campaign{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCampaignService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "returns 403 when selected clinic lacks accounting view grant",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("clinic_id", "2")
				setAccountingPermissionOnlyClinic(c, 1, "view")
			},
			svc: &mockCampaignService{
				listFn: func(_ context.Context, _ uint64) ([]model.Campaign, error) {
					t.Fatal("campaign service must not be reached")
					return nil, nil
				},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCampaignService{
				listFn: func(_ context.Context, _ uint64) ([]model.Campaign, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCampaignSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)

			h.ListCampaigns(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetCampaign ----

func TestGetCampaign(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockCampaignService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns campaign for valid id",
			paramID:  "4",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCampaignService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Campaign, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(4), id)
					return sampleCampaign(4), nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"id":4`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCampaignService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCampaignService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 403 when selected clinic lacks accounting view grant",
			paramID: "4",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("clinic_id", "2")
				setAccountingPermissionOnlyClinic(c, 1, "view")
			},
			svc: &mockCampaignService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Campaign, error) {
					t.Fatal("campaign service must not be reached")
					return nil, nil
				},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:     "returns 404 when campaign not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCampaignService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.Campaign, error) {
					return nil, apperrors.WrapNotFound("campaign", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCampaignSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetCampaign(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateCampaign ----

func TestCreateCampaign(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":           "夏の割引キャンペーン",
			"start_date":     "2026-07-01",
			"end_date":       "2026-08-31",
			"discount_type":  "rate",
			"discount_value": 10,
			"is_active":      true,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockCampaignService
		wantStatus int
		wantBody   string
		wantHeader bool
	}{
		{
			name:     "creates campaign successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCampaignService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateCampaignInput) (*model.Campaign, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "夏の割引キャンペーン", input.Name)
					assert.Equal(t, model.CampaignDiscountTypeRate, input.DiscountType)
					return sampleCampaign(8), nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"id":8`,
			wantHeader: true,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCampaignService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"start_date": "2026-07-01", "end_date": "2026-08-31", "discount_type": "rate"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCampaignService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCampaignService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for invalid start_date format",
			body: map[string]any{
				"name":          "テスト",
				"start_date":    "2026/07/01",
				"end_date":      "2026-08-31",
				"discount_type": "rate",
			},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCampaignService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 400 for invalid end_date format",
			body: map[string]any{
				"name":          "テスト",
				"start_date":    "2026-07-01",
				"end_date":      "2026/08/31",
				"discount_type": "rate",
			},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCampaignService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCampaignService{
				createFn: func(_ context.Context, _ uint64, _ *CreateCampaignInput) (*model.Campaign, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCampaignSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateCampaign(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantHeader {
				assert.Equal(t, "/v1/masters/campaigns/8", w.Header().Get("Location"))
			}
		})
	}
}

// ---- UpdateCampaign ----

func TestUpdateCampaign(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockCampaignService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates campaign successfully",
			paramID:  "1",
			body:     map[string]any{"name": "秋の割引キャンペーン"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCampaignService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *UpdateCampaignInput) (*model.Campaign, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, input.Name)
					assert.Equal(t, "秋の割引キャンペーン", *input.Name)
					m := sampleCampaign(1)
					m.Name = *input.Name
					return m, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"秋の割引キャンペーン"`,
		},
		{
			name:     "updates start_date and end_date",
			paramID:  "1",
			body:     map[string]any{"start_date": "2026-09-01", "end_date": "2026-09-30"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCampaignService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateCampaignInput) (*model.Campaign, error) {
					require.NotNil(t, input.StartDate)
					require.NotNil(t, input.EndDate)
					assert.Equal(t, "2026-09-01", input.StartDate.Format("2006-01-02"))
					assert.Equal(t, "2026-09-30", input.EndDate.Format("2006-01-02"))
					return sampleCampaign(1), nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCampaignService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCampaignService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCampaignService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid start_date format",
			paramID:    "1",
			body:       map[string]any{"start_date": "invalid"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCampaignService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when campaign not found",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCampaignService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateCampaignInput) (*model.Campaign, error) {
					return nil, apperrors.WrapNotFound("campaign", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCampaignSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateCampaign(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ReorderCampaigns ----

func newReorderCampaignsRouter(svc CampaignService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithCampaignSvc(svc)
	r.PATCH("/campaigns/reorder", func(c *gin.Context) {
		setClinicID(c)
	}, h.ReorderCampaigns)
	return r
}

func TestReorderCampaigns(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders campaigns successfully", func(t *testing.T) {
		svc := &mockCampaignService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{2, 1, 3}, ids)
				return nil
			},
		}
		router := newReorderCampaignsRouter(svc)
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{2, 1, 3}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/campaigns/reorder", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithCampaignSvc(&mockCampaignService{})
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderCampaigns(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		router := newReorderCampaignsRouter(&mockCampaignService{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/campaigns/reorder", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		svc := &mockCampaignService{
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return fmt.Errorf("db error")
			},
		}
		router := newReorderCampaignsRouter(svc)
		bodyBytes, err := json.Marshal(map[string]any{"ids": []int{1, 2}})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/campaigns/reorder", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ---- DeleteCampaign ----

func newDeleteCampaignRouter(svc CampaignService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithCampaignSvc(svc)
	r.DELETE("/campaigns/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteCampaign)
	return r
}

func TestDeleteCampaign(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockCampaignService
		wantStatus int
	}{
		{
			name:    "deletes campaign successfully",
			paramID: "1",
			svc: &mockCampaignService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockCampaignService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when campaign not found",
			paramID: "999",
			svc: &mockCampaignService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("campaign", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteCampaignRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/campaigns/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithCampaignSvc(&mockCampaignService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteCampaign(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
